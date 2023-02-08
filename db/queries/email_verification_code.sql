-- name: NewEmailVerificationCode :one
INSERT INTO email_verification (code, email, expiry)
VALUES ($1, $2, $3)
ON CONFLICT (email) DO UPDATE SET code = EXCLUDED.code, expiry = EXCLUDED.expiry
RETURNING *;

-- name: DeleteEmailVerificationCode :exec
DELETE FROM email_verification WHERE email = $1;

-- name: GetOtp :one
SELECT * FROM email_verification WHERE email = $1 LIMIT 1;