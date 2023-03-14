package api

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
	"github.com/kwandapchumba/go-bookmark-manager/mailjet"
	"github.com/kwandapchumba/go-bookmark-manager/middleware"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

func (h *BaseHandler) ShareCollection(w http.ResponseWriter, r *http.Request) {
	b := r.Context().Value("requestDetails").(*middleware.ShareCollectionRequestBody).Body

	p := r.Context().Value("requestDetails").(*middleware.ShareCollectionRequestBody).PayLoad

	collectionName := r.Context().Value("requestDetails").(*middleware.ShareCollectionRequestBody).CollectionName

	// save invite details in db
	// share id
	// shared collection
	// access level
	// shared with email address
	// share expiry
	// shared by email addresss
	// send invite email to each email

	q := sqlc.New(h.db)

	account, err := q.GetAccount(context.Background(), p.AccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err := errors.New("account not found")
			log.Println(err.Error())
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

	for _, email := range b.EmailsToInvite {
		token := uuid.NewString()

		encodedToken := base64.StdEncoding.EncodeToString([]byte(token))

		inviteEpiry := time.Now().UTC().Add(72 * time.Hour) // 3 days

		params := sqlc.CreateInviteParams{
			SharedCollectionID:      b.CollectionID,
			CollectionSharedByName:  account.Fullname,
			CollectionSharedByEmail: account.Email,
			CollectionSharedWith:    email,
			CollectionAccessLevel:   sqlc.CollectionAccessLevel(b.AccessLevel),
			InviteExpiry:            inviteEpiry, // 3 days
			InviteToken:             encodedToken,
		}

		invite, err := q.CreateInvite(context.Background(), params)
		if err != nil {

			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) {
				log.Printf("could not create invite %v", pgErr.Message)
				e := `duplicate key value violates unique constraint "invite_shared_collection_id_collection_shared_with_key"`
				if pgErr.Message == e {
					resendInvitation(email)
					return
				}
				util.Response(w, "something went wrong", http.StatusInternalServerError)
				return
			}

			log.Println(err)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		mail := mailjet.NewInviteUserMail(account.Fullname, account.Email, invite.CollectionSharedWith, token, collectionName, inviteEpiry)

		mail.SendInviteUserEmail()
	}

	// if len(b.EmailsToInvite) > 1 {
	// 	util.Response(w, fmt.Sprintf("%d people have been invited", len(b.EmailsToInvite)), http.StatusOK)
	// } else if len(b.EmailsToInvite) == 1 {
	// 	util.Response(w, fmt.Sprintf("%d person has been invited", len(b.EmailsToInvite)), http.StatusOK)
	// }

	util.JsonResponse(w, b.EmailsToInvite)
}

func resendInvitation(email string) {}
