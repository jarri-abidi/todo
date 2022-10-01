package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/jarri-abidi/todo/checklist"
	"github.com/jarri-abidi/todo/config"
	"github.com/jarri-abidi/todo/postgres"
)

func main() {
	logger := log.NewLogfmtLogger(os.Stderr)

	conf, err := config.Load("app.env")
	if err != nil {
		logger.Log("msg", "could not load config", "err", err)
		os.Exit(1)
	}

	db, err := postgres.NewDB(context.TODO(), conf.DBSource)
	if err != nil {
		logger.Log("msg", "could not connect to postgres", "err", err)
		os.Exit(1)
	}

	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint())
	if err != nil {
		logger.Log("msg", "could not init jaeger exporter", "err", err)
		os.Exit(1)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Log("msg", "could not shutdown tracer provider", "err", err)
		}
	}()

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

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT)
	go func() {
		<-quit
		if err := server.Shutdown(context.Background()); err != nil {
			logger.Log("terminated", err)
		}
	}()

	logger.Log("transport", "http", "address", conf.ServerAddress, "msg", "listening")
	logger.Log("terminated", server.ListenAndServe())
}
