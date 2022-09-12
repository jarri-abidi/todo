package todolist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/jarri-abidi/todolist/todos"
)

var ErrNonNumericTodoID = errors.New("todo id in path must be numeric")

type ErrInvalidRequestBody struct{ err error }

func (e ErrInvalidRequestBody) Error() string { return fmt.Sprintf("invalid request body: %v", e.err) }

func MakeHandler(s Service, logger log.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}

	saveTodoHandler := kithttp.NewServer(
		makeSaveTodoEndpoint(s),
		decodeSaveTodoRequest,
		encodeResponse,
		opts...,
	)

	listTodosHandler := kithttp.NewServer(
		makeListTodosEndpoint(s),
		kithttp.NopRequestDecoder,
		encodeResponse,
		opts...,
	)

	removeTodoHandler := kithttp.NewServer(
		makeRemoveTodoEndpoint(s),
		decodeRemoveTodoRequest,
		encodeResponse,
		opts...,
	)

	toggleTodoHandler := kithttp.NewServer(
		makeToggleTodoEndpoint(s),
		decodeToggleTodoRequest,
		encodeResponse,
		opts...,
	)

	r := mux.NewRouter()

	r.Handle("/todolist/v1/todos", saveTodoHandler).Methods("POST")
	r.Handle("/todolist/v1/todos", listTodosHandler).Methods("GET")
	r.Handle("/todolist/v1/todo/{id:[0-9]+}", removeTodoHandler).Methods("DELETE")
	r.Handle("/todolist/v1/todo/{id:[0-9]+}", toggleTodoHandler).Methods("PATCH")

	return r
}

func decodeSaveTodoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		Name string `json:"name"`
		Done bool   `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, ErrInvalidRequestBody{err}
	}
	return saveTodoRequest{Name: body.Name}, nil
}

func decodeRemoveTodoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, ErrNonNumericTodoID
	}
	return removeTodoRequest{id}, nil
}

func decodeToggleTodoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, ErrNonNumericTodoID
	}
	return toggleTodoRequest{id}, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}

	if response == nil {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

type errorer interface {
	error() error
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	switch err {
	case todos.ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
	case todos.ErrAlreadyExists:
		w.WriteHeader(http.StatusConflict)
	case ErrNonNumericTodoID:
		w.WriteHeader(http.StatusBadRequest)
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
