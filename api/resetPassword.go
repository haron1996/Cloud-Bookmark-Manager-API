package api

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

type createNewPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (c createNewPasswordRequest) validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Token, validation.Required.Error("token is required"), validation.Length(1, 255).Error("name must be between 1 and 255 characters long")),
		validation.Field(&c.Password, validation.Required.Error("password is required"), validation.Length(6, 1000).Error("password must be at least 6 characters long")),
	)
}

func (h *BaseHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	body := json.NewDecoder(r.Body)

	body.DisallowUnknownFields()

	var req createNewPasswordRequest

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
		util.Response(w, err.Error(), http.StatusBadRequest)
		return
	}

	q := sqlc.New(h.db)

	tokenHash := base64.StdEncoding.EncodeToString([]byte(req.Token))

	token, err := q.GetPasswordResetToken(r.Context(), tokenHash)
	if err != nil {
		log.Println(err)
		if errors.Is(err, sql.ErrNoRows) {
			// password reset token not found
			util.Response(w, "invalid password reset token", http.StatusUnauthorized)
			return
		} else {
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}
	}

	err = q.DeletePasswordResetToken(r.Context(), tokenHash)
	if err != nil {
		log.Println(err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	if time.Now().UTC().After(token.TokenExpiry) {
		// token has expired
		util.Response(w, "expired password reset token", http.StatusUnauthorized)
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		log.Println(err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	updatePasswordParams := sqlc.UpdatePasswordParams{
		AccountPassword: hashedPassword,
		ID:              token.AccountID,
	}

	err = q.UpdatePassword(r.Context(), updatePasswordParams)
	if err != nil {
		log.Println(err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	util.Response(w, "password update successfully", http.StatusOK)
}
