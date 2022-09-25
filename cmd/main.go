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

	"github.com/jarri-abidi/todolist/config"
	"github.com/jarri-abidi/todolist/inmem"
	"github.com/jarri-abidi/todolist/todolist"
)

func main() {
	logger := log.NewLogfmtLogger(os.Stderr)

	conf, err := config.Load("app.env")
	if err != nil {
		logger.Log("msg", "could not load config", "err", err)
	}

	var service todolist.Service
	service = todolist.NewService(inmem.NewTodoStore())
	service = todolist.LoggingMiddleware(logger)(service)

	mux := http.NewServeMux()
	mux.Handle("/todolist/v1/", todolist.MakeHandler(service, logger))
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:         conf.ServerAddress,
		WriteTimeout: conf.ServerWriteTimeout,
		ReadTimeout:  conf.ServerReadTimeout,
		IdleTimeout:  conf.ServerIdleTimeout,
		Handler:      mux,
	}

	fmt.Println(`
	 _____          _           __ _      _  
   	/__   \___   __| | ___     / /(_) ___| |_ 
	  / /\/ _ \ / _  |/ _ \   / / | |/ __| __|
	 / / | (_) | (_| | (_) | / /__| |\__ \ |_ 
	 \/   \___/ \__,_|\___/  \____/_|\___/\__|
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
