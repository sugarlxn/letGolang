-- 0002_add_table_prompt.sql
-- Migration: Add prompt metadata table for AI image generation
-- Created: 2026-01-25
-- Description: Creates prompt table to store generation parameters for each image.
--              Establishes 1:1 relationship between prompts and images, with user ownership tracking.
--              Supports full parameter capture including positive/negative prompts and inference steps.

-- ========================================
-- UP: Apply the schema
-- ========================================

-- prompts table: AI image generation metadata and parameters
-- Purpose: Store prompt configuration used to generate each image
-- Relationship: Each image has exactly ONE associated prompt (1:1 via UNIQUE constraint on image_id)
-- Ownership: Each prompt is owned by a user (many prompts per user)
CREATE TABLE IF NOT EXISTS prompts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,  -- Unique prompt record identifier
    user_id INTEGER NOT NULL,              -- Foreign key to users(id); prompt ownership for audit/filtering
    image_id INTEGER NOT NULL UNIQUE,      -- Foreign key to images(id); 1:1 relationship enforced by UNIQUE constraint
    prompt_text TEXT NOT NULL,             -- Positive prompt text used for image generation
    negative_prompt_text TEXT DEFAULT '',  -- Negative prompt to exclude unwanted elements (empty string if not used)
    inference_steps INTEGER NOT NULL,      -- Number of denoising steps in generation process (e.g., 20-50)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  -- Prompt creation/generation timestamp
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,    -- Cascade delete when user is removed
    FOREIGN KEY (image_id) REFERENCES images(id) ON DELETE CASCADE   -- Cascade delete when image is removed
);

-- Create index on user_id for efficient user-based prompt queries
CREATE INDEX IF NOT EXISTS idx_prompts_user_id ON prompts(user_id);

-- Create index on image_id for efficient image-to-prompt lookups (though UNIQUE already provides this)
CREATE INDEX IF NOT EXISTS idx_prompts_image_id ON prompts(image_id);

-- ========================================
-- DOWN: Rollback the schema
-- ========================================
-- Drop indexes and tables in reverse dependency order

-- DROP INDEX IF EXISTS idx_prompts_image_id;
-- DROP INDEX IF EXISTS idx_prompts_user_id;
-- DROP TABLE IF EXISTS prompts;
