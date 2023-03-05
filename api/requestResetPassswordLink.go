package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

type requestResetPasswordLinkRequest struct {
	Email string `json:"email"`
}

func (r requestResetPasswordLinkRequest) validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Email, is.Email),
	)
}

func (h *BaseHandler) RequestResetPasswordLink(w http.ResponseWriter, r *http.Request) {
	body := json.NewDecoder(r.Body)

	body.DisallowUnknownFields()

	var req requestResetPasswordLinkRequest

	err := body.Decode(&req)
	if err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("failed to decode request body with err: %v", e)
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		} else {
			log.Printf("failed to decode request body with err: %v", err)
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		}
	}

	err = req.validate()
	if err != nil {
		log.Printf("request validation error: %v", err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	q := sqlc.New(h.db)

	account, err := q.GetAccountByEmail(context.Background(), req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// account does not exist
			// send an account does not exist email with link to sign up
		} else {
			log.Printf("could not get account by email: %v", err)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}
	}

	log.Println(account)
}
