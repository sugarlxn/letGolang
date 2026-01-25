package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	httpSwagger "github.com/swaggo/http-swagger"

	_ "webserver/docs"

	_ "github.com/mattn/go-sqlite3"
)

//	@title			User Management API
//	@version		2.0
//	@description	This is a user management and todo list server with SQLite database, JWT authentication, and bcrypt password hashing.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8080
//	@BasePath	/

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

// Todo 代表一个简单的待办事项
//
//	@Description	Todo 代表一个简单的待办事项
type Todo struct {
	ID        int64     `json:"id"`         //	@Description	Todo ID
	UserID    int64     `json:"user_id"`    //	@Description	User ID who owns this todo
	Title     string    `json:"title"`      //	@Description	Todo title
	Completed bool      `json:"completed"`  //	@Description	Todo completion status
	CreatedAt time.Time `json:"created_at"` //	@Description	Todo creation time
}

// User 用户结构体
//
//	@Description	User 用户结构体
//	@ID				User
//	@Accept			json
//	@Produce		json
type User struct {
	ID        int64     `json:"id"`         //	@Description	User ID
	Username  string    `json:"username"`   //	@Description	User username
	Password  string    `json:"password"`   //	@Description	User password
	Phone     string    `json:"phone"`      //	@Description	User phone
	Email     string    `json:"email"`      //	@Description	User email
	CreatedAt time.Time `json:"created_at"` //	@Description	User creation time
}

// Image 图片结构体
//
//	@Description	Image 图片结构体
//	@ID				Image
//	@Accept			json
//	@Produce		json
type Image struct {
	ID        int64     `json:"id"`         //	@Description	Image ID
	UserID    int64     `json:"user_id"`    //	@Description	Image user ID
	ImageData []byte    `json:"image_data"` //	@Description	Image data
	CreatedAt time.Time `json:"created_at"` //	@Description	Image creation time
}

// Database 连接
var db *sql.DB

// JWT 密钥（生产环境应从环境变量读取）
var jwtSecret = []byte(getEnv("JWT_SECRET", "your-secret-key-change-in-production"))

// 日志记录器
var (
	infoLog  *log.Logger
	errorLog *log.Logger
)

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// JWT Claims
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
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
	errorLog.Printf("Error: %s (status: %d)", msg, status)
	writeJSON(w, status, map[string]string{"error": msg})
}

// hashPassword 使用 bcrypt 加密密码
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// checkPasswordHash 验证密码
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generateJWT 生成 JWT token
func generateJWT(userID int64, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// validateJWT 验证 JWT token
func validateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// authMiddleware JWT 认证中间件
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			errorResponse(w, http.StatusUnauthorized, "authorization header required")
			return
		}

		// 支持 "Bearer <token>" 格式
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			errorResponse(w, http.StatusUnauthorized, "invalid authorization format")
			return
		}

		claims, err := validateJWT(tokenString)
		if err != nil {
			errorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		// 将用户信息附加到请求头，供 handler 使用
		r.Header.Set("X-User-ID", fmt.Sprintf("%d", claims.UserID))
		r.Header.Set("X-Username", claims.Username)

		infoLog.Printf("Authenticated user: %s (ID: %d)", claims.Username, claims.UserID)
		next(w, r)
	}
}

// validateEmail 验证邮箱格式
func validateEmail(email string) bool {
	if email == "" {
		return true // 邮箱可选
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// validateUsername 验证用户名格式
func validateUsername(username string) bool {
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return usernameRegex.MatchString(username)
}

// validatePassword 验证密码强度
func validatePassword(password string) bool {
	return len(password) >= 6
}

// 初始化数据库
func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./test.db")
	if err != nil {
		log.Fatal(err)
	}

	// 创建用户表
	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT NOT NULL UNIQUE,
            password TEXT NOT NULL,
            phone TEXT,
            email TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );`)
	if err != nil {
		log.Fatal(err)
	}

	// 创建图片表
	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS images (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER NOT NULL,
            image_data BLOB NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (user_id) REFERENCES users(id)
        );`)
	if err != nil {
		log.Fatal(err)
	}

	// 创建 todos 表
	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS todos (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER NOT NULL,
            title TEXT NOT NULL,
            completed BOOLEAN DEFAULT 0,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (user_id) REFERENCES users(id)
        );`)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	// 初始化日志
	infoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	initDB()
	rand.Seed(time.Now().UnixNano())
	infoLog.Println("Server initialized successfully")
}

// handleHealth 简单健康检查
//
//	@Summary		Health check
//	@Description	Simple health check endpoint
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/health [get]
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleListTodos 处理 GET /todos
//
//	@Summary		List all todos
//	@Description	Get all todos from database
//	@Tags			todos
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	query		int	false	"Filter by user ID"
//	@Success		200		{array}		Todo
//	@Failure		401		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/todos [get]
func handleListTodos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	var rows *sql.Rows
	var err error

	if userIDStr != "" {
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, "invalid user_id")
			return
		}
		rows, err = db.Query("SELECT id, user_id, title, completed, created_at FROM todos WHERE user_id = ?", userID)
	} else {
		rows, err = db.Query("SELECT id, user_id, title, completed, created_at FROM todos")
	}

	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "database query failed")
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Completed, &todo.CreatedAt); err != nil {
			errorResponse(w, http.StatusInternalServerError, "scan failed")
			return
		}
		todos = append(todos, todo)
	}

	writeJSON(w, http.StatusOK, todos)
}

// handleGetTodo 处理 GET /todos/{id}
//
//	@Summary		Get a todo by ID
//	@Description	Get a specific todo by its ID
//	@Tags			todos
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Todo ID"
//	@Success		200	{object}	Todo
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/todos/{id} [get]
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

	var todo Todo
	err = db.QueryRow("SELECT id, user_id, title, completed, created_at FROM todos WHERE id = ?", id).Scan(
		&todo.ID, &todo.UserID, &todo.Title, &todo.Completed, &todo.CreatedAt)

	if err == sql.ErrNoRows {
		errorResponse(w, http.StatusNotFound, "todo not found")
		return
	} else if err != nil {
		errorResponse(w, http.StatusInternalServerError, "database query failed")
		return
	}

	writeJSON(w, http.StatusOK, todo)
}

// handleCreateTodo 处理 POST /todos
//
//	@Summary		Create a new todo
//	@Description	Create a new todo item
//	@Tags			todos
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			todo	body		Todo	true	"Todo object"
//	@Success		201		{object}	Todo
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/todos [post]
func handleCreateTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var input struct {
		UserID int64  `json:"user_id"`
		Title  string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if input.Title == "" {
		errorResponse(w, http.StatusBadRequest, "title is required")
		return
	}
	if input.UserID == 0 {
		errorResponse(w, http.StatusBadRequest, "user_id is required")
		return
	}

	// 检查用户是否存在
	var exists int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", input.UserID).Scan(&exists)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "database query failed")
		return
	}
	if exists == 0 {
		errorResponse(w, http.StatusBadRequest, "user does not exist")
		return
	}

	result, err := db.Exec(
		"INSERT INTO todos (user_id, title, completed) VALUES (?, ?, ?)",
		input.UserID, input.Title, false)

	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "database insert failed")
		return
	}

	id, _ := result.LastInsertId()

	var todo Todo
	err = db.QueryRow("SELECT id, user_id, title, completed, created_at FROM todos WHERE id = ?", id).Scan(
		&todo.ID, &todo.UserID, &todo.Title, &todo.Completed, &todo.CreatedAt)

	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "failed to retrieve created todo")
		return
	}

	writeJSON(w, http.StatusCreated, todo)
}

// handleUpdateTodo 处理 PUT /todos/{id}
//
//	@Summary		Update a todo
//	@Description	Update an existing todo item
//	@Tags			todos
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int		true	"Todo ID"
//	@Param			todo	body		Todo	true	"Todo object"
//	@Success		200		{object}	Todo
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/todos/{id} [put]
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

	// 检查todo是否存在
	var exists int
	err = db.QueryRow("SELECT COUNT(*) FROM todos WHERE id = ?", id).Scan(&exists)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "database query failed")
		return
	}
	if exists == 0 {
		errorResponse(w, http.StatusNotFound, "todo not found")
		return
	}

	// 构建更新语句
	query := "UPDATE todos SET "
	args := []interface{}{}
	updates := []string{}

	if input.Title != nil {
		updates = append(updates, "title = ?")
		args = append(args, *input.Title)
	}
	if input.Completed != nil {
		updates = append(updates, "completed = ?")
		args = append(args, *input.Completed)
	}

	if len(updates) == 0 {
		errorResponse(w, http.StatusBadRequest, "no fields to update")
		return
	}

	query += strings.Join(updates, ", ") + " WHERE id = ?"
	args = append(args, id)

	_, err = db.Exec(query, args...)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "database update failed")
		return
	}

	var todo Todo
	err = db.QueryRow("SELECT id, user_id, title, completed, created_at FROM todos WHERE id = ?", id).Scan(
		&todo.ID, &todo.UserID, &todo.Title, &todo.Completed, &todo.CreatedAt)

	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "failed to retrieve updated todo")
		return
	}

	writeJSON(w, http.StatusOK, todo)
}

// handleDeleteTodo 处理 DELETE /todos/{id}
//
//	@Summary		Delete a todo
//	@Description	Delete an existing todo item
//	@Tags			todos
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Todo ID"
//	@Success		204	{object}	nil
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/todos/{id} [delete]
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

	result, err := db.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "database delete failed")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		errorResponse(w, http.StatusNotFound, "todo not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleListUsers 处理 GET /users
//
//	@Summary		List all users
//	@Description	Get all users from database
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		User
//	@Failure		500	{object}	map[string]string
//	@Router			/users [get]
func handleListUsers(w http.ResponseWriter, _ *http.Request) {
	rows, err := db.Query("SELECT id, username, password, phone, email, created_at FROM users")
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "database query failed")
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Password, &user.Phone, &user.Email, &user.CreatedAt); err != nil {
			errorResponse(w, http.StatusInternalServerError, "scan failed")
			return
		}
		users = append(users, user)
	}

	writeJSON(w, http.StatusOK, users)
}

// handleGetUser 处理 GET /users/{id}
//
//	@Summary		Get a user by ID
//	@Description	Get a specific user by their ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	User
//	@Failure		400	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/users/{id} [get]
func handleGetUser(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/users/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid id")
		return
	}

	var user User
	err = db.QueryRow("SELECT id, username, password, phone, email, created_at FROM users WHERE id = ?", id).Scan(
		&user.ID, &user.Username, &user.Password, &user.Phone, &user.Email, &user.CreatedAt)

	if err == sql.ErrNoRows {
		errorResponse(w, http.StatusNotFound, "user not found")
		return
	} else if err != nil {
		errorResponse(w, http.StatusInternalServerError, "database query failed")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// handleCreateUser 处理 POST /users
//
//	@Summary		Create a new user
//	@Description	Create a new user in database
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			user	body		User	true	"User object"
//	@Success		201		{object}	User
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/users [post]
func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Phone    string `json:"phone"`
		Email    string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// 输入验证
	if input.Username == "" || input.Password == "" {
		errorResponse(w, http.StatusBadRequest, "username and password are required")
		return
	}

	if !validateUsername(input.Username) {
		errorResponse(w, http.StatusBadRequest, "username must be 3-20 characters and contain only letters, numbers, and underscores")
		return
	}

	if !validatePassword(input.Password) {
		errorResponse(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	if !validateEmail(input.Email) {
		errorResponse(w, http.StatusBadRequest, "invalid email format")
		return
	}

	// 加密密码
	hashedPassword, err := hashPassword(input.Password)
	if err != nil {
		errorLog.Printf("Failed to hash password: %v", err)
		errorResponse(w, http.StatusInternalServerError, "failed to process password")
		return
	}

	result, err := db.Exec(
		"INSERT INTO users (username, password, phone, email) VALUES (?, ?, ?, ?)",
		input.Username, hashedPassword, input.Phone, input.Email)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			errorResponse(w, http.StatusBadRequest, "username already exists")
		} else {
			errorLog.Printf("Database insert failed: %v", err)
			errorResponse(w, http.StatusInternalServerError, "database insert failed")
		}
		return
	}

	id, _ := result.LastInsertId()

	var user User
	err = db.QueryRow("SELECT id, username, password, phone, email, created_at FROM users WHERE id = ?", id).Scan(
		&user.ID, &user.Username, &user.Password, &user.Phone, &user.Email, &user.CreatedAt)

	if err != nil {
		errorLog.Printf("Failed to retrieve created user: %v", err)
		errorResponse(w, http.StatusInternalServerError, "failed to retrieve created user")
		return
	}

	// 不返回密码哈希
	user.Password = ""
	infoLog.Printf("User created: %s (ID: %d)", user.Username, user.ID)
	writeJSON(w, http.StatusCreated, user)
}

// handleUpdateUser 处理 PUT /users/{id}
//
//	@Summary		Update a user
//	@Description	Update an existing user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int		true	"User ID"
//	@Param			user	body		User	true	"User object"
//	@Success		200		{object}	User
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/users/{id} [put]
func handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/users/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid id")
		return
	}

	var input struct {
		Username *string `json:"username"`
		Password *string `json:"password"`
		Phone    *string `json:"phone"`
		Email    *string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// 检查用户是否存在
	var exists int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", id).Scan(&exists)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "database query failed")
		return
	}
	if exists == 0 {
		errorResponse(w, http.StatusNotFound, "user not found")
		return
	}

	// 构建更新语句
	query := "UPDATE users SET "
	args := []interface{}{}
	updates := []string{}

	if input.Username != nil {
		updates = append(updates, "username = ?")
		args = append(args, *input.Username)
	}
	if input.Password != nil {
		updates = append(updates, "password = ?")
		args = append(args, *input.Password)
	}
	if input.Phone != nil {
		updates = append(updates, "phone = ?")
		args = append(args, *input.Phone)
	}
	if input.Email != nil {
		updates = append(updates, "email = ?")
		args = append(args, *input.Email)
	}

	if len(updates) == 0 {
		errorResponse(w, http.StatusBadRequest, "no fields to update")
		return
	}

	query += strings.Join(updates, ", ") + " WHERE id = ?"
	args = append(args, id)

	_, err = db.Exec(query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			errorResponse(w, http.StatusBadRequest, "username already exists")
		} else {
			errorResponse(w, http.StatusInternalServerError, "database update failed")
		}
		return
	}

	var user User
	err = db.QueryRow("SELECT id, username, password, phone, email, created_at FROM users WHERE id = ?", id).Scan(
		&user.ID, &user.Username, &user.Password, &user.Phone, &user.Email, &user.CreatedAt)

	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "failed to retrieve updated user")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// handleDeleteUser 处理 DELETE /users/{id}
//
//	@Summary		Delete a user
//	@Description	Delete an existing user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		204	{object}	nil
//	@Failure		400	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/users/{id} [delete]
func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/users/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid id")
		return
	}

	result, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "database delete failed")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		errorResponse(w, http.StatusNotFound, "user not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleListImages 处理 GET /images
func handleListImages(w http.ResponseWriter, _ *http.Request) {
	errorResponse(w, http.StatusNotImplemented, "not implemented yet")
}

// handleGetImage 处理 GET /images/{id}
func handleGetImage(w http.ResponseWriter, _ *http.Request) {
	errorResponse(w, http.StatusNotImplemented, "not implemented yet")
}

// handleCreateImage 处理 POST /images
func handleCreateImage(w http.ResponseWriter, _ *http.Request) {
	errorResponse(w, http.StatusNotImplemented, "not implemented yet")
}

// handleUpdateImage 处理 PUT /images/{id}
func handleUpdateImage(w http.ResponseWriter, _ *http.Request) {
	errorResponse(w, http.StatusNotImplemented, "not implemented yet")
}

// handleDeleteImage 处理 DELETE /images/{id}
func handleDeleteImage(w http.ResponseWriter, _ *http.Request) {
	errorResponse(w, http.StatusNotImplemented, "not implemented yet")
}

// handleResetPassword 处理密码重置
func handleResetPassword(w http.ResponseWriter, _ *http.Request) {
	errorResponse(w, http.StatusNotImplemented, "not implemented yet")
}

// handleLogin 处理用户登录
//
//	@Summary		User login
//	@Description	Login with username and password to get JWT token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			login	body		LoginRequest	true	"Login credentials"
//	@Success		200		{object}	LoginResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/login [post]
func handleLogin(w http.ResponseWriter, r *http.Request) {
	var input LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if input.Username == "" || input.Password == "" {
		errorResponse(w, http.StatusBadRequest, "username and password are required")
		return
	}

	// 查询用户
	var user User
	err := db.QueryRow("SELECT id, username, password, phone, email, created_at FROM users WHERE username = ?", input.Username).Scan(
		&user.ID, &user.Username, &user.Password, &user.Phone, &user.Email, &user.CreatedAt)

	if err == sql.ErrNoRows {
		errorResponse(w, http.StatusUnauthorized, "invalid username or password")
		return
	} else if err != nil {
		errorLog.Printf("Database query failed: %v", err)
		errorResponse(w, http.StatusInternalServerError, "database query failed")
		return
	}

	// 验证密码
	if !checkPasswordHash(input.Password, user.Password) {
		errorResponse(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	// 生成 JWT token
	token, err := generateJWT(user.ID, user.Username)
	if err != nil {
		errorLog.Printf("Failed to generate JWT: %v", err)
		errorResponse(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	// 不返回密码
	user.Password = ""
	infoLog.Printf("User logged in: %s (ID: %d)", user.Username, user.ID)

	writeJSON(w, http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
	})
}

func main() {
	mux := http.NewServeMux()

	// 健康检查
	//	@Summary		Health check
	//	@Description	Simple health check endpoint
	//	@Tags			health
	//	@Accept			json
	//	@Produce		json
	//	@Success		200	{object}	map[string]string
	//	@Router			/health [get]
	mux.HandleFunc("/health", handleHealth)

	// 登录路由（不需要认证）
	//	@Summary		Login API
	//	@Description	User login endpoint
	//	@Tags			auth
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		handleLogin(w, r)
	})

	// 注册路由（不需要认证）
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		handleCreateUser(w, r)
	})

	// Todo 管理路由（需要认证）
	//	@Summary		Todo API
	//	@Description	RESTful Todo API endpoints (requires authentication)
	//	@Tags			todos
	//	@Security		BearerAuth
	mux.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
		// 需要认证
		authMiddleware(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				handleListTodos(w, r)
			case http.MethodPost:
				handleCreateTodo(w, r)
			default:
				w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
				errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
			}
		})(w, r)
	})

	mux.HandleFunc("/todos/", func(w http.ResponseWriter, r *http.Request) {
		// 需要认证
		authMiddleware(func(w http.ResponseWriter, r *http.Request) {
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
		})(w, r)
	})

	// 用户管理路由
	//	@Summary		User API
	//	@Description	RESTful User API endpoints
	//	@Tags			users
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleListUsers(w, r)
		case http.MethodPost:
			handleCreateUser(w, r)
		default:
			w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
			errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetUser(w, r)
		case http.MethodPut:
			handleUpdateUser(w, r)
		case http.MethodDelete:
			handleDeleteUser(w, r)
		default:
			w.Header().Set("Allow", http.MethodGet+", "+http.MethodPut+", "+http.MethodDelete)
			errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	// 图片管理路由
	//	@Summary		Image API
	//	@Description	RESTful Image API endpoints
	//	@Tags			images
	mux.HandleFunc("/images", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleListImages(w, r)
		case http.MethodPost:
			handleCreateImage(w, r)
		default:
			w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
			errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	mux.HandleFunc("/images/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetImage(w, r)
		case http.MethodPut:
			handleUpdateImage(w, r)
		case http.MethodDelete:
			handleDeleteImage(w, r)
		default:
			w.Header().Set("Allow", http.MethodGet+", "+http.MethodPut+", "+http.MethodDelete)
			errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	// 密码重置路由
	//	@Summary		Password Reset API
	//	@Description	RESTful Password Reset API endpoints
	//	@Tags			password
	mux.HandleFunc("/reset-password", handleResetPassword)

	// Swagger docs
	mux.HandleFunc("/docs/", httpSwagger.WrapHandler)

	addr := ":8080"
	log.Printf("Starting webserver on %s...", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
