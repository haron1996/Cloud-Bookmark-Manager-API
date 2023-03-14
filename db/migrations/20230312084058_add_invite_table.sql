-- +goose Up
DROP TABLE IF EXISTS invite CASCADE;
DROP TYPE IF EXISTS collection_access_level CASCADE;

CREATE TYPE collection_access_level AS ENUM ('view', 'edit', 'admin'); 

CREATE TABLE invite (
    invite_id BIGSERIAL PRIMARY KEY,
    invite_token TEXT NOT NULL,
    shared_collection_id TEXT NOT NULL,
    collection_shared_by_name TEXT NOT NULL,
    collection_shared_by_email TEXT NOT NULL,
    collection_shared_with TEXT NOT NULL,
    invite_expiry TIMESTAMPTZ NOT NULL,
    collection_access_level collection_access_level NOT NULL DEFAULT 'view',
    UNIQUE (shared_collection_id, collection_shared_with)
);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS invite CASCADE;
DROP TYPE IF EXISTS collection_access_level CASCADE;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
