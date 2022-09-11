package todolist

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"github.com/jarri-abidi/todolist/todos"
)

type saveTodoRequest struct {
	Name string
}

type saveTodoResponse struct {
	Todo *todos.Todo `json:"todo,omitempty"`
	Err  error       `json:"error,omitempty"`
}

func (r saveTodoResponse) error() error { return r.Err }

type listTodosResponse struct {
	Todos []todos.Todo `json:"todos"`
	Err   error        `json:"error,omitempty"`
}

func makeSaveTodoEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(saveTodoRequest)
		todo := &todos.Todo{Name: req.Name}
		err := s.Save(ctx, todo)
		return saveTodoResponse{todo, err}, nil
	}
}

func makeListTodosEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		todos, err := s.List(ctx)
		return listTodosResponse{todos, err}, nil
	}
}
