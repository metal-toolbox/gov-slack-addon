// Package configs contains configuration structures and utilities for the application.
package configs

import (
	"context"
	"net/url"
	"time"

	govclient "github.com/metal-toolbox/governor-api/pkg/client"
	govcfg "github.com/metal-toolbox/governor-api/pkg/configs"
	sdkcfg "github.com/metal-toolbox/governor-extension-sdk/pkg/configs"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	// DefaultReconcilerInterval is the default interval for the reconciler loop
	DefaultReconcilerInterval = 1 * time.Hour
	// DefaultNATSQueueSize is the default queue size for load balancing NATS consumers
	DefaultNATSQueueSize = 3
)

// AppConfig holds the application configuration
var AppConfig struct {
	govcfg.Configs `mapstructure:",squash"`

	DryRun     bool `mapstructure:"dryrun"`
	Audit      sdkcfg.Audit
	Tracing    sdkcfg.Tracing
	Logging    sdkcfg.Logging
	Governor   Governor
	Slack      Slack
	Reconciler Reconciler
	Server     sdkcfg.Server
	NATS       sdkcfg.NATSConfig
}

// Governor reuses the SDK's Governor config and adds the addon-specific
// application type slug.
type Governor struct {
	sdkcfg.Governor `mapstructure:",squash"`

	ApplicationType string `mapstructure:"application-type"`
}

// Slack holds Slack API configuration
type Slack struct {
	Token           string `mapstructure:"token"`
	UsergroupPrefix string `mapstructure:"usergroup-prefix"`
}

// Reconciler holds reconciler configuration
type Reconciler struct {
	Interval time.Duration `mapstructure:"interval"`
	Locking  bool          `mapstructure:"locking"`
}

// MustSlackFlags registers Slack related flags and binds them to viper
// Panics on error
func MustSlackFlags(v *viper.Viper, flags *pflag.FlagSet) {
	flags.Bool("dry-run", false, "do not make any changes, just log what would be done")
	viperBindFlag(v, "dryrun", flags.Lookup("dry-run"))
	flags.String("slack-token", "", "api token for slack")
	viperBindFlag(v, "slack.token", flags.Lookup("slack-token"))
	flags.String("slack-usergroup-prefix", "[Governor] ", "string to be prepended to slack usergroup names")
	viperBindFlag(v, "slack.usergroup-prefix", flags.Lookup("slack-usergroup-prefix"))
}

// MustGovernorFlags registers Governor related flags and binds them to viper
// Panics on error
func MustGovernorFlags(v *viper.Viper, flags *pflag.FlagSet) {
	sdkcfg.MustGovernorFlags(v, flags)

	flags.String("governor-application-type", "slack", "application type slug to be listening to events for")
	viperBindFlag(v, "governor.application-type", flags.Lookup("governor-application-type"))
}

// MustReconcilerFlags registers Reconciler related flags and binds them to viper
// Panics on error
func MustReconcilerFlags(v *viper.Viper, flags *pflag.FlagSet) {
	flags.Duration("reconciler-interval", DefaultReconcilerInterval, "interval for the reconciler loop")
	viperBindFlag(v, "reconciler.interval", flags.Lookup("reconciler-interval"))
	flags.Bool("reconciler-locking", false, "enable reconciler locking and leader election")
	viperBindFlag(v, "reconciler.locking", flags.Lookup("reconciler-locking"))
}

// viperBindFlag provides a wrapper around the viper bindings that handles error checks
func viperBindFlag(v *viper.Viper, name string, flag *pflag.Flag) {
	if err := v.BindPFlag(name, flag); err != nil {
		panic(err)
	}
}

// NewGovernorClient returns a governor API client based on auth configs
func NewGovernorClient(ctx context.Context, opts ...govclient.Option) (*govclient.Client, error) {
	var (
		ts  oauth2.TokenSource
		err error
	)

	if AppConfig.Governor.WorkloadIdentity {
		ts, err = AppConfig.WorkloadIdentity.ToTokenSource(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		cc := &clientcredentials.Config{
			ClientID:       AppConfig.Governor.ClientID,
			ClientSecret:   AppConfig.Governor.ClientSecret,
			TokenURL:       AppConfig.Governor.TokenURL,
			EndpointParams: url.Values{"audience": {AppConfig.Governor.Audience}},
			Scopes:         AppConfig.Governor.Scopes,
		}

		ts = cc.TokenSource(ctx)
	}

	opts = append(
		opts,
		govclient.WithURL(AppConfig.Governor.URL),
		govclient.WithTokenSource(ts),
	)

	return govclient.NewClient(opts...)
}
