package util

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

func ReturnFolder(ctx context.Context, folderID string) (*sqlc.Folder, error) {
	q := sqlc.New(connection.ConnectDB())

	folder, err := q.GetFolder(ctx, folderID)
	if err != nil {
		return &sqlc.Folder{}, err
	}

	return &folder, nil
}

func ReturnSharedCollection(ctx context.Context, collectionID string) (*sqlc.SharedCollection, error) {
	q := sqlc.New(connection.ConnectDB())

	sharedCollection, err := q.GetSharedCollection(ctx, collectionID)
	if err != nil {
		return &sqlc.SharedCollection{}, err
	}

	return &sharedCollection, nil
}

func ReturnSharedCollectionByCollectionIDandAccountID(ctx context.Context, collectionID string, accountID int64) (*sqlc.SharedCollection, error) {
	q := sqlc.New(connection.ConnectDB())

	arg := sqlc.GetSharedCollectionByCollectionIDandAccountIDParams{
		CollectionID:         collectionID,
		CollectionSharedWith: accountID,
	}

	sharedCollection, err := q.GetSharedCollectionByCollectionIDandAccountID(ctx, arg)
	if err != nil {
		return &sqlc.SharedCollection{}, err
	}

	return &sharedCollection, nil
}
