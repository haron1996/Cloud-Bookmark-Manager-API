package db

import (
	"context"

	"github.com/kwandapchumba/go-bookmark-manager/db/connection"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
)

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
