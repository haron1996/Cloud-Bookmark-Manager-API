package api

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kwandapchumba/go-bookmark-manager/auth"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

type getLinksAndFoldersResponse struct {
	Folders []returnFolder `json:"folders"`
	Links   []sqlc.Link    `json:"links"`
}

func newResponse(folders []returnFolder, links []sqlc.Link) *getLinksAndFoldersResponse {
	return &getLinksAndFoldersResponse{
		Folders: folders,
		Links:   links,
	}
}

type returnFolder struct {
	FolderID        string         `json:"folder_id"`
	AccountID       int64          `json:"account_id"`
	FolderName      string         `json:"folder_name"`
	Path            string         `json:"path"`
	Label           string         `json:"label"`
	Starred         bool           `json:"starred"`
	FolderCreatedAt string         `json:"folder_created_at"`
	FolderUpdatedAt string         `json:"folder_updated_at"`
	SubfolderOf     sql.NullString `json:"subfolder_of"`
	FolderDeletedAt sql.NullTime   `json:"folder_deleted_at"`
}

func newReturnedFolder(f sqlc.Folder) returnFolder {
	return returnFolder{
		FolderID:        f.FolderID,
		AccountID:       f.AccountID,
		FolderName:      f.FolderName,
		Path:            f.Path,
		Label:           f.Label,
		Starred:         f.Starred,
		FolderCreatedAt: strings.Join(strings.Split(strings.Split(f.FolderUpdatedAt.Local().Format(time.RFC3339), "T")[0], "-"), "/"),
		FolderUpdatedAt: strings.Join(strings.Split(strings.Split(f.FolderCreatedAt.Local().Format(time.RFC3339), "T")[0], "-"), "/"),
		SubfolderOf:     f.SubfolderOf,
		FolderDeletedAt: f.FolderDeletedAt,
	}
}

func (h *BaseHandler) GetLinksAndFolders(w http.ResponseWriter, r *http.Request) {
	folderID := chi.URLParam(r, "folderID")

	accountID := chi.URLParam(r, "accountID")

	payload := r.Context().Value("payload").(*auth.PayLoad)

	account_id, err := strconv.Atoi(accountID)
	if err != nil {
		ErrorInternalServerError(w, err)
		return
	}

	if int64(account_id) != payload.AccountID {
		log.Println("account IDs do not match")
		util.Response(w, errors.New("account IDs do not match").Error(), http.StatusUnauthorized)
		return
	}

	if folderID == "null" {
		getRootNodesAndLinks(h, payload.AccountID, w)
	} else {
		getFolderNodesAndLinks(h, payload.AccountID, folderID, w)
	}
}

func getRootNodesAndLinks(h *BaseHandler, accountID int64, w http.ResponseWriter) {
	q := sqlc.New(h.db)

	fs, err := q.GetRootFolders(context.Background(), accountID)
	if err != nil {
		log.Println(err)
		util.Response(w, err.Error(), 500)
		return
	}

	// nodes, err := q.GetRootNodes(context.Background(), accountID)
	// if err != nil {
	// 	ErrorInternalServerError(w, err)
	// 	return
	// }

	var rfs []returnFolder

	for _, f := range fs {
		folder := newReturnedFolder(f)
		rfs = append(rfs, folder)
	}

	links, err := q.GetRootLinks(context.Background(), accountID)
	if err != nil {
		ErrorInternalServerError(w, err)
		return
	}

	res := newResponse(rfs, links)

	util.JsonResponse(w, res)
}

func getFolderNodesAndLinks(h *BaseHandler, accountID int64, folderID string, w http.ResponseWriter) {
	q := sqlc.New(h.db)

	getFolderNodesParams := sqlc.GetFolderNodesParams{
		AccountID:   accountID,
		SubfolderOf: sql.NullString{String: folderID, Valid: true},
	}

	nodes, err := q.GetFolderNodes(context.Background(), getFolderNodesParams)
	if err != nil {
		ErrorInternalServerError(w, err)
		return
	}

	var rfs []returnFolder

	for _, n := range nodes {
		f := newReturnedFolder(n)
		rfs = append(rfs, f)
	}

	getFolderLinksParams := sqlc.GetFolderLinksParams{
		AccountID: accountID,
		FolderID:  sql.NullString{String: folderID, Valid: true},
	}

	links, err := q.GetFolderLinks(context.Background(), getFolderLinksParams)
	if err != nil {
		ErrorInternalServerError(w, err)
		return
	}

	res := newResponse(rfs, links)

	util.JsonResponse(w, res)
}
