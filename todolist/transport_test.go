package todolist_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jarri-abidi/todolist/inmem"
	"github.com/jarri-abidi/todolist/todolist"
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
			"1", http.StatusNoContent, ``, true,
		},
		{
			"Returns 404 and error msg for non-numeric id",
			"meow", http.StatusNotFound, `{"error":"resource not found"}`, false,
		},
		{
			"Returns 404 and error msg for id of todo that doesn't exist",
			"1337", http.StatusNotFound, `{"error":"todo not found"}`, false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			var (
				require = require.New(t)
				assert  = assert.New(t)
				svc     = todolist.NewService(inmem.NewTodoStore())
				handler = todolist.MakeHandler(svc, log.NewNopLogger())
			)

			todo := todos.Todo{Name: "Gaari ki service karwalo"}
			require.NoError(svc.Save(context.TODO(), &todo), "could not save todo")

			rec := httptest.NewRecorder()
			url := fmt.Sprintf("/todolist/v1/todo/%s", tc.TodoID)
			req, err := http.NewRequest("PATCH", url, nil)
			require.NoError(err, "could not create http request")

			handler.ServeHTTP(rec, req)

			if tc.ExpectedCode != http.StatusNoContent {
				assert.Equal(tc.ExpectedCode, rec.Result().StatusCode, "unexpected http status code")
				assert.JSONEq(tc.ExpectedRspBody, rec.Body.String(), " unexpected http response body")
			}

			todolist, err := svc.List(context.TODO())
			require.NoError(err) // could not list todos
			assert.Equal(tc.ExpectedDone, todolist[0].Done, "todo should be toggled")
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
				{ID: 1, Name: "Kachra phenk k ao", Done: false},
				{ID: 2, Name: "Gaari ki service karalo", Done: false},
				{ID: 3, Name: "Roti le ao", Done: false},
			},
		},
		{
			"Returns 200 and 1 todo if 1 todo exists",
			[]todos.Todo{
				{ID: 1, Name: "Kachra phenk k ao", Done: false},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			var (
				require = require.New(t)
				assert  = assert.New(t)
				svc     = todolist.NewService(inmem.NewTodoStore())
				handler = todolist.MakeHandler(svc, log.NewNopLogger())
			)

			for i := range tc.TodosInStore {
				require.NoError(svc.Save(context.TODO(), &tc.TodosInStore[i]), "could not save todo")
			}

			rec := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/todolist/v1/todos", nil)
			require.NoError(err, "could not create http request")

			handler.ServeHTTP(rec, req)

			assert.Equal(http.StatusOK, rec.Result().StatusCode, "unexpected http status code")
			expected, err := json.Marshal(map[string]interface{}{"todos": tc.TodosInStore})
			require.NoError(err)
			assert.JSONEq(string(expected), rec.Body.String(), "unexpected http response body")
		})
	}
}

func NoTestReplaceTodo(t *testing.T) {
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
			`{"error":"resource not found"}`,
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

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			var (
				require = require.New(t)
				assert  = assert.New(t)
				svc     = todolist.NewService(inmem.NewTodoStore())
				handler = todolist.MakeHandler(svc, log.NewNopLogger())
			)

			savedTodo := todos.Todo{Name: "Gaari ki service karwalo"}
			require.NoError(svc.Save(context.TODO(), &savedTodo), "could not save todo")

			rec := httptest.NewRecorder()
			url := fmt.Sprintf("/todolist/v1/todo/%s", tc.TodoID)
			req, err := http.NewRequest("PUT", url, strings.NewReader(tc.ReqBody))
			require.NoError(err, "could not create http request")

			handler.ServeHTTP(rec, req)

			assert.Equal(tc.ExpectedCode, rec.Result().StatusCode, "unexpected http status code")
			assert.JSONEq(tc.ExpectedRspBody, rec.Body.String(), "unexpected http response body")

			todolist, err := svc.List(context.TODO())
			require.NoError(err, "could not list todos")

			for _, todo := range todolist {
				if todo.ID == savedTodo.ID {
					assert.Equal(tc.ExpectedName, todo.Name, "expected Name to be updated")
					assert.Equal(tc.ExpectedDone, todo.Done, "expected Done to be updated")
					return
				}
			}
			t.Error("could not find todo after calling handler")
		})
	}
}
