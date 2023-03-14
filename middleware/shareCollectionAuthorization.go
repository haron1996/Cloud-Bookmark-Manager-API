package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/jackc/pgconn"
	"github.com/kwandapchumba/go-bookmark-manager/auth"
	"github.com/kwandapchumba/go-bookmark-manager/db"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

type shareCollectionRequest struct {
	CollectionID   string   `json:"collection_id"`
	AccessLevel    string   `json:"access_level"`
	EmailsToInvite []string `json:"emails_to_invite"`
}

func (s shareCollectionRequest) validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.CollectionID, validation.Required.Error("collection id required"), validation.Length(33, 33).Error("collecton id must be 33 characters long")),
		validation.Field(&s.AccessLevel, validation.Required.Error("access level required"), validation.In("view", "edit", "admin").Error(fmt.Sprintf(`access level must either be "%s", "%s", or "%s"`, "view", "edit", "admin"))),
		validation.Field(&s.EmailsToInvite, validation.Each(is.Email), validation.Required.Error("at least one email is required")),
	)
}

type ShareCollectionRequestBody struct {
	PayLoad        *auth.PayLoad
	Body           *shareCollectionRequest
	CollectionName string `json:"collecton_name"`
}

func newShareCollectionRequestBody(p auth.PayLoad, b shareCollectionRequest, collectionName string) *ShareCollectionRequestBody {
	return &ShareCollectionRequestBody{
		PayLoad:        &p,
		Body:           &b,
		CollectionName: collectionName,
	}
}

// check if user owns collection OR collection has been shared with them
func AuthorizeShareCollectionRequest() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			body := json.NewDecoder(r.Body)

			body.DisallowUnknownFields()

			var req shareCollectionRequest

			if err := body.Decode(&req); err != nil {
				var internalErr *json.SyntaxError
				if errors.As(err, &internalErr) {
					log.Printf("syntax error at byte offset %d", internalErr.Offset)
					util.Response(w, "something went wrong", http.StatusInternalServerError)
					return
				} else {
					log.Println(err)
					util.Response(w, "something went wrong", http.StatusInternalServerError)
					return
				}
			}

			if err := req.validate(); err != nil {
				log.Printf("invalid email address(es): %s", err.Error())
				util.Response(w, err.Error(), http.StatusUnauthorized)
				return
			}

			// check if folder exists
			folder, err := db.ReturnFolder(context.Background(), req.CollectionID)
			if err != nil {

				if errors.Is(err, sql.ErrNoRows) {
					log.Println("collection not found")
					util.Response(w, "folder not found", http.StatusUnauthorized)
					return
				}

				var pgErr *pgconn.PgError

				if errors.As(err, &pgErr) {
					log.Println(pgErr)
					util.Response(w, "something went wrong", http.StatusInternalServerError)
					return
				}

				log.Println(err)
				util.Response(w, "something went wrong", http.StatusInternalServerError)
				return
			}

			// check if user owns folder
			payload := r.Context().Value("payload").(*auth.PayLoad)

			if folder.AccountID == payload.AccountID {
				// user owns folder hence they are allowed to share
				// pass user and body to context for further processing

				rB := newShareCollectionRequestBody(*payload, req, folder.FolderName)

				ctx := context.WithValue(r.Context(), "requestDetails", rB)

				next.ServeHTTP(w, r.WithContext(ctx))

				return
			}

			// check if collection is shared
			sharedCollection, err := db.ReturnSharedCollection(context.Background(), req.CollectionID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					err := errors.New("this collection is not shared")
					log.Println(err)
					util.Response(w, err.Error(), http.StatusUnauthorized)
					return
				}

				var pgErr *pgconn.PgError

				if errors.As(err, &pgErr) {
					log.Println(pgErr)
					util.Response(w, "something went wrong", http.StatusInternalServerError)
					return
				}

				log.Println(err)
				util.Response(w, "something went wrong", http.StatusInternalServerError)
				return
			}

			// check if folder has been shared with user (request caller)
			sharedCollection, err = db.ReturnSharedCollectionByCollectionIDandAccountID(context.Background(), sharedCollection.CollectionID, payload.AccountID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					err := errors.New("this collection has not been shared with caller")
					log.Println(err)
					util.Response(w, err.Error(), http.StatusUnauthorized)
					return
				}

				var pgErr *pgconn.PgError

				if errors.As(err, &pgErr) {
					log.Println(pgErr)
					util.Response(w, "something went wrong", http.StatusInternalServerError)
					return
				}

				log.Println(err)
				util.Response(w, "something went wrong", http.StatusInternalServerError)
				return
			}

			// check if request caller has admin rights... it's only collection owners and admins that can share collection
			if sharedCollection.CollectionAccessLevel != "admin" {
				err := errors.New("request caller does not have admin rights hence cannot share collection")
				log.Println(err.Error())
				util.Response(w, err.Error(), http.StatusInternalServerError)
				return
			}

			rB := newShareCollectionRequestBody(*payload, req, folder.FolderName)

			ctx := context.WithValue(r.Context(), "requestDetails", rB)

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
