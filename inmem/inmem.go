package inmem

import (
	"context"
	"sync"

	"github.com/jarri-abidi/todolist/todos"
)

type todoStore struct {
	sync.RWMutex
	todolist []todos.Todo
	used     map[int64]bool
	counter  int64
}

func NewTodoStore() todos.Store {
	return &todoStore{todolist: []todos.Todo{}, used: make(map[int64]bool)}
}

func (ts *todoStore) Insert(_ context.Context, todo *todos.Todo) error {
	ts.Lock()
	defer ts.Unlock()

	if todo.ID != 0 {
		if ts.used[todo.ID] {
			return todos.ErrAlreadyExists
		}

		ts.used[todo.ID] = true
	} else {
		ts.counter++
		for ts.used[ts.counter] {
			ts.counter++
		}

		ts.used[todo.ID] = true
		todo.ID = ts.counter
	}

	ts.todolist = append(ts.todolist, *todo)

	return nil
}

func (ts *todoStore) FindAll(_ context.Context) ([]todos.Todo, error) {
	ts.RLock()
	defer ts.RUnlock()

	return ts.todolist, nil
}

func (ts *todoStore) FindByID(_ context.Context, id int64) (*todos.Todo, error) {
	ts.RLock()
	defer ts.RUnlock()

	for i, todo := range ts.todolist {
		if todo.ID == id {
			return &ts.todolist[i], nil
		}
	}
	return nil, todos.ErrNotFound
}

func (ts *todoStore) Update(_ context.Context, todo *todos.Todo) error {
	ts.Lock()
	defer ts.Unlock()

	for i, t := range ts.todolist {
		if t.ID == todo.ID {
			ts.todolist[i] = *todo
			return nil
		}
	}
	return todos.ErrNotFound
}

func (ts *todoStore) DeleteByID(_ context.Context, id int64) error {
	ts.Lock()
	defer ts.Unlock()

	for i, todo := range ts.todolist {
		if todo.ID == id {
			ts.todolist = append(ts.todolist[:i], ts.todolist[i+1:]...)
			return nil
		}
	}
	return todos.ErrNotFound
}
