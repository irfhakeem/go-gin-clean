-- Migration: create_outbox (rollback)

DROP TABLE IF EXISTS outbox_messages;
DROP TYPE IF EXISTS outbox_status;
