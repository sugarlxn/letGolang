# 快速命令参考卡

**保存此文件以便快速查阅**

---

## 🚀 常用命令（复制粘贴）

### 初始化（仅一次）

```bash
cd /home/lxn/letGolang/webserver
git add migrations/ docs/
git commit -m "feat: Establish database governance with migrations"
git push origin main
```

### 添加新表（最常用）

```bash
# 1. 创建迁移文件
touch migrations/0002_add_table_name.sql

# 2. 编辑文件并添加SQL
# 3. 本地测试
sqlite3 test.db < migrations/0002_add_table_name.sql

# 4. 验证
sqlite3 test.db ".tables"

# 5. 提交
git add migrations/0002_add_table_name.sql
git commit -m "feat: Add table_name table for [purpose]"
git push origin main
```

### 添加字段到现有表

```bash
touch migrations/0003_add_column.sql

# 使用 ALTER TABLE
# ALTER TABLE table_name ADD COLUMN column_name TYPE;

sqlite3 test.db < migrations/0003_add_column.sql
git add migrations/0003_add_column.sql
git commit -m "feat: Add column_name to table_name"
git push origin main
```

### 添加索引

```bash
touch migrations/0004_add_indexes.sql

# 使用 CREATE INDEX

sqlite3 test.db < migrations/0004_add_indexes.sql
git add migrations/0004_add_indexes.sql
git commit -m "perf: Add indexes for performance"
git push origin main
```

---

## 📋 SQL模板

### 创建表

```sql
-- 0002_add_feature.sql
-- Migration: Add [feature name]
-- Purpose: [Business reason]
-- Author: [Name]

-- ========================================
-- UP: Apply the schema
-- ========================================

CREATE TABLE IF NOT EXISTS table_name (
    id INTEGER PRIMARY KEY AUTOINCREMENT,           -- 唯一标识
    column1 TEXT NOT NULL,                          -- 描述
    column2 INTEGER,                                -- 描述
    user_id INTEGER NOT NULL,                       -- 外键引用
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    FOREIGN KEY (user_id) REFERENCES users(id)     -- 外键约束
);

-- ========================================
-- DOWN: Rollback the schema
-- ========================================
-- DROP TABLE IF EXISTS table_name;
```

### 修改表

```sql
-- 0003_modify_table.sql

-- ========================================
-- UP: Apply changes
-- ========================================

ALTER TABLE table_name ADD COLUMN new_column TEXT;

-- ========================================
-- DOWN: Rollback
-- ========================================
-- ALTER TABLE table_name DROP COLUMN new_column;
```

### 创建索引

```sql
-- 0004_add_index.sql

-- ========================================
-- UP: Create indexes
-- ========================================

CREATE INDEX IF NOT EXISTS idx_table_column ON table_name(column);

-- ========================================
-- DOWN: Drop indexes
-- ========================================
-- DROP INDEX IF EXISTS idx_table_column;
```

---

## ✅ 验证命令

```bash
# 所有表
sqlite3 /home/lxn/letGolang/webserver/test.db ".tables"

# 表结构
sqlite3 /home/lxn/letGolang/webserver/test.db "PRAGMA table_info(table_name);"

# 所有索引
sqlite3 /home/lxn/letGolang/webserver/test.db ".indices"

# 外键
sqlite3 /home/lxn/letGolang/webserver/test.db "PRAGMA foreign_key_list(table_name);"

# 行数
sqlite3 /home/lxn/letGolang/webserver/test.db "SELECT COUNT(*) FROM table_name;"
```

---

## 🔄 Git工作流

```bash
cd /home/lxn/letGolang/webserver

# 查看未跟踪的迁移
git status

# 查看迁移差异
git diff migrations/

# 查看提交历史
git log --oneline migrations/

# 查看特定迁移
git show HEAD:migrations/0001_init.sql
```

---

## 📊 备份和恢复

```bash
cd /home/lxn/letGolang/webserver

# 备份数据库
cp test.db test.db.$(date +%Y%m%d_%H%M%S).backup

# 恢复备份
cp test.db.backup test.db

# 重新初始化所有迁移
rm test.db
for file in migrations/*.sql; do
    sqlite3 test.db < "$file"
done
```

---

## 🎯 完整流程示例

```bash
# 场景：添加用户头像表

cd /home/lxn/letGolang/webserver

# 1. 创建迁移文件
touch migrations/0002_add_user_avatars.sql

# 2. 编辑迁移文件
cat > migrations/0002_add_user_avatars.sql << 'EOF'
-- 0002_add_user_avatars.sql
-- Migration: Add user avatars table
-- Purpose: Store user profile avatars
-- Author: Developer

-- UP
CREATE TABLE IF NOT EXISTS user_avatars (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    image_url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_user_avatars_user_id ON user_avatars(user_id);

-- DOWN
-- DROP TABLE IF EXISTS user_avatars;
-- DROP INDEX IF EXISTS idx_user_avatars_user_id;
EOF

# 3. 测试UP
sqlite3 test.db < migrations/0002_add_user_avatars.sql

# 4. 验证
sqlite3 test.db "PRAGMA table_info(user_avatars);"
sqlite3 test.db ".indices"

# 5. 测试DOWN（可选）
sqlite3 test.db "DROP TABLE IF EXISTS user_avatars;"
sqlite3 test.db ".tables"

# 6. 重新应用
sqlite3 test.db < migrations/0002_add_user_avatars.sql

# 7. 提交
git add migrations/0002_add_user_avatars.sql
git commit -m "feat: Add user_avatars table for profile images

- New table: user_avatars
- Stores avatar URL per user (1:1 relationship)
- Includes index for performance"

git push origin main
```

---

## 🚨 常见错误修复

### 错误1：迁移语法错误

```bash
# 测试迁移语法
sqlite3 :memory: < migrations/0002_add_table.sql

# 查看具体错误
sqlite3 test.db < migrations/0002_add_table.sql 2>&1
```

### 错误2：外键约束失败

```bash
# 检查外键
sqlite3 test.db "PRAGMA foreign_key_list(table_name);"

# 确保引用的表存在
sqlite3 test.db ".tables"
```

### 错误3：重复的表名

```bash
# 检查表是否已存在
sqlite3 test.db ".tables" | grep table_name

# 使用 CREATE TABLE IF NOT EXISTS
```

---

## 📝 提交信息模板

```
feat: Add table_name for feature_description

- New table: table_name
- Columns: col1, col2, col3
- Indexes: idx_table_col
- Relationships: FK to other_table

Reason: Business context
Impact: Users can now...
```

```
fix: Fix constraint in table_name

Issue: Description
Solution: Changed X to Y
Tested: Manual verification
```

```
perf: Add indexes for query optimization

- idx_table_col1: Improves WHERE queries
- idx_table_col2: Optimizes JOIN operations
```

```
data: Populate column_name with initial values

Context: Why we're populating
Change: Update all rows with default values
Result: All users have column_name set
```

---

## 📖 详细文档

- **完整迭代指南:** [docs/ITERATION_GUIDE.md](ITERATION_GUIDE.md)
- **治理规范:** [docs/db_schema.md](db_schema.md)
- **数据库字典:** [docs/db_overview.md](db_overview.md)
- **执行报告:** [docs/GOVERNANCE_REPORT.md](GOVERNANCE_REPORT.md)

---

**提示:** 将此文件收藏，以便快速查阅常用命令！
