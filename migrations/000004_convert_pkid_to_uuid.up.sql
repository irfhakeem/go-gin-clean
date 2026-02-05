-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Step 1: Add new UUID columns to users table
ALTER TABLE users ADD COLUMN id UUID DEFAULT gen_random_uuid();

-- Step 2: Populate the UUID column (already done by default)
-- UPDATE users SET id = gen_random_uuid() WHERE id IS NULL;

-- Step 3: Make id column NOT NULL
ALTER TABLE users ALTER COLUMN id SET NOT NULL;

-- Step 4: Add new UUID column to refresh_tokens table
ALTER TABLE refresh_tokens ADD COLUMN id UUID DEFAULT gen_random_uuid();
ALTER TABLE refresh_tokens ADD COLUMN user_id UUID;

-- Step 5: Map user_id in refresh_tokens to the new UUID from users
UPDATE refresh_tokens rt
SET user_id = u.id
FROM users u
WHERE rt.user_pkid = u.pkid;

-- Step 6: Make user_id NOT NULL
ALTER TABLE refresh_tokens ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE refresh_tokens ALTER COLUMN id SET NOT NULL;

-- Step 7: Drop old foreign key constraint
ALTER TABLE refresh_tokens DROP CONSTRAINT IF EXISTS fk_refresh_tokens_user;

-- Step 8: Drop old primary keys
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_pkey;
ALTER TABLE refresh_tokens DROP CONSTRAINT IF EXISTS refresh_tokens_pkey;

-- Step 9: Add new primary keys with UUID
ALTER TABLE users ADD PRIMARY KEY (id);
ALTER TABLE refresh_tokens ADD PRIMARY KEY (id);

-- Step 10: Add new foreign key constraint
ALTER TABLE refresh_tokens
    ADD CONSTRAINT fk_refresh_tokens_user
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

-- Step 11: Drop old columns
ALTER TABLE users DROP COLUMN pkid;
ALTER TABLE refresh_tokens DROP COLUMN user_pkid;

-- Step 12: Recreate indexes with new column names
DROP INDEX IF EXISTS idx_refresh_tokens_user_pkid;
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
