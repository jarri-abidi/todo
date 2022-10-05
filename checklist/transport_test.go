package checklist_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jarri-abidi/todo"
	"github.com/jarri-abidi/todo/checklist"
	"github.com/jarri-abidi/todo/inmem"
)

func TestToggleTask(t *testing.T) {
	tt := []struct {
		Name            string
		TaskID          string
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
			"Returns 404 and error msg for id of task that doesn't exist",
			"1337", http.StatusNotFound, `{"error":"task not found"}`, false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			var (
				require = require.New(t)
				assert  = assert.New(t)
				svc     = checklist.NewService(inmem.NewTaskRepository())
				handler = checklist.MakeHandler(svc, log.NewNopLogger())
			)

			task := todo.Task{Name: "Gaari ki service karwalo"}
			require.NoError(svc.Save(context.TODO(), &task), "could not save task")

			rec := httptest.NewRecorder()
			url := fmt.Sprintf("/checklist/v1/task/%s", tc.TaskID)
			req, err := http.NewRequest("PATCH", url, nil)
			require.NoError(err, "could not create http request")

			handler.ServeHTTP(rec, req)

			assert.Equal(tc.ExpectedCode, rec.Result().StatusCode, "unexpected http status code")
			if tc.ExpectedCode != http.StatusNoContent {
				assert.JSONEq(tc.ExpectedRspBody, rec.Body.String(), " unexpected http response body")
			}

			list, err := svc.List(context.TODO())
			require.NoError(err) // could not list tasks
			assert.Equal(tc.ExpectedDone, list[0].Done, "task should be toggled")
		})
	}
}

func TestListTasks(t *testing.T) {
	tt := []struct {
		Name        string
		TasksInRepo []todo.Task
		Expected    string
	}{
		{
			"Returns 200 and empty list if no tasks exist",
			[]todo.Task{},
			`[]`,
		},
		{
			"Returns 200 and 3 tasks if 3 tasks exist",
			[]todo.Task{
				{ID: 1, Name: "Kachra phenk k ao", Done: false},
				{ID: 2, Name: "Gaari ki service karalo", Done: false},
				{ID: 3, Name: "Roti le ao", Done: false},
			},
			`[
				{"id": 1, "name": "Kachra phenk k ao", "done": false},
				{"id": 2, "name": "Gaari ki service karalo", "done": false},
				{"id": 3, "name": "Roti le ao", "done": false}
			]`,
		},
		{
			"Returns 200 and 1 task if 1 task exists",
			[]todo.Task{
				{ID: 1, Name: "Kachra phenk k ao", Done: false},
			},
			`[
				{"id": 1, "name": "Kachra phenk k ao", "done": false}
			]`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			var (
				require = require.New(t)
				assert  = assert.New(t)
				svc     = checklist.NewService(inmem.NewTaskRepository())
				handler = checklist.MakeHandler(svc, log.NewNopLogger())
			)

			for i := range tc.TasksInRepo {
				require.NoError(svc.Save(context.TODO(), &tc.TasksInRepo[i]), "could not save task")
			}

			rec := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/checklist/v1/tasks", nil)
			require.NoError(err, "could not create http request")

			handler.ServeHTTP(rec, req)

			assert.Equal(http.StatusOK, rec.Result().StatusCode, "unexpected http status code")
			assert.JSONEq(tc.Expected, rec.Body.String(), "unexpected http response body")
		})
	}
}

func NoTestReplaceTask(t *testing.T) {
	tt := []struct {
		Name            string
		ReqBody         string
		TaskID          string
		ExpectedName    string
		ExpectedDone    bool
		ExpectedCode    int
		ExpectedRspBody string
	}{
		{
			"Returns 200 and updates task for valid request",
			`{"name":"Pawdo ko paani daal do","done":true}`,
			"1", "Pawdo ko paani daal do", true, http.StatusOK,
			`{"id": 1,"name":"Pawdo ko paani daal do","done":true}`,
		},
		{
			"Returns 201 and creates task for valid request if it doesn't exist",
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
				svc     = checklist.NewService(inmem.NewTaskRepository())
				handler = checklist.MakeHandler(svc, log.NewNopLogger())
			)

			savedTask := todo.Task{Name: "Gaari ki service karwalo"}
			require.NoError(svc.Save(context.TODO(), &savedTask), "could not save task")

			rec := httptest.NewRecorder()
			url := fmt.Sprintf("/checklist/v1/task/%s", tc.TaskID)
			req, err := http.NewRequest("PUT", url, strings.NewReader(tc.ReqBody))
			require.NoError(err, "could not create http request")

			handler.ServeHTTP(rec, req)

			assert.Equal(tc.ExpectedCode, rec.Result().StatusCode, "unexpected http status code")
			assert.JSONEq(tc.ExpectedRspBody, rec.Body.String(), "unexpected http response body")

			list, err := svc.List(context.TODO())
			require.NoError(err, "could not list tasks")

			for _, task := range list {
				if task.ID == savedTask.ID {
					assert.Equal(tc.ExpectedName, task.Name, "expected Name to be updated")
					assert.Equal(tc.ExpectedDone, task.Done, "expected Done to be updated")
					return
				}
			}
			t.Error("could not find task after calling handler")
		})
	}
}
