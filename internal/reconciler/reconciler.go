package reconciler

import (
	"context"
	"errors"
	"time"

	"github.com/metal-toolbox/auditevent"
	"go.uber.org/zap"

	"go.equinixmetal.net/governor-api/pkg/api/v1alpha1"
	governor "go.equinixmetal.net/governor-api/pkg/client"

	"github.com/equinixmetal/gov-slack-addon/internal/auctx"
	"github.com/equinixmetal/gov-slack-addon/internal/slack"
)

type govClientIface interface {
	Application(ctx context.Context, id string) (*v1alpha1.Application, error)
	Applications(ctx context.Context) ([]*v1alpha1.Application, error)
	ApplicationGroups(ctx context.Context, id string) ([]*v1alpha1.Group, error)
	Group(context.Context, string, bool) (*v1alpha1.Group, error)
	GroupMembers(ctx context.Context, id string) ([]*v1alpha1.GroupMember, error)
	User(context.Context, string, bool) (*v1alpha1.User, error)
	URL() string
}

// Reconciler reconciles with downstream system
type Reconciler struct {
	auditEventWriter *auditevent.EventWriter
	interval         time.Duration
	Client           *slack.Client
	GovernorClient   govClientIface
	Logger           *zap.Logger
	queue            string
	userGroupPrefix  string
	dryrun           bool
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

// New returns a new reconciler
func New(opts ...Option) *Reconciler {
	rec := Reconciler{
		Logger: zap.NewNop(),
	}

	for _, opt := range opts {
		opt(&rec)
	}

	rec.Logger.Debug("creating new reconciler")

	return &rec
}

// Run starts the reconciler loop
func (r *Reconciler) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	r.Logger.Info("starting reconciler loop",
		zap.Duration("interval", r.interval),
		zap.String("governor.url", r.GovernorClient.URL()),
		zap.Bool("dryrun", r.dryrun),
	)

	ws, err := r.Client.ListWorkspaces(ctx)
	if err != nil {
		r.Logger.Error(err.Error())
		panic(err)
	}

	r.Logger.Info("slack token has access to the following workspaces", zap.Any("workspaces", ws))

	for {
		select {
		case <-ticker.C:
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

			// if it's slack application, reconcile all of the groups linked to it
			for _, app := range apps {
				if app.Type != applicationTypeFilter {
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
