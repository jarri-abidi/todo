package checklist

import (
	"context"
	"fmt"

	"github.com/jarri-abidi/todo"
)

// Service is an application service that lets us interact with a list of tasks.
type Service interface {
	Save(context.Context, *todo.Task) error
	List(context.Context) ([]todo.Task, error)
	ToggleDone(ctx context.Context, id int64) error
	Remove(ctx context.Context, id int64) error
	Update(context.Context, *todo.Task) error
}

// Middleware describes a Service middleware.
type Middleware func(Service) Service

type service struct {
	repository todo.TaskRepository
}

func NewService(repository todo.TaskRepository) Service {
	return &service{repository: repository}
}

func (s *service) Save(ctx context.Context, task *todo.Task) error {
	if err := s.repository.Insert(ctx, task); err != nil {
		return fmt.Errorf("could not save task: %v", err)
	}
	return nil
}

func (s *service) List(ctx context.Context) ([]todo.Task, error) {
	list, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not list task: %v", err)
	}
	return list, nil
}

func (s *service) ToggleDone(ctx context.Context, id int64) error {
	task, err := s.repository.FindByID(ctx, id)
	if err == todo.ErrTaskNotFound {
		return err
	}
	if err != nil {
		return fmt.Errorf("could not find task: %v", err)
	}

	task.Done = !task.Done
	err = s.repository.Update(ctx, task)
	if err != nil {
		return fmt.Errorf("could not toggle task: %v", err)
	}
	return nil
}

func (s *service) Remove(ctx context.Context, id int64) error {
	err := s.repository.DeleteByID(ctx, id)
	if err == todo.ErrTaskNotFound {
		return err
	}
	if err != nil {
		return fmt.Errorf("could not delete task: %v", err)
	}
	return nil
}

func (s *service) Update(ctx context.Context, task *todo.Task) error {
	return nil
}
