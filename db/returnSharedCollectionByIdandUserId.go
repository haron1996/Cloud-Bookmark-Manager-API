package db

import (
	"context"

	"github.com/kwandapchumba/go-bookmark-manager/db/connection"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
)

func ReturnCollectionMemberByCollectionAndMemberIDs(ctx context.Context, collectionID string, accountID int64) (*sqlc.CollectionMember, error) {
	arg := sqlc.GetCollectionMemberByCollectionAndMemberIDsParams{
		CollectionID: collectionID,
		MemberID:     accountID,
	}
	collectionMember, err := sqlc.New(connection.ConnectDB()).GetCollectionMemberByCollectionAndMemberIDs(ctx, arg)
	if err != nil {
		return &sqlc.CollectionMember{}, err
	}

	return &collectionMember, nil
}
