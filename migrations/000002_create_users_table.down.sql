-- Drop indexes
DROP INDEX IF EXISTS idx_users_is_deleted;
DROP INDEX IF EXISTS idx_users_oauth;
DROP INDEX IF EXISTS idx_users_code;
DROP INDEX IF EXISTS idx_users_email;

-- Drop users table
DROP TABLE IF EXISTS users;
