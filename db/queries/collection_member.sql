-- name: AddNewCollectionMember :one
INSERT INTO collection_member (collection_id, member_id, collection_access_level) VALUES ($1, $2, $3) RETURNING *;

-- name: GetCollectionMemberByCollectionAndMemberIDs :one
SELECT * FROM collection_member WHERE collection_id = $1 AND member_id = $2 LIMIT 1;

-- name: CheckIfCollectionMemberWithCollectionAndMemberIDsExists :one
SELECT EXISTS (SELECT * FROM collection_member WHERE collection_id = $1 AND member_id = $2 LIMIT 1);

-- name: GetCollectionsSharedWithUser :many
SELECT * FROM collection_member WHERE member_id = $1;