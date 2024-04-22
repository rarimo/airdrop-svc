-- +migrate Up
CREATE TABLE participants
(
    nullifier  text PRIMARY KEY,
    address    text                        NOT NULL,
    created_at timestamp without time zone NOT NULL default NOW()
);

-- +migrate Down
DROP TABLE participants;