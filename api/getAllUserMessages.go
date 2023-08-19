package api

import (
	"log"
	"net/http"

	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

func (h *BaseHandler) GetAllUserMessages(w http.ResponseWriter, r *http.Request) {
	q := sqlc.New(h.db)

	messages, err := q.GetAllMessages(r.Context())
	if err != nil {
		log.Println(err)
		util.Response(w, "somenthing went wrong", http.StatusInternalServerError)
		return
	}

	util.JsonResponse(w, messages)
}
