package inmem

import (
	"fmt"
	"sync"

	"github.com/jarri-abidi/todolist/todos"
)

type todoStore struct {
	sync.RWMutex
	todolist []todos.Todo
	counter  int64
}

func NewTodoStore() todos.Store {
	return &todoStore{
		todolist: []todos.Todo{},
		counter:  1,
	}
}

func (ts *todoStore) incrementCounter() {
	ts.counter++
}

func (ts *todoStore) Insert(t *todos.Todo) error {
	ts.Lock()
	defer ts.Unlock()

	t.ID = ts.counter
	defer ts.incrementCounter()

	ts.todolist = append(ts.todolist, *t)

	return nil
}

func (ts *todoStore) FindAll() ([]todos.Todo, error) {
	return ts.todolist, nil
}

func (ts *todoStore) FindByID(id int64) (*todos.Todo, error) {
	for i, todo := range ts.todolist {
		if todo.ID == id {
			return &ts.todolist[i], nil
		}
	}
	return nil, fmt.Errorf("todo does not exist")
}

func (ts *todoStore) Update(t *todos.Todo) error {
	for i, todo := range ts.todolist {
		if todo.ID == t.ID {
			ts.todolist[i] = *t
			return nil
		}
	}
	return fmt.Errorf("todo does not exist")
}

func (ts *todoStore) DeleteByID(id int64) error {
	for i, todo := range ts.todolist {
		if todo.ID == id {
			ts.todolist = append(ts.todolist[:i], ts.todolist[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("todo does not exist")
}
