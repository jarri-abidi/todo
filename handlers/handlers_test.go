package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/matryer/is"

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

	is := is.New(t)
	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			var (
				is = is.New(t)
				s  = todos.NewService(inmem.NewTodoStore())
				h  = handlers.Handler{Service: s}
			)

			todo := todos.Todo{Name: "Gaari ki service karwalo"}
			is.NoErr(s.Save(&todo)) // could not save todo

			rec := httptest.NewRecorder()
			req, err := http.NewRequest("PATCH", "/todo/:id", nil)
			is.NoErr(err) // could not create http request
			req = mux.SetURLVars(req, map[string]string{"id": tc.TodoID})

			h.ToggleTodo(rec, req)

			is.Equal(rec.Result().StatusCode, tc.ExpectedCode) // unexpected HTTP status code
			is.Equal(rec.Body.String(), tc.ExpectedRspBody)    // unexpected HTTP response body

			todolist, err := s.List()
			is.NoErr(err)                               // could not list todos
			is.Equal(todolist[0].Done, tc.ExpectedDone) // todo should be toggled
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

	is := is.New(t)
	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			var (
				is = is.New(t)
				s  = todos.NewService(inmem.NewTodoStore())
				h  = handlers.Handler{Service: s}
			)

			for i := range tc.TodosInStore {
				is.NoErr(s.Save(&tc.TodosInStore[i])) // could not save todo
			}

			rec := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/todos", nil)
			is.NoErr(err) // could not create http request

			h.ListTodos(rec, req)

			is.Equal(rec.Result().StatusCode, http.StatusOK) // unexpected HTTP status code

			byt, err := json.Marshal(tc.TodosInStore)
			is.NoErr(err) // invalid test data
			expectedRspBody := string(byt)
			is.Equal(rec.Body.String(), expectedRspBody) // unexpected HTTP response body
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

	is := is.New(t)
	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			var (
				is = is.New(t)
				s  = todos.NewService(inmem.NewTodoStore())
				h  = handlers.Handler{Service: s}
			)

			todo := todos.Todo{Name: "Gaari ki service karwalo"}
			is.NoErr(s.Save(&todo)) // could not save todo

			rec := httptest.NewRecorder()
			req, err := http.NewRequest("PUT", "/todo/:id", strings.NewReader(tc.ReqBody))
			is.NoErr(err) // could not create http request
			req = mux.SetURLVars(req, map[string]string{"id": tc.TodoID})

			h.ReplaceTodo(rec, req)

			is.Equal(rec.Result().StatusCode, tc.ExpectedCode) // unexpected HTTP status code
			is.Equal(rec.Body.String(), tc.ExpectedRspBody)    // unexpected HTTP response body

			todolist, err := s.List()
			is.NoErr(err) // could not list todos

			for _, todo := range todolist {
				if strconv.FormatInt(todo.ID, 10) == tc.TodoID {
					is.Equal(todo.Name, tc.ExpectedName) // expected Name to be updated
					is.Equal(todo.Done, tc.ExpectedDone) // expected Done to be updated
					return
				}
			}
			is.Fail() // could not find todo after calling handler
		})
	}
}
