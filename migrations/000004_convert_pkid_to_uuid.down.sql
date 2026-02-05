-- Rollback: Convert UUID back to BIGSERIAL

-- Step 1: Add back old PKID columns
ALTER TABLE users ADD COLUMN pkid BIGSERIAL;
ALTER TABLE refresh_tokens ADD COLUMN pkid BIGSERIAL;
ALTER TABLE refresh_tokens ADD COLUMN user_pkid BIGINT;

-- Step 2: Map the data back (this will lose the original PKID values)
-- This is a destructive rollback and cannot restore original PKIDs
UPDATE refresh_tokens rt
SET user_pkid = (
    SELECT pkid FROM users u WHERE u.id = rt.user_id
)
WHERE user_id IS NOT NULL;

-- Step 3: Drop new foreign key
ALTER TABLE refresh_tokens DROP CONSTRAINT IF EXISTS fk_refresh_tokens_user;

-- Step 4: Drop new primary keys
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_pkey;
ALTER TABLE refresh_tokens DROP CONSTRAINT IF EXISTS refresh_tokens_pkey;

-- Step 5: Add back old primary keys
ALTER TABLE users ADD PRIMARY KEY (pkid);
ALTER TABLE refresh_tokens ADD PRIMARY KEY (pkid);

-- Step 6: Re-add foreign key with old columns
ALTER TABLE refresh_tokens
    ADD CONSTRAINT fk_refresh_tokens_user
    FOREIGN KEY (user_pkid)
    REFERENCES users(pkid)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

-- Step 7: Drop UUID columns
ALTER TABLE users DROP COLUMN id;
ALTER TABLE refresh_tokens DROP COLUMN id;
ALTER TABLE refresh_tokens DROP COLUMN user_id;

-- Step 8: Recreate old indexes
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
CREATE INDEX idx_refresh_tokens_user_pkid ON refresh_tokens(user_pkid);
