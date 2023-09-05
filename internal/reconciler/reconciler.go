package reconciler

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"github.com/metal-toolbox/auditevent"
	"go.uber.org/zap"

	"github.com/metal-toolbox/governor-api/pkg/api/v1alpha1"
	governor "github.com/metal-toolbox/governor-api/pkg/client"

	"github.com/metal-toolbox/gov-slack-addon/internal/auctx"
	"github.com/metal-toolbox/gov-slack-addon/internal/natslock"
	"github.com/metal-toolbox/gov-slack-addon/internal/slack"
)

type govClientIface interface {
	Application(ctx context.Context, id string) (*v1alpha1.Application, error)
	Applications(ctx context.Context) ([]*v1alpha1.Application, error)
	ApplicationTypes(ctx context.Context) ([]*v1alpha1.ApplicationType, error)
	ApplicationGroups(ctx context.Context, id string) ([]*v1alpha1.Group, error)
	Group(context.Context, string, bool) (*v1alpha1.Group, error)
	GroupMembers(ctx context.Context, id string) ([]*v1alpha1.GroupMember, error)
	User(context.Context, string, bool) (*v1alpha1.User, error)
	URL() string
}

// Reconciler reconciles with downstream system
type Reconciler struct {
	Client         *slack.Client
	GovernorClient govClientIface
	ID             uuid.UUID
	Locker         *natslock.Locker
	Logger         *zap.Logger

	auditEventWriter *auditevent.EventWriter
	dryrun           bool
	interval         time.Duration
	queue            string
	userGroupPrefix  string
	applicationType  string
}

// Option is a functional configuration option
type Option func(r *Reconciler)

// WithLogger sets logger
func WithLogger(l *zap.Logger) Option {
	return func(r *Reconciler) {
		r.Logger = l
	}
}

// WithAuditEventWriter sets auditEventWriter
func WithAuditEventWriter(a *auditevent.EventWriter) Option {
	return func(r *Reconciler) {
		r.auditEventWriter = a
	}
}

// WithClient sets slack client
func WithClient(c *slack.Client) Option {
	return func(r *Reconciler) {
		r.Client = c
	}
}

// WithGovernorClient sets governor api client
func WithGovernorClient(c *governor.Client) Option {
	return func(r *Reconciler) {
		r.GovernorClient = c
	}
}

// WithLocker sets the lead election locker
func WithLocker(l *natslock.Locker) Option {
	return func(r *Reconciler) {
		r.Locker = l
	}
}

// WithDryRun sets dryrun
func WithDryRun(d bool) Option {
	return func(r *Reconciler) {
		r.dryrun = d
	}
}

// WithQueue sets nats queue for events
func WithQueue(q string) Option {
	return func(r *Reconciler) {
		r.queue = q
	}
}

// WithInterval sets the reconciler interval
func WithInterval(i time.Duration) Option {
	return func(r *Reconciler) {
		r.interval = i
	}
}

// WithUserGroupPrefix sets a prefix for user group names
func WithUserGroupPrefix(p string) Option {
	return func(r *Reconciler) {
		r.userGroupPrefix = p
	}
}

// WithApplicationType sets the application type slug
func WithApplicationType(t string) Option {
	return func(r *Reconciler) {
		r.applicationType = t
	}
}

// New returns a new reconciler
func New(opts ...Option) *Reconciler {
	rec := Reconciler{
		Logger: zap.NewNop(),
	}

	for _, opt := range opts {
		opt(&rec)
	}

	var err error

	rec.ID, err = uuid.DefaultGenerator.NewV4()
	if err != nil {
		panic(err)
	}

	rec.Logger.Debug("creating new reconciler", zap.String("id", rec.ID.String()))

	return &rec
}

// Run starts the reconciler loop
func (r *Reconciler) Run(ctx context.Context) {
	r.Logger = r.Logger.With(zap.String("reconciler.id", r.ID.String()))

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	r.Logger.Info("starting reconciler loop",
		zap.Duration("interval", r.interval),
		zap.String("governor.url", r.GovernorClient.URL()),
		zap.Bool("dryrun", r.dryrun),
	)

	if r.Locker != nil {
		r.Logger.Info("using jetstream kv store for locking and leader election",
			zap.String("bucket", r.Locker.Name()),
			zap.String("ttl", r.Locker.TTL().String()),
		)
	}

	ws, err := r.Client.ListWorkspaces(ctx)
	if err != nil {
		r.Logger.Error(err.Error())
		panic(err)
	}

	r.Logger.Info("slack token has access to the following workspaces", zap.Any("workspaces", ws))

	for {
		select {
		case <-ticker.C:
			if r.Locker != nil {
				isLead, err := r.Locker.AcquireLead(r.ID)
				if err != nil {
					r.Logger.Error("error checking for leader lock", zap.Error(err))
					continue
				}

				if !isLead {
					r.Logger.Debug("not leader, skipping loop")
					continue
				}
			}

			r.Logger.Info("executing reconciler loop",
				zap.String("time", time.Now().UTC().Format(time.RFC3339)),
			)

			ctx = auctx.WithAuditEvent(ctx, auditevent.NewAuditEvent(
				"", // eventType to be populated later
				auditevent.EventSource{
					Type:  "local",
					Value: "ReconcileLoop",
					Extra: map[string]interface{}{
						"governor.url": r.GovernorClient.URL(),
					},
				},
				auditevent.OutcomeSucceeded,
				map[string]string{
					"event": "reconciler",
				},
				"gov-slack-addon",
			))

			apps, err := r.GovernorClient.Applications(ctx)
			if err != nil {
				r.Logger.Error("error listing governor applications", zap.Error(err))
				continue
			}

			r.Logger.Debug("got applications", zap.Any("applications list", apps))

			appTypes, err := r.GovernorClient.ApplicationTypes(ctx)
			if err != nil {
				r.Logger.Error("error listing governor application types")
				continue
			}

			var desiredAppTypeID string

			for _, appType := range appTypes {
				if appType.Slug == r.applicationType {
					desiredAppTypeID = appType.ID
				}
			}

			if desiredAppTypeID == "" {
				r.Logger.Error("could not find the specified application type in governor")
				continue
			}

			// if it's slack application, reconcile all of the groups linked to it
			for _, app := range apps {
				if app.TypeID.String != desiredAppTypeID {
					continue
				}

				groups, err := r.GovernorClient.ApplicationGroups(ctx, app.ID)
				if err != nil {
					r.Logger.Error("error listing groups", zap.Error(err))
					continue
				}

				r.Logger.Debug("got groups", zap.Any("groups list", groups), zap.String("application", app.Name))

				for _, g := range groups {
					if err := r.CreateUserGroup(ctx, g.ID, app.ID); err != nil {
						if !errors.Is(err, slack.ErrSlackGroupAlreadyExists) {
							r.Logger.Warn("error creating user group", zap.Error(err))
						}
					}

					if err := r.UpdateUserGroupMembers(ctx, g.ID, app.ID); err != nil {
						r.Logger.Warn("error updating user group members", zap.Error(err))
					}
				}
			}

			r.Logger.Info("finished reconciler loop",
				zap.String("time", time.Now().UTC().Format(time.RFC3339)),
			)

		case <-ctx.Done():
			r.Logger.Info("shutting down reconciler",
				zap.String("time", time.Now().UTC().Format(time.RFC3339)),
			)

			return
		}
	}
}

// Stop stops the reconciler loop and does any necessary cleanup
func (r *Reconciler) Stop() {
	if r.Locker != nil {
		if err := r.Locker.ReleaseLead(r.ID); err != nil {
			r.Logger.Error("error releasing leader lock", zap.Error(err))
		}
	}
}
