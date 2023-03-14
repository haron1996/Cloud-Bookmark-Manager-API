-- +goose Up
DROP TABLE IF EXISTS invite CASCADE;
DROP TYPE IF EXISTS collection_access_level CASCADE;

CREATE TYPE collection_access_level AS ENUM ('view', 'edit', 'admin'); 

CREATE TABLE invite (
    invite_id BIGSERIAL PRIMARY KEY,
    shared_collection_id TEXT NOT NULL,
    collection_shared_by TEXT NOT NULL,
    collection_shared_with TEXT NOT NULL,
    invite_expiry TIMESTAMPTZ NOT NULL,
    collection_access_level collection_access_level NOT NULL DEFAULT 'view',
    UNIQUE (shared_collection_id, collection_shared_with, collection_shared_by)
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
