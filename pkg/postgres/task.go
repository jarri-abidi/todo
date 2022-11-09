package postgres

import (
	"context"
	"database/sql"

	"github.com/jarri-abidi/todo/pkg/postgres/gen"
	"github.com/jarri-abidi/todo/pkg/todo"
)

type taskRepository struct {
	queries *gen.Queries
}

func NewTaskRepository(db *sql.DB) todo.TaskRepository {
	return &taskRepository{queries: gen.New(db)}
}

func (r *taskRepository) Insert(ctx context.Context, task *todo.Task) error {
	inserted, err := r.queries.InsertTask(ctx, task.Name)
	if err != nil {
		return err
	}
	task.ID = inserted.ID
	return nil
}

func (r *taskRepository) FindAll(ctx context.Context) ([]todo.Task, error) {
	var list []todo.Task
	tasks, err := r.queries.FindAllTasks(ctx)
	if err != nil {
		return list, err
	}

	for _, task := range tasks {
		list = append(list, todo.Task{ID: task.ID, Name: task.Name, Done: task.Done.Bool})
	}
	return list, nil
}

func (r *taskRepository) FindByID(ctx context.Context, id int64) (*todo.Task, error) {
	task, err := r.queries.FindTask(ctx, id)
	if err != nil {
		return nil, err
	}
	return &todo.Task{ID: task.ID, Name: task.Name, Done: task.Done.Bool}, nil
}

func (r *taskRepository) Update(ctx context.Context, task *todo.Task) error {
	return r.queries.UpdateTask(ctx, gen.UpdateTaskParams{
		ID: task.ID, Name: task.Name, Done: sql.NullBool{Bool: task.Done},
	})
}

func (r *taskRepository) DeleteByID(ctx context.Context, id int64) error {
	return r.queries.DeleteTask(ctx, id)
}
