-- +goose Up
DROP TABLE IF EXISTS password_reset_token;

CREATE TABLE password_reset_token (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGSERIAL NOT NULL UNIQUE,
    token_hash TEXT NOT NULL,
    token_expiry TIMESTAMPTZ NOT NULL,
    CONSTRAINT fk_account FOREIGN KEY (account_id) REFERENCES account (id) ON DELETE CASCADE
);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS password_reset_token;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
