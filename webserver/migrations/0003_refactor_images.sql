-- 0003_refactor_images.sql
-- Migration: Refactor images table schema to support AI image generation with metadata
-- Created: 2026-01-25
-- Description: Extends images table with image format, dimensions (width/height), optional image path
--              for production file storage, and prompt_id foreign key. Uses table recreation
--              (SQLite standard pattern) to maintain clean schema format.

-- ========================================
-- UP: Apply the schema refactoring
-- ========================================

-- Step 1: Create new images table with complete schema
-- Using table recreation pattern (SQLite standard for schema redesign)
-- This ensures clean schema formatting and proper foreign key constraints
CREATE TABLE IF NOT EXISTS images_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,         -- Unique image record identifier
    user_id INTEGER NOT NULL,                     -- Redundant user reference for fast access & auth checks
    prompt_id INTEGER NOT NULL,                   -- Source prompt that generated this image (FK to prompts.id)
    image_data BLOB,                              -- Binary image data (SQLite MVP stage, allow NULL for file storage)
    image_path TEXT,                              -- File path / object storage URL (production stage)
    image_format TEXT NOT NULL DEFAULT 'png',     -- Image format: png / jpg / webp
    width INTEGER NOT NULL DEFAULT 0,             -- Image width in pixels
    height INTEGER NOT NULL DEFAULT 0,            -- Image height in pixels
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Upload timestamp
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,      -- Cascade delete when user is removed
    FOREIGN KEY (prompt_id) REFERENCES prompts(id) ON DELETE CASCADE   -- Cascade delete when prompt is removed
);

-- Step 2: Copy existing data from old images table to new images table
-- Set default prompt_id = 1 for existing images (assumes first prompt exists)
-- This migration assumes data integrity and existing prompts
INSERT INTO images_new (id, user_id, prompt_id, image_data, image_path, image_format, width, height, created_at)
SELECT id, user_id, 1 AS prompt_id, image_data, NULL AS image_path, 'png' AS image_format, 0 AS width, 0 AS height, created_at
FROM images;

-- Step 3: Drop old images table
DROP TABLE images;

-- Step 4: Rename new images table to images
ALTER TABLE images_new RENAME TO images;

-- Step 5: Create indexes for performance optimization
-- Index on prompt_id for efficient lookups by prompt
CREATE INDEX IF NOT EXISTS idx_images_prompt_id ON images(prompt_id);

-- Index on user_id for efficient user-based image queries
CREATE INDEX IF NOT EXISTS idx_images_user_id ON images(user_id);

-- Index on created_at for time-based queries (recent images, etc.)
CREATE INDEX IF NOT EXISTS idx_images_created_at ON images(created_at DESC);

-- ========================================
-- DOWN: Rollback the schema refactoring
-- ========================================
-- Note: Rollback requires recreating the original schema

-- DROP INDEX IF EXISTS idx_images_created_at;
-- DROP INDEX IF EXISTS idx_images_user_id;
-- DROP INDEX IF EXISTS idx_images_prompt_id;
-- 
-- CREATE TABLE IF NOT EXISTS images_old (
--     id INTEGER PRIMARY KEY AUTOINCREMENT,
--     user_id INTEGER NOT NULL,
--     image_data BLOB NOT NULL,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     FOREIGN KEY (user_id) REFERENCES users(id)
-- );
-- 
-- INSERT INTO images_old (id, user_id, image_data, created_at)
-- SELECT id, user_id, image_data, created_at FROM images;
-- 
-- DROP TABLE images;
-- ALTER TABLE images_old RENAME TO images;
-- CREATE INDEX IF NOT EXISTS idx_images_user_id ON images(user_id);

