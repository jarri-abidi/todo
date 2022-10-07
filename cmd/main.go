package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-kit/log"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/jarri-abidi/todo/checklist"
	"github.com/jarri-abidi/todo/inmem"
	"github.com/jarri-abidi/todo/postgres"
)

func main() {
	logger := log.NewJSONLogger(os.Stderr)
	defer logger.Log("msg", "terminated")

	path, found := os.LookupEnv("TODOAPP_CONFIG_PATH")
	if found {
		if err := godotenv.Load(path); err != nil {
			logger.Log("msg", "could not load .env file", "path", path, "err", err)
		}
	}

	var config struct {
		ServerAddress              string        `envconfig:"SERVER_ADDRESS" default:"localhost:8085"`
		ServerWriteTimeout         time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"15s"`
		ServerReadTimeout          time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"15s"`
		ServerIdleTimeout          time.Duration `envconfig:"SERVER_IDLE_TIMEOUT" default:"60s"`
		GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT" default:"30s"`
		DBSource                   string        `envconfig:"DB_SOURCE"`
		DBConnectTimeout           time.Duration `envconfig:"DB_CONNECT_TIMEOUT"`
		OTELExporterJaegerEndpoint string        `envconfig:"OTEL_EXPORTER_JAEGER_ENDPOINT"`
	}
	if err := envconfig.Process("TODOAPP", &config); err != nil {
		logger.Log("msg", "could not load env vars", "err", err)
		os.Exit(1)
	}

	tasks := inmem.NewTaskRepository()

	if config.DBSource != "" {
		ctx, cancel := context.WithTimeout(context.Background(), config.DBConnectTimeout)
		defer cancel()
		db, err := postgres.NewDB(ctx, config.DBSource)
		if err != nil {
			logger.Log("msg", "could not connect to postgres", "err", err)
			os.Exit(1)
		}

		if err = postgres.Migrate("file://postgres/migrations", db); err != nil {
			logger.Log("msg", "could not run postgres schema migrations", "err", err)
			os.Exit(1)
		}

		tasks = postgres.NewTaskRepository(db)

		defer func() {
			if err := db.Close(); err != nil {
				logger.Log("msg", "could not close db connection", "err", err)
			}
		}()
	}

	if config.OTELExporterJaegerEndpoint != "" {
		exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.OTELExporterJaegerEndpoint)))
		if err != nil {
			logger.Log("msg", "could not create jaeger exporter", "err", err)
			os.Exit(1)
		}

		defer func() {
			if err := exporter.Shutdown(context.Background()); err != nil {
				logger.Log("msg", "could not shutdown traces exporter", "err", err)
			}
		}()

		tp := trace.NewTracerProvider(
			trace.WithSampler(trace.AlwaysSample()),
			trace.WithBatcher(exporter),
			trace.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("todo"),
				attribute.String("environment", "dev"),
			)),
		)
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(b3.New()))

		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				logger.Log("msg", "could not shutdown tracer provider", "err", err)
			}
		}()
	}

	var service checklist.Service
	service = checklist.NewService(tasks)
	service = checklist.LoggingMiddleware(logger)(service)

	mux := http.NewServeMux()
	mux.Handle("/checklist/v1/", checklist.MakeHandler(service, logger))
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:         config.ServerAddress,
		WriteTimeout: config.ServerWriteTimeout,
		ReadTimeout:  config.ServerReadTimeout,
		IdleTimeout:  config.ServerIdleTimeout,
		Handler:      mux,
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		logger.Log("transport", "http", "address", config.ServerAddress, "msg", "listening")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Log("transport", "http", "address", config.ServerAddress, "msg", "failed", "err", err)
			sig <- os.Interrupt // trigger shutdown of other resources
		}
	}()

	logger.Log("received", <-sig, "msg", "terminating")
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Log("msg", "could not shutdown http server", "err", err)
	}
}
