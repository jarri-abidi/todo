package todolist_test

import (
	"context"
	"testing"

	"github.com/jarri-abidi/todolist/inmem"
	"github.com/jarri-abidi/todolist/todolist"
	"github.com/jarri-abidi/todolist/todos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSave(t *testing.T) {
	var (
		assert = require.New(t)
		svc    = todolist.NewService(inmem.NewTodoStore())
	)

	todo := todos.Todo{Name: "Kachra phenk k ao", Done: false}

	assert.NoError(svc.Save(context.TODO(), &todo)) // could not save todo
}

func TestList(t *testing.T) {
	var (
		require = require.New(t)
		assert  = assert.New(t)
		svc     = todolist.NewService(inmem.NewTodoStore())
	)

	expected := []todos.Todo{
		{Name: "Kachra phenk k ao", Done: false},
		{Name: "Roti le kar ao", Done: false},
		{Name: "Geezer chala do", Done: false},
	}

	for i := range expected {
		require.NoError(svc.Save(context.TODO(), &expected[i])) // could not save todo
	}

	todolist, err := svc.List(context.TODO())
	require.NoError(err) // could not list todos

	for i := range todolist {
		assert.Equal(expected[i].ID, todolist[i].ID)     // IDs need to match
		assert.Equal(expected[i].Name, todolist[i].Name) // Names need to match
		assert.Equal(expected[i].Done, todolist[i].Done) // Done needs to match
	}
}

func TestToggleDone(t *testing.T) {
	var (
		require = require.New(t)
		assert  = assert.New(t)
		svc     = todolist.NewService(inmem.NewTodoStore())
	)

	todo := todos.Todo{Name: "Kachra phenk k ao", Done: false}
	require.NoError(svc.Save(context.TODO(), &todo)) // could not save todo

	require.NoError(svc.ToggleDone(context.TODO(), todo.ID)) // could not toggle todo

	todolist, err := svc.List(context.TODO())
	assert.NoError(err)           // could not list todos
	assert.True(todolist[0].Done) // expected todo to be done
}

func TestRemove(t *testing.T) {
	var (
		require = require.New(t)
		assert  = assert.New(t)
		svc     = todolist.NewService(inmem.NewTodoStore())
	)

	todo := todos.Todo{Name: "Kachra phenk k ao", Done: true}
	require.NoError(svc.Save(context.TODO(), &todo)) // could not save todo

	require.NoError(svc.Remove(context.TODO(), todo.ID)) // could not remove todo

	todolist, err := svc.List(context.TODO())
	assert.NoError(err)    // could not list todos
	assert.Empty(todolist) // expected list to be empty after removing todo
}

func NoTestUpdate(t *testing.T) {
	var (
		require = require.New(t)
		assert  = assert.New(t)
		svc     = todolist.NewService(inmem.NewTodoStore())
	)

	todo := todos.Todo{Name: "Internet ki complaint karo"}
	require.NoError(svc.Save(context.TODO(), &todo)) // could not save todo

	todo.Name = "Bijli* ki complaint karo"
	todo.Done = true
	require.NoError(svc.Update(context.TODO(), &todo)) // could not update todo

	todolist, err := svc.List(context.TODO())
	assert.NoError(err)                       // could not list todos
	assert.Equal(1, len(todolist))            // unexpected number of todos after update
	assert.Equal(todolist[0].ID, todo.ID)     // expected IDs to match
	assert.Equal(todolist[0].Name, todo.Name) // expected Name to be updated
	assert.Equal(todolist[0].Done, todo.Done) // expected Done to be updated
}
