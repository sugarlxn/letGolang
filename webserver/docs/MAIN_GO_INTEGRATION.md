// ============================================================
// 以下代码需要集成到 main.go 的 main() 函数中
// ============================================================

// 在 main() 函数中，在初始化数据库之后添加以下代码：

/*

// 初始化异步任务系统
if err := initAsyncSystem(); err != nil {
    errorLog.Fatalf("failed to initialize async system: %v", err)
}

// 启动后台清理任务
startBackgroundCleanup()

// 设置优雅关闭
setupGracefulShutdown()

*/

// ============================================================
// 然后在 mux.HandleFunc 路由部分添加以下路由：
// ============================================================

/*

// 异步任务相关路由
mux.HandleFunc("/api/v1/image/async", func(w http.ResponseWriter, r *http.Request) {
    authMiddleware(globalAsyncAPI.HandleSubmitImageTask)(w, r)
})

mux.HandleFunc("/api/v1/tasks", func(w http.ResponseWriter, r *http.Request) {
    authMiddleware(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Query().Get("task_id") != "" {
            globalAsyncAPI.HandleGetTaskStatus(w, r)
        } else {
            globalAsyncAPI.HandleGetUserTasks(w, r)
        }
    })(w, r)
})

// 语音转文字路由
mux.HandleFunc("/api/v1/speech/transcribe", func(w http.ResponseWriter, r *http.Request) {
    authMiddleware(globalAsyncAPI.HandleSpeechToText)(w, r)
})

mux.HandleFunc("/api/v1/speech/pcm", func(w http.ResponseWriter, r *http.Request) {
    authMiddleware(globalAsyncAPI.HandleSpeechToTextPCM)(w, r)
})

// 系统监控路由
mux.HandleFunc("/api/v1/system/stats", func(w http.ResponseWriter, r *http.Request) {
    authMiddleware(globalAsyncAPI.HandleSystemStats)(w, r)
})

*/

// ============================================================
// 或者使用更清晰的方式集成（推荐）：
// ============================================================

/*

// 注册异步任务 API 路由
func registerAsyncAPIRoutes(mux *http.ServeMux) {
    // 文生图异步接口
    mux.HandleFunc("/api/v1/image/async", func(w http.ResponseWriter, r *http.Request) {
        authMiddleware(globalAsyncAPI.HandleSubmitImageTask)(w, r)
    })

    // 任务查询接口
    mux.HandleFunc("/api/v1/tasks", func(w http.ResponseWriter, r *http.Request) {
        authMiddleware(func(w http.ResponseWriter, r *http.Request) {
            switch r.Method {
            case http.MethodGet:
                if r.URL.Query().Get("task_id") != "" {
                    globalAsyncAPI.HandleGetTaskStatus(w, r)
                } else {
                    globalAsyncAPI.HandleGetUserTasks(w, r)
                }
            default:
                errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
            }
        })(w, r)
    })

    // 语音转文字接口
    mux.HandleFunc("/api/v1/speech/transcribe", func(w http.ResponseWriter, r *http.Request) {
        authMiddleware(globalAsyncAPI.HandleSpeechToText)(w, r)
    })

    mux.HandleFunc("/api/v1/speech/pcm", func(w http.ResponseWriter, r *http.Request) {
        authMiddleware(globalAsyncAPI.HandleSpeechToTextPCM)(w, r)
    })

    // 系统监控接口
    mux.HandleFunc("/api/v1/system/stats", func(w http.ResponseWriter, r *http.Request) {
        authMiddleware(globalAsyncAPI.HandleSystemStats)(w, r)
    })

    infoLog.Println("Async API routes registered")
}

// 在 main() 中调用：
// registerAsyncAPIRoutes(mux)

*/
