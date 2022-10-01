package checklist_test

import (
	"context"
	"testing"

	"github.com/jarri-abidi/todo"
	"github.com/jarri-abidi/todo/checklist"
	"github.com/jarri-abidi/todo/inmem"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSave(t *testing.T) {
	var (
		assert = require.New(t)
		svc    = checklist.NewService(inmem.NewTaskRepository())
	)

	task := todo.Task{Name: "Kachra phenk k ao", Done: false}

	assert.NoError(svc.Save(context.TODO(), &task), "could not save task")
}

func TestList(t *testing.T) {
	var (
		require = require.New(t)
		assert  = assert.New(t)
		svc     = checklist.NewService(inmem.NewTaskRepository())
	)

	expected := []todo.Task{
		{Name: "Kachra phenk k ao", Done: false},
		{Name: "Roti le kar ao", Done: false},
		{Name: "Geezer chala do", Done: false},
	}

	for i := range expected {
		require.NoError(svc.Save(context.TODO(), &expected[i]), "could not save task")
	}

	list, err := svc.List(context.TODO())
	require.NoError(err, "could not list tasks")

	for i := range list {
		assert.Equal(expected[i].ID, list[i].ID, "IDs need to match")
		assert.Equal(expected[i].Name, list[i].Name, "Names need to match")
		assert.Equal(expected[i].Done, list[i].Done, "Done needs to match")
	}
}

func TestToggleDone(t *testing.T) {
	var (
		require = require.New(t)
		assert  = assert.New(t)
		svc     = checklist.NewService(inmem.NewTaskRepository())
	)

	task := todo.Task{Name: "Kachra phenk k ao", Done: false}
	require.NoError(svc.Save(context.TODO(), &task), "could not save task")

	require.NoError(svc.ToggleDone(context.TODO(), task.ID), "could not toggle task")

	list, err := svc.List(context.TODO())
	assert.NoError(err, "could not list tasks")
	assert.True(list[0].Done, "expected task to be done")
}

func TestRemove(t *testing.T) {
	var (
		require = require.New(t)
		assert  = assert.New(t)
		svc     = checklist.NewService(inmem.NewTaskRepository())
	)

	task := todo.Task{Name: "Kachra phenk k ao", Done: true}
	require.NoError(svc.Save(context.TODO(), &task), "could not save task")

	require.NoError(svc.Remove(context.TODO(), task.ID), "could not remove task")

	list, err := svc.List(context.TODO())
	assert.NoError(err, "could not list tasks")
	assert.Empty(list, "expected list to be empty after removing task")
}

func NoTestUpdate(t *testing.T) {
	var (
		require = require.New(t)
		assert  = assert.New(t)
		svc     = checklist.NewService(inmem.NewTaskRepository())
	)

	task := todo.Task{Name: "Internet ki complaint karo"}
	require.NoError(svc.Save(context.TODO(), &task), "could not save task")

	task.Name = "Bijli* ki complaint karo"
	task.Done = true
	require.NoError(svc.Update(context.TODO(), &task), "could not update task")

	list, err := svc.List(context.TODO())
	assert.NoError(err, "could not list tasks")
	assert.Equal(1, len(list), "unexpected number of tasks after update")
	assert.Equal(list[0].ID, task.ID, "expected IDs to match")
	assert.Equal(list[0].Name, task.Name, "expected Name to be updated")
	assert.Equal(list[0].Done, task.Done, "expected Done to be updated")
}
