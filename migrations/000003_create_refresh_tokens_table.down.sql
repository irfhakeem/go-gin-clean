DROP INDEX IF EXISTS idx_refresh_tokens_is_deleted;
DROP INDEX IF EXISTS idx_refresh_tokens_is_revoked;
DROP INDEX IF EXISTS idx_refresh_tokens_token;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;

DROP TABLE IF EXISTS refresh_tokens;
