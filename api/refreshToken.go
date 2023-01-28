package api

import (
	"log"
	"net/http"
)

func (h *BaseHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	log.Println("issue a new token")
}
