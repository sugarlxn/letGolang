# 异步任务处理架构 - 方案对比与分析

## 问题背景

**核心瓶颈：**
- 语音转文字：~100ms（毫秒级，快速）
- 文生图：~10-15s（慢速，占比90%）
- 并发用户：多用户同时请求
- 单实例吞吐量：~4-6 req/min（10s/次）

## 备选方案对比

### 方案 1：同步阻塞（原始方案）
```
用户请求 → 等待 10s+ → 返回结果
```

**优点：**
- 实现最简单
- 无额外依赖

**缺点：** ❌
- HTTP 超时（通常 30s）
- 用户体验差
- 服务器资源耗尽（连接堆积）
- 无法扩展

**吞吐量：** 3-6 req/min (单 server)

---

### 方案 2：线程池 + 队列（推荐方案 ✓）
```
用户请求 → 入队 → 返回 task_id → 轮询查询
         ↓
    Worker Pool（2-4 threads）
         ↓
    Load Balancer
         ↓
    Python Backend
```

**实现方式（当前使用）：**
- 使用 Go channel 作为队列
- 固定数量 goroutine worker
- LoadBalancer 轮询分配请求
- SQLite 持久化任务状态

**优点：** ✓
- 用户无需等待，立即返回 task_id
- Worker 数量可控，防止资源耗尽
- 支持多实例扩展
- 易于监控和告警
- 内存消耗低，毫秒级响应延迟

**缺点：**
- 需要轮询查询状态
- 内存队列重启丢失（可选持久化）

**吞吐量：** 24-30 req/min（4 workers × 3 instances）

**代码复杂度：** 中等

---

### 方案 3：消息队列（RabbitMQ/Redis）
```
用户请求 → Redis Queue → 返回 task_id
         ↓
    Worker Pool
         ↓
    Python Backend
         ↓
    WebSocket/Server Push
```

**优点：**
- 自动持久化，重启不丢失
- 支持分布式部署
- 消息可靠性高
- 支持多种 publish/subscribe 模式

**缺点：** ❌
- 增加运维复杂度（需要部署 Redis/RabbitMQ）
- 额外依赖和性能开销
- 对本项目来说是过度设计

**何时选择：**
- 多机器/微服务架构
- 需要消息可靠性和 exactly-once 语义
- 超大吞吐量（>1000 req/min）

---

### 方案 4：Job Queue + 定时任务（Apache Airflow）
```
用户请求 → Job Queue
    ↓
Airflow DAG（定向无环图）
    ↓
Executor（Local/Celery/Kubernetes）
    ↓
Python Backend
```

**优点：**
- 支持复杂工作流
- 内置任务重试和监控
- 适合 ETL 和数据处理流水线

**缺点：** ❌
- 严重过度设计
- 部署和学习成本高
- 性能开销大

**何时选择：** 不适合本场景

---

### 方案 5：异步函数库（Celery）
```
Python Celery Worker
    ↓
User Request → Celery Task
    ↓
Redis/RabbitMQ Broker
    ↓
Celery Worker
    ↓
Result Backend
```

**跨语言集成复杂**

**缺点：** ❌
- 主要为 Python，Go 集成复杂
- 通常用于 Python 工程

---

## 最终推荐方案

### 🏆 方案 2：Go Worker Pool + Channel + SQLite

**为什么选择？**
1. **完全符合需求**
   - 单机部署快速便利
   - 支持多实例扩展（无需中间件）
   - 毫秒级提交响应

2. **技术栈统一**
   - 全 Go 实现
   - 无外部依赖（SQLite 已使用）
   - 易于维护和调试

3. **性能优势**
   - 零 GC 压力（相比其他方案）
   - 原生 goroutine 轻量级
   - 毫秒级队列操作

4. **扩展路径清晰**
   ```
   第一阶段（当前）：Go Worker Pool（单机）
           ↓
   第二阶段：多实例 + LoadBalancer
           ↓
   第三阶段：Redis Queue（仅需替换队列实现）
   ```

---

## 性能指标对比

| 指标 | 同步阻塞 | Worker Pool | Redis Queue | 备注 |
|------|---------|------------|-----------|------|
| 提交延迟 | 10s+ | <5ms | 5-10ms | 用户感知延迟 |
| 吞吐量 | 3-6 req/min | 24-30 req/min | 50-100 req/min | 假设 10s 处理时间 |
| 并发连接 | 10-20 | 100+ | 1000+ | 服务器限制 |
| 内存占用 | 低 | 低 | 中等 | Worker Pool 友好 |
| 部署复杂度 | 最低 | 低 | 中等 | 无额外服务 |
| 可靠性 | 低 | 中等 | 高 | 队列持久化 |
| 扩展性 | ❌ | ✓✓ | ✓✓✓ | 多实例支持 |

---

## 实现细节

### 核心组件架构

```
┌─────────────────────────────────────────────────────┐
│                  HTTP Server (Go)                   │
│  ┌─────────────────────────────────────────────┐   │
│  │          AsyncAPIHandlers                   │   │
│  │  - POST /api/v1/image/async                 │   │
│  │  - GET /api/v1/tasks                        │   │
│  │  - POST /api/v1/speech/transcribe           │   │
│  └──────────────┬──────────────────────────────┘   │
│                 │                                   │
│                 ↓                                   │
│  ┌──────────────────────────────────────────────┐  │
│  │        TaskManager                           │  │
│  │  (SQLite + Memory Cache)                     │  │
│  └──────────────┬───────────────────────────────┘  │
│                 │                                   │
│                 ↓                                   │
│  ┌──────────────────────────────────────────────┐  │
│  │     WorkerPool (2-4 goroutines)              │  │
│  │                                              │  │
│  │  ┌──────────────────────────────────────┐   │  │
│  │  │  Channel Queue (capacity: 100-500)   │   │  │
│  │  └──────┬──────────────────────────┬───┘   │  │
│  │         │                          │        │  │
│  │    ┌────▼─────┐  ┌────────────┐   │        │  │
│  │    │Worker 1  │  │Worker N    │   │        │  │
│  │    └────┬─────┘  └────┬───────┘   │        │  │
│  │         │              │           │        │  │
│  │         ▼              ▼           │        │  │
│  │    ┌─────────────────────────┐    │        │  │
│  │    │  LoadBalancer           │    │        │  │
│  │    │  (Round-robin/Hash)     │    │        │  │
│  │    └─────────────────────────┘    │        │  │
│  └────────────────────────────────────┼────────┘  │
│                                       │            │
│                                       ↓            │
│                          HTTP Request to Backends  │
└─────────────────────────────────────────────────────┘
         ↓                    ↓                    ↓
    ┌─────────┐         ┌─────────┐         ┌─────────┐
    │Instance1│         │Instance2│         │Instance3│
    │ :8000   │         │ :8001   │         │ :8002   │
    └─────────┘         └─────────┘         └─────────┘
```

### 并发控制策略

**问题场景：**
```
100 个用户同时请求 → 10s 处理时间
= 需要 100 个并发连接
```

**解决方案：**
```
用户请求 (1ms 处理) → 任务队列 (Channel) 
                     ↓
                工作池控制并发
                ├─ Worker 1 处理实例 1
                ├─ Worker 2 处理实例 2  
                └─ Worker 3 处理实例 3
```

**关键参数：**
- `workerCount = num_instances × 2`
- `queueSize = expected_queue_depth`
- `timeout = 2 × avg_processing_time`

---

## 扩展演进路径

### Phase 1：单实例（当前）
```
Go Server → 1 GPU Instance
           吞吐：4-6 req/min
```

### Phase 2：多实例
```
Go Server → Instance 1 \
          → Instance 2 → 吞吐：24-30 req/min
          → Instance 3 /
```

### Phase 3：分布式队列（可选）
```
Go Server 1 \
Go Server 2 → Redis Queue → Worker Pool → Instances
Go Server 3 /
           吞吐：100+ req/min
```

---

## 代码完整性检查

✓ 已实现：
- [x] TaskManager - 任务持久化和查询
- [x] WorkerPool - 并发控制和处理
- [x] LoadBalancer - 实例路由和健康检查
- [x] AsyncAPIHandlers - RESTful API
- [x] Database migration - SQLite schema
- [x] System init - 初始化和优雅关闭
- [x] Error handling - 完整的错误处理
- [x] Logging - 结构化日志

---

## 监控和告警

### 关键指标
```
1. 队列长度：> 80% 容量 → 告警
2. 任务失败率：> 10% → 告警
3. 实例可用性：< 2 个 → 告警
4. 平均响应时间：> 20s → 调查
```

### Prometheus metrics（可选升级）
```go
queue_length_gauge
queue_capacity_gauge
task_processed_counter
task_failed_counter
instance_health_gauge
```

---

## 成本分析

| 方案 | 部署成本 | 运维成本 | 学习成本 | 总体评分 |
|------|---------|---------|---------|---------|
| 同步阻塞 | $$$ | $$ | $ | ⭐ |
| Worker Pool | $ | $$ | $$ | ⭐⭐⭐⭐⭐ |
| Redis Queue | $$ | $$$ | $$$ | ⭐⭐⭐ |
| Kubernetes | $$$ | $$$$ | $$$$ | ⭐⭐ |

---

## 总结

**当前方案（Worker Pool）完美匹配：**

✅ 需求：多用户异步请求处理
✅ 约束：单机/小集群部署  
✅ 工程：全 Go 技术栈
✅ 运维：最小化依赖
✅ 性能：充足的吞吐量和响应时间
✅ 扩展：清晰的升级路径

**下一步行动：**
1. 集成到 main.go
2. 执行数据库迁移
3. 配置环境变量
4. 部署多个 GPU 实例
5. 监控和优化
