package api

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgconn"
	"github.com/kwandapchumba/go-bookmark-manager/auth"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

func (h *BaseHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("refreshTokenCookie")
	if err != nil {
		log.Println(err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	payload, err := auth.VerifyToken(c.Value)
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

	res := newsession(account, accessToken, refreshToken)

	util.JsonResponse(w, res)
}
