package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/jackc/pgconn"
	"github.com/kwandapchumba/go-bookmark-manager/auth"
	"github.com/kwandapchumba/go-bookmark-manager/db"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

type createFolderRequest struct {
	FolderName string `json:"folder_name"`
	FolderID   string `json:"parent_folder_id"`
}

func (s createFolderRequest) validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.FolderName, validation.Required.When(s.FolderName == "").Error("folder name is required"), validation.Match(regexp.MustCompile("^[^?[\\]{}|\\\\`./!@#$%^&*()_-]+$")).Error("folder name must not have special characters"), validation.Length(1, 100).Error("folder name must be at least 1 character long")),
	)
}

type CreateFolderRequestBody struct {
	PayLoad *auth.PayLoad
	Body    *createFolderRequest
}

func newCreateFolderRequestBody(p auth.PayLoad, b createFolderRequest) *CreateFolderRequestBody {
	return &CreateFolderRequestBody{
		PayLoad: &p,
		Body:    &b,
	}
}

func AuthorizeCreateFolderRequest() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// get request body/content
			body := json.NewDecoder(r.Body)

			body.DisallowUnknownFields()

			var req createFolderRequest

			if err := body.Decode(&req); err != nil {
				if e, ok := err.(*json.SyntaxError); ok {
					log.Printf("syntax error at byte offset %d", e.Offset)
					util.Response(w, "something went wrong", http.StatusInternalServerError)
					return
				} else {
					log.Printf("error decoding request body to struct: %v", err)
					util.Response(w, "bad request", http.StatusBadRequest)
					return
				}
			}

			// validate request
			if err := req.validate(); err != nil {
				log.Println(err)
				util.Response(w, err.Error(), http.StatusBadRequest)
				return
			}

			// check if user wants to create a root folder
			if req.FolderID == "null" {
				payload := r.Context().Value("payload").(*auth.PayLoad)

				rB := newCreateFolderRequestBody(*payload, req)

				ctx := context.WithValue(r.Context(), "createFolderRequest", rB)

				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// get user payload from request
			payload := r.Context().Value("payload").(*auth.PayLoad)

			// get parent folder
			folder, err := db.ReturnFolder(r.Context(), req.FolderID)
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

			// check if user owns parent folder
			if folder.AccountID == payload.AccountID {
				rB := newCreateFolderRequestBody(*payload, req)

				ctx := context.WithValue(r.Context(), "createFolderRequest", rB)

				next.ServeHTTP(w, r.WithContext(ctx))

				return
			}

			// check if folder has been shared with user
			collectionMember, err := db.ReturnCollectionMemberByCollectionAndMemberIDs(context.Background(), folder.FolderID, payload.AccountID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					log.Printf("could not find collection member instance in createFolderAuthorization.go: %v", err)
					util.Response(w, "collection has not been shared with you", http.StatusUnauthorized)
					return
				}

				var pgErr *pgconn.PgError

				if errors.As(err, &pgErr) {
					log.Printf("could not get collection member in createFolderAuthorization.go... pgErr: %v", pgErr)
					util.Response(w, "something went wrong", http.StatusInternalServerError)
					return
				}

				log.Printf("could not get collection member in createFolderAuthorization.go... err: %v", err)
				util.Response(w, "something went wrong", http.StatusInternalServerError)
				return
			}

			// check if user has admin rights
			if collectionMember.CollectionAccessLevel != "admin" {
				log.Printf(`user is not allowed to create edit this collection: user access level is not "admin" but "%v"`, collectionMember.CollectionAccessLevel)
				util.Response(w, "access denied due to insufficient access level", http.StatusUnauthorized)
				return
			}

			// user has admin rights hence allowed to edit this collection
			rB := newCreateFolderRequestBody(*payload, req)

			ctx := context.WithValue(r.Context(), "myValues", rB)

			next.ServeHTTP(w, r.WithContext(ctx))

			return
		}

		return http.HandlerFunc(fn)
	}
}
