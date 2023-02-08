-- +goose Up
DROP TABLE IF EXISTS account_session;

CREATE TABLE account_session (
    refresh_token_id TEXT NOT NULL,
    account_id BIGSERIAL REFERENCES account(id) NOT NULL UNIQUE,
    issued_at TIMESTAMPTZ NOT NULL,
    expiry TIMESTAMPTZ NOT NULL,
    user_agent TEXT NOT NULL,
    client_ip TEXT NOT NULL
);
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE OR REPLACE FUNCTION update_account_last_login_trigger_fnc()
RETURNS TRIGGER AS
$BODY$
BEGIN
UPDATE account
SET last_login = NEW.issued_at
WHERE id = NEW.account_id;
RETURN NEW;
END;
$BODY$
LANGUAGE 'plpgsql';

CREATE TRIGGER update_account_last_login_trigger
AFTER INSERT OR UPDATE
ON account_session
FOR EACH ROW
EXECUTE PROCEDURE update_account_last_login_trigger_fnc();
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS account_session;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
