-- +migrate Up
CREATE TYPE tx_status_enum AS ENUM ('pending', 'completed', 'failed');

ALTER TABLE participants
    ADD COLUMN status     tx_status_enum              NOT NULL DEFAULT 'completed',
    ADD COLUMN updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    ADD COLUMN tx_hash    VARCHAR(64)                 NOT NULL DEFAULT '',
    ADD COLUMN amount     TEXT                        NOT NULL DEFAULT '0urmo';

-- +migrate Down
ALTER TABLE participants
    DROP COLUMN status,
    DROP COLUMN updated_at,
    DROP COLUMN tx_hash,
    DROP COLUMN amount;

DROP TYPE IF EXISTS tx_status_enum;
