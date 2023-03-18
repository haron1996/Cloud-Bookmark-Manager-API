-- name: CreateInvite :one
INSERT INTO member_invite (shared_collection_id, collection_shared_by_name, collection_shared_by_email, collection_shared_with, member_access_level, invite_expiry, invite_token)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT ON CONSTRAINT member_invite_shared_collection_id_collection_shared_with_key DO UPDATE SET  collection_shared_by_name = EXCLUDED.collection_shared_by_name, collection_shared_by_email = EXCLUDED.collection_shared_by_email, member_access_level = EXCLUDED.member_access_level, invite_expiry = EXCLUDED.invite_expiry, invite_token = EXCLUDED.invite_token 
RETURNING *;

-- name: GetInviteByToken :one
SELECT * FROM member_invite WHERE invite_token = $1 LIMIT 1;

-- name: DeleteInvite :exec
DELETE FROM member_invite WHERE invite_token = $1;