package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDeleteLogs_Run(t *testing.T) {
	var gotPath, gotMethod string
	var gotFilter string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		r.ParseForm()
		gotFilter = r.FormValue("filter")
		json.NewEncoder(w).Encode(map[string]string{
			"task_id": "task-123",
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	taskID, err := c.DeleteLogsRun("_time:<24h AND service:test")
	if err != nil {
		t.Fatalf("DeleteLogsRun() error = %v", err)
	}

	if gotMethod != "POST" {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/api/logs/delete/run_task" {
		t.Errorf("path = %q, want /api/logs/delete/run_task", gotPath)
	}
	if gotFilter != "_time:<24h AND service:test" {
		t.Errorf("filter = %q, want %q", gotFilter, "_time:<24h AND service:test")
	}
	if taskID != "task-123" {
		t.Errorf("taskID = %q, want %q", taskID, "task-123")
	}
}

func TestDeleteLogs_Stop(t *testing.T) {
	var gotPath, gotMethod string
	var gotTaskID string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		r.ParseForm()
		gotTaskID = r.FormValue("task_id")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	err := c.DeleteLogsStop("task-123")
	if err != nil {
		t.Fatalf("DeleteLogsStop() error = %v", err)
	}

	if gotMethod != "POST" {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/api/logs/delete/stop_task" {
		t.Errorf("path = %q, want /api/logs/delete/stop_task", gotPath)
	}
	if gotTaskID != "task-123" {
		t.Errorf("task_id = %q, want %q", gotTaskID, "task-123")
	}
}

func TestDeleteLogs_List(t *testing.T) {
	var gotPath, gotMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode([]map[string]string{
			{"task_id": "task-1", "filter": "_time:<24h", "status": "running", "progress": "50%"},
			{"task_id": "task-2", "filter": "service:test", "status": "running", "progress": "10%"},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	tasks, err := c.DeleteLogsList()
	if err != nil {
		t.Fatalf("DeleteLogsList() error = %v", err)
	}

	if gotMethod != "GET" {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/api/logs/delete/active_tasks" {
		t.Errorf("path = %q, want /api/logs/delete/active_tasks", gotPath)
	}
	if len(tasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(tasks))
	}
	if tasks[0].TaskID != "task-1" {
		t.Errorf("tasks[0].TaskID = %q, want %q", tasks[0].TaskID, "task-1")
	}
	if tasks[0].Status != "running" {
		t.Errorf("tasks[0].Status = %q, want %q", tasks[0].Status, "running")
	}
}

func TestDeleteLogs_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`deletion service unavailable`))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.DeleteLogsRun("invalid filter")
	if err == nil {
		t.Fatal("DeleteLogsRun() expected error for 500, got nil")
	}
	if !strings.Contains(err.Error(), "500") || !strings.Contains(err.Error(), "deletion service unavailable") {
		t.Errorf("unexpected error: %v", err)
	}
}
