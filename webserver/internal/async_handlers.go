package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"
)

// AsyncAPIHandlers 异步 API 处理器集合
type AsyncAPIHandlers struct {
	workerPool  *WorkerPool
	taskManager *TaskManager
	whisperSvc  SpeechToTextProvider
}

// NewAsyncAPIHandlers 创建异步 API 处理器
func NewAsyncAPIHandlers(wp *WorkerPool, tm *TaskManager, ws SpeechToTextProvider) *AsyncAPIHandlers {
	return &AsyncAPIHandlers{
		workerPool:  wp,
		taskManager: tm,
		whisperSvc:  ws,
	}
}

//
// ======================
// 文生图异步接口
// ======================
//

// SubmitImageTaskRequest 提交图片生成任务请求
type SubmitImageTaskRequest struct {
	Prompt         string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
}

// SubmitImageTaskResponse 提交图片生成任务响应
type SubmitImageTaskResponse struct {
	TaskID  string `json:"task_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// HandleSubmitImageTask 处理提交图片生成任务
//
//	@Summary		Submit image generation task
//	@Description	Submit an async image generation task and return task ID
//	@Tags			async-tasks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		SubmitImageTaskRequest	true	"Image generation request"
//	@Success		202		{object}	SubmitImageTaskResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/api/v1/image/async [post]
func (h *AsyncAPIHandlers) HandleSubmitImageTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 从请求头获取用户 ID
	userIDStr := r.Header.Get("X-User-ID")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	// 解析请求体
	var req SubmitImageTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Prompt == "" {
		errorResponse(w, http.StatusBadRequest, "prompt is required")
		return
	}

	// 创建任务
	task, err := h.taskManager.CreateTask(userID, req.Prompt)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "failed to create task")
		return
	}

	// 提交到 worker pool
	if err := h.workerPool.Submit(task); err != nil {
		errorResponse(w, http.StatusServiceUnavailable, "task queue is full, please try again later")
		return
	}

	// 返回任务 ID
	writeJSON(w, http.StatusAccepted, SubmitImageTaskResponse{
		TaskID:  task.ID,
		Status:  task.Status,
		Message: "task submitted successfully",
	})
}

// HandleGetTaskStatus 查询任务状态
//
//	@Summary		Get task status
//	@Description	Query the status of an async task by task ID
//	@Tags			async-tasks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			task_id	path		string	true	"Task ID"
//	@Success		200		{object}	ImageTask
//	@Failure		401		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Router			/api/v1/tasks/{task_id} [get]
func (h *AsyncAPIHandlers) HandleGetTaskStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 从请求头获取用户 ID（验证权限）
	userIDStr := r.Header.Get("X-User-ID")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	// 从 URL 获取任务 ID
	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		errorResponse(w, http.StatusBadRequest, "task_id is required")
		return
	}

	// 获取任务
	task, err := h.taskManager.GetTask(taskID)
	if err != nil {
		errorResponse(w, http.StatusNotFound, "task not found")
		return
	}

	// 验证任务所属用户
	if task.UserID != userID {
		errorResponse(w, http.StatusForbidden, "access denied")
		return
	}

	writeJSON(w, http.StatusOK, task)
}

// HandleGetUserTasks 获取用户的所有任务
//
//	@Summary		Get user tasks
//	@Description	Get all tasks for the authenticated user
//	@Tags			async-tasks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query		int	false	"Maximum number of tasks to return"	default(50)
//	@Success		200		{array}		ImageTask
//	@Failure		401		{object}	map[string]string
//	@Router			/api/v1/tasks [get]
func (h *AsyncAPIHandlers) HandleGetUserTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 从请求头获取用户 ID
	userIDStr := r.Header.Get("X-User-ID")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		errorResponse(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	// 获取 limit 参数
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// 获取任务列表
	tasks, err := h.taskManager.GetUserTasks(userID, limit)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "failed to get tasks")
		return
	}

	writeJSON(w, http.StatusOK, tasks)
}

//
// ======================
// 语音转文字接口
// ======================
//

// HandleSpeechToText 处理语音转文字（同步接口）
//
//	@Summary		Speech to text
//	@Description	Convert audio file to text
//	@Tags			speech
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			file	formData	file	true	"Audio file"
//	@Success		200		{object}	SpeechToTextResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Router			/api/v1/speech/transcribe [post]
func (h *AsyncAPIHandlers) HandleSpeechToText(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 解析 multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		errorResponse(w, http.StatusBadRequest, "failed to parse form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	// 读取文件数据
	audioData, err := io.ReadAll(file)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "failed to read file")
		return
	}

	// 调用语音识别服务
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := h.whisperSvc.TranscribeFile(ctx, audioData, header.Filename)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "transcription failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// HandleSpeechToTextPCM 处理 PCM 音频转文字（ESP32 专用）
//
//	@Summary		Speech to text (PCM)
//	@Description	Convert PCM audio data to text (for ESP32)
//	@Tags			speech
//	@Accept			application/octet-stream
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		[]byte	true	"PCM audio data"
//	@Success		200		{object}	SpeechToTextResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Router			/api/v1/speech/pcm [post]
func (h *AsyncAPIHandlers) HandleSpeechToTextPCM(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 读取 PCM 数据
	pcmData, err := io.ReadAll(r.Body)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "failed to read body")
		return
	}

	// 调用语音识别服务
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := h.whisperSvc.TranscribePCM(ctx, pcmData)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "transcription failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

//
// ======================
// 系统监控接口
// ======================
//

// HandleSystemStats 获取系统统计信息
//
//	@Summary		System statistics
//	@Description	Get system statistics including queue length and worker status
//	@Tags			system
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]interface{}
//	@Router			/api/v1/system/stats [get]
func (h *AsyncAPIHandlers) HandleSystemStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"queue_length":   h.workerPool.GetQueueLength(),
		"queue_capacity": h.workerPool.GetQueueCapacity(),
		"queue_usage":    float64(h.workerPool.GetQueueLength()) / float64(h.workerPool.GetQueueCapacity()) * 100,
	}

	writeJSON(w, http.StatusOK, stats)
}
