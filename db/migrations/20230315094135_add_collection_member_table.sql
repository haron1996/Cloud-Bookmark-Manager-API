-- +goose Up
DROP TABLE IF EXISTS collection_member;
DROP TYPE IF EXISTS collection_access_level CASCADE;

CREATE TYPE collection_access_level AS ENUM ('view', 'edit', 'admin'); 

CREATE TABLE collection_member (
  collection_id TEXT NOT NULL,
  member_id BIGSERIAL NOT NULL,
  join_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  collection_access_level collection_access_level NOT NULL,
  CONSTRAINT fk_collection FOREIGN KEY (collection_id) REFERENCES folder (folder_id) ON DELETE CASCADE,
  CONSTRAINT fk_account FOREIGN KEY (member_id) REFERENCES account (id) ON DELETE CASCADE,
  PRIMARY KEY (collection_id, member_id),
  -- When a unique constraint is defined on a group of columns, then the combination of those column values needs to be unique across the table.
  UNIQUE(collection_id, member_id)
);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
