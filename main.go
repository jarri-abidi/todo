package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jarri-abidi/todolist/handler"
	"github.com/jarri-abidi/todolist/inmem"
	"github.com/jarri-abidi/todolist/todos"
)

func main() {
	var (
		wait = *flag.Duration("graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
		port = *flag.Int("port", 8085, "the port on which the server should listen to - e.g. 8080 or 443")
	)
	flag.Parse()

	var (
		store   = inmem.NewTodoStore()
		service = todos.NewService(store)
		router  = handler.New(service)
		server  = newServer(port, router)
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
			os.Exit(0)
		case err := <-fail:
			log.Fatal(err)
		case <-quit:
			shutDown(server, wait)
		}
	}
}

func newServer(port int, handler http.Handler) *http.Server {
	return &http.Server{
		Addr: fmt.Sprintf("0.0.0.0:%d", port),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
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
