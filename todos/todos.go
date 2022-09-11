package todos

import (
	"context"
	"errors"
)

var (
	ErrNotFound      = errors.New("todo not found")
	ErrAlreadyExists = errors.New("todo already exists")
)

// Todo represents a task that may need to be performed.
// It is purely a domain entity with no context of usecases or applications.
type Todo struct {
	ID   int64
	Name string
	Done bool
}

// Store is the interface used to persist the Todo(s).
type Store interface {
	Insert(context.Context, *Todo) error
	FindAll(context.Context) ([]Todo, error)
	FindByID(ctx context.Context, id int64) (*Todo, error)
	Update(context.Context, *Todo) error
	DeleteByID(ctx context.Context, id int64) error
}
