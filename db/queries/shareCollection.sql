-- name: ShareCollection :one
INSERT INTO shared_collection (collection_id, collection_shared_by, collection_shared_with, collection_access_level)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSharedCollection :one
SELECT * FROM shared_collection WHERE collection_id = $1 LIMIT 1;

-- name: GetSharedCollectionByCollectionIDandAccountID :one
SELECT * FROM shared_collection WHERE collection_id = $1 AND collection_shared_with = $2 LIMIT 1;