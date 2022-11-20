package checklist_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jarri-abidi/todo/pkg/checklist"
	"github.com/jarri-abidi/todo/pkg/inmem"
	"github.com/jarri-abidi/todo/pkg/todo"
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
			Name:            "Returns 204 and toggles for valid request",
			TaskID:          "1",
			ExpectedCode:    http.StatusNoContent,
			ExpectedRspBody: ``,
			ExpectedDone:    true,
		},
		{
			Name:            "Returns 400 and error msg for non-numeric id",
			TaskID:          "meow",
			ExpectedCode:    http.StatusBadRequest,
			ExpectedRspBody: `{"error":"task id in path must be numeric"}`,
			ExpectedDone:    false,
		},
		{
			Name:            "Returns 404 and error msg for id of task that doesn't exist",
			TaskID:          "1337",
			ExpectedCode:    http.StatusNotFound,
			ExpectedRspBody: `{"error":"task not found"}`,
			ExpectedDone:    false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			var (
				require = require.New(t)
				assert  = assert.New(t)
				svc     = checklist.NewService(inmem.NewTaskRepository())
				handler = checklist.NewServer(svc, log.NewNopLogger())
			)

			task := todo.Task{Name: "Gaari ki service karwalo"}
			_, err := svc.Save(context.TODO(), task)
			require.NoError(err, "could not save task")

			rec := httptest.NewRecorder()
			url := fmt.Sprintf("/checklist/v1/task/%s", tc.TaskID)
			req, err := http.NewRequest("PATCH", url, nil)
			require.NoError(err, "could not create http request")

			handler.ServeHTTP(rec, req)

			assert.Equal(tc.ExpectedCode, rec.Result().StatusCode, "unexpected http status code")
			if tc.ExpectedCode != http.StatusNoContent {
				assert.JSONEq(tc.ExpectedRspBody, rec.Body.String(), "unexpected http response body")
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
			Name:        "Returns 200 and empty list if no tasks exist",
			TasksInRepo: []todo.Task{},
			Expected:    `[]`,
		},
		{
			Name: "Returns 200 and 3 tasks if 3 tasks exist",
			TasksInRepo: []todo.Task{
				{ID: 1, Name: "Kachra phenk k ao", Done: false},
				{ID: 2, Name: "Gaari ki service karalo", Done: false},
				{ID: 3, Name: "Roti le ao", Done: false},
			},
			Expected: `[
				{"id": 1, "name": "Kachra phenk k ao", "done": false},
				{"id": 2, "name": "Gaari ki service karalo", "done": false},
				{"id": 3, "name": "Roti le ao", "done": false}
			]`,
		},
		{
			Name: "Returns 200 and 1 task if 1 task exists",
			TasksInRepo: []todo.Task{
				{ID: 1, Name: "Kachra phenk k ao", Done: false},
			},
			Expected: `[
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
				handler = checklist.NewServer(svc, log.NewNopLogger())
			)

			for i := range tc.TasksInRepo {
				_, err := svc.Save(context.TODO(), tc.TasksInRepo[i])
				require.NoError(err, "could not save task")
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

func TestReplaceTask(t *testing.T) {
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
			Name:            "Returns 200 and updates task for valid request",
			ReqBody:         `{"name":"Pawdo ko paani daal do","done":true}`,
			TaskID:          "1",
			ExpectedName:    "Pawdo ko paani daal do",
			ExpectedDone:    true,
			ExpectedCode:    http.StatusOK,
			ExpectedRspBody: `{"id": 1,"name":"Pawdo ko paani daal do","done":true}`,
		},
		{
			Name:            "Returns 201 and creates task for valid request if it doesn't exist",
			ReqBody:         `{"name":"Pawdo ko paani daal do","done":true}`,
			TaskID:          "1337",
			ExpectedName:    "Pawdo ko paani daal do",
			ExpectedDone:    true,
			ExpectedCode:    http.StatusCreated,
			ExpectedRspBody: `{"id":1337,"name":"Pawdo ko paani daal do","done":true}`,
		},
		{
			Name:            "Returns 400 and error msg for non-numeric id",
			ReqBody:         `{"name": "Pawdo ko paani daal do", "done": true}`,
			TaskID:          "meow",
			ExpectedName:    "Gaari ki service karwalo",
			ExpectedDone:    false,
			ExpectedCode:    http.StatusBadRequest,
			ExpectedRspBody: `{"error":"task id in path must be numeric"}`,
		},
		{
			Name:            "Returns 400 and error msg for invalid json",
			ReqBody:         `>?!{"name": "ye kya horaha hai}`,
			TaskID:          "1",
			ExpectedName:    "Gaari ki service karwalo",
			ExpectedDone:    false,
			ExpectedCode:    http.StatusBadRequest,
			ExpectedRspBody: `{"error":"invalid request body: invalid character '>' looking for beginning of value"}`,
		},
		{
			Name:            "Returns 400 and error msg for blank name",
			ReqBody:         `{"name": "	", "done": true}`,
			TaskID:          "1",
			ExpectedName:    "Gaari ki service karwalo",
			ExpectedDone:    false,
			ExpectedCode:    http.StatusBadRequest,
			ExpectedRspBody: `{"error":"invalid request body: invalid character '\\t' in string literal"}`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			var (
				require = require.New(t)
				assert  = assert.New(t)
				svc     = checklist.NewService(inmem.NewTaskRepository())
				handler = checklist.NewServer(svc, log.NewNopLogger())
			)

			newTask := todo.Task{Name: "Gaari ki service karwalo"}
			_, err := svc.Save(context.TODO(), newTask)
			require.NoError(err, "could not save task")

			rec := httptest.NewRecorder()
			url := fmt.Sprintf("/checklist/v1/task/%s", tc.TaskID)
			req, err := http.NewRequest("PUT", url, strings.NewReader(tc.ReqBody))
			require.NoError(err, "could not create http request")

			handler.ServeHTTP(rec, req)

			assert.Equal(tc.ExpectedCode, rec.Result().StatusCode, "unexpected http status code")
			assert.JSONEq(tc.ExpectedRspBody, rec.Body.String(), "unexpected http response body")

			list, err := svc.List(context.TODO())
			require.NoError(err, "could not list tasks")

			if tc.ExpectedCode != 200 && tc.ExpectedCode != 201 {
				return
			}

			for _, task := range list {
				if strconv.FormatInt(task.ID, 10) == tc.TaskID {
					assert.Equal(tc.ExpectedName, task.Name, "expected Name to be updated")
					assert.Equal(tc.ExpectedDone, task.Done, "expected Done to be updated")
					return
				}
			}
			t.Error("could not find task after calling handler")
		})
	}
}
