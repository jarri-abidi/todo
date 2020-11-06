package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jarri-abidi/todolist/config"
	"github.com/jarri-abidi/todolist/handler"
	"github.com/jarri-abidi/todolist/inmem"
	"github.com/jarri-abidi/todolist/todos"
)

func main() {
	conf, err := config.Load(".", "app.env")
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	var (
		store             = inmem.NewTodoStore()
		service           = todos.NewService(store)
		router, resources = handler.New(service)
		server            = newServer(conf, router)
	)

	fmt.Println(`
	 _____          _           __ _     _   
   	/__   \___   __| | ___     / /(_)___| |_ 
	  / /\/ _ \ / _  |/ _ \   / / | / __| __|
	 / / | (_) | (_| | (_) | / /__| \__ \ |_ 
	 \/   \___/ \__,_|\___/  \____/_|___/\__|
	`)

	stop := make(chan error, 1)
	fail := make(chan error, 1)
	go start(server, stop, fail)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	for {
		select {
		case msg := <-stop:
			log.Println(msg)
			resources.Close()
			os.Exit(0)
		case err := <-fail:
			resources.Close()
			log.Fatal(err)
		case <-quit:
			shutDown(server, conf.GracefulShutdownTimeout)
		}
	}
}

func newServer(conf config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         conf.ServerAddress,
		WriteTimeout: conf.ServerWriteTimeout,
		ReadTimeout:  conf.ServerReadTimeout,
		IdleTimeout:  conf.ServerIdleTimeout,
		Handler:      handler,
	}
}

func start(srv *http.Server, stop, fail chan error) {
	log.Printf("Started HTTP server on %s\n", srv.Addr)
	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		stop <- err
		return
	}
	fail <- err
}

func shutDown(srv *http.Server, wait time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("could not shut down: %v", err)
	}
}
