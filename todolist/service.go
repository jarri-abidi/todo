package todolist

import (
	"context"
	"fmt"

	"github.com/jarri-abidi/todolist/todos"
)

// Service is an application service that lets us interact with a list of todos.
type Service interface {
	Save(context.Context, *todos.Todo) error
	List(context.Context) ([]todos.Todo, error)
	ToggleDone(ctx context.Context, id int64) error
	Remove(ctx context.Context, id int64) error
	Update(context.Context, *todos.Todo) error
}

type service struct {
	store todos.Store
}

func NewService(store todos.Store) Service {
	return &service{store: store}
}

func (s *service) Save(ctx context.Context, todo *todos.Todo) error {
	if err := s.store.Insert(ctx, todo); err != nil {
		return fmt.Errorf("could not save todo: %v", err)
	}
	return nil
}

func (s *service) List(ctx context.Context) ([]todos.Todo, error) {
	todos, err := s.store.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not list todos: %v", err)
	}
	return todos, nil
}

func (s *service) ToggleDone(ctx context.Context, id int64) error {
	todo, err := s.store.FindByID(ctx, id)
	if err == todos.ErrNotFound {
		return err
	}
	if err != nil {
		return fmt.Errorf("could not find todo: %v", err)
	}

	todo.Done = !todo.Done
	err = s.store.Update(ctx, todo)
	if err != nil {
		return fmt.Errorf("could not toggle todo: %v", err)
	}
	return nil
}

func (s *service) Remove(ctx context.Context, id int64) error {
	err := s.store.DeleteByID(ctx, id)
	if err == todos.ErrNotFound {
		return err
	}
	if err != nil {
		return fmt.Errorf("could not delete todo: %v", err)
	}
	return nil
}

func (s *service) Update(ctx context.Context, todo *todos.Todo) error {
	return nil
}
