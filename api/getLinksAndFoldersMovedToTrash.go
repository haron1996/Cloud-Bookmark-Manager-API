package api

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgconn"
	"github.com/kwandapchumba/go-bookmark-manager/auth"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

type response struct {
	Folders []sqlc.Folder `json:"folders"`
	Links   []sqlc.Link   `json:"links"`
}

func newRes(f []sqlc.Folder, l []sqlc.Link) *response {
	return &response{
		Folders: f,
		Links:   l,
	}
}

func (h *BaseHandler) GetFoldersAndLinksMovedToTrash(w http.ResponseWriter, r *http.Request) {
	accountID, err := strconv.Atoi(chi.URLParam(r, "accountID"))
	if err != nil {
		log.Println(err)
		ErrorInternalServerError(w, err)
		return
	}

	payload := r.Context().Value("payload").(*auth.PayLoad)

	if int64(accountID) != payload.AccountID {
		err := errors.New("account ids do not match")
		log.Println(err.Error())
		ErrorInternalServerError(w, err)
		return
	}

	q := sqlc.New(h.db)

	var folders []sqlc.Folder

	var links []sqlc.Link

	folders, err = q.GetFoldersMovedToTrash(r.Context(), payload.AccountID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Println(pgErr)
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		} else if errors.Is(err, sql.ErrNoRows) {
			folders = []sqlc.Folder{}
		} else {
			log.Println(err)
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		}
	}
	links, err = q.GetLinksMovedToTrash(r.Context(), payload.AccountID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Println(pgErr)
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		} else if errors.Is(err, sql.ErrNoRows) {
			links = []sqlc.Link{}
		} else {
			log.Println(err)
			util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
			return
		}
	}

	res := newRes(folders, links)

	util.JsonResponse(w, res)
}
