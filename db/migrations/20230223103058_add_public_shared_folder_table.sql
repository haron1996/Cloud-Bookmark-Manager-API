-- +goose Up
DROP TABLE IF EXISTS public_shared_collection;

CREATE TABLE public_shared_collection (
    collection_id TEXT NOT NULL PRIMARY KEY UNIQUE,
    collection_password CHAR(6) NOT NULL,
    collection_shared_by BIGSERIAL NOT NULL,
    collection_shared_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    collection_share_expiry TIMESTAMPTZ NULL,
    CONSTRAINT fk_folder FOREIGN KEY (collection_id) REFERENCES folder (folder_id) ON DELETE CASCADE,
    CONSTRAINT fk_account FOREIGN KEY (collection_shared_by) REFERENCES account (id) ON DELETE CASCADE
);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP CONSTRAINT IF EXISTS fk_folder;
DROP CONSTRAINT IF EXISTS fk_account;
DROP TABLE IF EXISTS public_shared_collection;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
