package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/jarri-abidi/todolist/handlers"
	"github.com/jarri-abidi/todolist/inmem"
	"github.com/jarri-abidi/todolist/todos"
)

func main() {
	var (
		store   = inmem.NewTodoStore()
		service = todos.NewService(store)
		handler = handlers.New(service)
		router  = mux.NewRouter()
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
	log.Print("Started HTTP server on port 8085")
	log.Fatal(http.ListenAndServe(":8085", router))
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
