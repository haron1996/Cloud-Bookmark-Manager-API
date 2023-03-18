package api

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/jackc/pgconn"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
	"github.com/kwandapchumba/go-bookmark-manager/middleware"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

func (h *BaseHandler) GetCollectionsSharedWithMe(w http.ResponseWriter, r *http.Request) {
	body := r.Context().Value("readRequestOnCollectionDetails").(*middleware.ReadRequestOnCollectionDetails)

	q := sqlc.New(h.db)

	collectionsMemberInstances, err := q.GetCollectionsSharedWithUser(context.Background(), body.Payload.AccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("found no collection member instances at getCollectionsSharedWithMe.go: %v", err)
			util.Response(w, "no collections have been shared with user yet", http.StatusNoContent)
			return
		}

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Printf("could not get collections shared with user at getCollectionsSharedWithMe.go with pgErr: %v", pgErr)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		log.Printf("could not get collections shared with user at getCollectionsSharedWithMe.go with rrr: %v", err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	// declared reponse
	var collectionsSharedWithMeResponse []sqlc.Folder

	for _, collectionMemberInstance := range collectionsMemberInstances {
		collection, err := q.GetFolder(context.Background(), collectionMemberInstance.CollectionID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err := errors.New("collection not found")
				log.Println(err)
				util.Response(w, err.Error(), http.StatusNotFound)
				return
			}

			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) {
				log.Printf("could not get collection at getCollectionsSharedWithMe.go with pgErr: %v", pgErr)
				util.Response(w, "something went wrong", http.StatusInternalServerError)
				return
			}

			log.Printf("could not get collection at getCollectionsSharedWithMe.go with err: %v", pgErr)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		// append collection to slice  of collections to return via json response
		collectionsSharedWithMeResponse = append(collectionsSharedWithMeResponse, collection)
	}

	util.JsonResponse(w, collectionsSharedWithMeResponse)
}
