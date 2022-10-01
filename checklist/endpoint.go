package checklist

import (
	"context"
	"encoding/json"

	"github.com/go-kit/kit/endpoint"
	"go.opentelemetry.io/contrib/instrumentation/github.com/go-kit/kit/otelkit"

	"github.com/jarri-abidi/todo"
)

type saveTaskRequest struct {
	Name string
}

type saveTaskResponse struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Done bool   `json:"done,omitempty"`
	Err  error  `json:"error,omitempty"`
}

func (r saveTaskResponse) Failed() error { return r.Err }

func makeSaveTaskEndpoint(s Service) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(saveTaskRequest)
		task := &todo.Task{Name: req.Name}
		err := s.Save(ctx, task)
		return saveTaskResponse{task.ID, task.Name, task.Done, err}, nil
	}

	return otelkit.EndpointMiddleware(otelkit.WithOperation("saveTask"))(ep)
}

type listTasksResponse struct {
	Tasks []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Done bool   `json:"done"`
	}
	Err error `json:"error,omitempty"`
}

func (r listTasksResponse) MarshalJSON() ([]byte, error) {
	if r.Failed() == nil {
		return json.Marshal(r.Tasks)
	}
	return json.Marshal(r)
}

func (r listTasksResponse) Failed() error { return r.Err }

func makeListTasksEndpoint(s Service) endpoint.Endpoint {
	ep := func(ctx context.Context, _ interface{}) (interface{}, error) {
		tasks, err := s.List(ctx)
		var resp = listTasksResponse{Tasks: make([]struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
			Done bool   `json:"done"`
		}, 0), Err: err}
		if err != nil {
			return resp, nil
		}
		for _, task := range tasks {
			resp.Tasks = append(resp.Tasks, struct {
				ID   int64  `json:"id"`
				Name string `json:"name"`
				Done bool   `json:"done"`
			}{ID: task.ID, Name: task.Name, Done: task.Done})
		}
		return resp, nil
	}

	return otelkit.EndpointMiddleware(otelkit.WithOperation("listTasks"))(ep)
}

type removeTaskRequest struct {
	id int64
}

func makeRemoveTaskEndpoint(s Service) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(removeTaskRequest)
		err := s.Remove(ctx, req.id)
		return nil, err
	}

	return otelkit.EndpointMiddleware(otelkit.WithOperation("removeTask"))(ep)
}

type toggleTaskRequest struct {
	id int64
}

func makeToggleTaskEndpoint(s Service) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(toggleTaskRequest)
		err := s.ToggleDone(ctx, req.id)
		return nil, err
	}

	return otelkit.EndpointMiddleware(otelkit.WithOperation("toggleTask"))(ep)
}
