-- name: NewMessage :one
INSERT INTO contact (account, message_body)
VALUES ($1, $2)
RETURNING *;