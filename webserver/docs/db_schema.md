# Database Schema Governance

**Last Updated:** 2026-01-25  
**Status:** ✅ Actively Governed  
**Governance Mode:** Migration-first with runtime validation

---

## Core Philosophy

This project implements **strict database governance** following these principles:

### 1. Migrations as Single Source of Truth

- **All schema changes** must flow through the `migrations/` directory
- Each migration is **immutable once applied** to production
- Migrations are **versioned and ordered** for reproducibility
- No direct SQL execution bypasses version control

### 2. No Manual Schema Changes

- ❌ DO NOT modify schema via direct database tools
- ❌ DO NOT alter tables/columns without migration files
- ✅ DO create migrations for every schema change
- ✅ DO review migrations before deployment

### 3. Documentation as Living Artifact

- Database schema auto-documents from migrations
- [Database Dictionary](db_overview.md) reflects current schema
- Annotations in migrations explain business logic
- Drift between code and database is treated as a defect

---

## Table Inventory

### Core Tables (Created in migrations/0001_init.sql)

| Table | Purpose | Owner | Status |
|-------|---------|-------|--------|
| **users** | Central user identity and authentication | Backend API | ✅ Active |
| **images** | User-uploaded image metadata and storage | Backend API | ✅ Active |
| **todos** | Task/todo management scoped to users | Backend API | ✅ Active |

---

## Comprehensive Schema Reference

👉 **For detailed column definitions, constraints, and defaults, see:** [Database Dictionary](db_overview.md)

---

## How to Manage Schema

### ➊ Adding a New Table

**Step 1:** Create migration file
```bash
touch migrations/000N_add_table_name.sql
```

**Step 2:** Write the migration (up/down sections)
```sql
-- 000N_add_table_name.sql
-- Migration: Add [table name]
-- Purpose: [Business reason for new table]

-- UP: Create table with annotations
CREATE TABLE new_table (
    id INTEGER PRIMARY KEY AUTOINCREMENT,  -- Unique identifier
    -- ... other columns with comments
);

-- DOWN: Rollback
-- DROP TABLE IF EXISTS new_table;
```

**Step 3:** Document the change in comments with:
- Purpose/business rationale
- Relationships to existing tables
- Ownership and lifecycle rules

**Step 4:** Test UP and DOWN sections locally

**Step 5:** Submit for code review with migration file attached

### ➋ Modifying a Column

**Step 1:** Create a new migration (SQLite has limited ALTER TABLE support)
```bash
touch migrations/000N_modify_column.sql
```

**Step 2:** Use standard migration pattern
```sql
-- Migration pattern for column modification in SQLite
-- (involves temporary table, data copy, drop/recreate)

CREATE TABLE table_name_new ( ... modified schema ... );
INSERT INTO table_name_new SELECT ... FROM table_name;
DROP TABLE table_name;
ALTER TABLE table_name_new RENAME TO table_name;
```

**Step 3:** Include detailed comments explaining the change

### ➌ Adding an Index

**Step 1:** Create a new migration
```bash
touch migrations/000N_add_indexes.sql
```

**Step 2:** Add index creation with purpose comment
```sql
-- UP: Add index for query performance
CREATE INDEX idx_todos_user_id ON todos(user_id);
CREATE INDEX idx_images_user_id ON images(user_id);

-- DOWN: Drop indexes
-- DROP INDEX IF EXISTS idx_todos_user_id;
-- DROP INDEX IF EXISTS idx_images_user_id;
```

---

## Code Review Expectations

### For All Migrations

**Before approving, reviewers MUST verify:**

- [ ] Migration has descriptive UP and DOWN sections
- [ ] Column comments explain business meaning
- [ ] Foreign key constraints are correct
- [ ] Uniqueness constraints match business rules
- [ ] Migration tested locally (UP and DOWN)
- [ ] No hardcoded credentials or sensitive data
- [ ] Compatible with production database version

### Common Issues to Flag

| Issue | Action |
|-------|--------|
| Missing comments | Request clarification |
| Incomplete DOWN section | Reject (rollback must work) |
| Breaking changes without warning | Discuss deployment strategy |
| No referential integrity | Question design necessity |
| Mixing DDL and DML | Suggest separate migrations |

---

## Go Model Mapping

### Generated Structs from Schema

```go
// User maps to users table
type User struct {
    ID        int64     `json:"id"`         // users.id
    Username  string    `json:"username"`   // users.username (UNIQUE)
    Password  string    `json:"password"`   // users.password (bcrypt-hashed)
    Phone     string    `json:"phone"`      // users.phone (optional)
    Email     string    `json:"email"`      // users.email (optional)
    CreatedAt time.Time `json:"created_at"` // users.created_at
}

// Image maps to images table
type Image struct {
    ID        int64     `json:"id"`         // images.id
    UserID    int64     `json:"user_id"`    // images.user_id (FK to users.id)
    ImageData []byte    `json:"image_data"` // images.image_data (BLOB)
    CreatedAt time.Time `json:"created_at"` // images.created_at
}

// Todo maps to todos table
type Todo struct {
    ID        int64     `json:"id"`         // todos.id
    UserID    int64     `json:"user_id"`    // todos.user_id (FK to users.id)
    Title     string    `json:"title"`      // todos.title
    Completed bool      `json:"completed"`  // todos.completed (0/1)
    CreatedAt time.Time `json:"created_at"` // todos.created_at
}
```

**Rule:** Models are READ-ONLY documentation from schema. Update models only when schema changes.

---

## Migration Directory Structure

```
migrations/
├── 0001_init.sql              # Initial schema (users, images, todos)
├── 000N_add_feature.sql       # Example future migration
└── README.md                  # (Optional) Migration guidelines
```

**Naming Convention:** `NNNN_brief_description.sql`
- NNNN = 4-digit sequential version (0001, 0002, 0003, ...)
- brief_description = kebab-case English summary

---

## Common Questions

**Q: Can I modify the database directly in development?**  
A: Only for prototyping. Always convert to migrations before committing to git.

**Q: What if a migration has a bug in production?**  
A: Create a new corrective migration (UP/DOWN). Never rewrite history.

**Q: How do I drop a table?**  
A: Create a migration with `DROP TABLE` in UP and `CREATE TABLE` in DOWN.

**Q: Should I version migrations in git?**  
A: Yes. Migrations are code and must be reviewed, versioned, and backed up.

**Q: How are migrations applied in production?**  
A: Recommend: CI/CD pipeline applies migrations before deployment. Manual: DBA reviews and applies with approval.

---

## Governance Checklist

- ✅ All tables defined in migrations directory
- ✅ All columns documented with comments
- ✅ Foreign keys protect referential integrity
- ✅ UNIQUE constraints match business rules
- ✅ Default values make sense for empty rows
- ✅ Migrations have working rollback (DOWN) sections
- ✅ Go structs match schema structure
- ✅ Schema changes require migration files
- ✅ Migrations are peer-reviewed before production
- ✅ Documentation synced with actual schema

---

**Questions?** Contact the database owner or review recent migrations for patterns.  
**Last audit:** 2026-01-25 | **Next audit:** When migrations change