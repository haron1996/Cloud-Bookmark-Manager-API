package middleware

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/kwandapchumba/go-bookmark-manager/auth"
	"github.com/kwandapchumba/go-bookmark-manager/db/connection"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

/*
Middleware performs some specific function on the HTTP request or response at a specific stage in the HTTP pipeline before or after the user defined controller. Middleware is a design pattern to eloquently add cross cutting concerns like logging, handling authentication without having many code contact points.
*/
func ReturnVerifiedUserToken() func(next http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			payload, err := getAndVerifyToken(r)
			if err != nil {
				log.Println(err)
				util.Response(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if payload != nil {
				q := sqlc.New(connection.ConnectDB())

				account, err := q.GetAccount(context.Background(), payload.AccountID)
				if err != nil {
					var pgErr *pgconn.PgError

					if errors.As(err, &pgErr) {
						log.Println(pgErr)
						err := errors.New("something went wrong")
						util.Response(w, err.Error(), 500)
					} else if errors.Is(err, sql.ErrNoRows) {
						err := errors.New("user with id not found")
						log.Println(err)
						util.Response(w, err.Error(), http.StatusUnauthorized)
						return
					} else {
						log.Println(err)
						err := errors.New("something went wrong")
						util.Response(w, err.Error(), 500)
					}
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
