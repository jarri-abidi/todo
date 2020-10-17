package todos

import (
	"errors"
	"fmt"
)

// Todo represents a single item on the todo-list.
type Todo struct {
	ID   int64
	Name string
	Done bool
}

// Service is the interface to interact with Todo(s).
type Service interface {
	Save(todo *Todo) error
	List() ([]Todo, error)
	ToggleDone(id int64) error
	Remove(id int64) error
	Update(todo *Todo) error
}

// Store is the interface used to persist the Todo(s).
type Store interface {
	Insert(todo *Todo) error
	FindAll() ([]Todo, error)
	FindByID(id int64) (*Todo, error)
	Update(todo *Todo) error
	DeleteByID(id int64) error
}

type todoService struct {
	store Store
}

var (
	ErrTodoNotFound      = errors.New("todo not found")
	ErrTodoAlreadyExists = errors.New("todo already exists")
)

func NewService(store Store) Service {
	return &todoService{store: store}
}

func (s *todoService) Save(todo *Todo) error {
	if err := s.store.Insert(todo); err != nil {
		return fmt.Errorf("could not save todo: %v", err)
	}

	return nil
}

func (s *todoService) List() ([]Todo, error) {
	todos, err := s.store.FindAll()
	if err != nil {
		return nil, fmt.Errorf("could not list todos: %v", err)
	}

	return todos, nil
}

func (s *todoService) ToggleDone(id int64) error {
	todo, err := s.store.FindByID(id)
	if err == ErrTodoNotFound {
		return err
	}
	if err != nil {
		return fmt.Errorf("could not find todo: %v", err)
	}

	todo.Done = !todo.Done

	err = s.store.Update(todo)
	if err != nil {
		return fmt.Errorf("could not toggle todo: %v", err)
	}

	return nil
}

func (s *todoService) Remove(id int64) error {
	err := s.store.DeleteByID(id)
	if err == ErrTodoNotFound {
		return err
	}
	if err != nil {
		return fmt.Errorf("could not delete todo: %v", err)
	}

	return nil
}

func (s *todoService) Update(todo *Todo) error {
	return nil
}
