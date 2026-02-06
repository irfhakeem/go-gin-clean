-- Migration: create_sequence_table
-- Created at: /home/hakim/.cache/go-build/33/3382315bb2fb1a9df4235e699d409792df5a2fc5b5113a85783e0c38e0940f37-d/main

-- Add your UP migration SQL here
CREATE TABLE IF NOT EXISTS sequences (
    entity_name VARCHAR(50) PRIMARY KEY,
    last_code VARCHAR(255) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO sequences (entity_name, last_code)
SELECT 'users', COALESCE((SELECT code FROM users ORDER BY id DESC LIMIT 1), '')
ON CONFLICT (entity_name) DO NOTHING;
