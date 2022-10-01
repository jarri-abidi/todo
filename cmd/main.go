package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-kit/log"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"

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

	var db *sql.DB
	{
		ctx, cancel := context.WithTimeout(context.Background(), conf.DBConnectTimeout)
		defer cancel()
		db, err = postgres.NewDB(ctx, conf.DBSource)
		if err != nil {
			logger.Log("msg", "could not connect to postgres", "err", err)
			os.Exit(1)
		}
	}

	var exporter trace.SpanExporter
	{
		exporter, err = jaeger.New(jaeger.WithCollectorEndpoint())
		if err != nil {
			logger.Log("msg", "could not create jaeger exporter", "err", err)
			os.Exit(1)
		}
	}

	var tp *trace.TracerProvider
	{
		tp = trace.NewTracerProvider(
			trace.WithSampler(trace.AlwaysSample()),
			trace.WithBatcher(exporter),
		)
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
			b3.New(),
		))
	}

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

	fmt.Println(`
	 _____          _       
   	/__   \___   __| | ___  
	  / /\/ _ \ / _  |/ _ \ 
	 / / | (_) | (_| | (_) |
	 \/   \___/ \__,_|\___/ 
	`)

	go func() {
		logger.Log("transport", "http", "address", conf.ServerAddress, "msg", "listening")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Log("transport", "http", "address", conf.ServerAddress, "msg", "failed", "err", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	logger.Log("received", <-sig, "msg", "terminating")

	if err := server.Shutdown(context.Background()); err != nil {
		logger.Log("msg", "could not shutdown http server", "err", err)
	}

	if err := db.Close(); err != nil {
		logger.Log("msg", "could not close db connection", "err", err)
	}

	if err := tp.Shutdown(context.Background()); err != nil {
		logger.Log("msg", "could not shutdown tracer provider", "err", err)
	}

	if err := exporter.Shutdown(context.Background()); err != nil {
		logger.Log("msg", "could not shutdown traces exporter", "err", err)
	}

	logger.Log("msg", "terminated")
}
