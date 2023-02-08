-- +goose Up
DROP TABLE IF EXISTS email_verification;

CREATE TABLE email_verification (
    code TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    expiry TIMESTAMPTZ NOT NULL
);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
