-- +goose Up
DROP TABLE IF EXISTS shared_collection CASCADE;
DROP TYPE IF EXISTS collection_access_level CASCADE;

CREATE TYPE collection_access_level AS ENUM ('view', 'edit');

CREATE TABLE shared_collection (
    collection_id TEXT NOT NULL PRIMARY KEY UNIQUE,
    collection_shared_by BIGSERIAL NOT NULL,
    collection_shared_with BIGSERIAL NOT NULL,
    collection_shared_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    collection_access_level collection_access_level NOT NULL DEFAULT 'view',
    CONSTRAINT fk_folder FOREIGN KEY (collection_id) REFERENCES folder (folder_id) ON DELETE CASCADE,
    CONSTRAINT fk_account_shared_by FOREIGN KEY (collection_shared_by) REFERENCES account (id) ON DELETE CASCADE,
    CONSTRAINT fk_account_shared_with FOREIGN KEY (collection_shared_with) REFERENCES account (id) ON DELETE CASCADE,
    UNIQUE (collection_id, collection_shared_with)
);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP CONSTRAINT IF EXISTS fk_folder_shared_by;
DROP CONSTRAINT IF EXISTS fk_account_shared_with;
DROP TABLE IF EXISTS shared_collection;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
