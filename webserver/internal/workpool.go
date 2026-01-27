package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// WorkerPool 工作池，管理多个 worker goroutine 处理任务
type WorkerPool struct {
	taskQueue    chan *ImageTask
	workerCount  int
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	imageClients []TextToImageProvider
	balancer     *LoadBalancer
	taskManager  *TaskManager
}

// NewWorkerPool 创建新的 worker pool
func NewWorkerPool(workerCount int, queueSize int, imageClients []TextToImageProvider, taskManager *TaskManager) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		taskQueue:    make(chan *ImageTask, queueSize),
		workerCount:  workerCount,
		ctx:          ctx,
		cancel:       cancel,
		imageClients: imageClients,
		balancer:     NewLoadBalancer(imageClients),
		taskManager:  taskManager,
	}
}

// Start 启动 worker pool
func (wp *WorkerPool) Start() {
	log.Printf("Starting worker pool with %d workers", wp.workerCount)
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// Stop 停止 worker pool
func (wp *WorkerPool) Stop() {
	log.Println("Stopping worker pool...")
	wp.cancel()
	close(wp.taskQueue)
	wp.wg.Wait()
	log.Println("Worker pool stopped")
}

// Submit 提交任务到队列
func (wp *WorkerPool) Submit(task *ImageTask) error {
	select {
	case wp.taskQueue <- task:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("task queue is full, timeout after 5s")
	}
}

// worker 处理任务的 goroutine
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	log.Printf("Worker %d started", id)

	for {
		select {
		case <-wp.ctx.Done():
			log.Printf("Worker %d stopped", id)
			return

		case task, ok := <-wp.taskQueue:
			if !ok {
				log.Printf("Worker %d: task queue closed", id)
				return
			}

			wp.processTask(id, task)
		}
	}
}

// processTask 处理单个任务
func (wp *WorkerPool) processTask(workerID int, task *ImageTask) {
	log.Printf("Worker %d processing task %s for user %d", workerID, task.ID, task.UserID)

	// 更新任务状态为运行中
	task.Status = TaskRunning
	task.UpdatedAt = time.Now()
	if err := wp.taskManager.UpdateTask(task); err != nil {
		log.Printf("Worker %d: failed to update task status: %v", workerID, err)
	}

	// 获取可用的图片生成客户端
	client := wp.balancer.GetNext()
	if client == nil {
		log.Printf("Worker %d: no available image generation client", workerID)
		task.Status = TaskFailed
		task.ErrorMsg = "no available image generation service"
		task.UpdatedAt = time.Now()
		_ = wp.taskManager.UpdateTask(task)
		return
	}

	// 执行图片生成
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	startTime := time.Now()
	resp, err := client.Generate(ctx, TextToImageRequest{
		Prompt: task.Prompt,
		Steps:  20,
	})
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("Worker %d: task %s failed after %.2fs: %v", workerID, task.ID, duration.Seconds(), err)
		task.Status = TaskFailed
		task.ErrorMsg = err.Error()
		task.UpdatedAt = time.Now()
		_ = wp.taskManager.UpdateTask(task)
		return
	}

	// 保存生成的图片到数据库
	imageID, err := wp.taskManager.SaveImage(task.UserID, task.Prompt, resp.ImageData, resp.MimeType)
	if err != nil {
		log.Printf("Worker %d: failed to save image: %v", workerID, err)
		task.Status = TaskFailed
		task.ErrorMsg = fmt.Sprintf("failed to save image: %v", err)
		task.UpdatedAt = time.Now()
		_ = wp.taskManager.UpdateTask(task)
		return
	}

	// 更新任务状态为完成
	task.Status = TaskDone
	task.ResultURL = fmt.Sprintf("/api/images/%d", imageID)
	task.UpdatedAt = time.Now()
	if err := wp.taskManager.UpdateTask(task); err != nil {
		log.Printf("Worker %d: failed to update task status: %v", workerID, err)
	}

	log.Printf("Worker %d: task %s completed in %.2fs", workerID, task.ID, duration.Seconds())
}

// GetQueueLength 获取队列长度
func (wp *WorkerPool) GetQueueLength() int {
	return len(wp.taskQueue)
}

// GetQueueCapacity 获取队列容量
func (wp *WorkerPool) GetQueueCapacity() int {
	return cap(wp.taskQueue)
}
