# 测试数据库初始化最佳实践

## 问题背景
在单元测试中，传统做法是在测试文件中硬编码 SQL 创建表语句。这种方式存在以下问题：
- ❌ 数据库 schema 变更时需要手动同步测试代码
- ❌ 生产环境和测试环境的 schema 可能不一致
- ❌ 维护成本高，容易出错
- ❌ 新增字段/表时需要修改多处代码

## 🏆 大厂解决方案对比

### 方案 1：复用 Migration 文件 ⭐⭐⭐⭐⭐ (推荐)

**适用公司**：Google, Uber, 字节跳动, Stripe

**原理**：
```
生产环境: migrations/*.sql → 生产数据库
测试环境: migrations/*.sql → 测试数据库
```

**优点**：
- ✅ 生产和测试使用相同 schema，100% 一致
- ✅ 新增 migration 文件，测试自动获得最新 schema
- ✅ 无需维护重复的 SQL 代码
- ✅ 支持复杂 migration（索引、约束、数据迁移）

**实现**：
```go
// testutil/db.go
func SetupTestDB(t *testing.T) *sql.DB {
    db := createTempDB(t)
    runMigrations(db, t)  // 自动读取并执行所有 migrations/*.sql
    return db
}

// main_test.go
func TestXxx(t *testing.T) {
    setupTestDB(t)  // 一行代码搞定
    // ... 测试逻辑
}
```

**目录结构**：
```
webserver/
├── migrations/
│   ├── 0001_init.sql          # 初始 schema
│   ├── 0002_add_table_prompt.sql
│   └── 0003_add_index.sql     # 新增 migration，测试自动使用
├── testutil/
│   └── db.go                  # migration 执行工具
└── main_test.go               # 测试文件
```

---

### 方案 2：ORM AutoMigrate ⭐⭐⭐

**适用公司**：中小型创业公司

**原理**：使用 GORM 等 ORM 的 AutoMigrate 功能

**优点**：
- ✅ 简单易用
- ✅ 自动同步 struct 定义

**缺点**：
- ❌ 无法处理复杂 migration（重命名列、数据迁移）
- ❌ 无法精确控制索引和约束
- ❌ 不适合已有项目（本项目使用原生 SQL）

**实现**（仅供参考，本项目不适用）：
```go
// 需要使用 GORM
db.AutoMigrate(&User{}, &Todo{}, &Image{})
```

---

### 方案 3：Testcontainers ⭐⭐⭐⭐

**适用公司**：Spotify, Netflix（需要真实数据库环境）

**原理**：在 Docker 容器中运行真实数据库

**优点**：
- ✅ 完全隔离的测试环境
- ✅ 支持 PostgreSQL, MySQL 等真实数据库
- ✅ 与生产环境完全一致

**缺点**：
- ❌ 需要 Docker 环境
- ❌ 测试速度较慢（容器启动开销）
- ❌ 本地开发环境配置复杂

**实现**（仅供参考）：
```go
// 需要 github.com/testcontainers/testcontainers-go
container, _ := postgres.RunContainer(ctx,
    testcontainers.WithImage("postgres:15"),
)
db = connectToContainer(container)
runMigrations(db)
```

---

### 方案 4：共享 Schema 文件 ⭐⭐

**原理**：将 schema 定义放在单独的 .sql 文件中

**优点**：
- ✅ 生产和测试共享 schema

**缺点**：
- ❌ 无法处理渐进式 migration
- ❌ 数据库版本管理困难
- ❌ 不支持回滚

---

## 本项目采用方案

✅ **方案 1：复用 Migration 文件**

### 使用方式

#### 1. 在测试中使用
```go
func TestSomething(t *testing.T) {
    setupTestDB(t)  // 自动运行所有 migrations
    
    // 你的测试逻辑
    // ...
}
```

#### 2. 添加新的数据库变更
只需在 `migrations/` 目录添加新文件：

```bash
# 创建新 migration
cat > migrations/0003_add_user_avatar.sql << 'EOF'
-- Add avatar column to users table
ALTER TABLE users ADD COLUMN avatar TEXT;
EOF
```

**无需修改测试代码**，下次运行测试时会自动应用！

#### 3. 可选：添加测试种子数据
```go
func TestWithSeedData(t *testing.T) {
    setupTestDB(t)
    testutil.SeedTestData(t, db)  // 插入测试数据
    
    // 测试逻辑
}
```

---

## Migration 文件编写规范

### 文件命名
```
migrations/
├── 0001_init.sql              # 四位数字 + 描述
├── 0002_add_table_prompt.sql
├── 0003_add_index.sql
└── 0004_alter_users.sql
```

### 文件内容结构
```sql
-- 0003_add_index.sql
-- Migration: Add performance indexes
-- Created: 2026-01-25
-- Description: Add indexes for frequently queried columns

-- ========================================
-- UP: Apply the schema
-- ========================================

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_todos_user_id ON todos(user_id);

-- ========================================
-- DOWN: Rollback the schema
-- ========================================

-- DROP INDEX IF EXISTS idx_todos_user_id;
-- DROP INDEX IF EXISTS idx_users_username;
```

**注意**：
- ✅ UP 部分：实际执行的 SQL（会被测试执行）
- ⚠️ DOWN 部分：注释掉（仅作文档，不执行）
- ✅ 使用 `IF NOT EXISTS` / `IF EXISTS` 确保幂等性

---

## 测试执行流程

```
┌─────────────────┐
│  运行 go test   │
└────────┬────────┘
         │
         ▼
┌─────────────────────┐
│  setupTestDB(t)     │  ← 每个测试都会调用
└────────┬────────────┘
         │
         ▼
┌──────────────────────────┐
│  testutil.SetupTestDB(t) │
└────────┬─────────────────┘
         │
         ├─► 创建临时数据库 (t.TempDir())
         │
         ├─► 读取 migrations/*.sql
         │   ├── 0001_init.sql
         │   └── 0002_add_table_prompt.sql
         │
         ├─► 按顺序执行每个 migration
         │   └── 清理注释、跳过 DOWN 部分
         │
         └─► 返回初始化完成的 DB
                  │
                  ▼
            ┌──────────────┐
            │  运行测试逻辑  │
            └──────────────┘
```

---

## 性能优化

### 当前实现
每个测试创建独立的临时数据库（完全隔离）

### 优化方案（可选）
```go
// 如果 migration 很多，可以考虑：
// 1. 使用 in-memory SQLite
db, _ := sql.Open("sqlite3", ":memory:")

// 2. 并行执行测试
func TestParallel(t *testing.T) {
    t.Parallel()  // 并行运行
    setupTestDB(t)
}
```

---

## 常见问题

### Q: Migration 执行失败怎么办？
```bash
# 检查 migration 文件语法
sqlite3 test.db < migrations/0001_init.sql

# 查看测试日志
go test -v  # 会显示每个 migration 的执行情况
```

### Q: 如何测试 migration 本身？
```go
func TestMigrations(t *testing.T) {
    db := testutil.SetupTestDB(t)
    
    // 验证表存在
    var count int
    err := db.QueryRow(`
        SELECT COUNT(*) FROM sqlite_master 
        WHERE type='table' AND name='users'
    `).Scan(&count)
    
    if count != 1 {
        t.Fatal("users table not created")
    }
}
```

### Q: 生产环境如何运行 migration？
```go
// main.go 中添加
func runMigrations(db *sql.DB) error {
    files, _ := os.ReadDir("./migrations")
    for _, file := range files {
        sql, _ := os.ReadFile("./migrations/" + file.Name())
        db.Exec(string(sql))
    }
    return nil
}

func main() {
    db := initDB()
    runMigrations(db)  // 启动时自动执行
}
```

---

## 总结

| 方案 | 一致性 | 易用性 | 灵活性 | 推荐度 |
|------|--------|--------|--------|---------|
| Migration 文件 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ✅ **推荐** |
| ORM AutoMigrate | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ | 适合新项目 |
| Testcontainers | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐ | 适合大型项目 |
| 共享 Schema | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐ | 不推荐 |

**本项目选择方案 1** 的原因：
1. ✅ 已有 migrations 目录和文件
2. ✅ 使用原生 SQL（不适合 ORM）
3. ✅ SQLite 适合快速测试（不需要容器）
4. ✅ 符合大厂最佳实践

---

## 参考资料

- [Google Testing Blog - Test Fixtures](https://testing.googleblog.com/)
- [Uber Go Style Guide - Testing](https://github.com/uber-go/guide)
- [Database Migration Best Practices](https://www.liquibase.org/get-started/best-practices)
