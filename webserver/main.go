package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Todo 代表一个简单的待办事项
// 在真实项目中你可以替换为自己业务的实体，例如用户、文章等。
type Todo struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// 使用内存存储 Todo，演示 REST API；生产环境请使用数据库。
var (
	todos  = make([]*Todo, 0)
	mu     sync.RWMutex
	idSeed int64
)

func init() {
	rand.Seed(time.Now().UnixNano())
	// 初始化一些示例数据
	idSeed = 3
	todos = []*Todo{
		{ID: 1, Title: "Learn Go", Completed: false},
		{ID: 2, Title: "Build a web server", Completed: false},
		{ID: 3, Title: "Write REST API", Completed: true},
	}
}

// writeJSON 是一个小工具函数，用于统一 JSON 返回
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// errorResponse 用于返回错误信息
func errorResponse(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// handleHealth 简单健康检查
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleListTodos 处理 GET /todos
func handleListTodos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	mu.RLock()
	defer mu.RUnlock()

	writeJSON(w, http.StatusOK, todos)
}

// handleGetTodo 处理 GET /todos/{id}
func handleGetTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/todos/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid id")
		return
	}

	mu.RLock()
	defer mu.RUnlock()

	for _, t := range todos {
		if t.ID == id {
			writeJSON(w, http.StatusOK, t)
			return
		}
	}

	errorResponse(w, http.StatusNotFound, "todo not found")
}

// handleCreateTodo 处理 POST /todos
func handleCreateTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var input struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if input.Title == "" {
		errorResponse(w, http.StatusBadRequest, "title is required")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	idSeed++
	newTodo := &Todo{
		ID:        idSeed,
		Title:     input.Title,
		Completed: false,
	}
	todos = append(todos, newTodo)

	writeJSON(w, http.StatusCreated, newTodo)
}

// handleUpdateTodo 处理 PUT /todos/{id}
func handleUpdateTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.Header().Set("Allow", http.MethodPut)
		errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/todos/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid id")
		return
	}

	var input struct {
		Title     *string `json:"title"`
		Completed *bool   `json:"completed"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	for _, t := range todos {
		if t.ID == id {
			if input.Title != nil {
				t.Title = *input.Title
			}
			if input.Completed != nil {
				t.Completed = *input.Completed
			}
			writeJSON(w, http.StatusOK, t)
			return
		}
	}

	errorResponse(w, http.StatusNotFound, "todo not found")
}

// handleDeleteTodo 处理 DELETE /todos/{id}
func handleDeleteTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", http.MethodDelete)
		errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/todos/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid id")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	for i, t := range todos {
		if t.ID == id {
			// 删除该元素
			todos = append(todos[:i], todos[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	errorResponse(w, http.StatusNotFound, "todo not found")
}

func main() {
	mux := http.NewServeMux()

	// 健康检查
	mux.HandleFunc("/health", handleHealth)

	// RESTful Todo 路由
	mux.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
		// 根据方法分发
		switch r.Method {
		case http.MethodGet:
			handleListTodos(w, r)
		case http.MethodPost:
			handleCreateTodo(w, r)
		default:
			w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
			errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	// 通过前缀匹配 /todos/ + 手动解析 ID，兼容当前 Go 版本
	mux.HandleFunc("/todos/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetTodo(w, r)
		case http.MethodPut:
			handleUpdateTodo(w, r)
		case http.MethodDelete:
			handleDeleteTodo(w, r)
		default:
			w.Header().Set("Allow", http.MethodGet+", "+http.MethodPut+", "+http.MethodDelete)
			errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	addr := ":8080"
	log.Printf("Starting webserver on %s...", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
