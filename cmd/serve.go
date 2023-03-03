package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/metal-toolbox/auditevent"
	audithelpers "github.com/metal-toolbox/auditevent/helpers"
	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/clientcredentials"

	governor "go.equinixmetal.net/governor-api/pkg/client"
	events "go.equinixmetal.net/governor-api/pkg/events/v1alpha1"

	"github.com/equinixmetal/gov-slack-addon/internal/natssrv"
	"github.com/equinixmetal/gov-slack-addon/internal/reconciler"
	"github.com/equinixmetal/gov-slack-addon/internal/slack"
)

// serveCmd starts the gov-slack-addon service
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "starts the gov-slack-addon service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return serve(cmd.Context(), viper.GetViper())
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().String("listen", "0.0.0.0:8000", "address to listen on")
	viperBindFlag("listen", serveCmd.Flags().Lookup("listen"))

	serveCmd.PersistentFlags().Bool("dry-run", false, "do not make any changes, just log what would be done")
	viperBindFlag("dryrun", serveCmd.PersistentFlags().Lookup("dry-run"))

	// Tracing Flags
	serveCmd.Flags().Bool("tracing", false, "enable tracing support")
	viperBindFlag("tracing.enabled", serveCmd.Flags().Lookup("tracing"))
	serveCmd.Flags().String("tracing-provider", "jaeger", "tracing provider to use")
	viperBindFlag("tracing.provider", serveCmd.Flags().Lookup("tracing-provider"))
	serveCmd.Flags().String("tracing-endpoint", "", "endpoint where traces are sent")
	viperBindFlag("tracing.endpoint", serveCmd.Flags().Lookup("tracing-endpoint"))
	serveCmd.Flags().String("tracing-environment", "production", "environment value in traces")
	viperBindFlag("tracing.environment", serveCmd.Flags().Lookup("tracing-environment"))

	serveCmd.Flags().String("audit-log-path", "/app-audit/audit.log", "file path to write audit logs to.")
	viperBindFlag("audit.log-path", serveCmd.Flags().Lookup("audit-log-path"))

	// Reconciler flags
	serveCmd.Flags().Duration("reconciler-interval", 1*time.Hour, "interval for the reconciler loop")
	viperBindFlag("reconciler.interval", serveCmd.Flags().Lookup("reconciler-interval"))

	// Slack related flags
	serveCmd.Flags().String("slack-token", "", "api token for slack")
	viperBindFlag("slack.token", serveCmd.Flags().Lookup("slack-token"))
	serveCmd.Flags().String("slack-usergroup-prefix", "[Governor] ", "string to be prepended to slack usergroup names")
	viperBindFlag("slack.usergroup-prefix", serveCmd.Flags().Lookup("slack-usergroup-prefix"))

	// Governor related flags
	serveCmd.Flags().String("governor-url", "https://api.iam.equinixmetal.net", "url of the governor api")
	viperBindFlag("governor.url", serveCmd.Flags().Lookup("governor-url"))
	serveCmd.Flags().String("governor-client-id", "gov-slack-addon-governor", "oauth client ID for client credentials flow")
	viperBindFlag("governor.client-id", serveCmd.Flags().Lookup("governor-client-id"))
	serveCmd.Flags().String("governor-client-secret", "", "oauth client secret for client credentials flow")
	viperBindFlag("governor.client-secret", serveCmd.Flags().Lookup("governor-client-secret"))
	serveCmd.Flags().String("governor-token-url", "http://hydra:4444/oauth2/token", "url used for client credential flow")
	viperBindFlag("governor.token-url", serveCmd.Flags().Lookup("governor-token-url"))
	serveCmd.Flags().String("governor-audience", "http://api:3001/", "oauth audience for client credential flow")
	viperBindFlag("governor.audience", serveCmd.Flags().Lookup("governor-audience"))

	// NATS related flags
	serveCmd.Flags().String("nats-url", "nats://127.0.0.1:4222", "NATS server connection url")
	viperBindFlag("nats.url", serveCmd.Flags().Lookup("nats-url"))
	serveCmd.Flags().String("nats-token", "", "NATS auth token")
	viperBindFlag("nats.token", serveCmd.Flags().Lookup("nats-token"))
	serveCmd.Flags().String("nats-nkey", "", "Path to the file containing the NATS nkey keypair")
	viperBindFlag("nats.nkey", serveCmd.Flags().Lookup("nats-nkey"))
	serveCmd.Flags().String("nats-subject-prefix", "equinixmetal.governor.events", "prefix for NATS subjects")
	viperBindFlag("nats.subject-prefix", serveCmd.Flags().Lookup("nats-subject-prefix"))
	serveCmd.Flags().String("nats-queue-group", "equinixmetal.governor.addons.gov-slack-addon", "queue group for load balancing messages across NATS consumers")
	viperBindFlag("nats.queue-group", serveCmd.Flags().Lookup("nats-queue-group"))
	serveCmd.Flags().Int("nats-queue-size", 3, "queue size for load balancing messages across NATS consumers") //nolint: gomnd
	viperBindFlag("nats.queue-size", serveCmd.Flags().Lookup("nats-queue-size"))
}

func serve(cmdCtx context.Context, v *viper.Viper) error {
	initTracing()

	if err := validateMandatoryFlags(); err != nil {
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(cmdCtx)

	go func() {
		<-c
		cancel()
	}()

	auditpath := viper.GetString("audit.log-path")

	if auditpath == "" {
		logger.Fatal("failed starting server. Audit log file path can't be empty")
	}

	// WARNING: This will block until the file is available;
	// make sure an initContainer creates the file
	auf, auerr := audithelpers.OpenAuditLogFileUntilSuccess(auditpath)
	if auerr != nil {
		logger.Fatalw("couldn't open audit file.", "error", auerr)
	}
	defer auf.Close()

	nc, err := newNATSConnection()
	if err != nil {
		logger.Fatalw("failed to create NATS client connection", "error", err)
	}

	natsClient, err := natssrv.NewNATSClient(
		natssrv.WithNATSLogger(logger.Desugar()),
		natssrv.WithNATSConn(nc),
		natssrv.WithNATSPrefix(viper.GetString("nats.subject-prefix")),
		natssrv.WithNATSSubject(events.GovernorApplicationsEventSubject),
		natssrv.WithNATSQueueGroup(viper.GetString("nats.queue-group"), viper.GetInt("nats.queue-size")),
	)
	if err != nil {
		logger.Fatalw("failed creating new NATS client", "error", err)
	}

	gc, err := governor.NewClient(
		governor.WithLogger(logger.Desugar()),
		governor.WithURL(viper.GetString("governor.url")),
		governor.WithClientCredentialConfig(&clientcredentials.Config{
			ClientID:       viper.GetString("governor.client-id"),
			ClientSecret:   viper.GetString("governor.client-secret"),
			TokenURL:       viper.GetString("governor.token-url"),
			EndpointParams: url.Values{"audience": {viper.GetString("governor.audience")}},
			Scopes: []string{
				"read:governor:users",
				"read:governor:groups",
				"read:governor:applications",
			},
		}),
	)
	if err != nil {
		return err
	}

	sc := slack.NewClient(
		slack.WithLogger(logger.Desugar()),
		slack.WithToken(viper.GetString("slack.token")),
	)

	rec := reconciler.New(
		reconciler.WithAuditEventWriter(auditevent.NewDefaultAuditEventWriter(auf)),
		reconciler.WithClient(sc),
		reconciler.WithGovernorClient(gc),
		reconciler.WithLogger(logger.Desugar()),
		reconciler.WithInterval(viper.GetDuration("reconciler.interval")),
		reconciler.WithUserGroupPrefix(viper.GetString("slack.usergroup-prefix")),
		reconciler.WithDryRun(viper.GetBool("dryrun")),
	)

	server := &natssrv.Server{
		Debug:           viper.GetBool("logging.debug"),
		Listen:          viper.GetString("listen"),
		Logger:          logger.Desugar(),
		AuditFileWriter: auf,
		NATSClient:      natsClient,
		Reconciler:      rec,
	}

	logger.Infow("starting server",
		"address", viper.GetString("listen"),
		"governor-url", viper.GetString("governor.url"),
		"slack-usergroup-prefix", viper.GetString("slack.usergroup-prefix"),
		"dryrun", viper.GetBool("dryrun"),
	)

	if err := server.Run(ctx); err != nil {
		logger.Fatalw("failed starting server", "error", err)
	}

	return nil
}

// newNATSConnection creates a new NATS connection
func newNATSConnection() (*nats.Conn, error) {
	opts := []nats.Option{}

	if viper.GetBool("development") {
		logger.Debug("enabling development settings")

		opts = append(opts, nats.Token(viper.GetString("nats.token")))
	} else {
		opt, err := nats.NkeyOptionFromSeed(viper.GetString("nats-nkey"))
		if err != nil {
			return nil, err
		}

		opts = append(opts, opt)
	}

	return nats.Connect(
		viper.GetString("nats-url"),
		opts...,
	)
}

// validateMandatoryFlags collects the mandatory flag validation
func validateMandatoryFlags() error {
	errs := []string{}

	if viper.GetString("nats.url") == "" {
		errs = append(errs, ErrNATSURLRequired.Error())
	}

	if viper.GetString("nats.token") == "" && viper.GetString("nats.nkey") == "" {
		errs = append(errs, ErrNATSAuthRequired.Error())
	}

	if viper.GetString("slack.token") == "" {
		errs = append(errs, ErrSlackTokenRequired.Error())
	}

	if viper.GetString("governor.url") == "" {
		errs = append(errs, ErrGovernorURLRequired.Error())
	}

	if viper.GetString("governor.client-id") == "" {
		errs = append(errs, ErrGovernorClientIDRequired.Error())
	}

	if viper.GetString("governor.client-secret") == "" {
		errs = append(errs, ErrGovernorClientSecretRequired.Error())
	}

	if viper.GetString("governor.token-url") == "" {
		errs = append(errs, ErrGovernorClientTokenURLRequired.Error())
	}

	if viper.GetString("governor.audience") == "" {
		errs = append(errs, ErrGovernorClientAudienceRequired.Error())
	}

	if len(errs) == 0 {
		return nil
	}

	return fmt.Errorf(strings.Join(errs, "\n")) //nolint:goerr113
}
