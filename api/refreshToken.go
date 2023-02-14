package api

import (
	"context"
	"database/sql"
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

type refreshToken struct {
	RefreshToken string `json:"refresh_token"`
}

func (s refreshToken) Validate(reqValidationChan chan error) error {
	returnVal := validation.ValidateStruct(&s,
		validation.Field(&s.RefreshToken, validation.Required.Error("refresh token required")),
	)
	reqValidationChan <- returnVal

	return returnVal
}

func (h *BaseHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshTC, err := r.Cookie("refresh_token_cookie")
	if err != nil {
		log.Println(err)
		return
	}

	// body := json.NewDecoder(r.Body)

	// body.DisallowUnknownFields()

	// var req refreshToken

	// err = body.Decode(&req)
	// if err != nil {
	// 	if e, ok := err.(*json.SyntaxError); ok {
	// 		log.Printf("syntax error at byte offset %d", e.Offset)
	// 		util.Response(w, internalServerError, http.StatusInternalServerError)
	// 		return
	// 	} else {
	// 		log.Printf("error decoding request body to struct: %v", err)
	// 		util.Response(w, badRequest, http.StatusBadRequest)
	// 		return
	// 	}
	// }

	// reqValidationChan := make(chan error, 1)

	// var wg sync.WaitGroup

	// wg.Add(1)

	// go func() {
	// 	defer wg.Done()

	// 	req.Validate(reqValidationChan)
	// }()

	// requestValidationErr := <-reqValidationChan
	// if requestValidationErr != nil {
	// 	if e, ok := requestValidationErr.(validation.InternalError); ok {
	// 		log.Println(e)
	// 		util.Response(w, internalServerError, http.StatusInternalServerError)
	// 		return
	// 	} else {
	// 		log.Println(requestValidationErr)
	// 		util.Response(w, requestValidationErr.Error(), http.StatusInternalServerError)
	// 		return
	// 	}
	// }

	// wg.Wait()

	payload, err := auth.VerifyToken(refreshTC.Value)
	if err != nil {
		util.Response(w, err.Error(), http.StatusUnauthorized)
		return
	}

	queries := sqlc.New(h.db)

	account, err := queries.GetAccount(context.Background(), payload.AccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			util.Response(w, "account not found", http.StatusUnauthorized)
			return
		}

		util.Response(w, "something went wrong", http.StatusInternalServerError)
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
		Name:     "refresh_token_cookie",
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

	_, err = queries.CreateAccountSession(context.Background(), createAccountSessionParams)
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

	account, err = queries.GetAccount(context.Background(), account.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			util.Response(w, "account not found", http.StatusUnauthorized)
			return
		}

		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	res := newsession(account, accessToken, refreshToken, accessTokenPayload.Expiry)

	log.Printf("res: %v", res)

	util.JsonResponse(w, res)
}
