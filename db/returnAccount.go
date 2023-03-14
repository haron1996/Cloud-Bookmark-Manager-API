package db

import (
	"context"

	"github.com/kwandapchumba/go-bookmark-manager/db/connection"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
)

func ReturnAccount(ctx context.Context, accountID int64) (*sqlc.Account, error) {
	q := sqlc.New(connection.ConnectDB())

	account, err := q.GetAccount(ctx, accountID)
	if err != nil {
		return &sqlc.Account{}, err
	}

	return &account, nil
}
