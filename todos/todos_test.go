package todos_test

import (
	"testing"

	"github.com/jarri-abidi/todolist/inmem"
	"github.com/jarri-abidi/todolist/todos"
)

func TestSave(t *testing.T) {
	svc := todos.NewService(inmem.NewTodoStore())

	todo := todos.Todo{
		Name: "Kachra phenk k ao",
		Done: false,
	}

	err := svc.Save(&todo)
	if err != nil {
		t.Fatalf("could not save todo: %v", err)
	}
}

func TestList(t *testing.T) {
	svc := todos.NewService(inmem.NewTodoStore())

	expected := []todos.Todo{
		{
			Name: "Kachra phenk k ao",
			Done: false,
		},
		{
			Name: "Roti le kar ao",
			Done: false,
		},
		{
			Name: "Geezer chala do",
			Done: false,
		},
	}

	for i := range expected {
		err := svc.Save(&expected[i])
		if err != nil {
			t.Fatalf("could not save todo: %v", err)
		}
	}

	todolist, err := svc.List()
	if err != nil {
		t.Fatalf("could not list todos: %v", err)
	}

	for i := range todolist {
		if todolist[i].ID != expected[i].ID {
			t.Fatalf("ID does not match for todo at index %d", i)
		}
		if todolist[i].Name != expected[i].Name {
			t.Fatalf("Name does not match for todo at index %d", i)
		}
		if todolist[i].Done != expected[i].Done {
			t.Fatalf("Done does not match for todo at index %d", i)
		}
	}
}

func TestToggleDone(t *testing.T) {
	svc := todos.NewService(inmem.NewTodoStore())

	todo := todos.Todo{
		Name: "Kachra phenk k ao",
		Done: false,
	}

	err := svc.Save(&todo)
	if err != nil {
		t.Fatalf("could not save todo: %v", err)
	}

	err = svc.ToggleDone(todo.ID)
	if err != nil {
		t.Fatalf("could not toggle todo: %v", err)
	}

	todolist, err := svc.List()
	if err != nil {
		t.Fatalf("could not list todos: %v", err)
	}

	if todolist[0].Done != true {
		t.Fatal("expected true, got false")
	}
}

func TestRemove(t *testing.T) {
	svc := todos.NewService(inmem.NewTodoStore())

	todo := todos.Todo{
		Name: "Kachra phenk k ao",
		Done: true,
	}

	err := svc.Save(&todo)
	if err != nil {
		t.Fatalf("could not save todo: %v", err)
	}

	err = svc.Remove(todo.ID)
	if err != nil {
		t.Fatalf("could not remove todo: %v", err)
	}

	todolist, err := svc.List()
	if err != nil {
		t.Fatalf("could not list todos: %v", err)
	}

	if len(todolist) > 0 {
		t.Fatal("expected list to be empty after removing todo")
	}
}

func TestUpdate(t *testing.T) {
	svc := todos.NewService(inmem.NewTodoStore())

	todo := todos.Todo{Name: "Internet ki complaint karo"}
	if err := svc.Save(&todo); err != nil {
		t.Fatalf("could not save todo: %v", err)
	}

	todo.Name = "Bijli* ki complaint karo"
	todo.Done = true
	if err := svc.Update(&todo); err != nil {
		t.Fatalf("could not update todo: %v", err)
	}

	todolist, err := svc.List()
	if err != nil {
		t.Fatalf("could not list todos: %v", err)
	}

	if len(todolist) != 1 {
		t.Fatalf("unexpected number of todos after update")
	}

	if todo.ID != todolist[0].ID {
		t.Fatalf("expected IDs to match")
	}

	if todo.Done != todolist[0].Done {
		t.Fatalf("expected Done to be updated")
	}

	if todo.Name != todolist[0].Name {
		t.Fatalf("expected Name to be updated")
	}
}
