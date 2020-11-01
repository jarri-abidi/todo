package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/jarri-abidi/todolist/todos"
)

type handler struct {
	*mux.Router
	service todos.Service
}

// New creates and returns an http.Handler using gorilla/mux.
func New(svc todos.Service) (http.Handler, io.Closer) {
	tracer := initTracer()
	h := handler{
		Router:  mux.NewRouter(),
		service: svc,
	}

	h.Use(commonMiddleware)
	h.Use(tracingMiddleware)
	h.Use(metricsMiddleware)
	h.Handle("/metrics", metricsHandler)
	h.HandleFunc("/todos", h.ListTodos).Methods("GET")
	h.HandleFunc("/todos", h.SaveTodo).Methods("POST")
	h.HandleFunc("/todo/{id:[0-9]+}", h.RemoveTodo).Methods("DELETE")
	h.HandleFunc("/todo/{id:[0-9]+}", h.ToggleTodo).Methods("PATCH")
	h.HandleFunc("/todo/{id:[0-9]+}", h.ReplaceTodo).Methods("PUT")
	h.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	return &h, tracer
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotFound, "page not found")
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.WriteHeader(code)

	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("could not encode http response: %v", err)
		return
	}

	if _, err := w.Write(response); err != nil {
		log.Printf("could not write http response: %v", err)
	}
}

func bindFromJSON(r *http.Request, dest interface{}) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		return errors.New("Invalid request body")
	}
	return nil
}

func (h *handler) ToggleTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid todo ID")
		return
	}

	err = h.service.ToggleDone(id)
	if err == todos.ErrTodoNotFound {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func (h *handler) SaveTodo(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Name string `json:"name"`
		Done bool   `json:"done"`
	}

	var req request
	if err := bindFromJSON(r, &req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	todo := todos.Todo{Name: req.Name, Done: req.Done}
	if err := h.service.Save(&todo); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, todo)
}

func (h *handler) RemoveTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid todo ID")
		return
	}

	err = h.service.Remove(id)
	if err == todos.ErrTodoNotFound {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func (h *handler) ListTodos(w http.ResponseWriter, r *http.Request) {
	todolist, err := h.service.List()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, todolist)
}

func (h *handler) ReplaceTodo(w http.ResponseWriter, r *http.Request) {}
