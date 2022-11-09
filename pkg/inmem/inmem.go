package inmem

import (
	"context"
	"sync"

	"github.com/jarri-abidi/todo/pkg/todo"
)

type taskRepository struct {
	sync.RWMutex
	tasklist []todo.Task
	used     map[int64]bool
	counter  int64
}

// NewTaskRepository returns an in-memory implementation of todo.TaskRepository.
// This lets us run tests and start the app locally without a persistent database.
func NewTaskRepository() todo.TaskRepository {
	return &taskRepository{tasklist: []todo.Task{}, used: make(map[int64]bool)}
}

func (ts *taskRepository) Insert(_ context.Context, task *todo.Task) error {
	ts.Lock()
	defer ts.Unlock()

	if task.ID != 0 {
		if ts.used[task.ID] {
			return todo.ErrTaskAlreadyExists
		}

		ts.used[task.ID] = true
		ts.tasklist = append(ts.tasklist, *task)
		return nil
	}

	ts.counter++
	for ts.used[ts.counter] {
		ts.counter++
	}
	ts.used[task.ID] = true
	task.ID = ts.counter
	ts.tasklist = append(ts.tasklist, *task)
	return nil
}

func (ts *taskRepository) FindAll(_ context.Context) ([]todo.Task, error) {
	ts.RLock()
	defer ts.RUnlock()

	return ts.tasklist, nil
}

func (ts *taskRepository) FindByID(_ context.Context, id int64) (*todo.Task, error) {
	ts.RLock()
	defer ts.RUnlock()

	for i, task := range ts.tasklist {
		if task.ID == id {
			return &ts.tasklist[i], nil
		}
	}
	return nil, todo.ErrTaskNotFound
}

func (ts *taskRepository) Update(_ context.Context, task *todo.Task) error {
	ts.Lock()
	defer ts.Unlock()

	for i, t := range ts.tasklist {
		if t.ID == task.ID {
			ts.tasklist[i] = *task
			return nil
		}
	}
	return todo.ErrTaskNotFound
}

func (ts *taskRepository) DeleteByID(_ context.Context, id int64) error {
	ts.Lock()
	defer ts.Unlock()

	for i, task := range ts.tasklist {
		if task.ID == id {
			ts.tasklist = append(ts.tasklist[:i], ts.tasklist[i+1:]...)
			return nil
		}
	}
	return todo.ErrTaskNotFound
}
