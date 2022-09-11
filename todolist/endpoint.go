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
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Done bool   `json:"done,omitempty"`
	Err  error  `json:"error,omitempty"`
}

func (r saveTodoResponse) error() error { return r.Err }

func makeSaveTodoEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(saveTodoRequest)
		todo := &todos.Todo{Name: req.Name}
		err := s.Save(ctx, todo)
		return saveTodoResponse{todo.ID, todo.Name, todo.Done, err}, nil
	}
}
