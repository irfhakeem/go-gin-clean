-- Migration: edit_user_code_col
-- Created at: /tmp/go-build82322524/b001/exe/main

-- Add your UP migration SQL here
ALTER TABLE users
    ALTER COLUMN code TYPE VARCHAR(15);
