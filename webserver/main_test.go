package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func resetState() {
	mu.Lock()
	defer mu.Unlock()
	idSeed = 3
	todos = []*Todo{
		{ID: 1, Title: "Learn Go", Completed: false},
		{ID: 2, Title: "Build a web server", Completed: false},
		{ID: 3, Title: "Write REST API", Completed: true},
	}
}

func TestHandleHealthOK(t *testing.T) {
	resetState()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handleHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var payload map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", payload["status"])
	}
}

func TestHandleListTodosOK(t *testing.T) {
	resetState()

	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	rr := httptest.NewRecorder()

	handleListTodos(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var list []*Todo
	if err := json.Unmarshal(rr.Body.Bytes(), &list); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 todos, got %d", len(list))
	}
}

func TestHandleListTodosMethodNotAllowed(t *testing.T) {
	resetState()

	req := httptest.NewRequest(http.MethodPost, "/todos", nil)
	rr := httptest.NewRecorder()

	handleListTodos(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rr.Code)
	}
	if allow := rr.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("expected Allow header %q, got %q", http.MethodGet, allow)
	}
}

func TestHandleGetTodoFound(t *testing.T) {
	resetState()

	req := httptest.NewRequest(http.MethodGet, "/todos/2", nil)
	rr := httptest.NewRecorder()

	handleGetTodo(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var todo Todo
	if err := json.Unmarshal(rr.Body.Bytes(), &todo); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if todo.ID != 2 {
		t.Fatalf("expected todo ID 2, got %d", todo.ID)
	}
}

func TestHandleGetTodoNotFound(t *testing.T) {
	resetState()

	req := httptest.NewRequest(http.MethodGet, "/todos/99", nil)
	rr := httptest.NewRecorder()

	handleGetTodo(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}

	var payload map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload["error"] != "todo not found" {
		t.Fatalf("expected error 'todo not found', got %q", payload["error"])
	}
}

func TestHandleCreateTodoCreatesNewItem(t *testing.T) {
	resetState()

	body := strings.NewReader(`{"title":"Ship project"}`)
	req := httptest.NewRequest(http.MethodPost, "/todos", body)
	rr := httptest.NewRecorder()

	handleCreateTodo(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}

	var todo Todo
	if err := json.Unmarshal(rr.Body.Bytes(), &todo); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if todo.ID != 4 {
		t.Fatalf("expected new todo ID 4, got %d", todo.ID)
	}
	if todo.Title != "Ship project" {
		t.Fatalf("unexpected todo title %q", todo.Title)
	}
	if todo.Completed {
		t.Fatal("expected new todo to be incomplete")
	}

	mu.RLock()
	if len(todos) != 4 {
		mu.RUnlock()
		t.Fatalf("expected 4 todos in memory, got %d", len(todos))
	}
	mu.RUnlock()
}

func TestHandleUpdateTodoUpdatesFields(t *testing.T) {
	resetState()

	body := strings.NewReader(`{"title":"Updated", "completed":true}`)
	req := httptest.NewRequest(http.MethodPut, "/todos/2", body)
	rr := httptest.NewRecorder()

	handleUpdateTodo(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var updatedResp Todo
	if err := json.Unmarshal(rr.Body.Bytes(), &updatedResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if updatedResp.ID != 2 {
		t.Fatalf("expected todo ID 2, got %d", updatedResp.ID)
	}

	mu.RLock()
	var stored Todo
	for _, t := range todos {
		if t.ID == 2 {
			stored = *t
			break
		}
	}
	mu.RUnlock()

	if stored.ID == 0 {
		t.Fatal("todo with ID 2 not found after update")
	}
	if stored.Title != "Updated" {
		t.Fatalf("expected title 'Updated', got %q", stored.Title)
	}
	if !stored.Completed {
		t.Fatal("expected todo to be marked completed")
	}
}

func TestHandleDeleteTodoRemovesItem(t *testing.T) {
	resetState()

	req := httptest.NewRequest(http.MethodDelete, "/todos/2", nil)
	rr := httptest.NewRecorder()

	handleDeleteTodo(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Fatal("expected empty response body")
	}

	mu.RLock()
	defer mu.RUnlock()
	if len(todos) != 2 {
		t.Fatalf("expected 2 todos remaining, got %d", len(todos))
	}
	for _, todo := range todos {
		if todo.ID == 2 {
			t.Fatal("todo ID 2 still present after deletion")
		}
	}
}
