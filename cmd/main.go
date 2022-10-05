package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-kit/log"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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
	"github.com/jarri-abidi/todo/config"
	"github.com/jarri-abidi/todo/postgres"
)

func main() {
	logger := log.NewJSONLogger(os.Stderr)

	conf, err := config.Load("app.env")
	if err != nil {
		logger.Log("msg", "could not load config", "err", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), conf.DBConnectTimeout)
	defer cancel()
	db, err := postgres.NewDB(ctx, conf.DBSource)
	if err != nil {
		logger.Log("msg", "could not connect to postgres", "err", err)
		os.Exit(1)
	}

	if err = postgres.Migrate("file://postgres/migrations", db); err != nil {
		logger.Log("msg", "could not run postgres schema migrations", "err", err)
		os.Exit(1)
	}

	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint())
	if err != nil {
		logger.Log("msg", "could not create jaeger exporter", "err", err)
		os.Exit(1)
	}

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

	tasks := postgres.NewTaskRepository(db)

	var service checklist.Service
	service = checklist.NewService(tasks)
	service = checklist.LoggingMiddleware(logger)(service)

	mux := http.NewServeMux()
	mux.Handle("/checklist/v1/", checklist.MakeHandler(service, logger))
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:         conf.ServerAddress,
		WriteTimeout: conf.ServerWriteTimeout,
		ReadTimeout:  conf.ServerReadTimeout,
		IdleTimeout:  conf.ServerIdleTimeout,
		Handler:      mux,
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		logger.Log("transport", "http", "address", conf.ServerAddress, "msg", "listening")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Log("transport", "http", "address", conf.ServerAddress, "msg", "failed", "err", err)
			sig <- os.Interrupt // trigger shutdown of other resources
		}
	}()

	logger.Log("received", <-sig, "msg", "terminating")

	if err = server.Shutdown(context.Background()); err != nil {
		logger.Log("msg", "could not shutdown http server", "err", err)
	}

	if err = db.Close(); err != nil {
		logger.Log("msg", "could not close db connection", "err", err)
	}

	if err = tp.Shutdown(context.Background()); err != nil {
		logger.Log("msg", "could not shutdown tracer provider", "err", err)
	}

	if err = exporter.Shutdown(context.Background()); err != nil {
		logger.Log("msg", "could not shutdown traces exporter", "err", err)
	}

	logger.Log("msg", "terminated")
}
