package middleware

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgconn"
	"github.com/kwandapchumba/go-bookmark-manager/auth"
	"github.com/kwandapchumba/go-bookmark-manager/db"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

type ReadRequestOnCollectionDetails struct {
	FolderID  string       `json:"folderid"`
	AccountID int64        `json:"accountid"`
	Payload   auth.PayLoad `json:"payload"`
}

func newReadRequestOnCollectionDetails(folderid string, accountid int64, payload auth.PayLoad) *ReadRequestOnCollectionDetails {
	return &ReadRequestOnCollectionDetails{
		FolderID:  folderid,
		AccountID: accountid,
		Payload:   payload,
	}
}

func AuthorizeReadRequestOnCollection() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			payload := r.Context().Value("payload").(*auth.PayLoad)

			// log.Printf("PAYLOAD at authorizeReadRequestOnCollection.go: %v", payload)

			folderID := chi.URLParam(r, "folderID")

			accountID := chi.URLParam(r, "accountID")

			account_id, err := strconv.Atoi(accountID)
			if err != nil {
				log.Printf("could not convert account id from url to int64 at authorizeReadRequestOnCollection.go: %v", err)
				util.Response(w, "something went wrong", http.StatusInternalServerError)
				return
			}

			if int64(account_id) != payload.AccountID {
				log.Println("account IDs do not match")
				util.Response(w, errors.New("account IDs do not match").Error(), http.StatusUnauthorized)
				return
			}

			if folderID == "null" {
				// user owns folder hence is allowed to read from it hence no further checks hence pass request details to context!
				body := newReadRequestOnCollectionDetails(folderID, int64(account_id), *payload)

				ctx := context.WithValue(r.Context(), "readRequestOnCollectionDetails", body)

				next.ServeHTTP(w, r.WithContext(ctx))

				return
			}

			collection, err := db.ReturnFolder(r.Context(), folderID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					err := errors.New("collection not found")
					log.Println(err)
					util.Response(w, err.Error(), http.StatusNotFound)
					return
				}

				var pgErr *pgconn.PgError

				if errors.As(err, &pgErr) {
					log.Printf("could not get collection at authorizeReadRequestOnCollection.go with pgErr: %v", pgErr)
					util.Response(w, "something went wrong", http.StatusInternalServerError)
					return
				}

				log.Printf("could not get collection at authorizeReadRequestOnCollection.go with err: %v", pgErr)
				util.Response(w, "something went wrong", http.StatusInternalServerError)
				return
			}

			// check if collection belongs to user
			if collection.AccountID == payload.AccountID {
				// collection belongs to user hence user is allowed to read from it hence no further check hence pass request details to context
				body := newReadRequestOnCollectionDetails(folderID, int64(account_id), *payload)

				ctx := context.WithValue(r.Context(), "readRequestOnCollectionDetails", body)

				next.ServeHTTP(w, r.WithContext(ctx))

				return
			}

			// user does not own folder hence check if folder has been shared with them
			collectionMember, err := db.ReturnCollectionMemberByCollectionAndMemberIDs(r.Context(), collection.FolderID, payload.AccountID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {

					// this specific folder has not been shared with them
					// check if any of it's ancestors has been shared with used
					ancestorsOfFolder, err := db.ReturnAncestorsOfFolder(r.Context(), collection.FolderID)
					if err != nil {
						if errors.Is(err, sql.ErrNoRows) {
							log.Printf("authorizeReadRequestOnCollection: found no ancestors of folder: %v", err)
							util.Response(w, "no ancestors of folder found", http.StatusNoContent)
							return
						}

						var pgErr *pgconn.PgError

						if errors.As(err, &pgErr) {
							log.Printf("authorizeReadRequestOnCollection: could not get ancestors of folder: pgErr: %v", pgErr)
							util.Response(w, "something went wrong", http.StatusInternalServerError)
							return
						}

						log.Printf("authorizeReadRequestOnCollection: could not get ancestors of folder: err: %v", err)
						util.Response(w, "something went wrong", http.StatusInternalServerError)
						return
					}

					for _, ancestorOfFolder := range ancestorsOfFolder {
						val, err := db.CheckIfCollectionMemberExists(r.Context(), ancestorOfFolder.FolderID, payload.AccountID)
						if err != nil {
							var pgErr *pgconn.PgError

							if errors.As(err, &pgErr) {
								log.Printf("authorizeReadRequestOnCollecton.go: could not check of collection member exists: %v", pgErr)
								util.Response(w, "something went wrong", http.StatusInternalServerError)
								return
							}
							log.Printf("authorizeReadRequestOnCollecton.go: could not check of collection member exists: %v", err)
							util.Response(w, "something went wrong", http.StatusInternalServerError)
							return
						}

						if val {
							body := newReadRequestOnCollectionDetails(collection.FolderID, int64(account_id), *payload)

							ctx := context.WithValue(r.Context(), "readRequestOnCollectionDetails", body)

							next.ServeHTTP(w, r.WithContext(ctx))

							return
						}
					}

					return
				}

				var pgErr *pgconn.PgError

				if errors.As(err, &pgErr) {
					log.Printf("could not get collection member in authorizeReadRequestOnCollection.go... pgErr: %v", pgErr)
					util.Response(w, "something went wrong", http.StatusInternalServerError)
					return
				}

				log.Printf("could not get collection member in authorizeReadRequestOnCollection.go... err: %v", err)
				util.Response(w, "something went wrong", http.StatusInternalServerError)
				return
			}

			// collection has been shared with user hence pass request details to context
			body := newReadRequestOnCollectionDetails(collectionMember.CollectionID, int64(account_id), *payload)

			ctx := context.WithValue(r.Context(), "readRequestOnCollectionDetails", body)

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
