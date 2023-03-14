package db

import (
	"context"

	"github.com/kwandapchumba/go-bookmark-manager/db/connection"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
)

func ReturnSharedCollection(ctx context.Context, collectionID string) (*sqlc.SharedCollection, error) {
	q := sqlc.New(connection.ConnectDB())

	sharedCollection, err := q.GetSharedCollection(ctx, collectionID)
	if err != nil {
		return &sqlc.SharedCollection{}, err
	}

	return &sharedCollection, nil
}
