package todolist

import (
	"context"
	"time"

	"github.com/go-kit/log"

	"github.com/jarri-abidi/todolist/todos"
)

// LoggingMiddleware takes a logger as a dependency and returns a service Middleware.
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(s Service) Service { return &loggingMiddleware{logger, s} }
}

type loggingMiddleware struct {
	logger log.Logger
	Service
}

func (s *loggingMiddleware) Save(ctx context.Context, todo *todos.Todo) (err error) {
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

func (s *loggingMiddleware) List(ctx context.Context) (todos []todos.Todo, err error) {
	defer func(begin time.Time) {
		s.logger.Log(
			"method", "list",
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return s.Service.List(ctx)
}

func (s *loggingMiddleware) Remove(ctx context.Context, id int64) (err error) {
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

func (s *loggingMiddleware) ToggleDone(ctx context.Context, id int64) (err error) {
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
