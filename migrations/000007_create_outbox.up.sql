-- Migration: create_outbox

CREATE TYPE outbox_status AS ENUM ('pending', 'processing', 'published', 'failed');

CREATE TABLE outbox_messages (
    pkid           BIGSERIAL     PRIMARY KEY,
    aggregate_type VARCHAR(100)  NOT NULL,
    aggregate_id   VARCHAR(100)  NOT NULL,
    event_type     VARCHAR(100)  NOT NULL,
    exchange       VARCHAR(100)  NOT NULL DEFAULT 'pc_main_event_bus',
    payload        JSONB         NOT NULL,
    status         outbox_status NOT NULL DEFAULT 'pending',
    retry_count    INT           NOT NULL DEFAULT 0,
    max_retries    INT           NOT NULL DEFAULT 3,
    last_error     TEXT,
    created_at     TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    published_at   TIMESTAMP
);

CREATE INDEX idx_outbox_messages_status_created ON outbox_messages (status, created_at)
    WHERE status IN ('pending', 'processing');
