-- name: CreateAccountSession :one
INSERT INTO account_session (refresh_token_id, account_id, issued_at, expiry, user_agent, client_ip)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (account_id) DO UPDATE SET refresh_token_id = EXCLUDED.refresh_token_id, issued_at = EXCLUDED.issued_at, expiry = EXCLUDED.expiry, user_agent = EXCLUDED.user_agent, client_ip = EXCLUDED.client_ip
RETURNING *; 