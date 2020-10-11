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
		t.Fatal(err)
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
			Name: "geezer chala do",
			Done: false,
		},
	}

	for i := range expected {
		err := svc.Save(&expected[i])
		if err != nil {
			t.Fatal(err)
		}
	}

	todolist, err := svc.List()
	if err != nil {
		t.Fatal(err)
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
		t.Fatal(err)
	}

	err = svc.ToggleDone(todo.ID)
	if err != nil {
		t.Fatal(err)
	}

	todolist, err := svc.List()
	if err != nil {
		t.Fatal(err)
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
		t.Fatal(err)
	}

	err = svc.Remove(todo.ID)
	if err != nil {
		t.Fatal(err)
	}

	todolist, err := svc.List()
	if err != nil {
		t.Fatal(err)
	}

	if len(todolist) > 0 {
		t.Fatal("expected list to be empty after removing todo")
	}
}
