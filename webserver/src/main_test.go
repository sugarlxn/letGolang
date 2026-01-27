package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"webserver/testutil"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB 使用 migrations 文件自动初始化测试数据库
// 优势：
// 1. 与生产环境 schema 完全一致
// 2. 新增字段/表时无需修改测试代码
// 3. 遵循大厂最佳实践
func setupTestDB(t *testing.T) {
	t.Helper()

	// 设置测试环境变量
	jwtSecret = []byte("test-secret")
	infoLog = log.New(io.Discard, "", 0)
	errorLog = log.New(io.Discard, "", 0)

	// 使用 testutil 自动运行所有 migrations
	// 这样当你在 migrations/ 中添加新的 SQL 文件时，测试会自动使用最新的 schema
	db = testutil.SetupTestDB(t)
}

func createTestUser(t *testing.T, username string) User {
	t.Helper()

	hashed, err := hashPassword("password123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	res, err := db.Exec(
		"INSERT INTO users (username, password, phone, email) VALUES (?, ?, ?, ?)",
		username, hashed, "1234567890", fmt.Sprintf("%s@example.com", username),
	)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	id, _ := res.LastInsertId()
	return User{ID: id, Username: username, Password: "", Phone: "1234567890", Email: fmt.Sprintf("%s@example.com", username)}
}

func insertTodo(t *testing.T, userID int64, title string, completed bool) int64 {
	t.Helper()

	res, err := db.Exec("INSERT INTO todos (user_id, title, completed) VALUES (?, ?, ?)", userID, title, completed)
	if err != nil {
		t.Fatalf("failed to insert todo: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func bearerFor(t *testing.T, user User) string {
	t.Helper()
	token, err := generateJWT(user.ID, user.Username)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	return "Bearer " + token
}

func insertPrompt(t *testing.T, userID int64, imageID int64, prompt string, negative_prompt_text string, inferencer_steps int64) int64 {
	t.Helper()

	res, err := db.Exec("INSERT INTO prompts (user_id, image_id, prompt_text, negative_prompt_text, inference_steps) VALUES (?, ?, ?, ?, ?)", userID, imageID, prompt, negative_prompt_text, inferencer_steps)
	if err != nil {
		t.Fatalf("failed to insert prompt: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func TestHandleHealthOK(t *testing.T) {
	setupTestDB(t)

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

func TestHandleCreateUser(t *testing.T) {
	setupTestDB(t)

	body := strings.NewReader(`{"username":"alice","password":"secret123","email":"a@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	rr := httptest.NewRecorder()

	handleCreateUser(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}

	var user User
	if err := json.Unmarshal(rr.Body.Bytes(), &user); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if user.Username != "alice" {
		t.Fatalf("unexpected username %q", user.Username)
	}
	if user.Password != "" {
		t.Fatalf("password should not be returned in response")
	}

	var storedPassword string
	if err := db.QueryRow("SELECT password FROM users WHERE id = ?", user.ID).Scan(&storedPassword); err != nil {
		t.Fatalf("failed to load stored user: %v", err)
	}
	if storedPassword == "" || storedPassword == "secret123" {
		t.Fatalf("password should be stored as hashed value")
	}
}

func TestHandleLogin(t *testing.T) {
	setupTestDB(t)
	user := createTestUser(t, "bob")

	body := strings.NewReader(`{"username":"bob","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	rr := httptest.NewRecorder()

	handleLogin(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp LoginResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Token == "" {
		t.Fatalf("expected a JWT token in response")
	}
	if resp.User.Password != "" {
		t.Fatalf("password should not be returned in login response")
	}

	claims, err := validateJWT(resp.Token)
	if err != nil {
		t.Fatalf("token validation failed: %v", err)
	}
	if claims.UserID != user.ID || claims.Username != user.Username {
		t.Fatalf("claims mismatch: got user_id=%d username=%s", claims.UserID, claims.Username)
	}
}

func TestAuthMiddlewareRejectsMissingHeader(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	rr := httptest.NewRecorder()

	authMiddleware(handleListTodos)(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}

func TestHandleCreateTodoWithAuth(t *testing.T) {
	setupTestDB(t)
	user := createTestUser(t, "carol")

	body := strings.NewReader(fmt.Sprintf(`{"user_id":%d,"title":"Ship project"}`, user.ID))
	req := httptest.NewRequest(http.MethodPost, "/todos", body)
	req.Header.Set("Authorization", bearerFor(t, user))
	rr := httptest.NewRecorder()

	authMiddleware(handleCreateTodo)(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}

	var todo Todo
	if err := json.Unmarshal(rr.Body.Bytes(), &todo); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if todo.Title != "Ship project" || todo.UserID != user.ID || todo.Completed {
		t.Fatalf("unexpected todo response: %+v", todo)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM todos WHERE id = ?", todo.ID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("todo not persisted")
	}
}

func TestHandleListTodosFiltersByUser(t *testing.T) {
	setupTestDB(t)
	userA := createTestUser(t, "dave")
	userB := createTestUser(t, "erin")

	insertTodo(t, userA.ID, "A1", false)
	insertTodo(t, userA.ID, "A2", true)
	insertTodo(t, userB.ID, "B1", false)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/todos?user_id=%d", userA.ID), nil)
	req.Header.Set("Authorization", bearerFor(t, userA))
	rr := httptest.NewRecorder()

	authMiddleware(handleListTodos)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var todos []Todo
	if err := json.Unmarshal(rr.Body.Bytes(), &todos); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(todos) != 2 {
		t.Fatalf("expected 2 todos for user A, got %d", len(todos))
	}
	for _, todo := range todos {
		if todo.UserID != userA.ID {
			t.Fatalf("expected todos only for user A, got user_id=%d", todo.UserID)
		}
	}
}

func TestHandleUpdateTodo(t *testing.T) {
	setupTestDB(t)
	user := createTestUser(t, "frank")
	todoID := insertTodo(t, user.ID, "Initial", false)

	body := strings.NewReader(`{"title":"Updated","completed":true}`)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/todos/%d", todoID), body)
	req.Header.Set("Authorization", bearerFor(t, user))
	rr := httptest.NewRecorder()

	authMiddleware(handleUpdateTodo)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var todo Todo
	if err := json.Unmarshal(rr.Body.Bytes(), &todo); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if todo.ID != todoID || todo.Title != "Updated" || !todo.Completed {
		t.Fatalf("unexpected todo response: %+v", todo)
	}

	var stored Todo
	if err := db.QueryRow("SELECT id, user_id, title, completed, created_at FROM todos WHERE id = ?", todoID).
		Scan(&stored.ID, &stored.UserID, &stored.Title, &stored.Completed, &stored.CreatedAt); err != nil {
		t.Fatalf("failed to load todo from db: %v", err)
	}
	if stored.Title != "Updated" || !stored.Completed {
		t.Fatalf("todo not updated in db: %+v", stored)
	}
}

func TestHandleDeleteTodo(t *testing.T) {
	setupTestDB(t)
	user := createTestUser(t, "gina")
	todoID := insertTodo(t, user.ID, "Disposable", false)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/todos/%d", todoID), nil)
	req.Header.Set("Authorization", bearerFor(t, user))
	rr := httptest.NewRecorder()

	authMiddleware(handleDeleteTodo)(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Fatalf("expected empty response body")
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM todos WHERE id = ?", todoID).Scan(&count); err != nil {
		t.Fatalf("failed to query todo: %v", err)
	}
	if count != 0 {
		t.Fatalf("todo not deleted")
	}
}

func TestHandleListPrompts(t *testing.T) {
	setupTestDB(t)
	user := createTestUser(t, "hannah")

	// Insert test prompts
	insertPrompt(t, user.ID, 1, "Prompt 1", "negative prompt 1", 20)
	insertPrompt(t, user.ID, 2, "Prompt 2", "negative prompt 2", 25)

	req := httptest.NewRequest(http.MethodGet, "/prompts", nil)
	req.Header.Set("Authorization", bearerFor(t, user))
	rr := httptest.NewRecorder()

	authMiddleware(handleListPrompts)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var prompts []Prompt
	if err := json.NewDecoder(rr.Body).Decode(&prompts); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(prompts) != 2 {
		t.Fatalf("expected 2 prompts, got %d", len(prompts))
	}

	// Verify the prompts are in the correct order (newest first)
	if prompts[0].PromptText != "Prompt 1" || prompts[1].PromptText != "Prompt 2" {
		t.Fatalf("expected prompts in order: Prompt 1, Prompt 2; got: %s, %s", prompts[0].PromptText, prompts[1].PromptText)
	}
	// Verify that negative prompt text and inference steps are correctly stored
	if prompts[0].NegativePromptText != "negative prompt 1" || prompts[0].InferenceSteps != 20 {
		t.Fatalf("expected prompt 1 details to match; got negative_prompt_text=%s, inference_steps=%d", prompts[0].NegativePromptText, prompts[0].InferenceSteps)
	}

}
