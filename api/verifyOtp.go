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
	"github.com/kwandapchumba/go-bookmark-manager/auth"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

type otpVerification struct {
	Code  string `json:"code"`
	Email string `json:"email"`
}

func (s otpVerification) Validate(requestValidationChan chan error) error {
	err := validation.ValidateStruct(&s,
		validation.Field(&s.Email, validation.Required.Error("email address is required"), is.Email.Error("email must be valid email address")),
		validation.Field(&s.Code, validation.Required.Error("verification code is required"), validation.Length(6, 6).Error("verification code must be 6 characters long")),
	)

	requestValidationChan <- err

	return err
}

func (h *BaseHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	body := json.NewDecoder(r.Body)

	body.DisallowUnknownFields()

	var req otpVerification

	err := body.Decode(&req)
	if err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		} else {
			log.Printf("error decoding request body to struct: %v", err)
			util.Response(w, badRequest, http.StatusBadRequest)
			return
		}
	}

	reqValidationChan := make(chan error, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		req.Validate(reqValidationChan)
	}()

	requestValidationErr := <-reqValidationChan
	if requestValidationErr != nil {
		if e, ok := requestValidationErr.(validation.InternalError); ok {
			log.Println(e)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		} else {
			log.Println(requestValidationErr)
			util.Response(w, requestValidationErr.Error(), http.StatusInternalServerError)
			return
		}
	}

	wg.Wait()

	queries := sqlc.New(h.db)

	otp, err := queries.GetOtp(context.Background(), req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			util.Response(w, "otp was not found", http.StatusNotFound)
			return
		} else {
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}
	}

	if otp.Code != req.Code {
		util.Response(w, "invalid code", http.StatusUnauthorized)
		return
	}

	if time.Now().UTC().After(otp.Expiry) {
		util.Response(w, "code has expired", http.StatusUnauthorized)
		return
	}

	if err = queries.UpdateAccountEmailVerificationStatus(context.Background(), otp.Email); err != nil {
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	account, err := queries.GetAccountByEmail(context.Background(), otp.Email)
	if err != nil {
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Printf("failed to load config file with err: %v", err)
		util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
		return
	}

	if err = queries.DeleteEmailVerificationCode(context.Background(), account.Email); err != nil {
		util.Response(w, "something went wrong", http.StatusInternalServerError)
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

	_, err = queries.CreateAccountSession(context.Background(), createAccountSessionParams)
	if err != nil {
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	// refreshTokenCookie := http.Cookie{
	// 	Name:     "refresh_token",
	// 	Value:    refreshToken,
	// 	Path:     "/public/refresh_token",
	// 	Expires:  refreshTokenPayload.Expiry,
	// 	Secure:   true,
	// 	SameSite: http.SameSite(http.SameSiteStrictMode),
	// 	HttpOnly: true,
	// }

	// http.SetCookie(w, &refreshTokenCookie)

	account, err = queries.GetAccountByEmail(context.Background(), otp.Email)
	if err != nil {
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	res := newsession(account, accessToken, refreshToken, accessTokenPayload.Expiry)

	util.JsonResponse(w, res)
}
