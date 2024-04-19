-- +migrate Up
CREATE TYPE tx_status AS ENUM ('pending', 'completed');

CREATE TABLE participants
(
    nullifier  text PRIMARY KEY,
    address    text                        NOT NULL,
    status tx_status NOT NULL,
    created_at timestamp without time zone NOT NULL default NOW(),
    updated_at timestamp without time zone NOT NULL default NOW()
);

-- +migrate Down
DROP TABLE participants;
DROP TYPE tx_status;
