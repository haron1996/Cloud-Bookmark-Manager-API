package db

import (
	"context"

	"github.com/kwandapchumba/go-bookmark-manager/db/connection"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
)

func ReturnFolder(ctx context.Context, folderID string) (*sqlc.Folder, error) {
	q := sqlc.New(connection.ConnectDB())

	folder, err := q.GetFolder(ctx, folderID)
	if err != nil {
		return &sqlc.Folder{}, err
	}

	return &folder, nil
}
