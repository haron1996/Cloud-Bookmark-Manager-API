package api

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgconn"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

type getCollectionAndInviterNamesResponse struct {
	CollectionName string `json:"collection_name"`
	InviterName    string `json:"inviter_name"`
}

func newGetCollectionAndInviterNamesResponse(collectionName, inviterName string) *getCollectionAndInviterNamesResponse {
	return &getCollectionAndInviterNamesResponse{
		CollectionName: collectionName,
		InviterName:    inviterName,
	}
}

func (h *BaseHandler) GetCollectionAndInviterNames(w http.ResponseWriter, r *http.Request) {
	q := sqlc.New(h.db)

	token, err := q.GetInviteByToken(context.Background(), base64.StdEncoding.EncodeToString([]byte(chi.URLParam(r, "inviteToken"))))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err := errors.New("invite not found")
			log.Println(err)
			util.Response(w, "invite not found", http.StatusNotFound)
			return
		}

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Println(pgErr.Message)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		log.Println(err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	collection, err := q.GetFolder(context.Background(), token.SharedCollectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err := errors.New("collection not found")
			log.Println(err)
			util.Response(w, err.Error(), http.StatusNotFound)
			return
		}

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Println(pgErr.Message)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		log.Println(err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	account, err := q.GetAccountByEmail(context.Background(), token.CollectionSharedByEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err := errors.New("account not found")
			log.Println(err)
			util.Response(w, err.Error(), http.StatusNotFound)
			return
		}

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Println(pgErr.Message)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		log.Println(err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	if token.CollectionSharedByName != account.Fullname {
		err := errors.New("account name and invite token name do not match")
		log.Println(err)
		util.Response(w, err.Error(), http.StatusUnauthorized)
		return
	}

	res := newGetCollectionAndInviterNamesResponse(collection.FolderName, account.Fullname)

	util.JsonResponse(w, res)
}
