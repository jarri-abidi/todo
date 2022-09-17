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

func makeSaveTodoEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(saveTodoRequest)
		todo := &todos.Todo{Name: req.Name}
		err := s.Save(ctx, todo)
		return saveTodoResponse{todo, err}, nil
	}
}

type listTodosResponse struct {
	Todos []todos.Todo `json:"todos"`
	Err   error        `json:"error,omitempty"`
}

func (r listTodosResponse) error() error { return r.Err }

func makeListTodosEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		todos, err := s.List(ctx)
		return listTodosResponse{todos, err}, nil
	}
}

type removeTodoRequest struct {
	id int64
}

func makeRemoveTodoEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(removeTodoRequest)
		err := s.Remove(ctx, req.id)
		return nil, err
	}
}

type toggleTodoRequest struct {
	id int64
}

func makeToggleTodoEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(toggleTodoRequest)
		err := s.ToggleDone(ctx, req.id)
		return nil, err
	}
}
