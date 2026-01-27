# 异步任务系统架构文档

## 系统概述

本系统实现了一个高性能的异步任务处理架构，专为处理耗时的文生图和语音转文字请求设计。

## 核心组件

### 1. TaskManager (taskqueue.go)
**职责：** 任务持久化和状态管理
- 创建任务并分配唯一 ID
- 更新任务状态（QUEUED → RUNNING → DONE/FAILED）
- 提供任务查询接口
- 内存缓存 + SQLite 持久化

**关键方法：**
```go
CreateTask(userID, prompt) (*ImageTask, error)
UpdateTask(*ImageTask) error
GetTask(taskID) (*ImageTask, error)
GetUserTasks(userID, limit) ([]*ImageTask, error)
SaveImage(userID, prompt, imageData, mimeType) (imageID, error)
```

### 2. WorkerPool (workpool.go)
**职责：** 管理 worker goroutine 池处理任务
- 固定数量的 worker 并发处理任务
- 从 channel 队列消费任务
- 调用负载均衡器选择实例
- 自动重试和错误处理

**配置参数：**
- `workerCount`: Worker 数量（推荐：2-4）
- `queueSize`: 队列容量（推荐：100-500）

### 3. LoadBalancer (loadbalancer.go)
**职责：** 将请求路由到不同的文生图实例
- **轮询策略：** 均匀分配负载
- **用户哈希：** 会话亲和性（同一用户→同一实例）
- **健康检查：** 定期 ping 实例，自动摘除故障节点
- **热更新：** 动态添加/移除实例

**支持策略：**
```go
GetNext()                          // 轮询
GetByStrategy("user-hash", userID) // 用户哈希
```

### 4. AsyncAPIHandlers (async_handlers.go)
**职责：** 提供 RESTful API 接口

#### 文生图异步接口
- `POST /api/v1/image/async` - 提交任务，返回 task_id
- `GET /api/v1/tasks?task_id=xxx` - 查询任务状态
- `GET /api/v1/tasks` - 获取用户所有任务

#### 语音转文字同步接口
- `POST /api/v1/speech/transcribe` - 上传音频文件转文字
- `POST /api/v1/speech/pcm` - PCM 数据转文字（ESP32）

#### 系统监控接口
- `GET /api/v1/system/stats` - 队列状态、使用率

## 数据流

```
用户请求
   ↓
[异步 API]
   ↓
[TaskManager] 创建任务 → SQLite
   ↓
[WorkerPool] 提交到队列 (channel)
   ↓
[Worker Goroutine] 消费任务
   ↓
[LoadBalancer] 选择实例
   ↓
[QwenImageGGUF] HTTP 调用 Python 后端
   ↓
[TaskManager] 保存结果 → SQLite
   ↓
[用户轮询] 查询任务状态
```

## 使用示例

### 1. 提交文生图任务

```bash
# 登录获取 token
TOKEN=$(curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user1","password":"pass123"}' \
  | jq -r '.token')

# 提交任务
curl -X POST http://localhost:8080/api/v1/image/async \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "A futuristic city with flying cars",
    "negative_prompt": "blurry, low quality"
  }'

# 响应
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "QUEUED",
  "message": "task submitted successfully"
}
```

### 2. 查询任务状态

```bash
curl -X GET "http://localhost:8080/api/v1/tasks?task_id=550e8400..." \
  -H "Authorization: Bearer $TOKEN"

# 响应（进行中）
{
  "task_id": "550e8400...",
  "user_id": 1,
  "prompt": "A futuristic city...",
  "status": "RUNNING",
  "created_at": "2026-01-27T10:00:00Z",
  "updated_at": "2026-01-27T10:00:05Z"
}

# 响应（完成）
{
  "task_id": "550e8400...",
  "status": "DONE",
  "result_url": "/api/images/123",
  "created_at": "2026-01-27T10:00:00Z",
  "updated_at": "2026-01-27T10:00:15Z"
}
```

### 3. 语音转文字

```bash
# 上传音频文件
curl -X POST http://localhost:8080/api/v1/speech/transcribe \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@audio.wav"

# 响应
{
  "code": 0,
  "language": "zh",
  "language_probability": 0.98,
  "segments": [
    {"start": 0.0, "end": 2.5, "text": "你好世界"},
    {"start": 2.5, "end": 5.0, "text": "这是一个测试"}
  ]
}
```

## 性能优化建议

### 1. 扩展文生图实例
```go
// 在 system_init.go 中添加更多实例
imageBaseURLs := []string{
    "http://localhost:8000",
    "http://localhost:8001",  // 新增
    "http://localhost:8002",  // 新增
}
```

### 2. 调整 Worker 数量
```go
// 根据实例数量调整 worker
// 建议：worker 数量 = 实例数量 * 2
workerCount := len(imageClients) * 2
```

### 3. 增加队列容量
```go
// 高峰期可增大队列容量
queueSize := 500  // 默认 100
```

### 4. 启用持久化队列
如需重启不丢失任务，可将 channel 替换为 Redis：
```go
// TODO: 使用 Redis 作为队列
// import "github.com/go-redis/redis/v8"
```

## 监控指标

### 队列状态
```bash
curl http://localhost:8080/api/v1/system/stats
{
  "queue_length": 5,
  "queue_capacity": 100,
  "queue_usage": 5.0
}
```

### 建议告警阈值
- 队列使用率 > 80%：需要增加 worker 或实例
- 任务失败率 > 10%：检查后端服务健康状态
- 平均处理时间 > 20s：考虑优化推理参数

## 部署架构

### 单机部署（当前）
```
┌─────────────────────────────────────┐
│  Webserver (Go)                     │
│  ├─ Worker Pool (2 workers)        │
│  ├─ Task Queue (channel, 100)      │
│  └─ LoadBalancer                    │
└─────────────────────────────────────┘
              ↓
     ┌────────────────┐
     │ Python Backend │
     │ (QwenImageGGUF)│
     └────────────────┘
```

### 多实例部署（推荐）
```
┌─────────────────────────────────────┐
│  Webserver (Go)                     │
│  ├─ Worker Pool (4 workers)        │
│  ├─ Task Queue (channel, 200)      │
│  └─ LoadBalancer (Round-robin)     │
└─────────────────────────────────────┘
       ↓         ↓         ↓
   ┌─────┐   ┌─────┐   ┌─────┐
   │GPU 1│   │GPU 2│   │GPU 3│
   └─────┘   └─────┘   └─────┘
```

### 高可用部署（企业级）
```
       ┌─────────────┐
       │  Nginx LB   │
       └─────────────┘
         ↓         ↓
   ┌──────────┐ ┌──────────┐
   │ Go Web 1 │ │ Go Web 2 │
   └──────────┘ └──────────┘
         ↓         ↓
   ┌────────────────────┐
   │  Redis (Queue)     │
   └────────────────────┘
         ↓         ↓
   ┌──────────┐ ┌──────────┐
   │ Worker 1 │ │ Worker 2 │
   └──────────┘ └──────────┘
         ↓         ↓
       GPU Pool
```

## 故障恢复

### 后端服务异常
- LoadBalancer 自动摘除故障实例
- 30s 后自动重试健康检查
- 任务标记为 FAILED，用户可重新提交

### Webserver 重启
- 内存队列丢失（未处理任务）
- SQLite 中的任务状态保留
- 重启后用户可重新提交 QUEUED 任务

### 数据库故障
- TaskManager 提供内存缓存降级
- 建议定期备份 SQLite

## 安全考虑

1. **速率限制**：建议添加 per-user 速率限制
2. **任务超时**：默认 120s，可根据实际情况调整
3. **队列容量**：防止恶意用户填满队列
4. **认证鉴权**：所有 API 需要 JWT token

## 后续优化方向

1. **任务优先级队列**：VIP 用户优先处理
2. **任务取消功能**：支持用户取消排队任务
3. **WebSocket 推送**：任务完成主动通知
4. **分布式追踪**：集成 OpenTelemetry
5. **Prometheus 监控**：导出 metrics
