-- +goose Up
DROP TABLE IF EXISTS contact;

CREATE TABLE contact (
    id BIGSERIAL NOT NULL PRIMARY KEY,
    account BIGSERIAL REFERENCES account(id) NOT NULL,
    message_body TEXT NOT NULL
);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS contact;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
