package reconciler

import (
	"time"

	"github.com/metal-toolbox/auditevent"
	"go.uber.org/zap"

	governor "go.equinixmetal.net/governor-api/pkg/client"
)

// Reconciler reconciles with downstream system
type Reconciler struct {
	auditEventWriter *auditevent.EventWriter
	interval         time.Duration
	// Client           *slack.Client
	GovernorClient *governor.Client
	Logger         *zap.Logger
	queue          string
	dryrun         bool
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
// func WithClient(c *slack.Client) Option {
// 	return func(r *Reconciler) {
// 		r.Client = c
// 	}
// }

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
