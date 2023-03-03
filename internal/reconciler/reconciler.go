package reconciler

import (
	"context"
	"time"

	"github.com/metal-toolbox/auditevent"
	"go.uber.org/zap"

	governor "go.equinixmetal.net/governor-api/pkg/client"

	"github.com/equinixmetal/gov-slack-addon/internal/auctx"
	"github.com/equinixmetal/gov-slack-addon/internal/slack"
)

// Reconciler reconciles with downstream system
type Reconciler struct {
	auditEventWriter *auditevent.EventWriter
	interval         time.Duration
	Client           *slack.Client
	GovernorClient   *governor.Client
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

			// TODO: implement reconciler logic

		case <-ctx.Done():
			r.Logger.Info("shutting down reconciler",
				zap.String("time", time.Now().UTC().Format(time.RFC3339)),
			)

			return
		}
	}
}
