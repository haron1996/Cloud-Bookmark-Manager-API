package api

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgconn"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

func (h *BaseHandler) GetCollection(w http.ResponseWriter, r *http.Request) {
	folder, err := sqlc.New(h.db).GetFolder(context.Background(), chi.URLParam(r, "collectionID"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err := errors.New("collection not found")
			log.Println(err)
			util.Response(w, err.Error(), http.StatusNotFound)
			return
		}

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Println(pgErr.Message)
			util.Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		log.Println(err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	util.JsonResponse(w, folder)
}
