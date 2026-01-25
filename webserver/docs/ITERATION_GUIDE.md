# 数据库治理迭代指南

**指南版本:** v1.0  
**最后更新:** 2026-01-25  
**适用范围:** User Management API 后端项目

---

## 概述

本指南说明如何在未来迭代中维护和扩展数据库架构。所有操作遵循**迁移优先**的原则。

---

## 快速命令参考

### 初始设置（首次）

```bash
# 进入项目目录
cd /home/lxn/letGolang/webserver

# 初始化git仓库（如果尚未初始化）
git init

# 提交当前的迁移和文档
git add migrations/ docs/
git commit -m "feat: Establish database governance with migrations

- Create migrations/0001_init.sql with full schema annotations
- Generate database dictionary (db_overview.md)
- Create governance procedures guide (db_schema.md)
- Add audit trail and validation reports"
```

---

## 场景1：添加新表

### 命令步骤

```bash
# 1. 创建新的迁移文件
touch /home/lxn/letGolang/webserver/migrations/0002_add_user_sessions.sql

# 2. 编辑迁移文件（使用编辑器打开）
# vi/nano/code migrations/0002_add_user_sessions.sql
```

### SQL模板

```sql
-- 0002_add_user_sessions.sql
-- Migration: Add user_sessions table for JWT token tracking
-- Created: 2026-01-25
-- Purpose: Track active user sessions and token validity
-- Author: [Your Name]

-- ========================================
-- UP: Apply the schema
-- ========================================

-- user_sessions table: JWT token and session management
CREATE TABLE IF NOT EXISTS user_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,    -- Unique session identifier
    user_id INTEGER NOT NULL,                -- Foreign key to users(id); session ownership
    token TEXT NOT NULL UNIQUE,              -- JWT token value; must be unique
    refresh_token TEXT,                      -- Optional refresh token for token rotation
    expires_at TIMESTAMP NOT NULL,           -- Token expiration timestamp
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  -- Session creation time
    FOREIGN KEY (user_id) REFERENCES users(id)      -- Enforce referential integrity
);

-- ========================================
-- DOWN: Rollback the schema
-- ========================================
-- DROP TABLE IF EXISTS user_sessions;
```

### 验证和应用

```bash
# 3. 测试UP部分（本地验证）
cd /home/lxn/letGolang/webserver
sqlite3 test.db < migrations/0002_add_user_sessions.sql

# 4. 验证表已创建
sqlite3 test.db ".tables"
# 应该显示: images  sessions  todos  users

# 5. 检查表结构
sqlite3 test.db "PRAGMA table_info(user_sessions);"

# 6. 测试DOWN部分（回滚验证）
sqlite3 test.db "DROP TABLE IF EXISTS user_sessions;"
sqlite3 test.db ".tables"
# 应该回到: images  todos  users
```

### 代码审查和提交

```bash
# 7. 查看文件差异
git diff migrations/0002_add_user_sessions.sql

# 8. 重新应用迁移（准备提交）
sqlite3 test.db < migrations/0002_add_user_sessions.sql

# 9. 提交代码
git add migrations/0002_add_user_sessions.sql
git commit -m "feat: Add user_sessions table for JWT session tracking

- New table: user_sessions
- Tracks active JWT tokens and refresh tokens
- Links to users via foreign key
- Includes expiration timestamp for session lifecycle"

# 10. 推送到远程仓库
git push origin main
```

---

## 场景2：修改现有列（添加字段）

### 命令步骤

```bash
# 1. 创建新的迁移文件
touch /home/lxn/letGolang/webserver/migrations/0003_add_user_profile.sql
```

### SQL模板

```sql
-- 0003_add_user_profile.sql
-- Migration: Add profile fields to users table
-- Created: 2026-01-25
-- Purpose: Support user profile information
-- Author: [Your Name]

-- ========================================
-- UP: Apply the schema
-- ========================================

-- Add new columns to users table
ALTER TABLE users ADD COLUMN bio TEXT;                      -- User biography
ALTER TABLE users ADD COLUMN avatar_url TEXT;               -- URL to user avatar image
ALTER TABLE users ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;  -- Last update timestamp

-- ========================================
-- DOWN: Rollback the schema
-- ========================================
-- SQLite doesn't support DROP COLUMN easily, so we use a workaround:
-- Create temporary table with original columns
-- CREATE TABLE users_backup AS SELECT id, username, password, phone, email, created_at FROM users;
-- DROP TABLE users;
-- CREATE TABLE users (
--     id INTEGER PRIMARY KEY AUTOINCREMENT,
--     username TEXT NOT NULL UNIQUE,
--     password TEXT NOT NULL,
--     phone TEXT,
--     email TEXT,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );
-- INSERT INTO users SELECT * FROM users_backup;
-- DROP TABLE users_backup;
```

### 验证和应用

```bash
# 2. 测试迁移
cd /home/lxn/letGolang/webserver
sqlite3 test.db < migrations/0003_add_user_profile.sql

# 3. 验证新列
sqlite3 test.db "PRAGMA table_info(users);"
# 应该显示新增的列: bio, avatar_url, updated_at

# 4. 验证数据完整性
sqlite3 test.db "SELECT COUNT(*) FROM users;"
# 应该显示原有数据数量

# 5. 提交
git add migrations/0003_add_user_profile.sql
git commit -m "feat: Add user profile fields to users table

- Add bio (TEXT) for user biography
- Add avatar_url (TEXT) for profile image
- Add updated_at (TIMESTAMP) for modification tracking"
```

---

## 场景3：添加索引优化查询性能

### 命令步骤

```bash
# 1. 创建新的迁移文件
touch /home/lxn/letGolang/webserver/migrations/0004_add_indexes.sql
```

### SQL模板

```sql
-- 0004_add_indexes.sql
-- Migration: Add indexes for query performance optimization
-- Created: 2026-01-25
-- Purpose: Improve query performance on frequently filtered columns
-- Author: [Your Name]

-- ========================================
-- UP: Create indexes
-- ========================================

-- Index for user login queries (username lookups)
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Indexes for foreign key lookups (common WHERE conditions)
CREATE INDEX IF NOT EXISTS idx_images_user_id ON images(user_id);
CREATE INDEX IF NOT EXISTS idx_todos_user_id ON todos(user_id);

-- Composite index for common query pattern: todos by user and date
CREATE INDEX IF NOT EXISTS idx_todos_user_date ON todos(user_id, created_at DESC);

-- ========================================
-- DOWN: Drop indexes
-- ========================================
-- DROP INDEX IF EXISTS idx_users_username;
-- DROP INDEX IF EXISTS idx_images_user_id;
-- DROP INDEX IF EXISTS idx_todos_user_id;
-- DROP INDEX IF EXISTS idx_todos_user_date;
```

### 验证和应用

```bash
# 2. 测试迁移
cd /home/lxn/letGolang/webserver
sqlite3 test.db < migrations/0004_add_indexes.sql

# 3. 验证索引已创建
sqlite3 test.db ".indices"
# 应该显示: idx_images_user_id  idx_todos_user_date  idx_todos_user_id  idx_users_username

# 4. 提交
git add migrations/0004_add_indexes.sql
git commit -m "perf: Add indexes for query optimization

- idx_users_username: Improve user login lookups
- idx_images_user_id: Optimize image queries by owner
- idx_todos_user_id: Optimize todo queries by owner
- idx_todos_user_date: Optimize date-range queries"
```

---

## 场景4：添加约束和验证

### 命令步骤

```bash
# 1. 创建新的迁移文件
touch /home/lxn/letGolang/webserver/migrations/0005_add_constraints.sql
```

### SQL模板

```sql
-- 0005_add_constraints.sql
-- Migration: Add business logic constraints
-- Created: 2026-01-25
-- Purpose: Enforce data validation at database level
-- Author: [Your Name]

-- ========================================
-- UP: Add constraints via new tables
-- ========================================

-- SQLite doesn't support ALTER TABLE ADD CONSTRAINT easily,
-- so we use CHECK constraints with new table creation

CREATE TABLE IF NOT EXISTS users_v2 (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE CHECK (LENGTH(username) >= 3 AND LENGTH(username) <= 50),
    password TEXT NOT NULL CHECK (LENGTH(password) >= 8),
    phone TEXT,
    email TEXT UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Migrate data from old table to new table
INSERT INTO users_v2 SELECT * FROM users;

-- Drop old table and rename new one
DROP TABLE users;
ALTER TABLE users_v2 RENAME TO users;

-- ========================================
-- DOWN: Revert to original table
-- ========================================
-- (Implementation depends on before/after state)
```

### 验证和应用

```bash
# 2. 测试迁移
cd /home/lxn/letGolang/webserver
sqlite3 test.db < migrations/0005_add_constraints.sql

# 3. 测试约束是否生效
sqlite3 test.db "INSERT INTO users (username, password) VALUES ('ab', 'short');"
# 应该失败：username太短，password太短

sqlite3 test.db "INSERT INTO users (username, password) VALUES ('valid_user', 'validpassword123');"
# 应该成功

# 4. 提交
git add migrations/0005_add_constraints.sql
git commit -m "feat: Add validation constraints to users table

- Username: 3-50 characters required
- Password: Minimum 8 characters required
- Email: Unique constraint (if provided)"
```

---

## 场景5：执行数据迁移（添加/修改数据）

### 命令步骤

```bash
# 1. 创建新的迁移文件
touch /home/lxn/letGolang/webserver/migrations/0006_populate_audit_columns.sql
```

### SQL模板

```sql
-- 0006_populate_audit_columns.sql
-- Migration: Populate audit columns with existing data
-- Created: 2026-01-25
-- Purpose: Initialize updated_at column for all existing users
-- Author: [Your Name]

-- ========================================
-- UP: Populate audit columns
-- ========================================

-- Set updated_at to created_at for all existing users
UPDATE users SET updated_at = created_at WHERE updated_at IS NULL;

-- ========================================
-- DOWN: Reset audit columns
-- ========================================
-- UPDATE users SET updated_at = NULL;
```

### 验证和应用

```bash
# 2. 测试迁移
cd /home/lxn/letGolang/webserver
sqlite3 test.db < migrations/0006_populate_audit_columns.sql

# 3. 验证数据已更新
sqlite3 test.db "SELECT id, username, created_at, updated_at FROM users LIMIT 5;"

# 4. 提交
git add migrations/0006_populate_audit_columns.sql
git commit -m "data: Populate updated_at audit column

- Initialize updated_at with created_at for all users
- Prepares for future audit trail tracking"
```

---

## 定期维护流程

### 周期：每次部署前

```bash
# 1. 检查所有未应用的迁移
ls -la /home/lxn/letGolang/webserver/migrations/

# 2. 备份数据库
cp /home/lxn/letGolang/webserver/test.db /home/lxn/letGolang/webserver/test.db.backup

# 3. 应用所有待处理的迁移
for file in /home/lxn/letGolang/webserver/migrations/*.sql; do
    echo "Applying: $file"
    sqlite3 /home/lxn/letGolang/webserver/test.db < "$file"
done

# 4. 验证所有表
sqlite3 /home/lxn/letGolang/webserver/test.db ".tables"

# 5. 生成最新的数据库文档
# （可选）重新运行文档生成脚本
```

### 周期：每月审计

```bash
# 1. 验证数据库状态与迁移一致
cd /home/lxn/letGolang/webserver

# 检查所有表
echo "Tables in database:"
sqlite3 test.db "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name;"

# 检查所有索引
echo "Indexes in database:"
sqlite3 test.db ".indices"

# 检查外键
echo "Foreign keys:"
sqlite3 test.db "PRAGMA foreign_key_list(images);"
sqlite3 test.db "PRAGMA foreign_key_list(todos);"

# 2. 更新文档
# 编辑 docs/db_overview.md（如果架构改变）
# 编辑 docs/db_schema.md（如果流程改变）

# 3. 生成审计报告
git log --oneline migrations/
```

---

## 完整的迭代工作流示例

### 场景：添加新的评论功能

```bash
# ===== STEP 1: 计划 =====
# 需求：添加评论表，支持对todos的评论
# 用户故事：用户可以给任何todo添加评论

# ===== STEP 2: 创建迁移 =====
touch /home/lxn/letGolang/webserver/migrations/0002_add_comments.sql

# ===== STEP 3: 编写SQL迁移 =====
cat > /home/lxn/letGolang/webserver/migrations/0002_add_comments.sql << 'EOF'
-- 0002_add_comments.sql
-- Migration: Add comments table for todo discussions
-- Created: 2026-01-25
-- Purpose: Allow users to comment on todos
-- Author: Development Team

-- ========================================
-- UP: Create comments table
-- ========================================

CREATE TABLE IF NOT EXISTS comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,              -- Unique comment identifier
    todo_id INTEGER NOT NULL,                          -- Foreign key to todos(id); comment target
    user_id INTEGER NOT NULL,                          -- Foreign key to users(id); comment author
    content TEXT NOT NULL,                             -- Comment text content
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,    -- Comment creation timestamp
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,    -- Last edit timestamp
    FOREIGN KEY (todo_id) REFERENCES todos(id),        -- Link to todo
    FOREIGN KEY (user_id) REFERENCES users(id)         -- Link to user
);

-- Add index for common query: comments for a todo
CREATE INDEX IF NOT EXISTS idx_comments_todo_id ON comments(todo_id);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);

-- ========================================
-- DOWN: Rollback
-- ========================================
-- DROP TABLE IF EXISTS comments;
EOF

# ===== STEP 4: 本地测试 =====
cd /home/lxn/letGolang/webserver
sqlite3 test.db < migrations/0002_add_comments.sql

# 验证表创建
sqlite3 test.db "PRAGMA table_info(comments);"

# 测试插入
sqlite3 test.db "INSERT INTO comments (todo_id, user_id, content) VALUES (1, 1, 'Great task!');"

# 测试查询
sqlite3 test.db "SELECT * FROM comments;"

# 测试回滚
sqlite3 test.db "DROP TABLE IF EXISTS comments; DROP INDEX IF EXISTS idx_comments_todo_id; DROP INDEX IF EXISTS idx_comments_user_id;"

# 重新应用迁移
sqlite3 test.db < migrations/0002_add_comments.sql

# ===== STEP 5: Go 模型更新 =====
# 在 main.go 中添加 Comment 结构体（可选，因为迁移是真实来源）
# type Comment struct {
#     ID        int64     `json:"id"`
#     TodoID    int64     `json:"todo_id"`
#     UserID    int64     `json:"user_id"`
#     Content   string    `json:"content"`
#     CreatedAt time.Time `json:"created_at"`
#     UpdatedAt time.Time `json:"updated_at"`
# }

# ===== STEP 6: 代码审查 =====
git diff migrations/0002_add_comments.sql

# ===== STEP 7: 提交 =====
git add migrations/0002_add_comments.sql
git commit -m "feat: Add comments table for todo discussions

- New table: comments (links todos and users)
- Tracks comment content with timestamps
- Includes indexes for query optimization
- Supports many-to-many comment pattern

Schema changes:
- CREATE TABLE comments
- CREATE INDEX idx_comments_todo_id
- CREATE INDEX idx_comments_user_id

Allows users to collaborate on todos via comments."

# ===== STEP 8: 更新文档 =====
# 编辑 docs/db_overview.md，在"Table: todos"之后添加comments
# 编辑 docs/db_schema.md，在表清单中添加comments

# ===== STEP 9: 推送和部署 =====
git push origin main

# 在生产环境部署时，确保迁移在应用启动前执行
# 创建部署脚本或在CI/CD中集成
```

---

## 常见任务命令速查

### 查看数据库当前状态

```bash
cd /home/lxn/letGolang/webserver

# 所有表
sqlite3 test.db ".tables"

# 所有索引
sqlite3 test.db ".indices"

# 某表结构
sqlite3 test.db "PRAGMA table_info(users);"

# 外键关系
sqlite3 test.db "PRAGMA foreign_key_list(images);"

# 数据库大小
ls -lh test.db

# 表行数
sqlite3 test.db "SELECT name, COUNT(*) as count FROM sqlite_master WHERE type='table' GROUP BY name;"
```

### 安全地测试迁移

```bash
cd /home/lxn/letGolang/webserver

# 1. 备份原始数据库
cp test.db test.db.backup

# 2. 在副本上测试新迁移
cp test.db test_test.db
sqlite3 test_test.db < migrations/0002_add_comments.sql

# 3. 验证测试副本
sqlite3 test_test.db ".tables"

# 4. 清理测试副本
rm test_test.db

# 5. 在生产数据库上应用
sqlite3 test.db < migrations/0002_add_comments.sql

# 6. 保存备份
# 保留 test.db.backup 作为检查点
```

### 查看迁移历史

```bash
cd /home/lxn/letGolang/webserver

# 列出所有迁移
ls -1 migrations/*.sql | sort

# 查看迁移提交历史
git log --oneline -- migrations/

# 查看具体迁移的内容
cat migrations/0001_init.sql
```

### 团队协作

```bash
cd /home/lxn/letGolang/webserver

# 拉取最新迁移
git pull origin main

# 应用最新迁移到本地
for file in $(git diff --name-only HEAD origin/main -- migrations/); do
    sqlite3 test.db < "$file"
done

# 查看待提交的迁移
git status migrations/

# 查看迁移差异
git diff migrations/
```

---

## 故障排查

### 问题1：迁移应用失败

```bash
# 检查迁移文件语法
sqlite3 :memory: < migrations/0002_add_comments.sql

# 如果失败，查看具体错误
sqlite3 test.db ".mode list"
sqlite3 test.db < migrations/0002_add_comments.sql 2>&1

# 回滚到上一个备份
cp test.db.backup test.db
```

### 问题2：忘记备份

```bash
# 从git历史恢复
git log --oneline

# 查看特定提交时的迁移状态
git show commit_hash:migrations/0001_init.sql
```

### 问题3：数据丢失

```bash
# 查看是否有备份
ls -la test.db*

# 从备份恢复
cp test.db.backup test.db

# 或者重新初始化所有迁移
rm test.db
for file in migrations/*.sql; do
    sqlite3 test.db < "$file"
done
```

---

## 检查清单

### 每次创建迁移时：

- [ ] 创建新的迁移文件：`000N_description.sql`
- [ ] 编写详细的注释说明目的
- [ ] 包含完整的UP部分（CREATE/ALTER语句）
- [ ] 包含完整的DOWN部分（DROP语句或回滚逻辑）
- [ ] 所有表和列都有注释说明业务含义
- [ ] 本地测试UP和DOWN部分
- [ ] 验证数据完整性
- [ ] 检查外键关系

### 提交前：

- [ ] 运行 `git diff` 检查改动
- [ ] 清晰的提交信息（feat/fix/perf/data）
- [ ] 提交仅包含迁移文件（不包含数据库文件）
- [ ] 可选：更新文档（db_overview.md, db_schema.md）

### 部署前：

- [ ] 备份生产数据库
- [ ] 在测试环境验证迁移
- [ ] 验证回滚能否正确执行
- [ ] 通知团队部署计划
- [ ] 保存迁移前的数据库快照

---

## 参考资源

- **迁移设计指南:** [docs/db_schema.md](docs/db_schema.md)
- **数据库字典:** [docs/db_overview.md](docs/db_overview.md)
- **完整报告:** [docs/GOVERNANCE_REPORT.md](docs/GOVERNANCE_REPORT.md)
- **初始迁移:** `migrations/0001_init.sql`

---

**最后提醒：** 始终遵循"迁移优先"原则。所有架构更改必须通过版本控制的迁移文件执行。
