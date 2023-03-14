-- name: NewAccount :one
INSERT INTO account (fullname, email, account_password)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetAccount :one
SELECT * FROM account
WHERE id = $1
LIMIT 1;

-- name: GetAccountByEmail :one
SELECT * FROM account
WHERE email = $1
LIMIT 1;

-- name: GetAllAccounts :many
SELECT * FROM account;

-- name: UpdateLastLogin :one
UPDATE account
SET last_login = $1
WHERE id = $2
RETURNING *;

-- name: UpdatePassword :exec
UPDATE account SET account_password = $1 WHERE id = $2;

-- name: UpdateAccountEmailVerificationStatus :exec
UPDATE account SET email_verified = 'TRUE' WHERE email = $1;

-- name: EmailExists :one
SELECT EXISTS (SELECT * FROM account WHERE email = $1 LIMIT 1);

-- name: GetAccountLastLogin :one
SELECT Date(last_login) FROM account WHERE id = $1 LIMIT 1;