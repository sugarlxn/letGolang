# Database Dictionary

**Last Updated:** 2026-01-25  
**Database Type:** SQLite  
**Database File:** `test.db`  
**Schema Source:** Runtime introspection + migrations/0001_init.sql

---

## Table: images

**Purpose:** User-uploaded image metadata and binary storage

| Column | Type | Not Null | Primary Key | Default | Constraints |
|--------|------|----------|-------------|---------|-------------|
| id | INTEGER | Yes | Yes | | Auto-increment |
| user_id | INTEGER | Yes | No | | FK → users(id) |
| image_data | BLOB | Yes | No | | Binary image payload |
| created_at | TIMESTAMP | No | No | CURRENT_TIMESTAMP | Upload timestamp |

**Foreign Keys:**
- `user_id` → `users.id` (NO ACTION, NO ACTION)

---

## Table: todos

**Purpose:** Task/todo item management scoped to users

| Column | Type | Not Null | Primary Key | Default | Constraints |
|--------|------|----------|-------------|---------|-------------|
| id | INTEGER | Yes | Yes | | Auto-increment |
| user_id | INTEGER | Yes | No | | FK → users(id) |
| title | TEXT | Yes | No | | Task description |
| completed | BOOLEAN | No | No | 0 | 0 = incomplete, 1 = complete |
| created_at | TIMESTAMP | No | No | CURRENT_TIMESTAMP | Task creation timestamp |

**Foreign Keys:**
- `user_id` → `users.id` (NO ACTION, NO ACTION)

---

## Table: users

**Purpose:** Central user identity and authentication store

| Column | Type | Not Null | Primary Key | Default | Constraints |
|--------|------|----------|-------------|---------|-------------|
| id | INTEGER | Yes | Yes | | Auto-increment |
| username | TEXT | Yes | No | | UNIQUE; login identifier |
| password | TEXT | Yes | No | | bcrypt-hashed (never plaintext) |
| phone | TEXT | No | No | | Optional phone number |
| email | TEXT | No | No | | Optional email address |
| created_at | TIMESTAMP | No | No | CURRENT_TIMESTAMP | Account creation timestamp |

**Unique Constraints:**
- `username` (enforced at database level)

---

## Data Dictionary Legend

- **Type:** SQLite data type (INTEGER, TEXT, BLOB, BOOLEAN, TIMESTAMP)
- **Not Null:** Column cannot store NULL values
- **Primary Key:** Unique identifier for the row (id column auto-incremented)
- **Default:** Default value assigned if no value provided
- **Constraints:** Uniqueness, foreign key references, or business rules

---

## Referential Integrity Map

```
users (root table)
├── images.user_id → users.id
└── todos.user_id → users.id
```

**Deletion Policy:** NO ACTION (prevents deletion of users with related images/todos)

---

## Schema Statistics

| Metric | Value |
|--------|-------|
| Total Tables | 3 |
| Total Columns | 16 |
| Primary Keys | 3 |
| Foreign Keys | 2 |
| Unique Constraints | 1 |
| TIMESTAMP columns | 3 |
| BLOB columns | 1 |

---

**This dictionary is auto-generated from schema introspection. Do not edit manually.**  
**For schema changes, create a new migration in `migrations/` directory.**
