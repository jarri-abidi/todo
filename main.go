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

	"github.com/gorilla/mux"

	"github.com/jarri-abidi/todolist/handlers"
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
		handler = handlers.New(service)
		router  = mux.NewRouter()
		server  = newServer(port, router)
	)

	router.Use(commonMiddleware)
	router.HandleFunc("/todos", handler.ListTodos).Methods("GET")
	router.HandleFunc("/todos", handler.SaveTodo).Methods("POST")
	router.HandleFunc("/todo/{id:[0-9]+}", handler.RemoveTodo).Methods("DELETE")
	router.HandleFunc("/todo/{id:[0-9]+}", handler.ToggleTodo).Methods("PATCH")
	router.HandleFunc("/todo/{id:[0-9]+}", handler.ReplaceTodo).Methods("PUT")

	fmt.Println(`
	 _____          _           __ _     _   
   	/__   \___   __| | ___     / /(_)___| |_ 
	  / /\/ _ \ / _  |/ _ \   / / | / __| __|
	 / / | (_) | (_| | (_) | / /__| \__ \ |_ 
	 \/   \___/ \__,_|\___/  \____/_|___/\__|
	`)

	done := make(chan error, 1)
	fail := make(chan error, 1)
	go func() {
		log.Printf("Started HTTP server on %s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				done <- err
				return
			}
			fail <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	for {
		select {
		case msg := <-done:
			log.Println(msg)
			os.Exit(0)
		case err := <-fail:
			log.Fatal(err)
		case <-quit:
			shutDown(server, wait)
		}
	}
}

func newServer(port int, router http.Handler) *http.Server {
	return &http.Server{
		Addr: fmt.Sprintf("0.0.0.0:%d", port),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func shutDown(srv *http.Server, wait time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("could not shut down: %v", err)
	}
}
