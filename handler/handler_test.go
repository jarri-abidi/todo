package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matryer/is"

	"github.com/jarri-abidi/todolist/handler"
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
			"Returns 404 and error msg for non-numeric id",
			"meow", http.StatusNotFound, `{"error":"page not found"}`, false,
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
				h  = handler.New(s)
			)

			todo := todos.Todo{Name: "Gaari ki service karwalo"}
			is.NoErr(s.Save(&todo)) // could not save todo

			rec := httptest.NewRecorder()
			url := fmt.Sprintf("/todo/%s", tc.TodoID)
			req, err := http.NewRequest("PATCH", url, nil)
			is.NoErr(err) // could not create http request

			h.ServeHTTP(rec, req)

			is.Equal(rec.Result().StatusCode, tc.ExpectedCode) // unexpected http status code
			is.Equal(rec.Body.String(), tc.ExpectedRspBody)    // unexpected http response body

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
				h  = handler.New(s)
			)

			for i := range tc.TodosInStore {
				is.NoErr(s.Save(&tc.TodosInStore[i])) // could not save todo
			}

			rec := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/todos", nil)
			is.NoErr(err) // could not create http request

			h.ServeHTTP(rec, req)

			is.Equal(rec.Result().StatusCode, http.StatusOK) // unexpected http status code

			byt, err := json.Marshal(tc.TodosInStore)
			is.NoErr(err) // invalid test data
			expectedRspBody := string(byt)
			is.Equal(rec.Body.String(), expectedRspBody) // unexpected http response body
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
			`{"name":"Pawdo ko paani daal do","done":true}`,
			"1", "Pawdo ko paani daal do", true, http.StatusOK,
			`{"id": 1,"name":"Pawdo ko paani daal do","done":true}`,
		},
		{
			"Returns 201 and creates todo for valid request if it doesn't exist",
			`{"name":"Pawdo ko paani daal do","done":true}`,
			"1337", "Pawdo ko paani daal do", false, http.StatusCreated,
			`{"id":1337,"name":"Pawdo ko paani daal do","done":true}`,
		},
		{
			"Returns 404 and error msg for non-numeric id",
			`{"name": "Pawdo ko paani daal do", "done": true}`,
			"meow", "Gaari ki service karwalo", false, http.StatusNotFound,
			`{"error":"page not found"}`,
		},
		{
			"Returns 400 and error msg for invalid json",
			`>?!{"name": "ye kya horaha hai}`,
			"1", "Gaari ki service karwalo", false, http.StatusBadRequest,
			`{"error":"invalid request body"}`,
		},
		{
			"Returns 400 and error msg for blank name",
			`{"name": "	", "done": true}`,
			"1", "Gaari ki service karwalo", false, http.StatusBadRequest,
			`{"error":"name cannot be blank"}`,
		},
	}

	is := is.New(t)
	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			var (
				is = is.New(t)
				s  = todos.NewService(inmem.NewTodoStore())
				h  = handler.New(s)
			)

			savedTodo := todos.Todo{Name: "Gaari ki service karwalo"}
			is.NoErr(s.Save(&savedTodo)) // could not save todo

			rec := httptest.NewRecorder()
			url := fmt.Sprintf("/todo/%s", tc.TodoID)
			req, err := http.NewRequest("PUT", url, strings.NewReader(tc.ReqBody))
			is.NoErr(err) // could not create http request

			h.ServeHTTP(rec, req)

			is.Equal(rec.Result().StatusCode, tc.ExpectedCode) // unexpected http status code
			is.Equal(rec.Body.String(), tc.ExpectedRspBody)    // unexpected http response body

			todolist, err := s.List()
			is.NoErr(err) // could not list todos

			for _, todo := range todolist {
				if todo.ID == savedTodo.ID {
					is.Equal(todo.Name, tc.ExpectedName) // expected Name to be updated
					is.Equal(todo.Done, tc.ExpectedDone) // expected Done to be updated
					return
				}
			}
			is.Fail() // could not find todo after calling handler
		})
	}
}
