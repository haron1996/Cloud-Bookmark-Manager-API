-- name: AddLink :one
INSERT INTO link (link_id, link_title, link_hostname, link_url, link_favicon, account_id, folder_id, link_thumbnail)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetRootLinks :many
SELECT * FROM link WHERE account_id = $1 AND folder_id IS NULL AND deleted_at IS NULL ORDER BY added_at DESC;

-- name: GetFolderLinks :many
SELECT * FROM link WHERE account_id = $1 AND folder_id = $2 AND deleted_at IS NULL ORDER BY added_at DESC;

-- name: RenameLink :one
UPDATE link SET link_title = $1 WHERE link_id = $2 RETURNING *;

-- name: MoveLinkToFolder :one
UPDATE link SET folder_id = $1 WHERE link_id = $2 RETURNING *;

-- name: MoveLinkToRoot :one
UPDATE link SET folder_id = NULL WHERE link_id = $1 RETURNING *;

-- name: MoveLinkToTrash :one
UPDATE link SET deleted_at = CURRENT_TIMESTAMP WHERE link_id = $1 RETURNING *;

-- name: RestoreLinkFromTrash :one
UPDATE link SET deleted_at = NULL WHERE link_id = $1 RETURNING *;

-- name: GetLinksMovedToTrash :many
SELECT * FROM link WHERE deleted_at IS NOT NULL AND account_id = $1 ORDER BY deleted_at DESC;

-- name: DeleteLinkForever :one
DELETE FROM link WHERE link_id = $1 RETURNING *;

-- name: SearchLinks :many
SELECT *
FROM link
WHERE textsearchable_index_col @@ plainto_tsquery($1) AND account_id = $2 AND deleted_at IS NULL
ORDER BY added_at DESC;

-- name: SearchLinkz :many
SELECT *
FROM link
WHERE link_title ILIKE $1 AND account_id = $2 AND deleted_at IS NULL
ORDER BY added_at DESC;

-- name: GetLink :one
SELECT * FROM link
WHERE link_id = $1
LIMIT 1;

-- name: GetLinksByUserID :many
SELECT * FROM link WHERE account_id = $1;