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
	"github.com/jackc/pgconn"
	"github.com/kwandapchumba/go-bookmark-manager/auth"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

type acceptInviteRequest struct {
	Token    string `json:"token"`
	Fullname string `json:"fullname"`
	Password string `json:"password"`
}

func (a acceptInviteRequest) validate() error {
	return validation.ValidateStruct(&a,
		// Name cannot be empty, and the length must be between 5 and 20.
		validation.Field(&a.Fullname, validation.Required.Error("name required"), validation.Length(1, 255).Error("name must be between 1 and 255 characters long")),
		// Emails are optional, but if given must be valid.
		validation.Field(&a.Token, validation.Required.Error("token is required"), validation.Length(1, 1000).Error("token must be at least one character long")),
		validation.Field(&a.Password, validation.Required.Error("password is required"), validation.Length(6, 1000).Error("password must be at least 6 characters long")),
	)
}

func (h *BaseHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	body := json.NewDecoder(r.Body)

	body.DisallowUnknownFields()

	var req acceptInviteRequest

	if err := body.Decode(&req); err != nil {
		var syntaxErr *json.SyntaxError

		if errors.As(err, &syntaxErr) {
			log.Printf("failed to decode request body with err: %v", syntaxErr)
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		} else {
			log.Printf("failed to decode request body with err: %v", err)
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := req.validate(); err != nil {
		log.Printf("bad request: %v", err)
		util.Response(w, err.Error(), http.StatusBadRequest)
		return
	}

	q := sqlc.New(h.db)

	token, err := q.GetInviteByToken(r.Context(), base64.StdEncoding.EncodeToString([]byte(req.Token)))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err := errors.New("invite token not found")
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

	if time.Now().UTC().After(token.InviteExpiry) {
		err := errors.New("invite token is expired")
		log.Println(err)
		util.Response(w, err.Error(), http.StatusUnauthorized)
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		log.Printf("failed to hash password at acceptInvite.go: %v", err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	account, err := q.NewAccount(r.Context(), sqlc.NewAccountParams{
		Fullname:        req.Fullname,
		Email:           token.CollectionSharedWith,
		AccountPassword: hashedPassword,
	})
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Printf("failed to create new account for invited user: %v", pgErr)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		log.Printf("failed to create new account for invited user: %v", err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	if err := q.UpdateAccountEmailVerificationStatus(r.Context(), account.Email); err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Printf("could not update email verification status at accceptInvite.go: %v", pgErr)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		log.Printf("could not update email verification status at accceptInvite.go: %v", err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	collection, err := q.GetFolder(r.Context(), token.SharedCollectionID)
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

	if _, err := q.AddNewCollectionMember(r.Context(), sqlc.AddNewCollectionMemberParams{
		CollectionID:          collection.FolderID,
		MemberID:              account.ID,
		CollectionAccessLevel: sqlc.CollectionAccessLevel(token.MemberAccessLevel),
	}); err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Printf("could not add new collection member at acceptInvite.go: %v", pgErr)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		log.Printf("could not add new collection member at acceptInvite.go: %v", err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	accessToken, refreshToken, refreshTokenCookie := auth.LoginUser(account, q, r.Context())

	account, err = q.GetAccount(r.Context(), account.ID)
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

	res := newsession(account, accessToken, refreshToken)

	if err := q.DeleteInvite(r.Context(), base64.StdEncoding.EncodeToString([]byte(req.Token))); err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Printf("could not delete invite token at acceptInvite.go: %v", pgErr)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		log.Printf("could not delete invite token at acceptInvite.go: %v", err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &refreshTokenCookie)

	util.JsonResponse(w, res)
}
