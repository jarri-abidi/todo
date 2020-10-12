package inmem

import (
	"sync"

	"github.com/jarri-abidi/todolist/todos"
)

type todoStore struct {
	sync.RWMutex
	todolist []todos.Todo
	counter  int64
}

func NewTodoStore() todos.Store {
	return &todoStore{todolist: []todos.Todo{}}
}

func (ts *todoStore) Insert(t *todos.Todo) error {
	ts.Lock()
	defer ts.Unlock()

	ts.counter++
	t.ID = ts.counter

	ts.todolist = append(ts.todolist, *t)

	return nil
}

func (ts *todoStore) FindAll() ([]todos.Todo, error) {
	ts.RLock()
	defer ts.RUnlock()

	return ts.todolist, nil
}

func (ts *todoStore) FindByID(id int64) (*todos.Todo, error) {
	ts.RLock()
	defer ts.RUnlock()

	for i, todo := range ts.todolist {
		if todo.ID == id {
			return &ts.todolist[i], nil
		}
	}
	return nil, todos.ErrTodoNotFound
}

func (ts *todoStore) Update(t *todos.Todo) error {
	ts.Lock()
	defer ts.Unlock()

	for i, todo := range ts.todolist {
		if todo.ID == t.ID {
			ts.todolist[i] = *t
			return nil
		}
	}
	return todos.ErrTodoNotFound
}

func (ts *todoStore) DeleteByID(id int64) error {
	ts.Lock()
	defer ts.Unlock()

	for i, todo := range ts.todolist {
		if todo.ID == id {
			ts.todolist = append(ts.todolist[:i], ts.todolist[i+1:]...)
			return nil
		}
	}
	return todos.ErrTodoNotFound
}
