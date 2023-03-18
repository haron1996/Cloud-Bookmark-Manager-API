package api

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgconn"
	"github.com/kwandapchumba/go-bookmark-manager/db"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

func (h *BaseHandler) CheckIfFolderHasBeenSharedWithUser(w http.ResponseWriter, r *http.Request) {
	memberID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		log.Printf("could not convert user id from url to int at checkIfFolderHasBeenSharedWithUser.go: %v", err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	folderID := chi.URLParam(r, "folderID")

	q := sqlc.New(h.db)

	folder, err := q.GetFolder(context.Background(), folderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("checkIFFolderHasBeenSharedWithUser.go: folder not found: %v", err)
			util.Response(w, "folder not found", http.StatusNoContent)
			return
		}

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Printf("checkIfFolderHasBeenSharedWithUser.go: could not get folder: pgErr %v", pgErr)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		log.Printf("checkIfFolderHasBeenSharedWithUser.go: could not get folder: err %v", err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	if folder.AccountID == int64(memberID) {
		log.Println("checkIfFolderHasBeenSharedWithUser.go: collection be belongs to user")
		util.Response(w, "collection be belongs to user", http.StatusFound)
		return
	}

	args := sqlc.GetCollectionMemberByCollectionAndMemberIDsParams{
		CollectionID: folderID,
		MemberID:     int64(memberID),
	}

	collectionMemberInstance, err := q.GetCollectionMemberByCollectionAndMemberIDs(context.Background(), args)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("collectiom member instance not found")
			// no instance of collection member found
			// return empty collection member instance
			// this specific folder has not been shared with user.... check if one of it's ancestors has already been shared with user.... if so, return that ancestor folder that has been shared

			ancestorsOfFolder, err := q.GetFolderAncestors(context.Background(), folder.Label)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					log.Printf("checkIfFolderHasBeenSharedWithUser.go: found no ancestors of folder: %v", err)
					util.Response(w, "no ancestors of folder found", http.StatusNoContent)
					return
				}

				var pgErr *pgconn.PgError

				if errors.As(err, &pgErr) {
					log.Printf("checkIfFolderHasBeenSharedWithUser.go: could not get ancestors of folder: pgErr: %v", pgErr)
					util.Response(w, "something went wrong", http.StatusInternalServerError)
					return
				}

				log.Printf("checkIfFolderHasBeenSharedWithUser.go: could not get ancestors of folder: err: %v", err)
				util.Response(w, "something went wrong", http.StatusInternalServerError)
				return
			}

			for _, ancancestorOfFolder := range ancestorsOfFolder {
				val, err := db.CheckIfCollectionMemberExists(context.Background(), ancancestorOfFolder.FolderID, args.MemberID)
				if err != nil {
					var pgErr *pgconn.PgError

					if errors.As(err, &pgErr) {
						log.Printf("checkIfFolderHasBeenSharedWithUser.go: could not check of collection member exists: %v", pgErr)
						util.Response(w, "something went wrong", http.StatusInternalServerError)
						return
					}
					log.Printf("checkIfFolderHasBeenSharedWithUser.go: could not check of collection member exists: %v", err)
					util.Response(w, "something went wrong", http.StatusInternalServerError)
					return
				}

				if val {
					cm, _ := q.GetCollectionMemberByCollectionAndMemberIDs(context.Background(), sqlc.GetCollectionMemberByCollectionAndMemberIDsParams{
						CollectionID: ancancestorOfFolder.FolderID,
						MemberID:     int64(memberID),
					})

					util.JsonResponse(w, cm)
					return
				}
			}
			return
		}

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Printf("could not get collection member instance at checkIfFolderHasBeenSharedWithUser.go with pgErr: %v", pgErr)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		log.Printf("could not get collection member instance at checkIfFolderHasBeenSharedWithUser.go with err: %v", err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	util.JsonResponse(w, collectionMemberInstance)
}
