## 使用 Go 在本地启动一个简单的 RESTful Web Server

本目录包含一个使用 Go 标准库 net/http 实现的示例 Web Server，提供基于内存的 Todo REST API，适合作为学习和扩展自己后端服务的模板。

### 1. 运行环境准备

- 已安装 Go（建议 Go 1.22+，至少 1.21）
- 当前工作目录为项目根目录：`/home/lxn/letGolang`

进入 webserver 目录：

```bash
cd webserver
```

### 2. 启动 Web Server

在 webserver 目录下执行：

```bash
# 直接运行
go run .

# 或先构建再运行
go build -o webserver
./webserver
```

默认监听地址为：

- http://localhost:8080

成功启动后，终端会打印类似信息：

```text
Starting webserver on :8080...
```

### 3. API 概览

当前示例实现了一个简单的 Todo REST API，使用内存存储：

- `GET    /health`              健康检查
- `GET    /todos`               获取所有 todo
- `GET    /todos/{id}`          根据 id 获取单个 todo
- `POST   /todos`               新增 todo
- `PUT    /todos/{id}`          更新 todo（标题或完成状态）
- `DELETE /todos/{id}`          删除 todo

Todo 结构体定义如下：

```go
type Todo struct {
		ID        int64  `json:"id"`
		Title     string `json:"title"`
		Completed bool   `json:"completed"`
}
```

> 注意：所有数据保存在内存中，重启服务后会丢失，仅用于演示 REST 风格的 API。

### 4. 常用请求示例（使用 curl）

#### 4.1 健康检查

```bash
curl http://localhost:8080/health
```

返回示例：

```json
{"status":"ok"}
```

#### 4.2 获取所有 Todo（GET /todos）

```bash
curl http://localhost:8080/todos
```

返回示例：

```json
[
	{"id":1,"title":"Learn Go","completed":false},
	{"id":2,"title":"Build a web server","completed":false},
	{"id":3,"title":"Write REST API","completed":true}
]
```

#### 4.3 新增 Todo（POST /todos）

```bash
curl -X POST http://localhost:8080/todos \
	-H "Content-Type: application/json" \
	-d '{"title": "Read Go doc"}'
```

可能的返回：

```json
{"id":4,"title":"Read Go doc","completed":false}
```

#### 4.4 获取单个 Todo（GET /todos/{id}）

```bash
curl http://localhost:8080/todos/1
```

返回示例：

```json
{"id":1,"title":"Learn Go","completed":false}
```

#### 4.5 更新 Todo（PUT /todos/{id}）

示例 1：只更新标题

```bash
curl -X PUT http://localhost:8080/todos/1 \
	-H "Content-Type: application/json" \
	-d '{"title": "Learn Go basics"}'
```

示例 2：只更新完成状态

```bash
curl -X PUT http://localhost:8080/todos/1 \
	-H "Content-Type: application/json" \
	-d '{"completed": true}'
```

示例 3：同时更新标题和状态

```bash
curl -X PUT http://localhost:8080/todos/1 \
	-H "Content-Type: application/json" \
	-d '{"title": "Learn Go","completed": true}'
```

#### 4.6 删除 Todo（DELETE /todos/{id}）

```bash
curl -X DELETE http://localhost:8080/todos/1 -v
```

成功时返回 HTTP 204 No Content，无 Body。

### 5. 如何基于此示例实现自己的 REST API

本示例重点演示：

1. 路由与 Handler 绑定
2. 解析请求 JSON
3. 返回 JSON 响应
4. 使用内存数据结构模拟持久化层

#### 5.1 定义自己的数据结构

比如你要实现一个简单的“文章”服务，可以定义：

```go
type Article struct {
		ID      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
}
```

然后参考 Todo 的实现，增加：

- 全局切片或 map 存储 Article
- `handleListArticles`、`handleCreateArticle` 等 Handler

#### 5.2 编写路由

本项目使用 Go1.22+ 支持的模式路由（`/todos/{id}`）：

```go
mux.HandleFunc("/todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		// 根据 Method 分发到不同的处理函数
})
```

如果你的 Go 版本不支持模式路由，可以使用第三方路由库（如 gorilla/mux、chi 等），或自己解析 URL Path。

#### 5.3 统一 JSON 返回

可以复用本项目中的工具函数：

```go
func writeJSON(w http.ResponseWriter, status int, data interface{})
func errorResponse(w http.ResponseWriter, status int, msg string)
```

它们帮助你统一设置 `Content-Type`、状态码以及错误返回格式。

### 6. 常见问题

1. 端口被占用
	 - 修改 main.go 中的 `addr := ":8080"`，例如改为 `":9090"`。

2. 无法使用 /todos/{id} 路由
	 - 请确认 Go 版本是否为 1.22+，或参考 5.2 节使用第三方路由库。

3. 数据为何重启后丢失？
	 - 示例使用内存切片存储数据，重启进程后内存会重置。若需要持久化，请替换为数据库（例如 MySQL、PostgreSQL、SQLite 等）。

---

你可以在此基础上继续扩展认证、数据库访问、中间件（日志、跨域等），逐步搭建完整的后端 REST API 服务。