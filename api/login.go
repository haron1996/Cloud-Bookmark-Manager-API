package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-ozzo/ozzo-validation/is"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/jackc/pgconn"
	"github.com/kwandapchumba/go-bookmark-manager/auth"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

type signin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s signin) Validate(requestValidationChan chan error) error {
	err := validation.ValidateStruct(&s,
		validation.Field(&s.Email, validation.Required.Error("email address is required"), is.Email.Error("email must be valid email address")),
		validation.Field(&s.Password, validation.Required.Error("password is required"), validation.Length(6, 1000).Error("password must be at least 6 characters long")),
	)

	requestValidationChan <- err

	return err
}

func (h *BaseHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	log.Printf("request body: %v", r.Body)
	data := json.NewDecoder(r.Body)

	data.DisallowUnknownFields()

	var req signin

	if err := data.Decode(&req); err != nil {
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

	requestValidationChan := make(chan error, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		req.Validate(requestValidationChan)
	}()

	wg.Wait()

	err := <-requestValidationChan
	if err != nil {
		if e, ok := err.(validation.InternalError); ok {
			// an internal error happened
			log.Printf("an internal server error occured while validation request body: %v", e.InternalError())
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		} else {
			log.Printf("invalid request: %v", err)
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		}
	}

	q := sqlc.New(h.db)

	account, err := q.GetAccountByEmail(context.Background(), req.Email)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			log.Println(err)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		} else if errors.Is(err, sql.ErrNoRows) {
			log.Println("email not found")
			util.Response(w, "invalid email", http.StatusUnauthorized)
			return
		} else {
			log.Println(err)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}
	}

	if !util.CompareHash(req.Password, account.AccountPassword) {
		log.Println("invalid password")
		util.Response(w, "invalid password", http.StatusUnauthorized)
		return
	}

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Printf("failed to load config file with err: %v", err)
		util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
		return
	}

	accessToken, accessTokenPayload, err := auth.CreateToken(account.ID, time.Now(), config.Access_Token_Duration)
	if err != nil {
		log.Printf("failed to create access token with err: %v", err)
		util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
		return
	}

	refreshToken, refreshTokenPayload, err := auth.CreateToken(account.ID, accessTokenPayload.IssuedAt, config.Refresh_Token_Duration)
	if err != nil {
		log.Printf("failed to create refresh token with err: %v", err)
		util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
		return
	}

	refreshTokenCookie := http.Cookie{
		Name:     "refreshTokenCookie",
		Value:    refreshToken,
		Path:     "/",
		Expires:  refreshTokenPayload.Expiry,
		Secure:   true,
		SameSite: http.SameSite(http.SameSiteStrictMode),
		HttpOnly: true,
	}

	http.SetCookie(w, &refreshTokenCookie)

	createAccountSessionParams := sqlc.CreateAccountSessionParams{
		RefreshTokenID: refreshTokenPayload.ID,
		AccountID:      account.ID,
		IssuedAt:       refreshTokenPayload.IssuedAt,
		Expiry:         refreshTokenPayload.Expiry,
		UserAgent:      "",
		ClientIp:       "",
	}

	_, err = q.CreateAccountSession(context.Background(), createAccountSessionParams)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Printf("failed co create session with err: %s", pgErr.Message)
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		} else {
			log.Printf("failed to create session with error: %s", err.Error())
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		}
	}

	account, err = q.GetAccount(context.Background(), account.ID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Printf("failed to get created account with err: %v", pgErr)
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		} else if errors.Is(err, sql.ErrNoRows) {
			log.Println("account not found")
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		} else {
			log.Printf("failed to get created account with err: %v", err)
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		}
	}

	res := newsession(account, accessToken, refreshToken, accessTokenPayload.Expiry)

	util.JsonResponse(w, res)
}
