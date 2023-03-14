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

			if err := req.validate(); err != nil {
				log.Println(err)
				util.Response(w, err.Error(), http.StatusBadRequest)
				return
			}

			// check if parent folder id exists... if not, no further checks
			if req.FolderID == "null" {
				// means folder is root
				payload := r.Context().Value("payload").(*auth.PayLoad)

				rB := newCreateFolderRequestBody(*payload, req)

				ctx := context.WithValue(r.Context(), "myValues", rB)

				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				// means folder is a child folder
				// let's check if folder exists
				payload := r.Context().Value("payload").(*auth.PayLoad)

				folder, err := db.ReturnFolder(r.Context(), req.FolderID)
				if err != nil {
					util.Response(w, "parent folder not found", http.StatusNotFound)
					return
				}

				// folder exists
				if folder.AccountID == payload.AccountID {
					// means user owns folder... hence can edit
					rB := newCreateFolderRequestBody(*payload, req)

					ctx := context.WithValue(r.Context(), "myValues", rB)

					next.ServeHTTP(w, r.WithContext(ctx))

					return
				}

				// user does not own folder
				if folder.AccountID != payload.AccountID {
					// mmmhh... let's check if this folder is shared..
					// do this by getting shared collection by collection id (folder id)
					sharedCollection, err := db.ReturnSharedCollection(context.Background(), folder.FolderID)
					if err != nil {
						// if folder is not shared, means user has no access
						if errors.Is(err, sql.ErrNoRows) {
							util.Response(w, "folder is not shared", http.StatusUnauthorized)
						} else {
							util.Response(w, "something went wrong", http.StatusInternalServerError)
						}
					}

					// folder is shared
					// let's check if this user has access to this shared collection
					sharedCollection, err = db.ReturnSharedCollectionByCollectionIDandAccountID(r.Context(), sharedCollection.CollectionID, payload.AccountID)
					if err != nil {
						if errors.Is(err, sql.ErrNoRows) {
							// means folder has not been shared with this user
							util.Response(w, "this folder has not been shared with this user", http.StatusUnauthorized)
						} else {
							util.Response(w, "something went wrong", http.StatusInternalServerError)
							return
						}
					}

					// folder has been shared with this user
					// let's check if permission is edit

					if sharedCollection.CollectionAccessLevel == "view" {
						// means user is not allowed to edit this folder
						util.Response(w, "no edit access", http.StatusUnauthorized)
						return
					}

					if sharedCollection.CollectionAccessLevel == "edit" {
						// user is allowed to edit this folder
						rB := newCreateFolderRequestBody(*payload, req)

						ctx := context.WithValue(r.Context(), "myValues", rB)

						next.ServeHTTP(w, r.WithContext(ctx))

						return
					}

					log.Println("folder access level is none of the above!")
				}
			}
			// check if user owns parent folder id... if true, no further checks
			// check if folder exists in shared collectins... if not, unauthorized error
			// check if folder has been shared with user... if not, return access denied
			// check if user has access to edit shared collection... if not, unauthorized error
			// all good, proceed

			// payload := r.Context().Value("payload").(*auth.PayLoad)

			// ctx := context.WithValue(r.Context(), "authorizedPayload", payload)

			// next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
