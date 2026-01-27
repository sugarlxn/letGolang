-- Migration: Create image_tasks table
-- Description: Store async image generation tasks
-- Created: 2026-01-27

CREATE TABLE IF NOT EXISTS image_tasks (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    prompt TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'QUEUED',
    result_url TEXT,
    error_msg TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_image_tasks_user_id ON image_tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_image_tasks_status ON image_tasks(status);
CREATE INDEX IF NOT EXISTS idx_image_tasks_created_at ON image_tasks(created_at);
