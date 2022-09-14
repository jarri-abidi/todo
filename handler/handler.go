package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/jarri-abidi/todolist/todolist"
)

type handler struct {
	*mux.Router
	service todolist.Service
}

// New creates and returns an http.Handler using gorilla/mux.
func New(svc todolist.Service) (http.Handler, io.Closer) {
	tracer := initTracer()
	h := handler{
		Router:  mux.NewRouter(),
		service: svc,
	}

	h.Use(loggingMiddleware)
	h.Use(tracingMiddleware)
	h.Use(metricsMiddleware)
	h.Handle("/metrics", metricsHandler)
	h.NotFoundHandler = loggingMiddleware(http.HandlerFunc(notFound))
	h.MethodNotAllowedHandler = loggingMiddleware(http.HandlerFunc(methodNotAllowed))

	return &h, tracer
}

func loggingMiddleware(next http.Handler) http.Handler {
	return handlers.CustomLoggingHandler(os.Stdout, next, func(writer io.Writer, params handlers.LogFormatterParams) {
		log.Printf(`"%s %s %s" %d`, params.Request.Method, params.Request.URL, params.Request.Proto, params.StatusCode)
	})
}

func notFound(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotFound, "resource not found")
}

func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	if payload == nil {
		w.WriteHeader(code)
		return
	}

	w.Header().Add("Content-Type", "application/json")
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
