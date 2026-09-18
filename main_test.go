package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateTaskAndConsume(t *testing.T) {
	queue := make(chan Task, 1)
	handler := newHandler(Producer{queue: queue})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"name":" Send welcome email "}`)))
	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var task Task
	if err := json.NewDecoder(response.Body).Decode(&task); err != nil {
		t.Fatal(err)
	}
	if task.ID == "" || task.Name != "Send welcome email" || task.Topic != taskTopic || task.CreatedAt.IsZero() {
		t.Fatalf("unexpected task: %+v", task)
	}
	close(queue)
	var output bytes.Buffer
	consume(queue, log.New(&output, "", 0))
	if !strings.Contains(output.String(), "id="+task.ID) || !strings.Contains(output.String(), `name="Send welcome email"`) {
		t.Fatalf("consumer did not process the accepted task: %s", output.String())
	}
}

func TestInvalidTasksAreNotPublished(t *testing.T) {
	for _, body := range []string{
		``, `{`, `null`, `{}`, `{"name":" "}`, `{"name":123}`,
		`{"name":"test","extra":true}`, `{"name":"test"} {}`,
		`{"name":"` + strings.Repeat("a", 201) + `"}`,
		`{"name":"test"}` + strings.Repeat(" ", 4096),
	} {
		t.Run(body[:min(len(body), 40)], func(t *testing.T) {
			queue := make(chan Task, 1)
			response := httptest.NewRecorder()
			newHandler(Producer{queue: queue}).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(body)))
			if response.Code != http.StatusBadRequest || len(queue) != 0 {
				t.Fatalf("status = %d, queued = %d", response.Code, len(queue))
			}
		})
	}
}

func TestFullQueue(t *testing.T) {
	queue := make(chan Task, 1)
	queue <- Task{ID: "existing"}
	response := httptest.NewRecorder()
	newHandler(Producer{queue: queue}).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"name":"test"}`)))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", response.Code)
	}
	if task := <-queue; task.ID != "existing" {
		t.Fatalf("existing task was replaced: %+v", task)
	}
}
