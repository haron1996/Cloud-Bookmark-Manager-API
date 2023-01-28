-- name: CreateAccountSession :one
INSERT INTO account_session (refresh_token_id, account_id, issued_at, expiry, user_agent, client_ip)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

