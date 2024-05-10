-- +migrate Up
CREATE TYPE tx_status_enum AS ENUM ('pending', 'completed', 'failed');

CREATE TABLE airdrops
(
    id         uuid PRIMARY KEY                     DEFAULT gen_random_uuid(),
    nullifier  text                        NOT NULL,
    address    text                        NOT NULL,
    tx_hash    text,
    amount     text                        NOT NULL,
    status     tx_status_enum              NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp without time zone NOT NULL DEFAULT NOW()
);

-- +migrate Down
DROP TABLE airdrops;
DROP TYPE tx_status_enum;