-- +goose Up
DROP TABLE IF EXISTS public_shared_collection;

CREATE TYPE collection_access_level AS ENUM ('view', 'edit');

CREATE TABLE public_shared_collection (
    collection_id TEXT NOT NULL PRIMARY KEY UNIQUE,
    collection_password CHAR(6) NOT NULL,
    collection_shared_by BIGSERIAL NOT NULL,
    collection_shared_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    collection_share_expiry TIMESTAMPTZ NULL,
    collection_access_level collection_access_level NOT NULL DEFAULT 'view',
    CONSTRAINT fk_folder FOREIGN KEY (collection_id) REFERENCES folder (folder_id) ON DELETE CASCADE,
    CONSTRAINT fk_account FOREIGN KEY (collection_shared_by) REFERENCES account (id) ON DELETE CASCADE
);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS public_shared_collection;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
