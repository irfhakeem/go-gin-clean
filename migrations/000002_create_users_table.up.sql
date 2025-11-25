-- Create users table
CREATE TABLE users (
    pkid BIGSERIAL PRIMARY KEY,
    code VARCHAR(11) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255),
    avatar VARCHAR(255) DEFAULT '',
    gender gender DEFAULT NULL,
    role role DEFAULT 'User',
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    is_verified BOOLEAN DEFAULT FALSE NOT NULL,
    oauth_provider VARCHAR(50),
    oauth_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,
    is_deleted BOOLEAN DEFAULT FALSE
);

-- Create indexes for better query performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_code ON users(code);
CREATE INDEX idx_users_oauth ON users(oauth_provider, oauth_id);
CREATE INDEX idx_users_is_deleted ON users(is_deleted);
