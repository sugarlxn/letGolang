package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 全局变量：异步任务系统组件
var (
	globalTaskManager *TaskManager
	globalWorkerPool  *WorkerPool
	globalAsyncAPI    *AsyncAPIHandlers
)

// initAsyncSystem 初始化异步任务系统
func initAsyncSystem() error {
	log.Println("Initializing async task system...")

	// 1. 初始化 TaskManager
	globalTaskManager = NewTaskManager(db)

	// 2. 初始化文生图客户端实例
	imageBaseURLs := []string{
		getEnv("IMAGE_GEN_URL_1", "http://localhost:8000"), // 第一个实例
		// 后续可以添加更多实例
		// getEnv("IMAGE_GEN_URL_2", "http://localhost:8001"),
		// getEnv("IMAGE_GEN_URL_3", "http://localhost:8002"),
	}

	var imageClients []TextToImageProvider
	for i, baseURL := range imageBaseURLs {
		client, err := NewQwenImageGGUF(baseURL, nil)
		if err != nil {
			log.Printf("Warning: failed to create image client %d (%s): %v", i, baseURL, err)
			continue
		}
		imageClients = append(imageClients, client)
		log.Printf("Initialized image generation client %d: %s", i, baseURL)
	}

	if len(imageClients) == 0 {
		log.Println("Warning: no image generation clients available")
	}

	// 3. 初始化语音转文字客户端
	whisperURL := getEnv("WHISPER_URL", "http://localhost:8001")
	whisperClient, err := NewFastWhisperService(whisperURL, nil)
	if err != nil {
		log.Printf("Warning: failed to create whisper client: %v", err)
	} else {
		log.Printf("Initialized speech-to-text client: %s", whisperURL)
	}

	// 4. 初始化 WorkerPool
	workerCount := 2 // 并发 worker 数量
	queueSize := 100 // 队列容量
	globalWorkerPool = NewWorkerPool(workerCount, queueSize, imageClients, globalTaskManager)
	globalWorkerPool.Start()

	// 5. 初始化异步 API 处理器
	globalAsyncAPI = NewAsyncAPIHandlers(globalWorkerPool, globalTaskManager, whisperClient)

	log.Println("Async task system initialized successfully")
	return nil
}

// shutdownAsyncSystem 优雅关闭异步任务系统
func shutdownAsyncSystem() {
	log.Println("Shutting down async task system...")

	if globalWorkerPool != nil {
		globalWorkerPool.Stop()
	}

	log.Println("Async task system shutdown complete")
}

// setupGracefulShutdown 设置优雅关闭
func setupGracefulShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v", sig)
		shutdownAsyncSystem()
		if db != nil {
			db.Close()
		}
		os.Exit(0)
	}()
}

// startBackgroundCleanup 启动后台清理任务
func startBackgroundCleanup() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			// 清理 7 天前的已完成/失败任务
			if err := globalTaskManager.CleanupOldTasks(7 * 24 * time.Hour); err != nil {
				log.Printf("Background cleanup error: %v", err)
			}
		}
	}()

	log.Println("Background cleanup task started")
}
