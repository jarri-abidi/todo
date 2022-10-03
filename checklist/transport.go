package checklist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/jarri-abidi/todo"
	"github.com/pkg/errors"
)

var ErrNonNumericTaskID = errors.New("task id in path must be numeric")
var ErrResourceNotFound = errors.New("resource not found")
var ErrMethodNotAllowed = errors.New("method not allowed")

type ErrInvalidRequestBody struct{ err error }

func (e ErrInvalidRequestBody) Error() string { return fmt.Sprintf("invalid request body: %v", e.err) }

func MakeHandler(s Service, logger log.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}

	saveTaskHandler := kithttp.NewServer(
		makeSaveTaskEndpoint(s),
		decodeSaveTaskRequest,
		encodeResponse,
		opts...,
	)

	listTasksHandler := kithttp.NewServer(
		makeListTasksEndpoint(s),
		kithttp.NopRequestDecoder,
		encodeResponse,
		opts...,
	)

	removeTaskHandler := kithttp.NewServer(
		makeRemoveTaskEndpoint(s),
		decodeRemoveTaskRequest,
		encodeResponse,
		opts...,
	)

	toggleTaskHandler := kithttp.NewServer(
		makeToggleTaskEndpoint(s),
		decodeToggleTaskRequest,
		encodeResponse,
		opts...,
	)

	r := mux.NewRouter()

	r.Handle("/checklist/v1/tasks", saveTaskHandler).Methods("POST")
	r.Handle("/checklist/v1/tasks", listTasksHandler).Methods("GET")
	r.Handle("/checklist/v1/task/{id:[0-9]+}", removeTaskHandler).Methods("DELETE")
	r.Handle("/checklist/v1/task/{id:[0-9]+}", toggleTaskHandler).Methods("PATCH")

	r.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowed)
	r.NotFoundHandler = http.HandlerFunc(notFound)

	return r
}

func decodeSaveTaskRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request saveTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, ErrInvalidRequestBody{err}
	}
	return request, nil
}

func decodeRemoveTaskRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, ErrNonNumericTaskID
	}
	return removeTaskRequest{id}, nil
}

func decodeToggleTaskRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, ErrNonNumericTaskID
	}
	return toggleTaskRequest{id}, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(endpoint.Failer); ok && e.Failed() != nil {
		encodeError(ctx, e.Failed(), w)
		return nil
	}

	if response == nil {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

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

func notFound(w http.ResponseWriter, r *http.Request) {
	encodeError(context.Background(), ErrResourceNotFound, w)
}

func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	encodeError(context.Background(), ErrMethodNotAllowed, w)
}
