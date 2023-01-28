-- +goose Up
DROP TABLE IF EXISTS link;

CREATE TABLE link (
    link_id TEXT NOT NULL PRIMARY KEY,
    link_title TEXT NOT NULL,
    link_thumbnail TEXT NOT NULL,
    link_favicon TEXT NOT NULL,
    link_hostname TEXT NOT NULL,
    link_url TEXT NOT NULL,
    link_notes TEXT NOT NULL DEFAULT '',
    account_id BIGSERIAL NOT NULL,
    folder_id TEXT,
    added_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_folder FOREIGN KEY (folder_id) REFERENCES folder (folder_id) ON DELETE CASCADE,
    CONSTRAINT fk_account FOREIGN KEY (account_id) REFERENCES account (id) ON DELETE CASCADE
);

ALTER TABLE link ADD COLUMN textsearchable_index_col tsvector GENERATED ALWAYS AS (to_tsvector('english', link_title)) STORED;

CREATE INDEX link_title_search_idx ON folder USING GIN (textsearchable_index_col);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
ALTER TABLE IF EXISTS link DROP CONSTRAINT IF EXISTS fk_folder;
ALTER TABLE IF EXISTS link DROP CONSTRAINT IF EXISTS fk_account;
DROP TABLE IF EXISTS link;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
