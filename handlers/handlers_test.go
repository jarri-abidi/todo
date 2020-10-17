package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jarri-abidi/todolist/handlers"
	"github.com/jarri-abidi/todolist/inmem"
	"github.com/jarri-abidi/todolist/todos"
)

func TestToggleTodo(t *testing.T) {
	tt := []struct {
		Name            string
		TodoID          string
		ExpectedCode    int
		ExpectedRspBody string
		ExpectedDone    bool
	}{
		{
			"Returns 204 and toggles for valid request",
			"1", http.StatusNoContent, `null`, true,
		},
		{
			"Returns 400 and error msg for non-numeric id",
			"meow", http.StatusBadRequest, `{"error":"Invalid todo ID"}`, false,
		},
		{
			"Returns 404 and error msg for id of todo that doesn't exist",
			"1337", http.StatusNotFound, `{"error":"todo not found"}`, false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			s := todos.NewService(inmem.NewTodoStore())
			h := handlers.Handler{Service: s}

			todo := todos.Todo{Name: "Gaari ki service karwalo"}
			if err := s.Save(&todo); err != nil {
				t.Fatalf("could not save todo: %v", err)
			}

			rec := httptest.NewRecorder()
			req, err := http.NewRequest("PATCH", "/todo/:id", nil)
			if err != nil {
				t.Fatalf("could not create http request: %v", err)
			}
			req = mux.SetURLVars(req, map[string]string{"id": tc.TodoID})

			h.ToggleTodo(rec, req)

			if rec.Result().StatusCode != tc.ExpectedCode {
				t.Fatalf("expected code %d, got: %d", tc.ExpectedCode, rec.Result().StatusCode)
			}

			if rec.Body.String() != tc.ExpectedRspBody {
				t.Fatalf("expected response %s, got: %s", tc.ExpectedRspBody, rec.Body.String())
			}

			todolist, err := s.List()
			if err != nil {
				t.Fatalf("could not list todos: %v", err)
			}

			if todolist[0].Done != tc.ExpectedDone {
				t.Fatalf("expected todo done %t, got: %t", tc.ExpectedDone, todolist[0].Done)
			}
		})
	}
}

func TestListTodos(t *testing.T) {
	tt := []struct {
		Name         string
		TodosInStore []todos.Todo
	}{
		{
			"Returns 200 and empty list if no todos exist",
			[]todos.Todo{},
		},
		{
			"Returns 200 and 3 todos if 3 todos exist",
			[]todos.Todo{
				{1, "Kachra phenk k ao", false},
				{2, "Gaari ki service karalo", false},
				{3, "Roti le ao", false},
			},
		},
		{
			"Returns 200 and 1 todo if 1 todo exists",
			[]todos.Todo{
				{1, "Kachra phenk k ao", false},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			s := todos.NewService(inmem.NewTodoStore())
			h := handlers.Handler{Service: s}

			for _, todo := range tc.TodosInStore {
				if err := s.Save(&todo); err != nil {
					t.Fatalf("could not save todo: %v", err)
				}
			}

			rec := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/todos", nil)
			if err != nil {
				t.Fatalf("could not create http request: %v", err)
			}

			h.ListTodos(rec, req)

			if rec.Result().StatusCode != http.StatusOK {
				t.Fatalf("expected code %d, got: %d", http.StatusOK, rec.Result().StatusCode)
			}

			byt, err := json.Marshal(tc.TodosInStore)
			if err != nil {
				t.Fatalf("invalid test data")
			}

			expectedRspBody := string(byt)
			if rec.Body.String() != expectedRspBody {
				t.Fatalf("expected response %s, got: %s", expectedRspBody, rec.Body.String())
			}
		})
	}
}

func TestReplaceTodo(t *testing.T) {
	tt := []struct {
		Name            string
		ReqBody         string
		TodoID          string
		ExpectedName    string
		ExpectedDone    bool
		ExpectedCode    int
		ExpectedRspBody string
	}{
		{
			"Returns 200 and updates todo for valid request",
			`{"name": "Pawdo ko paani daal do", "done": true}`,
			"1", "Pawdo ko paani daal do", true, http.StatusOK,
			`{"id": 1, "name": "Pawdo ko paani daal do", "done": true}`,
		},
		{
			"Returns 201 and creates todo for valid request if it doesn't exist",
			`{"name": "Pawdo ko paani daal do", "done": true}`,
			"1337", "Pawdo ko paani daal do", false, http.StatusCreated,
			`{"id": 1337, "name": "Pawdo ko paani daal do", "done": true}`,
		},
		{
			"Returns 400 and error msg for non-numeric id",
			`{"name": "Pawdo ko paani daal do", "done": true}`,
			"meow", "Gaari ki service karwalo", false, http.StatusBadRequest,
			`{"error": "Invalid todo ID"}`,
		},
		{
			"Returns 400 and error msg for invalid json",
			`>?!{"name": "ye kya horaha hai}`,
			"1", "Gaari ki service karwalo", false, http.StatusBadRequest,
			`{"error": "Invalid request body"}`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			s := todos.NewService(inmem.NewTodoStore())
			h := handlers.Handler{Service: s}

			todo := todos.Todo{Name: "Gaari ki service karwalo"}
			if err := s.Save(&todo); err != nil {
				t.Fatalf("could not save todo: %v", err)
			}

			rec := httptest.NewRecorder()
			req, err := http.NewRequest("PUT", "/todo/:id", strings.NewReader(tc.ReqBody))
			if err != nil {
				t.Fatalf("could not create http request: %v", err)
			}
			req = mux.SetURLVars(req, map[string]string{"id": tc.TodoID})

			h.ReplaceTodo(rec, req)

			if rec.Result().StatusCode != tc.ExpectedCode {
				t.Fatalf("expected code %d, got: %d", tc.ExpectedCode, rec.Result().StatusCode)
			}

			if rec.Body.String() != tc.ExpectedRspBody {
				t.Fatalf("expected response %s, got: %s", tc.ExpectedRspBody, rec.Body.String())
			}

			todolist, err := s.List()
			if err != nil {
				t.Fatalf("could not list todos: %v", err)
			}

			for _, todo := range todolist {
				if strconv.FormatInt(todo.ID, 10) == tc.TodoID {
					if todo.Name != tc.ExpectedName {
						t.Fatalf("expected todo name %s, got: %s", tc.ExpectedName, todo.Name)
					}
					if todo.Done != tc.ExpectedDone {
						t.Fatalf("expected todo done %t, got: %t", tc.ExpectedDone, todo.Done)
					}
					return
				}
			}
			t.Fatalf("could not find todo after calling handler")
		})
	}
}
