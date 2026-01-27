package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ======================
// Task & Status
// ======================

const (
	TaskQueued  = "QUEUED"
	TaskRunning = "RUNNING"
	TaskDone    = "DONE"
	TaskFailed  = "FAILED"
)

type ImageTask struct {
	ID        string    `json:"task_id"`
	UserID    int64     `json:"user_id"`
	Prompt    string    `json:"prompt"`
	Status    string    `json:"status"`
	ResultURL string    `json:"result_url,omitempty"`
	ErrorMsg  string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ======================
// TaskManager
// ======================

// TaskManager 管理任务的持久化和查询
type TaskManager struct {
	db    *sql.DB
	mu    sync.RWMutex
	cache map[string]*ImageTask
}

// NewTaskManager 创建新的 TaskManager
func NewTaskManager(db *sql.DB) *TaskManager {
	return &TaskManager{
		db:    db,
		cache: make(map[string]*ImageTask),
	}
}

// CreateTask 创建新任务
func (tm *TaskManager) CreateTask(userID int64, prompt string) (*ImageTask, error) {
	task := &ImageTask{
		ID:        uuid.New().String(),
		UserID:    userID,
		Prompt:    prompt,
		Status:    TaskQueued,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := tm.db.Exec(`
		INSERT INTO image_tasks (id, user_id, prompt, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, task.ID, task.UserID, task.Prompt, task.Status, task.CreatedAt, task.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	tm.mu.Lock()
	tm.cache[task.ID] = task
	tm.mu.Unlock()

	log.Printf("Created task %s for user %d: %s", task.ID, userID, prompt)
	return task, nil
}

// UpdateTask 更新任务状态
func (tm *TaskManager) UpdateTask(task *ImageTask) error {
	task.UpdatedAt = time.Now()

	_, err := tm.db.Exec(`
		UPDATE image_tasks 
		SET status = ?, result_url = ?, error_msg = ?, updated_at = ?
		WHERE id = ?
	`, task.Status, task.ResultURL, task.ErrorMsg, task.UpdatedAt, task.ID)

	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	tm.mu.Lock()
	tm.cache[task.ID] = task
	tm.mu.Unlock()

	return nil
}

// GetTask 获取任务
func (tm *TaskManager) GetTask(taskID string) (*ImageTask, error) {
	tm.mu.RLock()
	if task, ok := tm.cache[taskID]; ok {
		tm.mu.RUnlock()
		return task, nil
	}
	tm.mu.RUnlock()

	var task ImageTask
	err := tm.db.QueryRow(`
		SELECT id, user_id, prompt, status, result_url, error_msg, created_at, updated_at
		FROM image_tasks
		WHERE id = ?
	`, taskID).Scan(
		&task.ID, &task.UserID, &task.Prompt, &task.Status,
		&task.ResultURL, &task.ErrorMsg, &task.CreatedAt, &task.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	tm.mu.Lock()
	tm.cache[taskID] = &task
	tm.mu.Unlock()

	return &task, nil
}

// GetUserTasks 获取用户的所有任务
func (tm *TaskManager) GetUserTasks(userID int64, limit int) ([]*ImageTask, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := tm.db.Query(`
		SELECT id, user_id, prompt, status, result_url, error_msg, created_at, updated_at
		FROM image_tasks
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, userID, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to get user tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*ImageTask
	for rows.Next() {
		var task ImageTask
		err := rows.Scan(
			&task.ID, &task.UserID, &task.Prompt, &task.Status,
			&task.ResultURL, &task.ErrorMsg, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

// SaveImage 保存生成的图片到数据库
func (tm *TaskManager) SaveImage(userID int64, prompt string, imageData []byte, mimeType string) (int64, error) {
	format := "jpeg"
	if mimeType == "image/png" {
		format = "png"
	}

	result, err := tm.db.Exec(`
		INSERT INTO prompts (user_id, prompt_text, created_at)
		VALUES (?, ?, ?)
	`, userID, prompt, time.Now())

	if err != nil {
		return 0, fmt.Errorf("failed to create prompt: %w", err)
	}

	promptID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get prompt id: %w", err)
	}

	result, err = tm.db.Exec(`
		INSERT INTO images (user_id, prompt_id, image_data, image_format, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, userID, promptID, imageData, format, time.Now())

	if err != nil {
		return 0, fmt.Errorf("failed to save image: %w", err)
	}

	imageID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get image id: %w", err)
	}

	log.Printf("Saved image %d for user %d (prompt_id: %d)", imageID, userID, promptID)
	return imageID, nil
}

// CleanupOldTasks 清理旧任务
func (tm *TaskManager) CleanupOldTasks(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	result, err := tm.db.Exec(`
		DELETE FROM image_tasks
		WHERE created_at < ? AND status IN (?, ?)
	`, cutoff, TaskDone, TaskFailed)

	if err != nil {
		return fmt.Errorf("failed to cleanup old tasks: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("Cleaned up %d old tasks", rowsAffected)
	}

	tm.mu.Lock()
	for id, task := range tm.cache {
		if task.CreatedAt.Before(cutoff) {
			delete(tm.cache, id)
		}
	}
	tm.mu.Unlock()

	return nil
}
