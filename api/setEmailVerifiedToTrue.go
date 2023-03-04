package api

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
)

func (h *BaseHandler) SetEmailVerifiedToTrue(w http.ResponseWriter, r *http.Request) {
	q := sqlc.New(h.db)

	err := q.UpdateAccountEmailVerificationStatus(context.Background(), chi.URLParam(r, "email"))
	if err != nil {
		panic(err)
	}
}
