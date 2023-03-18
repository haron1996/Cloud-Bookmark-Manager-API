package auth

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgconn"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

func LoginUser(account sqlc.Account, q *sqlc.Queries) (string, string, http.Cookie) {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatalf("could not load config at login.go: %v", err)
		return "", "", http.Cookie{}
	}

	accessToken, accessTokenPayload, err := CreateToken(account.ID, time.Now().UTC(), config.Access_Token_Duration)
	if err != nil {
		log.Fatalf("could not create access token at login.go: %v", err)
		return "", "", http.Cookie{}
	}

	refreshToken, refreshTokenPayload, err := CreateToken(account.ID, accessTokenPayload.IssuedAt, config.Refresh_Token_Duration)
	if err != nil {
		log.Fatalf("could not create refresh token at login.go: %v", err)
		return "", "", http.Cookie{}
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
			log.Fatalf("could not create account session at login.go: %v", pgErr)
			return "", "", http.Cookie{}
		}

		log.Fatalf("could not create account session at auth/login.go: %v", err)
		return "", "", http.Cookie{}
	}

	return accessToken, refreshToken, refreshTokenCookie
}
