## 使用 Go 在本地启动一个 RESTful Web Server

本目录包含一个使用 Go 标准库 net/http 和 SQLite 数据库实现的示例 Web Server，提供 Todo 和 User 管理的 REST API，适合作为学习和扩展自己后端服务的模板。

### 功能特性

- ✅ 用户管理（增删改查）
- ✅ Todo 管理（增删改查，支持按用户过滤）
- ✅ SQLite 数据库持久化存储
- ✅ **bcrypt 密码加密**
- ✅ **JWT 身份认证**
- ✅ **输入验证（用户名、邮箱、密码强度）**
- ✅ **结构化日志记录**
- ✅ Swagger API 文档
- ✅ RESTful 架构设计

### 1. 运行环境准备

- 已安装 Go（建议 Go 1.21+）
- 当前工作目录为项目根目录：`/home/lxn/letGolang`

进入 webserver 目录：

```bash
cd webserver
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 启动 Web Server

在 webserver 目录下执行：

```bash
# 直接运行
go run main.go

# 或先构建再运行
go build -o webserver main.go
./webserver
```

默认监听地址为：

- Web Server: http://localhost:8080
- Swagger 文档: http://localhost:8080/docs/

成功启动后，终端会打印类似信息：

```text
Starting webserver on :8080...
```

### 4. API 概览

#### 4.1 健康检查

- `GET /health` - 健康检查

#### 4.2 认证 API（无需 Token）

- `POST /register` - 注册新用户
- `POST /login` - 用户登录，获取 JWT token

#### 4.3 用户管理 API

- `GET    /users` - 获取所有用户
- `GET    /users/{id}` - 根据 id 获取单个用户
- `PUT    /users/{id}` - 更新用户信息
- `DELETE /users/{id}` - 删除用户

User 结构体定义：

```go
type User struct {
    ID        int64     `json:"id"`
    Username  string    `json:"username"`
    Password  string    `json:"password"`
    Phone     string    `json:"phone"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}
```

#### 4.4 Todo 管理 API（需要 JWT 认证）

- `GET    /todos` - 获取所有 todo（支持 `?user_id=xxx` 参数按用户过滤）
- `GET    /todos/{id}` - 根据 id 获取单个 todo
- `POST   /todos` - 新增 todo
- `PUT    /todos/{id}` - 更新 todo（标题或完成状态）
- `DELETE /todos/{id}` - 删除 todo

> **注意：** 所有 Todo 相关接口需要在请求头中携带 JWT token：`Authorization: Bearer <token>`

Todo 结构体定义：

```go
type Todo struct {
    ID        int64     `json:"id"`
    UserID    int64     `json:"user_id"`
    Title     string    `json:"title"`
    Completed bool      `json:"completed"`
    CreatedAt time.Time `json:"created_at"`
}
```

### 5. 常用请求示例（使用 curl）

#### 5.1 健康检查

```bash
curl http://localhost:8080/health
```

返回示例：

```json
{"status":"ok"}
```

#### 5.2 用户管理

**注册用户：**

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "secret123",
    "phone": "1234567890",
    "email": "john@example.com"
  }'
```

**用户登录（获取 JWT token）：**

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "secret123"
  }'
```

返回示例：

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "john_doe",
    "phone": "1234567890",
    "email": "john@example.com",
    "created_at": "2026-01-23T22:00:00Z"
  }
}
```

**获取所有用户：**

```bash
curl http://localhost:8080/users
```

**获取单个用户：**

```bash
curl http://localhost:8080/users/1
```

**更新用户：**

```bash
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newemail@example.com"
  }'
```

**删除用户：**

```bash
curl -X DELETE http://localhost:8080/users/1
```

#### 5.3 Todo 管理（需要 JWT Token）

**重要：** 在请求头中添加 `Authorization: Bearer <your_jwt_token>`

**创建 Todo：**

```bash
# 先获取 token
TOKEN=$(curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe","password":"secret123"}' \
  | jq -r '.token')

# 使用 token 创建 todo
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "user_id": 1,
    "title": "Learn Go programming"
  }'
```

**获取所有 Todo：**

```bash
curl http://localhost:8080/todos \
  -H "Authorization: Bearer $TOKEN"
```

**按用户 ID 过滤 Todo：**

```bash
curl "http://localhost:8080/todos?user_id=1" \
  -H "Authorization: Bearer $TOKEN"
```

**获取单个 Todo：**

```bash
curl http://localhost:8080/todos/1 \
  -H "Authorization: Bearer $TOKEN"
```

**更新 Todo：**

```bash
curl -X PUT http://localhost:8080/todos/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Learn Go and build APIs",
    "completed": true
  }'
```

**删除 Todo：**

```bash
curl -X DELETE http://localhost:8080/todos/1 \
  -H "Authorization: Bearer $TOKEN"
```


### 6. 数据库说明

本项目使用 SQLite 数据库进行持久化存储，数据库文件为 `test.db`。

#### 6.1 数据库表结构

**users 表：**

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    phone TEXT,
    email TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**todos 表：**

```sql
CREATE TABLE todos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    completed BOOLEAN DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

**images 表（预留）：**

```sql
CREATE TABLE images (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    image_data BLOB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

#### 6.2 数据库操作

查看数据库内容：

```bash
# 进入 SQLite 命令行
sqlite3 test.db

# 查看所有表
.tables

# 查看用户
SELECT * FROM users;

# 查看 todos
SELECT * FROM todos;

# 退出
.quit
```

### 7. Swagger API 文档

本项目集成了 Swagger/OpenAPI 文档，可以通过浏览器访问：

**访问地址：** http://localhost:8080/docs/

在 Swagger UI 中可以：
- 查看所有 API 接口
- 查看请求/响应参数
- 在线测试 API

#### 7.1 更新 Swagger 文档

修改代码后需要重新生成 Swagger 文档：

```bash
swag init
```

### 8. 项目结构

```
webserver/
├── main.go           # 主程序文件
├── main_test.go      # 测试文件
├── go.mod            # Go 模块定义
├── readme.md         # 项目文档
├── test.db           # SQLite 数据库文件
└── docs/             # Swagger 文档目录
    ├── docs.go
    ├── swagger.json
    └── swagger.yaml
```

### 9. 测试

运行单元测试：

```bash
go test -v
```

### 10. 技术栈

- **编程语言：** Go 1.24+
- **Web 框架：** 标准库 `net/http`
- **数据库：** SQLite3
- **数据库驱动：** `github.com/mattn/go-sqlite3`
- **密码加密：** bcrypt (`golang.org/x/crypto/bcrypt`)
- **身份认证：** JWT (`github.com/golang-jwt/jwt/v5`)
- **API 文档：** Swagger/OpenAPI
- **文档生成：** `github.com/swaggo/swag`

### 11. 安全特性

#### 11.1 密码安全

- 使用 bcrypt 算法对密码进行加密存储
- 密码至少 6 个字符
- 登录时验证加密密码
- API 响应中不返回密码哈希

#### 11.2 JWT 认证

- 登录成功后返回 JWT token
- Token 有效期 24 小时
- Todo 相关接口需要 Bearer Token 认证
- Token 包含用户 ID 和用户名信息

#### 11.3 输入验证

- **用户名：** 3-20 个字符，只允许字母、数字和下划线
- **邮箱：** 标准邮箱格式验证
- **密码：** 至少 6 个字符

#### 11.4 环境变量

支持通过环境变量配置 JWT 密钥：

```bash
export JWT_SECRET="your-super-secret-key"
./webserver
```

### 12. 注意事项

1. ✅ **密码安全：** 已使用 bcrypt 加密
2. ✅ **认证授权：** 已实现 JWT 认证
3. ✅ **输入验证：** 已添加用户名、邮箱、密码验证
4. ✅ **错误处理：** 已完善错误日志记录
5. ⚠️ **并发安全：** SQLite 在高并发场景下可能存在性能瓶颈，生产环境建议使用 PostgreSQL/MySQL
6. ⚠️ **HTTPS：** 生产环境应使用 HTTPS
7. ⚠️ **CORS：** 如需前端调用，请添加 CORS 中间件

### 13. 扩展建议

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

---

## 作者

学习项目 - Go Web Server 示例

## 更新日志

- v3.0 - 添加 bcrypt 密码加密、JWT 认证、输入验证、日志记录
- v2.0 - 添加 SQLite 数据库持久化，用户与 Todo 关联
- v1.0 - 初始版本，内存存储

