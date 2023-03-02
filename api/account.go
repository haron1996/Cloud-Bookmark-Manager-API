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

type signUp struct {
	FullName     string `json:"full_name"`
	EmailAddress string `json:"email_address"`
	Password     string `json:"password"`
}

type session struct {
	Account      sqlc.Account
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
}

func newsession(account sqlc.Account, accessToken, refreshToken string, expiry time.Time) session {
	return session{
		Account:      account,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiry,
	}
}

func (s signUp) Validate(requestValidationChan chan error) error {
	err := validation.ValidateStruct(&s,
		// Name cannot be empty, and the length must be between 5 and 20.
		validation.Field(&s.FullName, validation.Required.Error("name required"), validation.Length(1, 255).Error("name must be between 1 and 255 characters long")),
		// Emails are optional, but if given must be valid.
		validation.Field(&s.EmailAddress, validation.Required.Error("email address is required"), is.Email.Error("email must be valid email address")),
		validation.Field(&s.Password, validation.Required.Error("password is required")),
	)

	requestValidationChan <- err

	return err
}

func (h *BaseHandler) NewAccount(w http.ResponseWriter, r *http.Request) {
	body := json.NewDecoder(r.Body)

	body.DisallowUnknownFields()

	var req signUp

	if err := body.Decode(&req); err != nil {
		log.Printf("failed to decode request with error %v", err)
		ErrorInvalidRequest(w, err)
		return
	}

	requestValidationChan := make(chan error, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		req.Validate(requestValidationChan)
	}()

	validationErr := <-requestValidationChan
	if validationErr != nil {
		ErrorInvalidRequest(w, validationErr)
		return
	}

	wg.Wait()

	q := sqlc.New(h.db)

	emailExists, err := q.EmailExists(context.Background(), req.EmailAddress)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Println(pgErr)
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		} else {
			log.Println(err)
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		}
	}

	if emailExists {
		log.Println("email already exists")
		util.Response(w, errors.New("email already exists").Error(), http.StatusConflict)
		return
	}

	var p string

	p, err = util.HashPassword(req.Password)
	if err != nil {
		log.Printf("failed to hash password with error: %s", err)
		util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
		return
	}

	arg := sqlc.NewAccountParams{
		Fullname:        req.FullName,
		Email:           req.EmailAddress,
		AccountPassword: p,
	}

	account, err := q.NewAccount(context.Background(), arg)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Println(pgErr)
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		} else {
			log.Println(err)
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		}
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

// sign up with google

type continueWithGoogle struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

func (c continueWithGoogle) Validate(errChan chan error) error {
	err := validation.ValidateStruct(&c,
		validation.Field(&c.Name, validation.Required.Error("name required")),
		validation.Field(&c.Email, validation.Required.Error("email required"), is.Email),
		validation.Field(&c.Picture, validation.Required.Error("profile picture required"), is.URL),
	)

	errChan <- err

	return err
}

func (h *BaseHandler) ContinueWithGoogle(w http.ResponseWriter, r *http.Request) {
	rBody := json.NewDecoder(r.Body)

	rBody.DisallowUnknownFields()

	var req continueWithGoogle

	if err := rBody.Decode(&req); err != nil {
		ErrorDecodingRequest(w, err)
		return
	}

	errChan := make(chan error, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		req.Validate(errChan)
	}()

	wg.Wait()

	q := sqlc.New(h.db)

	email := req.Email

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Println(err)
		return
	}

	existingAccount, err := q.GetAccountByEmail(context.Background(), email)
	if err != nil {
		var pgErr *pgconn.PgError

		switch {
		case errors.As(err, &pgErr):
			ErrorInternalServerError(w, pgErr)
			return
		case errors.Is(err, sql.ErrNoRows):
			createAccountParams := sqlc.NewAccountParams{
				Fullname: req.Name,
				Email:    req.Email,
				// Picture:  req.Picture,
			}

			createAccount(createAccountParams, q, w, h, config)

			return
		default:
			ErrorInternalServerError(w, err)
			return
		}
	}

	if existingAccount != (sqlc.Account{}) {
		loginUser(existingAccount, w, h, config)
	}
}

func createAccount(arg sqlc.NewAccountParams, q *sqlc.Queries, w http.ResponseWriter, h *BaseHandler, config util.Config) {
	account, err := q.NewAccount(context.Background(), arg)
	if err != nil {
		var pgErr *pgconn.PgError

		switch {
		case errors.As(err, &pgErr):
			ErrorInternalServerError(w, pgErr)
			return
		default:
			ErrorInternalServerError(w, err)
			return
		}
	}

	if err := q.UpdateAccountEmailVerificationStatus(context.Background(), arg.Email); err != nil {
		log.Println(err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	loginUser(account, w, h, config)
}

func loginUser(account sqlc.Account, w http.ResponseWriter, h *BaseHandler, config util.Config) {
	accessToken, accessTokenPayload, err := auth.CreateToken(account.ID, time.Now().UTC(), config.Access_Token_Duration)
	if err != nil {
		ErrorInternalServerError(w, err)
		return
	}

	q := sqlc.New(h.db)

	// updateLastLoginParams := sqlc.UpdateLastLoginParams{
	// 	LastLogin: accessTokenPayload.IssuedAt,
	// 	ID:        accessTokenPayload.AccountID,
	// }

	// if _, err := q.UpdateLastLogin(context.Background(), updateLastLoginParams); err != nil {
	// 	ErrorInternalServerError(w, err)
	// 	return
	// }

	refreshToken, refreshTokenPayload, err := auth.CreateToken(account.ID, accessTokenPayload.IssuedAt, config.Refresh_Token_Duration)
	if err != nil {
		ErrorInternalServerError(w, err)
		return
	}

	createAccountSessionParams := sqlc.CreateAccountSessionParams{
		RefreshTokenID: refreshTokenPayload.ID,
		AccountID:      account.ID,
		IssuedAt:       refreshTokenPayload.IssuedAt,
		Expiry:         refreshTokenPayload.Expiry,
		UserAgent:      "",
		ClientIp:       "",
	}

	if _, err := q.CreateAccountSession(context.Background(), createAccountSessionParams); err != nil {
		ErrorInternalServerError(w, err)
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

	newAccount, err := q.GetAccount(context.Background(), account.ID)
	if err != nil {
		ErrorInternalServerError(w, err)
		return
	}

	res := newsession(newAccount, accessToken, refreshToken, accessTokenPayload.Expiry)

	util.JsonResponse(w, res)
}

func (h *BaseHandler) GetAllAccounts(w http.ResponseWriter, r *http.Request) {
	q := sqlc.New(h.db)

	accounts, err := q.GetAllAccounts(context.Background())
	if err != nil {
		log.Println(err)
		util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
		return
	}

	if len(accounts) == 0 {
		util.Response(w, errors.New("no accounts found").Error(), http.StatusNotFound)
		return
	}

	util.JsonResponse(w, accounts)
}
