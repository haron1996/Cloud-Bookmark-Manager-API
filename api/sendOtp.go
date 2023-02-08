package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-ozzo/ozzo-validation/is"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
	"github.com/kwandapchumba/go-bookmark-manager/mailjet"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

type sendOtpRequest struct {
	Email string `json:"email"`
}

func (s sendOtpRequest) Validate(requestValidationChan chan error) error {
	err := validation.ValidateStruct(&s,
		validation.Field(&s.Email, validation.Required.Error("email address is required"), is.Email.Error("email must be valid email address")),
	)

	requestValidationChan <- err

	return err
}

func (h *BaseHandler) SendOTP(w http.ResponseWriter, r *http.Request) {
	body := json.NewDecoder(r.Body)

	body.DisallowUnknownFields()

	var req sendOtpRequest

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

	name := strings.Split(req.Email, "@")[0]

	code := util.OTPGenerator()

	mail := mailjet.NewMail(req.Email, name, code)

	mail.SendEmailVificationMail()

	queries := sqlc.New(h.db)

	params := sqlc.NewEmailVerificationCodeParams{
		Code:   code,
		Email:  req.Email,
		Expiry: time.Now().UTC().Add(30 * time.Minute),
	}

	if _, err = queries.NewEmailVerificationCode(context.Background(), params); err != nil {
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	util.Response(w, "verification code has been sent", http.StatusOK)
}
