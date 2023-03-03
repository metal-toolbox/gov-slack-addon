package natssrv

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/equinixmetal/gov-slack-addon/internal/reconciler"
)

// Server implements the HTTP Server
type Server struct {
	Logger          *zap.Logger
	Listen          string
	Debug           bool
	AuditFileWriter io.Writer
	NATSClient      *NATSClient
	Reconciler      *reconciler.Reconciler
}

var (
	readTimeout     = 10 * time.Second
	writeTimeout    = 20 * time.Second
	corsMaxAge      = 12 * time.Hour
	shutdownTimeout = 5 * time.Second
)

func (s *Server) setup() *gin.Engine {
	// Setup default gin router
	r := gin.New()

	r.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowAllOrigins:  true,
		AllowCredentials: true,
		MaxAge:           corsMaxAge,
	}))

	p := ginprometheus.NewPrometheus("gin")

	// Remove any params from the URL string to keep the number of labels down
	p.ReqCntURLLabelMappingFn = func(c *gin.Context) string {
		return c.FullPath()
	}

	p.Use(r)

	customLogger := s.Logger.With(zap.String("component", "httpsrv"))
	r.Use(
		ginzap.GinzapWithConfig(customLogger, &ginzap.Config{
			TimeFormat: time.RFC3339,
			SkipPaths:  []string{"/healthz", "/healthz/readiness", "/healthz/liveness"},
			UTC:        true,
		}),
	)

	r.Use(ginzap.RecoveryWithZap(s.Logger.With(zap.String("component", "httpsrv")), true))

	tp := otel.GetTracerProvider()
	if tp != nil {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}

		r.Use(otelgin.Middleware(hostname, otelgin.WithTracerProvider(tp)))
	}

	// Health endpoints
	r.GET("/healthz", s.livenessCheck)
	r.GET("/healthz/liveness", s.livenessCheck)
	r.GET("/healthz/readiness", s.readinessCheck)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "invalid request - route not found"})
	})

	return r
}

// NewServer returns a configured server
func (s *Server) NewServer() *http.Server {
	if !s.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	return &http.Server{
		Handler:      s.setup(),
		Addr:         s.Listen,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}
}

// Run will start the server listening on the specified address
func (s *Server) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	httpsrv := s.NewServer()

	go func() {
		if err := httpsrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	go s.Reconciler.Run(ctx)

	if err := s.registerSubscriptionHandlers(); err != nil {
		panic(err)
	}

	// TEST - remove
	_, err := s.Reconciler.Client.ListWorkspaces(ctx)
	if err != nil {
		s.Logger.Error(err.Error())
	}

	// _, err = s.Reconciler.Client.CreateUserGroup(ctx, "Group from the API", "apigroup1", "Test Group #1", "T04RSB7NSJ1")
	// if err != nil {
	// 	s.Logger.Error(err.Error())
	// }

	// desc := "Changed this!!!"
	// hand := "api-group-1"
	// _, err = s.Reconciler.Client.UpdateUserGroup(ctx, "S04S08DFDDJ", "T04RSB7NSJ1", slack.UserGroupReq{
	// 	Description: &desc,
	// 	Handle:      &hand,
	// })
	// if err != nil {
	// 	s.Logger.Error(err.Error())
	// }

	// _, err = s.Reconciler.Client.GetUserGroups(ctx, "T04PFBXHFPZ", false)
	// if err != nil {
	// 	s.Logger.Error(err.Error())
	// }

	// if err := s.Reconciler.CreateUserGroup(ctx, "ad788310-17c3-4759-b6f8-b65991f5785a", "b8afd34d-378e-4de6-944f-04ec3b4da767"); err != nil {
	// 	s.Logger.Error(err.Error())
	// }

	// if err := s.Reconciler.DeleteUserGroup(ctx, "ad788310-17c3-4759-b6f8-b65991f5785a", "b8afd34d-378e-4de6-944f-04ec3b4da767"); err != nil {
	// 	s.Logger.Error(err.Error())
	// }

	// _, err = s.Reconciler.Client.GetUserGroupMembers(ctx, "S04RBJ53NR4", "T04PFBXHFPZ", false)
	// if err != nil {
	// 	s.Logger.Error(err.Error())
	// }

	// _, err = s.Reconciler.Client.UpdateUserGroupMembers(ctx, "S04S08DFDDJ", "T04RSB7NSJ1", []string{"U04R2FDHM71"})
	// if err != nil {
	// 	s.Logger.Error(err.Error())
	// }

	// _, err = s.Reconciler.Client.DisableUserGroup(ctx, "S04S08DFDDJ", "T04RSB7NSJ1")
	// if err != nil {
	// 	s.Logger.Error(err.Error())
	// }

	// _, err = s.Reconciler.Client.GetUser(ctx, "U04MW61D7FZ")
	// if err != nil {
	// 	if errors.Is(err, slack.ErrSlackUserNotFound) {
	// 		s.Logger.Info("WHOOOOOOOO no such user")
	// 	} else {
	// 		s.Logger.Error(err.Error())
	// 	}
	// }

	// _, err = s.Reconciler.Client.GetUserByEmail(ctx, "tgrozev@equinix.com")
	// if err != nil {
	// 	s.Logger.Error(err.Error())
	// }

	<-ctx.Done()

	ctxShutDown, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer func() {
		cancel()
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()

		if err := s.shutdownSubscriptions(); err != nil {
			s.Logger.Warn("error shutting down subscription", zap.Error(err))
		}
	}()

	if err := httpsrv.Shutdown(ctxShutDown); err != nil {
		return err
	}

	// wait for clean shutdown
	wg.Wait()

	s.Logger.Info("server shutdown cleanly", zap.String("time", time.Now().UTC().Format(time.RFC3339)))

	return nil
}

// livenessCheck ensures that the server is up and responding
func (s *Server) livenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "UP",
	})
}

// readinessCheck ensures that the server is up and that we are able to process requests.
func (s *Server) readinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "UP",
	})
}
