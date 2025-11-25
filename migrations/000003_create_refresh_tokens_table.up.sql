-- Create refresh_tokens table
CREATE TABLE refresh_tokens (
    pkid BIGSERIAL PRIMARY KEY,
    user_pkid BIGINT NOT NULL,
    token VARCHAR(500) UNIQUE NOT NULL,
    expiry_at TIMESTAMP NOT NULL,
    is_revoked BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    CONSTRAINT fk_refresh_tokens_user
        FOREIGN KEY (user_pkid)
        REFERENCES users(pkid)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX idx_refresh_tokens_user_pkid ON refresh_tokens(user_pkid);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_is_revoked ON refresh_tokens(is_revoked);
CREATE INDEX idx_refresh_tokens_is_deleted ON refresh_tokens(is_deleted);
