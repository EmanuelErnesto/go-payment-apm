package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"go.elastic.co/apm/v2"
	"go.elastic.co/apm/v2/transport"
	"go.uber.org/zap"

	_ "github.com/EmanuelErnesto/go-payment-apm/docs"
	"github.com/EmanuelErnesto/go-payment-apm/internal/application/command"
	"github.com/EmanuelErnesto/go-payment-apm/internal/application/query"
	"github.com/EmanuelErnesto/go-payment-apm/internal/infrastructure/httpapi"
	"github.com/EmanuelErnesto/go-payment-apm/internal/infrastructure/logger"
	"github.com/EmanuelErnesto/go-payment-apm/internal/infrastructure/persistence/postgres"
)

const shutdownTimeout = 15 * time.Second

// @title			Payments API
// @version		1.0
// @description	Simple payments API instrumented with Elastic APM.
// @BasePath		/
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	log := logger.New(logger.Options{
		Level:       cfg.Log.Level.Zap(),
		ServiceName: cfg.APM.ServiceName,
		Environment: cfg.APM.Environment,
		Output:      os.Stdout,
	})
	defer func() { _ = log.Sync() }()
	zap.ReplaceGlobals(log)

	tracer, err := newTracer(cfg.APM)
	if err != nil {
		return fmt.Errorf("building apm tracer: %w", err)
	}
	// The tracer buffers events; flushing on exit avoids losing the last transactions.
	defer func() {
		tracer.Flush(nil)
		tracer.Close()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := postgres.Open(ctx, cfg.Database.DSN())
	if err != nil {
		return err
	}
	defer db.Close()

	paymentHandler := httpapi.NewPaymentHandler(
		command.NewCreatePaymentHandler(postgres.NewPaymentCommandRepository(db), newPaymentID, time.Now),
		query.NewListPaymentsHandler(postgres.NewPaymentQueryRepository(db)),
	)

	router, err := httpapi.NewRouter(httpapi.RouterConfig{
		Mode:           cfg.HTTP.Mode,
		SwaggerEnabled: cfg.HTTP.SwaggerEnabled,
		SimulatedDelay: cfg.HTTP.SimulatedDelay,
	}, paymentHandler, tracer, log)
	if err != nil {
		return fmt.Errorf("building router: %w", err)
	}

	server := &http.Server{
		Addr:              ":" + cfg.HTTP.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	return serve(ctx, server, log)
}

// newTracer builds the APM tracer from explicit configuration instead of
// letting the agent read ELASTIC_APM_* behind our back. It is also set as the
// default tracer because apmsql resolves spans through the default tracer.
func newTracer(cfg apmConfig) (*apm.Tracer, error) {
	httpTransport, err := transport.NewHTTPTransport(transport.HTTPTransportOptions{
		ServerURLs:  []*url.URL{cfg.ServerURL.URL},
		SecretToken: cfg.SecretToken,
	})
	if err != nil {
		return nil, err
	}

	tracer, err := apm.NewTracerOptions(apm.TracerOptions{
		ServiceName:        cfg.ServiceName,
		ServiceEnvironment: cfg.Environment,
		Transport:          httpTransport,
	})
	if err != nil {
		return nil, err
	}
	apm.SetDefaultTracer(tracer)
	return tracer, nil
}

func serve(ctx context.Context, server *http.Server, log *zap.Logger) error {
	serverErr := make(chan error, 1)
	go func() {
		log.Info("http server started", zap.String("addr", server.Addr))
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	log.Info("shutting down http server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}

// UUIDv7 is time-ordered, which keeps the primary key index append-friendly.
func newPaymentID() string {
	return uuid.Must(uuid.NewV7()).String()
}
