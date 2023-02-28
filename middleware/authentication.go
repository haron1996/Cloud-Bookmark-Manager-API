package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/kwandapchumba/go-bookmark-manager/auth"
	dbutils "github.com/kwandapchumba/go-bookmark-manager/db/dbutils"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

/*
Middleware performs some specific function on the HTTP request or response at a specific stage in the HTTP pipeline before or after the user defined controller. Middleware is a design pattern to eloquently add cross cutting concerns like logging, handling authentication without having many code contact points.
*/

func AuthenticateRequest() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			payload, err := getAndVerifyToken(r)
			if err != nil {
				log.Println(err)
				util.Response(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if payload != nil {

				account, err := dbutils.ReturnAccount(context.Background(), payload.AccountID)
				if err != nil {
					log.Println(err)
					util.Response(w, errors.New("unauthorized").Error(), http.StatusUnauthorized)
					return
				}

				if payload.IssuedAt.Unix() != account.LastLogin.Unix() {
					err := errors.New("invalid token")
					log.Println(err)
					util.Response(w, err.Error(), http.StatusUnauthorized)
					return
				}

				ctx := context.WithValue(r.Context(), "payload", payload)

				next.ServeHTTP(w, r.WithContext(ctx))
			}
		}

		return http.HandlerFunc(fn)
	}
}

func getAndVerifyToken(r *http.Request) (*auth.PayLoad, error) {
	token := r.Header.Get("authorization")

	if token == "" {
		log.Println("token is empty!")
		return nil, errors.New("token is empty")
	}

	splitToken := strings.Split(token, "Bearer")

	if len(splitToken) != 2 {
		log.Println("bearer not in proper format")
		err := errors.New("bearer token is not in proper format")
		return nil, err
	}

	token = splitToken[1]

	payload, err := auth.VerifyToken(token)
	if err != nil {
		return nil, err
	}

	return payload, nil
}
