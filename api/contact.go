package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kwandapchumba/go-bookmark-manager/auth"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
	"github.com/kwandapchumba/go-bookmark-manager/mailjet"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

type contactSupportRequest struct {
	Message string `json:"message"`
}

func (s contactSupportRequest) validate(reqValidationChan chan error) error {
	returnVal := validation.ValidateStruct(&s,
		validation.Field(&s.Message, validation.Required.When(s.Message == "").Error("message is required")),
	)
	reqValidationChan <- returnVal
	return returnVal
}

func (h *BaseHandler) ContactSupport(w http.ResponseWriter, r *http.Request) {
	body := json.NewDecoder(r.Body)

	body.DisallowUnknownFields()

	var req contactSupportRequest

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

		req.validate(reqValidationChan)
	}()

	wg.Wait()

	payload := r.Context().Value("payload").(*auth.PayLoad)

	queries := sqlc.New(h.db)

	account, err := queries.GetAccount(context.Background(), payload.AccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println(err)
			util.Response(w, "account not found", http.StatusUnauthorized)
			return
		} else {
			log.Println(err)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}
	}

	newMailRequest := &mailjet.EmailSupportRequest{
		FromEmail: account.Email,
		FromName:  account.Fullname,
		Subject:   req.Message,
		TextPart:  req.Message,
	}

	if err := newMailRequest.EmailSupport(); err != nil {
		log.Println(err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	newMessageParams := sqlc.NewMessageParams{
		Account:     account.ID,
		MessageBody: req.Message,
	}

	if _, err := queries.NewMessage(context.Background(), newMessageParams); err != nil {
		log.Println(err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	util.Response(w, "message sent", http.StatusOK)
}
