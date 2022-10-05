package checklist

import (
	"context"
	"encoding/json"

	"github.com/go-kit/kit/endpoint"
	"go.opentelemetry.io/contrib/instrumentation/github.com/go-kit/kit/otelkit"

	"github.com/jarri-abidi/todo"
)

func makeSaveTaskEndpoint(s Service) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(saveTaskRequest)
		task := &todo.Task{Name: req.Name}
		err := s.Save(ctx, task)
		return saveTaskResponse{taskResponse{task.ID, task.Name, task.Done}, err}, nil
	}

	return otelkit.EndpointMiddleware(otelkit.WithOperation("saveTask"))(ep)
}

func makeListTasksEndpoint(s Service) endpoint.Endpoint {
	ep := func(ctx context.Context, _ interface{}) (interface{}, error) {
		tasks, err := s.List(ctx)
		var resp = listTasksResponse{Tasks: make([]taskResponse, 0, len(tasks)), Err: err}
		if err != nil {
			return resp, nil
		}
		for _, task := range tasks {
			resp.Tasks = append(resp.Tasks, taskResponse{ID: task.ID, Name: task.Name, Done: task.Done})
		}
		return resp, nil
	}

	return otelkit.EndpointMiddleware(otelkit.WithOperation("listTasks"))(ep)
}

func makeRemoveTaskEndpoint(s Service) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(removeTaskRequest)
		err := s.Remove(ctx, req.ID)
		return emptyResponse{Err: err}, nil
	}

	return otelkit.EndpointMiddleware(otelkit.WithOperation("removeTask"))(ep)
}

func makeToggleTaskEndpoint(s Service) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(toggleTaskRequest)
		err := s.ToggleDone(ctx, req.ID)
		return emptyResponse{Err: err}, nil
	}

	return otelkit.EndpointMiddleware(otelkit.WithOperation("toggleTask"))(ep)
}

type taskResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Done bool   `json:"done"`
}

type saveTaskRequest struct {
	Name string `json:"name"`
}

type saveTaskResponse struct {
	taskResponse
	Err error `json:"error,omitempty"`
}

func (r saveTaskResponse) Failed() error { return r.Err }

type listTasksResponse struct {
	Tasks []taskResponse
	Err   error `json:"error,omitempty"`
}

func (r listTasksResponse) MarshalJSON() ([]byte, error) {
	if r.Failed() == nil {
		return json.Marshal(r.Tasks)
	}
	return json.Marshal(r)
}

func (r listTasksResponse) Failed() error { return r.Err }

type removeTaskRequest struct {
	ID int64
}

type toggleTaskRequest struct {
	ID int64
}

type emptyResponse struct {
	Err error `json:"error,omitempty"`
}

func (r emptyResponse) Failed() error { return r.Err }
