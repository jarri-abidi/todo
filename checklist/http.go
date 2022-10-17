package checklist

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-kit/log"
	"github.com/matryer/way"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/jarri-abidi/todo"
)

func NewServer(service Service, logger log.Logger) http.Handler {
	s := server{service: service}

	var handleSaveTask http.Handler
	handleSaveTask = s.handleSaveTask()
	handleSaveTask = httpLoggingMiddleware(logger, "handleSaveTask")(handleSaveTask)
	handleSaveTask = otelhttp.NewHandler(handleSaveTask, "handleSaveTask")
	// wire other middlewares here

	var handleListTasks http.Handler
	handleListTasks = s.handleListTasks()
	handleListTasks = httpLoggingMiddleware(logger, "handleListTasks")(handleListTasks)
	handleListTasks = otelhttp.NewHandler(handleListTasks, "handleListTasks")

	var handleRemoveTask http.Handler
	handleRemoveTask = s.handleRemoveTask()
	handleRemoveTask = httpLoggingMiddleware(logger, "handleRemoveTask")(handleRemoveTask)
	handleRemoveTask = otelhttp.NewHandler(handleRemoveTask, "handleRemoveTask")

	var handleToggleTask http.Handler
	handleToggleTask = s.handleToggleTask()
	handleToggleTask = httpLoggingMiddleware(logger, "handleToggleTask")(handleToggleTask)
	handleToggleTask = otelhttp.NewHandler(handleToggleTask, "handleToggleTask")

	var handleUpdateTask http.Handler
	handleUpdateTask = s.handleUpdateTask()
	handleUpdateTask = httpLoggingMiddleware(logger, "handleUpdateTask")(handleUpdateTask)
	handleUpdateTask = otelhttp.NewHandler(handleUpdateTask, "handleUpdateTask")

	router := way.NewRouter()

	router.Handle("POST", "/checklist/v1/tasks", handleSaveTask)
	router.Handle("GET", "/checklist/v1/tasks", handleListTasks)
	router.Handle("DELETE", "/checklist/v1/task/:id", handleRemoveTask)
	router.Handle("PATCH", "/checklist/v1/task/:id", handleToggleTask)
	router.Handle("PUT", "/checklist/v1/task/:id", handleUpdateTask)

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { writeError(w, ErrResourceNotFound) })

	return router
}

const (
	contentTypeKey   = "Content-Type"
	contentTypeValue = "application/json; charset=utf-8"
)

var (
	ErrNonNumericTaskID = errors.New("task id in path must be numeric")
	ErrResourceNotFound = errors.New("resource not found")
	ErrMethodNotAllowed = errors.New("method not allowed")
)

type ErrInvalidRequestBody struct{ err error }

func (e ErrInvalidRequestBody) Error() string { return fmt.Sprintf("invalid request body: %v", e.err) }

type server struct {
	service Service
}

func (s *server) handleSaveTask() http.HandlerFunc {
	type request struct {
		Name string `json:"name"`
	}
	type response struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Done bool   `json:"done"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, ErrInvalidRequestBody{err})
			return
		}

		task, err := s.service.Save(r.Context(), todo.Task{Name: req.Name})
		if err != nil {
			writeError(w, err)
			return
		}
		w.Header().Set(contentTypeKey, contentTypeValue)
		json.NewEncoder(w).Encode(response{ID: task.ID, Name: task.Name, Done: task.Done})
	}
}

func (s *server) handleListTasks() http.HandlerFunc {
	type task struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Done bool   `json:"done"`
	}
	type response []task
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := s.service.List(r.Context())
		if err != nil {
			writeError(w, err)
			return
		}

		resp := make(response, 0, len(list))
		for _, v := range list {
			resp = append(resp, task{ID: v.ID, Name: v.Name, Done: v.Done})
		}
		w.Header().Set(contentTypeKey, contentTypeValue)
		json.NewEncoder(w).Encode(resp)
	}
}

func (s *server) handleRemoveTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(way.Param(r.Context(), "id"), 10, 64)
		if err != nil {
			writeError(w, ErrNonNumericTaskID)
			return
		}

		if err := s.service.Remove(r.Context(), id); err != nil {
			writeError(w, err)
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *server) handleToggleTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(way.Param(r.Context(), "id"), 10, 64)
		if err != nil {
			writeError(w, ErrNonNumericTaskID)
			return
		}

		if err := s.service.ToggleDone(r.Context(), id); err != nil {
			writeError(w, err)
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *server) handleUpdateTask() http.HandlerFunc {
	type request struct {
		Name string `json:"name"`
		Done bool   `json:"done"`
	}
	type response struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Done bool   `json:"done"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(way.Param(r.Context(), "id"), 10, 64)
		if err != nil {
			writeError(w, ErrNonNumericTaskID)
			return
		}

		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, ErrInvalidRequestBody{err})
			return
		}

		task, err := s.service.Update(r.Context(), todo.Task{ID: id, Name: req.Name, Done: req.Done})
		if err != nil {
			writeError(w, err)
		}
		w.Header().Set(contentTypeKey, contentTypeValue)
		json.NewEncoder(w).Encode(response{ID: task.ID, Name: task.Name, Done: task.Done})
	}
}

func writeError(w http.ResponseWriter, err error) {
	w.Header().Set(contentTypeKey, contentTypeValue)

	switch err {
	case ErrResourceNotFound, todo.ErrTaskNotFound:
		w.WriteHeader(http.StatusNotFound)
	case todo.ErrTaskAlreadyExists:
		w.WriteHeader(http.StatusConflict)
	case ErrNonNumericTaskID:
		w.WriteHeader(http.StatusBadRequest)
	case ErrMethodNotAllowed:
		w.WriteHeader(http.StatusMethodNotAllowed)
	default:
		switch err.(type) {
		case ErrInvalidRequestBody:
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func httpLoggingMiddleware(logger log.Logger, operation string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			begin := time.Now()
			lrw := &loggingResponseWriter{w, http.StatusOK}
			next.ServeHTTP(lrw, r)
			logger.Log(
				"operation", operation,
				"method", r.Method,
				"path", r.URL.Path,
				"took", time.Since(begin),
				"status", lrw.statusCode,
			)
		})
	}
}
