package checklist

import (
	"context"
	"encoding/json"

	"github.com/go-kit/kit/endpoint"
	"go.opentelemetry.io/contrib/instrumentation/github.com/go-kit/kit/otelkit"

	"github.com/jarri-abidi/todo"
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

func (r saveTodoResponse) Failed() error { return r.Err }

func makeSaveTodoEndpoint(s Service) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(saveTodoRequest)
		task := &todo.Task{Name: req.Name}
		err := s.Save(ctx, task)
		return saveTodoResponse{task.ID, task.Name, task.Done, err}, nil
	}

	return otelkit.EndpointMiddleware(otelkit.WithOperation("saveTodo"))(ep)
}

type listTodosResponse struct {
	Todos []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Done bool   `json:"done"`
	}
	Err error `json:"error,omitempty"`
}

func (r listTodosResponse) MarshalJSON() ([]byte, error) {
	if r.Failed() == nil {
		return json.Marshal(r.Todos)
	}
	return json.Marshal(r)
}

func (r listTodosResponse) Failed() error { return r.Err }

func makeListTodosEndpoint(s Service) endpoint.Endpoint {
	ep := func(ctx context.Context, _ interface{}) (interface{}, error) {
		tasks, err := s.List(ctx)
		var resp = listTodosResponse{Todos: make([]struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
			Done bool   `json:"done"`
		}, 0), Err: err}
		if err != nil {
			return resp, nil
		}
		for _, task := range tasks {
			resp.Todos = append(resp.Todos, struct {
				ID   int64  `json:"id"`
				Name string `json:"name"`
				Done bool   `json:"done"`
			}{ID: task.ID, Name: task.Name, Done: task.Done})
		}
		return resp, nil
	}

	return otelkit.EndpointMiddleware(otelkit.WithOperation("listTodos"))(ep)
}

type removeTodoRequest struct {
	id int64
}

func makeRemoveTodoEndpoint(s Service) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(removeTodoRequest)
		err := s.Remove(ctx, req.id)
		return nil, err
	}

	return otelkit.EndpointMiddleware(otelkit.WithOperation("removeTodo"))(ep)
}

type toggleTodoRequest struct {
	id int64
}

func makeToggleTodoEndpoint(s Service) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(toggleTodoRequest)
		err := s.ToggleDone(ctx, req.id)
		return nil, err
	}

	return otelkit.EndpointMiddleware(otelkit.WithOperation("toggleTodo"))(ep)
}
