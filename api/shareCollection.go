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
	b := r.Context().Value("shareCollectionRequest").(*middleware.ShareCollectionRequestBody).Body

	p := r.Context().Value("shareCollectionRequest").(*middleware.ShareCollectionRequestBody).PayLoad

	collectionName := r.Context().Value("shareCollectionRequest").(*middleware.ShareCollectionRequestBody).CollectionName

	// save invite details in db
	// share id
	// shared collection
	// access level
	// shared with email address
	// share expiry
	// shared by email addresss
	// send invite email to each email

	q := sqlc.New(h.db)

	// what if user owns the folder and yet wants to share with themselves? check if so
	// what if invited user already is a member of collection?

	inviterAccount, err := q.GetAccount(context.Background(), p.AccountID)
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

		if inviterAccount.Email == email {
			log.Println("user wants to invite themselves to a folder they own? abomination")
			util.Response(w, "you cannot share collection with yourself", http.StatusConflict)
			return
		}

		// TODO... check if email has an account associated with it already
		accountInvited, err := q.GetAccountByEmail(context.Background(), email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// email is not registered yet
				token := uuid.NewString()

				encodedToken := base64.StdEncoding.EncodeToString([]byte(token))

				inviteEpiry := time.Now().UTC().Add(72 * time.Hour) // 3 days

				params := sqlc.CreateInviteParams{
					SharedCollectionID:      b.CollectionID,
					CollectionSharedByName:  inviterAccount.Fullname,
					CollectionSharedByEmail: inviterAccount.Email,
					CollectionSharedWith:    email,
					MemberAccessLevel:       sqlc.AccessLevel(b.AccessLevel),
					InviteExpiry:            inviteEpiry, // 3 days
					InviteToken:             encodedToken,
				}

				invite, err := q.CreateInvite(context.Background(), params)
				if err != nil {

					var pgErr *pgconn.PgError

					if errors.As(err, &pgErr) {

						e := `duplicate key value violates unique constraint "invite_shared_collection_id_collection_shared_with_key"`
						if pgErr.Message == e {
							resendInvitation(email)
							return
						}

						log.Printf("could not create invite: pgErr: %v", pgErr)
						util.Response(w, "something went wrong", http.StatusInternalServerError)
						return
					}

					log.Println(err)
					util.Response(w, "something went wrong", http.StatusInternalServerError)
					return
				}

				mail := mailjet.NewInviteUserMail(inviterAccount.Fullname, inviterAccount.Email, invite.CollectionSharedWith, token, collectionName, inviteEpiry)

				mail.SendInviteUserEmail()

				util.JsonResponse(w, b.EmailsToInvite)

				return
			}

			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) {
				log.Printf("could not get account by email at shareCollection.go with pgErr: %v", pgErr)
				util.Response(w, "something went wrong", http.StatusInternalServerError)
				return
			}

			log.Printf("could not get account by email at shareCollection.go with pgErr: %v", pgErr)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}
		// if email has an account associated with it already, create collection member instance, send an email with a link to the shared collection

		// first this collection has already been shared with them before sharing
		collectionMemberInstance, err := q.GetCollectionMemberByCollectionAndMemberIDs(context.Background(), sqlc.GetCollectionMemberByCollectionAndMemberIDsParams{
			CollectionID: b.CollectionID,
			MemberID:     accountInvited.ID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// invited user is not an existing folder member hence invite them
				folder, err := q.GetFolder(context.Background(), b.CollectionID)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						log.Printf("no folder found.... shareCollection.go: %v", err)
						util.Response(w, "folder not found", http.StatusNoContent)
						return
					}

					var pgErr *pgconn.PgError

					if errors.As(err, &pgErr) {
						log.Printf("could not get folder at shareCollection.go with pgErr: %v", pgErr)
						util.Response(w, "something went wrong", http.StatusInternalServerError)
						return
					}

					log.Printf("could not get folder at shareCollection.go with err: %v", err)
					util.Response(w, "something went wrong", http.StatusInternalServerError)
					return
				}

				newCollectionMemberParams := sqlc.AddNewCollectionMemberParams{
					CollectionID:          folder.FolderID,
					MemberID:              accountInvited.ID,
					CollectionAccessLevel: sqlc.CollectionAccessLevel(b.AccessLevel),
				}

				_, err = q.AddNewCollectionMember(context.Background(), newCollectionMemberParams)
				if err != nil {
					var pgErr *pgconn.PgError

					if errors.As(err, &pgErr) {
						log.Printf("could not create new collecton member at shareCollection.go due to pgErr: %v", pgErr)
						util.Response(w, "something went wrong", http.StatusInternalServerError)
						return
					}

					log.Printf("could not create new collecton member at shareCollection.go due to err: %v", err)
					util.Response(w, "something went wrong", http.StatusInternalServerError)
					return
				}

				mail := mailjet.NewCollectionHasBeenSharedWithYou(accountInvited.Email, inviterAccount.Fullname, inviterAccount.Email, folder.FolderName, folder.FolderID)

				mail.SendACollectionHasBeenSharedWithYouMail()

				return
			}

			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) {
				log.Printf("shareCollection.go: coludl not get collection member instance, pgErr %v", pgErr)
				util.Response(w, "something went wrong", http.StatusInternalServerError)
				return
			}

			log.Printf("shareCollection.go: coludl not get collection member instance, err, %v", err)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		// account to be invited to folder already exists so...
		log.Printf("invited account is already a collection member: %v", collectionMemberInstance)

	}

	log.Printf("execution reached here? ... emails to invite = %v", b.EmailsToInvite)

	util.JsonResponse(w, b.EmailsToInvite)
}

func resendInvitation(email string) {}
