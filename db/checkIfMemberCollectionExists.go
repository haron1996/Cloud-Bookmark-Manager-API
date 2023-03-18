package db

import (
	"context"

	"github.com/kwandapchumba/go-bookmark-manager/db/connection"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
)

func CheckIfCollectionMemberExists(ctx context.Context, folderID string, accountID int64) (bool, error) {
	value, err := sqlc.New(connection.ConnectDB()).CheckIfCollectionMemberWithCollectionAndMemberIDsExists(ctx, sqlc.CheckIfCollectionMemberWithCollectionAndMemberIDsExistsParams{
		CollectionID: folderID,
		MemberID:     accountID,
	})
	if err != nil {
		return false, err
	}
	return value, nil
}
