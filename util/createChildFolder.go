package util

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgconn"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
)

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

func CreateChildFolder(q *sqlc.Queries, w http.ResponseWriter, r *http.Request, folder_name, folder_id string, account_id int64) {
	parentFolder, err := q.GetFolder(r.Context(), folder_id)
	if err != nil {
		var pgErr *pgconn.PgError

		switch {
		case errors.As(err, &pgErr):
			log.Println(pgErr)
			Response(w, "something went wrong", http.StatusInternalServerError)
			return
		case errors.Is(err, sql.ErrNoRows):
			log.Println(sql.ErrNoRows.Error())
			Response(w, "not found", http.StatusNotFound)
			return
		default:
			log.Println(err)
			Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}
	}

	parentFolderPath := parentFolder.Path

	stringChan := make(chan string, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		RandomStringGenerator(stringChan)
	}()

	folderLabelChan := make(chan string, 1)

	wg.Add(1)

	go func() {
		defer wg.Done()

		GenFolderLabel(folderLabelChan)
	}()

	label := <-folderLabelChan

	path := strings.Join([]string{parentFolderPath, label}, ".")

	folderID := <-stringChan

	arg := sqlc.CreateFolderParams{
		FolderID:    folderID,
		FolderName:  folder_name,
		SubfolderOf: sql.NullString{String: parentFolder.FolderID, Valid: true},
		AccountID:   account_id,
		Path:        path,
		Label:       label,
	}

	folder, err := q.CreateFolder(r.Context(), arg)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Println(pgErr)
			Response(w, "something went wrong", http.StatusInternalServerError)
			return
		} else {
			log.Println(err)
			Response(w, "something went wrong", http.StatusInternalServerError)
			return
		}
	}

	wg.Wait()

	JsonResponse(w, newReturnedFolder(folder))
}
