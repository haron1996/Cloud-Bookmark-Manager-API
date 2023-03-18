-- +goose Up
DROP TABLE IF EXISTS member_invite CASCADE;
DROP TYPE IF EXISTS access_level CASCADE;

CREATE TYPE access_level AS ENUM ('view', 'edit', 'admin'); 

CREATE TABLE member_invite (
    invite_id BIGSERIAL PRIMARY KEY,
    invite_token TEXT NOT NULL,
    shared_collection_id TEXT NOT NULL,
    collection_shared_by_name TEXT NOT NULL,
    collection_shared_by_email TEXT NOT NULL,
    collection_shared_with TEXT NOT NULL,
    invite_expiry TIMESTAMPTZ NOT NULL,
    member_access_level access_level NOT NULL,
    UNIQUE (shared_collection_id, collection_shared_with)
);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS member_invite CASCADE;
DROP TYPE IF EXISTS access_level CASCADE;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
