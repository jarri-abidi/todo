package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jarri-abidi/todolist/handlers"
	"github.com/jarri-abidi/todolist/inmem"
	"github.com/jarri-abidi/todolist/todos"
)

func main() {
	h := handlers.Handler{Service: todos.NewService(inmem.NewTodoStore())}

	router := mux.NewRouter()
	router.Use(commonMiddleware)
	router.HandleFunc("/todos", h.ListTodos).Methods("GET")
	router.HandleFunc("/todos", h.SaveTodo).Methods("POST")
	router.HandleFunc("/todo/{id:[0-9]+}", h.RemoveTodo).Methods("DELETE")
	router.HandleFunc("/todo/{id:[0-9]+}", h.ToggleTodo).Methods("PATCH")

        log.Print(`
 _____          _           __ _     _   
/__   \___   __| | ___     / /(_)___| |_ 
  / /\/ _ \ / _` |/ _ \   / / | / __| __|
 / / | (_) | (_| | (_) | / /__| \__ \ |_ 
 \/   \___/ \__,_|\___/  \____/_|___/\__|
                                         
Starting HTTP server on port 8085.
        `)
	log.Fatal(http.ListenAndServe(":8085", router))
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
