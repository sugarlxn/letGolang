-- 0001_init.sql
-- Migration: Initialize core schema for User Management API
-- Created: 2026-01-25
-- Description: Creates foundational tables for users, images, and todos with foreign key constraints

-- ========================================
-- UP: Apply the schema
-- ========================================

-- users table: Central user identity and authentication
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,  -- Unique user identifier
    username TEXT NOT NULL UNIQUE,          -- Login identifier; must be unique across all users
    password TEXT NOT NULL,                 -- bcrypt-hashed password; never store plaintext
    phone TEXT,                             -- Optional user phone number
    email TEXT,                             -- Optional user email address
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP  -- Account creation timestamp
);

-- images table: User-uploaded image storage
CREATE TABLE IF NOT EXISTS images (
    id INTEGER PRIMARY KEY AUTOINCREMENT,  -- Unique image record identifier
    user_id INTEGER NOT NULL,              -- Foreign key to users(id); image ownership
    image_data BLOB NOT NULL,              -- Binary image data payload
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  -- Upload timestamp
    FOREIGN KEY (user_id) REFERENCES users(id)  -- Enforce referential integrity
);

-- todos table: Task management for users
CREATE TABLE IF NOT EXISTS todos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,  -- Unique todo identifier
    user_id INTEGER NOT NULL,              -- Foreign key to users(id); task ownership
    title TEXT NOT NULL,                   -- Todo item description
    completed BOOLEAN DEFAULT 0,           -- Completion flag (0=incomplete, 1=complete)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  -- Task creation timestamp
    FOREIGN KEY (user_id) REFERENCES users(id)  -- Enforce referential integrity
);

-- ========================================
-- DOWN: Rollback the schema
-- ========================================
-- Drop tables in reverse dependency order (todos before users, images before users)

-- DROP TABLE IF EXISTS todos;
-- DROP TABLE IF EXISTS images;
-- DROP TABLE IF EXISTS users;
