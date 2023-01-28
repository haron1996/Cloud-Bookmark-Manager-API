-- +goose Up
DROP TABLE IF EXISTS account;

CREATE TABLE account (
    id BIGSERIAL PRIMARY KEY,
    fullname TEXT NOT NULL,
    email TEXT NOT NULL CONSTRAINT email_must_be_unique UNIQUE,
    email_verified BOOLEAN NOT NULL DEFAULT 'false',
    picture TEXT NOT NULL DEFAULT '',
    account_password TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    intention TEXT,
    last_login TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX email_idx ON account(LOWER(email));
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS account;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
