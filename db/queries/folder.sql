-- name: CreateFolder :one
INSERT INTO folder (folder_id, folder_name, subfolder_of, account_id, path, label)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetRootNodes :many
SELECT * FROM folder
WHERE account_id = $1 AND subfolder_of IS NULL AND folder_deleted_at IS NULL
ORDER BY folder_created_at DESC;

-- name: GetFolderNodes :many
SELECT * FROM folder
WHERE subfolder_of = $1 AND folder_deleted_at IS NULL
ORDER BY folder_created_at DESC;

-- name: GetFolderAncestors :many
SELECT * FROM folder
WHERE folder.path @> (
  SELECT path FROM folder as f
  WHERE f.label = $1
)
ORDER BY path;

-- name: GetFolder :one
SELECT * FROM folder
WHERE folder_id = $1
LIMIT 1;

-- name: GetFolderByFolderAndAccountIds :one
SELECT * FROM folder
WHERE folder_id = $1 AND account_id = $2
LIMIT 1;

-- name: StarFolder :one
UPDATE folder
SET starred = 'true'
WHERE folder_id = $1
RETURNING *;

-- name: UnstarFolder :one
UPDATE folder
SET starred = 'false'
WHERE folder_id = $1
RETURNING *;

-- name: RenameFolder :one
UPDATE folder
SET folder_name = $1
WHERE folder_id = $2
RETURNING *;

-- name: MoveFolderToTrash :one
UPDATE folder
SET folder_deleted_at = CURRENT_TIMESTAMP
WHERE folder_id = $1
RETURNING *;

-- name: MoveFolder :many
UPDATE folder SET path = (SELECT path FROM folder WHERE folder.label = $1) || SUBPATH(path, NLEVEL((SELECT path FROM folder WHERE folder.label = $2))-1) WHERE path <@ (SELECT path FROM folder WHERE folder.label = $3) RETURNING *;

-- name: UpdateFolderSubfolderOf :one
UPDATE folder
SET subfolder_of = $1
WHERE folder_id = $2
RETURNING *;

-- name: MoveFoldersToRoot :many
UPDATE folder SET path = SUBPATH(path, NLEVEL((SELECT path FROM folder WHERE folder.label = $1))-1) WHERE path <@ (
SELECT path FROM folder WHERE folder.label = $2
) RETURNING *;

-- name: UpdateParentFolderToNull :exec
UPDATE folder
SET subfolder_of = NULL
WHERE folder_id = $1;

-- name: ToggleFolderStarred :one
UPDATE folder SET starred = NOT starred WHERE folder_id = $1 RETURNING *;

-- name: GetRootFolders :many
SELECT * FROM folder WHERE NLEVEL(path) = 1 AND account_id = $1 AND folder_deleted_at IS NULL ORDER BY folder_created_at DESC;

-- name: GetFoldersMovedToTrash :many
SELECT * FROM folder WHERE folder_deleted_at IS NOT NULL AND account_id = $1 ORDER BY folder_deleted_at DESC;

-- name: RestoreFolderFromTrash :one
UPDATE folder SET folder_deleted_at = NULL WHERE folder_id = $1 RETURNING *;

-- name: DeleteFolderForever :many
DELETE FROM folder where path <@ (SELECT path FROM folder where folder.folder_id = $1) RETURNING *;

-- name: SearchFolders :many
SELECT *
FROM folder
WHERE textsearchable_index_col @@ plainto_tsquery($1) AND account_id = $2 AND folder_deleted_at IS NULL
ORDER BY folder_created_at DESC;

-- name: SearchFolderz :many
SELECT *
FROM folder
WHERE folder_name ILIKE $1 AND account_id = $2 AND folder_deleted_at IS NULL
ORDER BY folder_created_at DESC;