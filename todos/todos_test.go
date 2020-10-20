package todos_test

import (
	"testing"

	"github.com/matryer/is"

	"github.com/jarri-abidi/todolist/inmem"
	"github.com/jarri-abidi/todolist/todos"
)

func TestSave(t *testing.T) {
	var (
		is  = is.New(t)
		svc = todos.NewService(inmem.NewTodoStore())
	)

	todo := todos.Todo{Name: "Kachra phenk k ao", Done: false}

	is.NoErr(svc.Save(&todo)) // could not save todo
}

func TestList(t *testing.T) {
	var (
		is  = is.New(t)
		svc = todos.NewService(inmem.NewTodoStore())
	)

	expected := []todos.Todo{
		{Name: "Kachra phenk k ao", Done: false},
		{Name: "Roti le kar ao", Done: false},
		{Name: "Geezer chala do", Done: false},
	}

	for i := range expected {
		is.NoErr(svc.Save(&expected[i])) // could not save todo
	}

	todolist, err := svc.List()
	is.NoErr(err) // could not list todos

	for i := range todolist {
		is.Equal(todolist[i].ID, expected[i].ID)     // IDs need to match
		is.Equal(todolist[i].Name, expected[i].Name) // Names need to match
		is.Equal(todolist[i].Done, expected[i].Done) // Done needs to match
	}
}

func TestToggleDone(t *testing.T) {
	var (
		is  = is.New(t)
		svc = todos.NewService(inmem.NewTodoStore())
	)

	todo := todos.Todo{Name: "Kachra phenk k ao", Done: false}
	is.NoErr(svc.Save(&todo)) // could not save todo

	is.NoErr(svc.ToggleDone(todo.ID)) // could not toggle todo

	todolist, err := svc.List()
	is.NoErr(err)             // could not list todos
	is.True(todolist[0].Done) // expected todo to be done
}

func TestRemove(t *testing.T) {
	var (
		is  = is.New(t)
		svc = todos.NewService(inmem.NewTodoStore())
	)

	todo := todos.Todo{Name: "Kachra phenk k ao", Done: true}
	is.NoErr(svc.Save(&todo)) // could not save todo

	is.NoErr(svc.Remove(todo.ID)) // could not remove todo

	todolist, err := svc.List()
	is.NoErr(err)              // could not list todos
	is.Equal(len(todolist), 0) // expected list to be empty after removing todo
}

func TestUpdate(t *testing.T) {
	var (
		is  = is.New(t)
		svc = todos.NewService(inmem.NewTodoStore())
	)

	todo := todos.Todo{Name: "Internet ki complaint karo"}
	is.NoErr(svc.Save(&todo)) // could not save todo

	todo.Name = "Bijli* ki complaint karo"
	todo.Done = true
	is.NoErr(svc.Update(&todo)) // could not update todo

	todolist, err := svc.List()
	is.NoErr(err)                         // could not list todos
	is.Equal(len(todolist), 1)            // unexpected number of todos after update
	is.Equal(todo.ID, todolist[0].ID)     // expected IDs to match
	is.Equal(todo.Name, todolist[0].Name) // expected Name to be updated
	is.Equal(todo.Done, todolist[0].Done) // expected Done to be updated
}
