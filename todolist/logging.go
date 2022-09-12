package todolist

import (
	"context"
	"time"

	"github.com/go-kit/log"

	"github.com/jarri-abidi/todolist/todos"
)

type loggingService struct {
	logger log.Logger
	Service
}

// NewLoggingService returns a new instance of a logging Service.
func NewLoggingService(logger log.Logger, s Service) Service {
	return &loggingService{logger, s}
}

func (s *loggingService) Save(ctx context.Context, todo *todos.Todo) (err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "save",
			"name", todo.Name,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.Service.Save(ctx, todo)
}

func (s *loggingService) List(ctx context.Context) (todos []todos.Todo, err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "list",
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.Service.List(ctx)
}

func (s *loggingService) Remove(ctx context.Context, id int64) (err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "remove",
			"id", id,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.Service.Remove(ctx, id)
}

func (s *loggingService) ToggleDone(ctx context.Context, id int64) (err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "toggle_done",
			"id", id,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.Service.ToggleDone(ctx, id)
}
