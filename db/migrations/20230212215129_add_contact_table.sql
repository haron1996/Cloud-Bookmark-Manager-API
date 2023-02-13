-- +goose Up
DROP TABLE IF EXISTS contact;

CREATE TABLE contact (
    id BIGSERIAL NOT NULL PRIMARY KEY,
    account_id BIGSERIAL REFERENCES account(id) NOT NULL UNIQUE,
    message TEXT NOT NULL
);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS contact;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
