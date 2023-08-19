package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/jackc/pgconn"
	"github.com/kwandapchumba/go-bookmark-manager/auth"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

type token struct {
	Token string `json:"token"`
}

func (t token) Validate(requestValidationChan chan error) error {
	err := validation.ValidateStruct(&t,
		validation.Field(&t.Token, validation.Required.Error("token is required")),
	)

	requestValidationChan <- err

	return err
}

func (h *BaseHandler) CheckIfIsAuthenticated(w http.ResponseWriter, r *http.Request) {
	data := json.NewDecoder(r.Body)

	data.DisallowUnknownFields()

	var req token

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

	payload, err := auth.VerifyToken(req.Token)
	if err != nil {
		log.Println(err)
		util.Response(w, err.Error(), http.StatusUnauthorized)
		return
	}

	q := sqlc.New(h.db)

	account, err := q.GetAccount(r.Context(), int64(payload.AccountID))
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Printf("failed to get account with err: %v", pgErr)
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		} else if errors.Is(err, sql.ErrNoRows) {
			log.Println("account not found")
			util.Response(w, errors.New("account not found").Error(), http.StatusUnauthorized)
			return
		} else {
			log.Printf("failed to get account with err: %v", err)
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		}
	}

	if account.LastLogin.Unix() != payload.IssuedAt.Unix() {
		util.Response(w, "user not logged in", http.StatusUnauthorized)
		return
	}

	util.Response(w, "user logged in", http.StatusOK)
}
