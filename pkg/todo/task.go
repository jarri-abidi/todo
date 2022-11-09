package todo

import (
	"context"
	"errors"
)

var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrTaskAlreadyExists = errors.New("task already exists")
)

// Task represents a task that may need to be performed.
// It is purely a domain entity with no context of usecases or applications.
type Task struct {
	ID   int64
	Name string
	Done bool
}

// TaskRepository is the interface used to persist the Task(s).
type TaskRepository interface {
	Insert(context.Context, *Task) error
	FindAll(context.Context) ([]Task, error)
	FindByID(ctx context.Context, id int64) (*Task, error)
	Update(context.Context, *Task) error
	DeleteByID(ctx context.Context, id int64) error
}
