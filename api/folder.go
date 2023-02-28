package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/jackc/pgconn"
	"github.com/kwandapchumba/go-bookmark-manager/auth"
	"github.com/kwandapchumba/go-bookmark-manager/db/sqlc"
	"github.com/kwandapchumba/go-bookmark-manager/middleware"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

// CREATE FOLDER
type createRootFolderRequest struct {
	FolderName string `json:"folder_name"`
	FolderID   string `json:"parent_folder_id"`
}

func (s createRootFolderRequest) validate(reqValidationChan chan error) error {
	returnVal := validation.ValidateStruct(&s,
		validation.Field(&s.FolderName, validation.Required.When(s.FolderName == "").Error("Folder name is required"), validation.Match(regexp.MustCompile("^[^?[\\]{}|\\\\`./!@#$%^&*()_-]+$")).Error("Folder name must not have special characters"), validation.Length(1, 100).Error("Folder name must be at least 1 character long")),
		// validation.Field(&s.FolderID, validation.Length(33, 33).Error("folder id must be 33 characters long")),
	)
	reqValidationChan <- returnVal
	return returnVal
}

func (h *BaseHandler) CreateFolder(w http.ResponseWriter, r *http.Request) {
	// log.Printf("authorized payload: %v", r.Context().Value("authorizedPayload").(*auth.PayLoad))

	requestBody := r.Context().Value("myValues").(*middleware.RequestBody).Body

	authorizedPayload := r.Context().Value("myValues").(*middleware.RequestBody).PayLoad

	// log.Println(rB)

	// rBody := json.NewDecoder(r.Body)

	// rBody.DisallowUnknownFields()

	// var req createRootFolderRequest

	// err := rBody.Decode(&req)
	// if err != nil {
	// 	if e, ok := err.(*json.SyntaxError); ok {
	// 		log.Printf("syntax error at byte offset %d", e.Offset)
	// 		util.Response(w, internalServerError, http.StatusInternalServerError)
	// 		return
	// 	} else {
	// 		log.Printf("error decoding request body to struct: %v", err)
	// 		util.Response(w, badRequest, http.StatusBadRequest)
	// 		return
	// 	}
	// }

	// reqValidationChan := make(chan error, 1)

	// var wg sync.WaitGroup

	// wg.Add(1)

	// go func() {
	// 	defer wg.Done()

	// 	req.validate(reqValidationChan)
	// }()

	// requestValidationErr := <-reqValidationChan
	// if requestValidationErr != nil {
	// 	if e, ok := requestValidationErr.(validation.InternalError); ok {
	// 		log.Println(e)
	// 		util.Response(w, internalServerError, http.StatusInternalServerError)
	// 		return
	// 	} else {
	// 		log.Println(requestValidationErr)
	// 		util.Response(w, requestValidationErr.Error(), http.StatusInternalServerError)
	// 		return
	// 	}
	// }

	// payload := r.Context().Value("payload").(*auth.PayLoad)

	queries := sqlc.New(h.db)

	if requestBody.FolderID != "null" {
		util.CreateChildFolder(context.Background(), queries, w, r, requestBody.FolderName, requestBody.FolderID, authorizedPayload.AccountID)
		return
	}

	stringChan := make(chan string, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		util.RandomStringGenerator(stringChan)
	}()

	folderLabelChan := make(chan string, 1)

	wg.Add(1)

	go func() {
		defer wg.Done()

		util.GenFolderLabel(folderLabelChan)
	}()

	folderID := <-stringChan

	folderLabel := <-folderLabelChan

	folderParams := sqlc.CreateFolderParams{
		FolderID:    folderID,
		FolderName:  requestBody.FolderName,
		SubfolderOf: sql.NullString{},
		AccountID:   authorizedPayload.AccountID,
		Path:        folderLabel,
		Label:       folderLabel,
	}

	folder, err := queries.CreateFolder(context.Background(), folderParams)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Println(pgErr)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		} else {
			log.Print(err)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		}
	}

	rf := newReturnedFolder(folder)

	util.JsonResponse(w, rf)

	wg.Wait()
}

// CREATE CHILD FOLDER
type createChildFolderRequest struct {
	FolderName   string `json:"folder_name"`
	ParentFolder string `json:"parent_folder"`
}

func (s createChildFolderRequest) validate(reqValidationChan chan error) error {
	returnVal := validation.ValidateStruct(&s,
		validation.Field(&s.FolderName, validation.Required.When(s.FolderName == "").Error("Folder name is required"), validation.Length(1, 1000).Error("Folder name must be at least 1 character long"), validation.Match(regexp.MustCompile("^[^?[\\]{}|\\\\`./!@#$%^&*()_-]+$")).Error("Folder name must not have special characters")),
		validation.Field(&s.ParentFolder, validation.Required.When(s.ParentFolder == "").Error("Parent folder id is required"), validation.Length(33, 33).Error("Parent folder id must be 33 characters long")),
	)
	reqValidationChan <- returnVal

	return returnVal
}

func (h *BaseHandler) CreateChildFolder(w http.ResponseWriter, r *http.Request) {
	rBody := json.NewDecoder(r.Body)

	rBody.DisallowUnknownFields()

	var req createChildFolderRequest

	err := rBody.Decode(&req)
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

	payload := r.Context().Value("payload").(*auth.PayLoad)

	q := sqlc.New(h.db)

	parentFolder, err := q.GetFolder(context.Background(), req.ParentFolder)
	if err != nil {
		var pgErr *pgconn.PgError

		switch {
		case errors.As(err, &pgErr):
			log.Println(pgErr)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		case errors.Is(err, sql.ErrNoRows):
			log.Println(sql.ErrNoRows.Error())
			util.Response(w, "not found", http.StatusNotFound)
			return
		default:
			log.Println(err)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		}
	}

	parentFolderPath := parentFolder.Path

	stringChan := make(chan string, 1)

	wg.Add(1)

	go func() {
		defer wg.Done()

		util.RandomStringGenerator(stringChan)
	}()

	folderLabel := <-stringChan

	folderID := <-stringChan

	path := strings.Join([]string{parentFolderPath, folderLabel}, ".")

	arg := sqlc.CreateFolderParams{
		FolderID:    folderID,
		FolderName:  req.FolderName,
		SubfolderOf: sql.NullString{String: req.ParentFolder, Valid: true},
		AccountID:   payload.AccountID,
		Path:        path,
		Label:       folderLabel,
	}

	createdChildFolder, err := q.CreateFolder(context.Background(), arg)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Println(pgErr)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		} else {
			log.Println(err)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		}
	}

	wg.Wait()

	util.JsonResponse(w, createdChildFolder)
}

func (h *BaseHandler) GetRootFolders(w http.ResponseWriter, r *http.Request) {
	account_id := chi.URLParam(r, "accountID")

	log.Printf("ACCOUNT ID: %s", account_id)

	payload := r.Context().Value("payload").(*auth.PayLoad)

	q := sqlc.New(h.db)

	folders, err := q.GetRootNodes(context.Background(), payload.AccountID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Println(pgErr)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		} else {
			log.Println(err)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		}
	}

	util.JsonResponse(w, folders)
}

func (h *BaseHandler) GetFolderChildren(w http.ResponseWriter, r *http.Request) {
	log.Println("Getting folder children...")

	account_id := chi.URLParam(r, "accountID")

	folderID := chi.URLParam(r, "folderID")

	payload := r.Context().Value("payload").(*auth.PayLoad)

	accountID, err := strconv.Atoi(account_id)
	if err != nil {
		ErrorInternalServerError(w, err)
		return
	}

	if payload.AccountID != int64(accountID) {
		log.Println("unauthorized")
		util.Response(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	q := sqlc.New(h.db)

	folder, err := q.GetFolder(context.Background(), folderID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Println(pgErr)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		} else if errors.Is(err, sql.ErrNoRows) {
			log.Println("folder not found")
			util.Response(w, "folder not found", http.StatusNotFound)
			return
		} else {
			log.Println(err)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		}
	}

	if folder.AccountID != payload.AccountID {
		log.Println("unauthorized")
		util.Response(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if folder.AccountID != int64(accountID) {
		log.Println("unauthorized")
		util.Response(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	childrenFolders, err := q.GetFolderNodes(context.Background(), sqlc.GetFolderNodesParams{
		AccountID:   payload.AccountID,
		SubfolderOf: sql.NullString{String: folderID, Valid: true},
	})
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Println(pgErr)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		} else if errors.Is(err, sql.ErrNoRows) {
			log.Println(sql.ErrNoRows.Error())
			util.Response(w, "no child folders found", http.StatusNotFound)
			return
		} else {
			log.Println(err)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return

		}
	}

	util.JsonResponse(w, childrenFolders)
}

// GET FOLDER ANCESTORS
func (h *BaseHandler) GetFolderAncestors(w http.ResponseWriter, r *http.Request) {
	folderID := chi.URLParam(r, "folderID")

	q := sqlc.New(h.db)

	folder, err := q.GetFolder(context.Background(), folderID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			log.Println(pgErr)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		} else if errors.Is(err, sql.ErrNoRows) {
			log.Println(sql.ErrNoRows.Error())
			util.Response(w, errors.New("parent folder not in database").Error(), http.StatusNotFound)
			return
		} else {
			log.Println(err)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		}
	}

	label := folder.Label

	ancestors, err := q.GetFolderAncestors(context.Background(), label)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			log.Println(pgErr)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		} else if errors.Is(err, sql.ErrNoRows) {
			log.Println(sql.ErrNoRows.Error())
			util.Response(w, "folder ancestors not found", http.StatusNotFound)
			return
		} else {
			log.Println(err)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		}
	}

	util.JsonResponse(w, ancestors)
}

// STAR FOLDER
type starFoldersReq struct {
	FolderIDs []string `json:"folder_ids"`
}

func (s starFoldersReq) Validate(reqValidationChan chan error) error {
	validationErr := validation.ValidateStruct(&s,
		validation.Field(&s.FolderIDs, validation.Each(validation.Length(33, 33)), validation.Required),
	)

	reqValidationChan <- validationErr

	return validationErr
}

func (h *BaseHandler) StarFolders(w http.ResponseWriter, r *http.Request) {
	// get and validate folder id
	reqBody := json.NewDecoder(r.Body)

	reqBody.DisallowUnknownFields()

	var req starFoldersReq

	if err := reqBody.Decode(&req); err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("JSON syntax error occurred at offset byte: %d", e.Offset)
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

	reqValidationErr := <-reqValidationChan
	if reqValidationErr != nil {
		if e, ok := reqValidationErr.(validation.InternalError); ok {
			log.Println(e)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		} else {
			log.Println(reqValidationErr)
			util.Response(w, reqValidationErr.Error(), http.StatusInternalServerError)
			return
		}
	}

	// get folder
	q := sqlc.New(h.db)

	var starredFolders []sqlc.Folder

	for _, fid := range req.FolderIDs {
		folder, err := q.GetFolder(context.Background(), fid)
		if err != nil {
			var pgErr *pgconn.PgError

			switch {
			case errors.As(err, &pgErr):
				log.Println(pgErr)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			case errors.Is(err, sql.ErrNoRows):
				log.Println("folder not found")
				util.Response(w, "folder not found", http.StatusNotFound)
				return
			default:
				log.Println(err)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return

			}
		}

		// check if folder belongs to caller
		payload := r.Context().Value("payload").(*auth.PayLoad)

		if folder.AccountID != payload.AccountID {
			util.Response(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// toggle folder star status
		starredFolder, err := q.StarFolder(context.Background(), folder.FolderID)
		if err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) {
				log.Println(pgErr)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			} else {
				log.Println(err)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			}
		}

		starredFolders = append(starredFolders, starredFolder)
	}

	// wait for go routines to finish
	wg.Wait()

	// return starred of folders
	util.JsonResponse(w, starredFolders)
}

// UNSTAR FOLDERS
type unStarFoldersReq struct {
	FolderIDs []string `json:"folder_ids"`
}

func (s unStarFoldersReq) Validate(reqValidationChan chan error) error {
	validationErr := validation.ValidateStruct(&s,
		validation.Field(&s.FolderIDs, validation.Each(validation.Length(33, 33)), validation.Required),
	)

	reqValidationChan <- validationErr

	return validationErr
}

func (h *BaseHandler) UnstarFolders(w http.ResponseWriter, r *http.Request) {
	reqBody := json.NewDecoder(r.Body)

	reqBody.DisallowUnknownFields()

	var req unStarFoldersReq

	if err := reqBody.Decode(&req); err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("JSON syntax error occurred at offset byte: %d", e.Offset)
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

	reqValidationErr := <-reqValidationChan
	if reqValidationErr != nil {
		if e, ok := reqValidationErr.(validation.InternalError); ok {
			log.Println(e)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		} else {
			log.Println(reqValidationErr)
			util.Response(w, reqValidationErr.Error(), http.StatusInternalServerError)
			return
		}
	}

	wg.Wait()

	q := sqlc.New(h.db)

	var unstarredFolders []sqlc.Folder

	for _, folderID := range req.FolderIDs {
		// check if each folder exists
		folder, err := q.GetFolder(context.Background(), folderID)
		if err != nil {
			var pgErr *pgconn.PgError

			switch {
			case errors.As(err, &pgErr):
				log.Println(pgErr)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			case errors.Is(err, sql.ErrNoRows):
				log.Println(sql.ErrNoRows.Error())
				util.Response(w, "folder not found", http.StatusNotFound)
				return
			default:
				log.Println(err)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			}
		}

		// check if user is authorized ie is the owner of the folders
		payload := r.Context().Value("payload").(*auth.PayLoad)

		if payload.AccountID != folder.AccountID {
			log.Println("unauthorized")
			util.Response(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// unstar each folder
		unstarredFolder, err := q.UnstarFolder(context.Background(), folder.FolderID)
		if err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) {
				log.Println(pgErr)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			} else {
				log.Println(err)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			}
		}

		unstarredFolders = append(unstarredFolders, unstarredFolder)
	}

	// return unstarred folders
	util.JsonResponse(w, unstarredFolders)
}

// TOGGLE FOLDER STARRED
type toggleFolderStarredReq struct {
	FolderIDs []string `json:"folder_ids"`
}

func (t toggleFolderStarredReq) Validate(rValidationChan chan error) error {
	validationErr := validation.ValidateStruct(&t,
		validation.Field(&t.FolderIDs, validation.Each(validation.Length(33, 33).Error("each folder id must be 33 characters long")), validation.Required.Error("folder id/ids required")),
	)

	rValidationChan <- validationErr

	return validationErr
}

func (h *BaseHandler) ToggleFolderStarred(w http.ResponseWriter, r *http.Request) {
	rBody := json.NewDecoder(r.Body)

	rBody.DisallowUnknownFields()

	var req toggleFolderStarredReq

	if err := rBody.Decode(&req); err != nil {
		ErrorInvalidRequest(w, err)
		return
	}

	rValidationChan := make(chan error, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		req.Validate(rValidationChan)
	}()

	wg.Wait()

	if err := <-rValidationChan; err != nil {
		ErrorInvalidRequest(w, err)
		return
	}

	q := sqlc.New(h.db)

	var foldersStarred []sqlc.Folder

	for _, folderID := range req.FolderIDs {
		folderStarred, err := q.ToggleFolderStarred(context.Background(), folderID)
		if err != nil {
			ErrorInternalServerError(w, err)
			return
		}

		foldersStarred = append(foldersStarred, folderStarred)
	}

	util.JsonResponse(w, foldersStarred)
}

// RENAME FOLDER
type renameFolder struct {
	NewFolderName string `json:"new_folder_name"`
	FolderID      string `json:"folder_id"`
}

func (s renameFolder) Validate(reqValidationChan chan error) error {
	validationErr := validation.ValidateStruct(&s,
		validation.Field(&s.NewFolderName, validation.Required.When(s.NewFolderName == "").Error("New folder name cannot be empty!"), validation.Length(1, 200).Error("Folder name must be atleast 1 character long"), validation.Match(regexp.MustCompile("^[^?[\\]{}|\\\\`./!@$%^&*()_]+$")).Error("Folder name must not have special characters")),
		validation.Field(&s.FolderID, validation.Required.When(s.FolderID == "").Error("Folder id is required"), validation.Length(33, 33).Error("Folder ID must be 33 characters long")),
	)

	reqValidationChan <- validationErr

	return validationErr
}

func (h *BaseHandler) RenameFolder(w http.ResponseWriter, r *http.Request) {
	reqBody := json.NewDecoder(r.Body)

	reqBody.DisallowUnknownFields()

	var req renameFolder

	if err := reqBody.Decode(&req); err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("JSON syntax error occurred at offset byte: %d", e.Offset)
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

	wg.Wait()

	reqValidationErr := <-reqValidationChan
	if reqValidationErr != nil {
		if e, ok := reqValidationErr.(validation.InternalError); ok {
			log.Println(e)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		} else {
			log.Println(reqValidationErr)
			util.Response(w, reqValidationErr.Error(), http.StatusInternalServerError)
			return
		}
	}

	payload := r.Context().Value("payload").(*auth.PayLoad)

	q := sqlc.New(h.db)

	folder, err := q.GetFolder(context.Background(), req.FolderID)
	if err != nil {
		var pgErr *pgconn.PgError

		switch {
		case errors.As(err, &pgErr):
			log.Println(pgErr)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		case errors.Is(err, sql.ErrNoRows):
			log.Println("folder to rename not found")
			util.Response(w, sql.ErrNoRows.Error(), http.StatusNotFound)
			return
		default:
			log.Println(err)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		}
	}

	if folder.AccountID != payload.AccountID {
		log.Println("user is unauthorized")
		util.Response(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	arg := sqlc.RenameFolderParams{
		FolderName: req.NewFolderName,
		FolderID:   req.FolderID,
	}

	renamedFolder, err := q.RenameFolder(context.Background(), arg)
	if err != nil {
		var pgErr *pgconn.PgError

		switch {
		case errors.As(err, &pgErr):
			log.Println(pgErr)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		default:
			log.Println(err)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		}
	}

	util.JsonResponse(w, newReturnedFolder(renamedFolder))
}

type moveFoldersToTrash struct {
	FolderIDs []string `json:"folder_ids"`
}

func (s moveFoldersToTrash) Validate(reqValidationChan chan error) error {
	validationErr := validation.ValidateStruct(&s,
		validation.Field(&s.FolderIDs, validation.Required.Error("folder ids requiured"), validation.Each(validation.Length(33, 33).Error("folder id must be 33 characters long"))),
	)

	reqValidationChan <- validationErr

	return validationErr
}

func (h *BaseHandler) MoveFoldersToTrash(w http.ResponseWriter, r *http.Request) {
	rBody := json.NewDecoder(r.Body)

	rBody.DisallowUnknownFields()

	var req moveFoldersToTrash

	if err := rBody.Decode(&req); err != nil {
		ErrorDecodingRequest(w, err)
		return
	}

	reqValidationChan := make(chan error, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		req.Validate(reqValidationChan)
	}()

	wg.Wait()

	reqValidationErr := <-reqValidationChan
	if reqValidationErr != nil {
		ErrorInvalidRequest(w, reqValidationErr)
		return
	}

	payload := r.Context().Value("payload").(*auth.PayLoad)

	q := sqlc.New(h.db)

	var trashedFolders []sqlc.Folder

	for _, folderID := range req.FolderIDs {
		folder, err := q.GetFolder(context.Background(), folderID)
		if err != nil {
			var pgErr *pgconn.PgError

			switch {
			case errors.As(err, &pgErr):
				log.Println(pgErr)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			case errors.Is(err, sql.ErrNoRows):
				log.Println("folder not found")
				util.Response(w, "folder not found", http.StatusNotFound)
				return
			default:
				log.Println(err)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			}
		}

		if folder.AccountID != payload.AccountID {
			log.Println("user unauthorized to delete this folder")
			util.Response(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		trashedFolder, err := q.MoveFolderToTrash(context.Background(), folder.FolderID)
		if err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) {
				log.Println(pgErr)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			} else {
				log.Println(err)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			}
		}

		trashedFolders = append(trashedFolders, trashedFolder)
	}

	util.JsonResponse(w, trashedFolders)
}

func (h *BaseHandler) GetFolder(w http.ResponseWriter, r *http.Request) {
	folderID := chi.URLParam(r, "folderID")
	payload := r.Context().Value("payload").(*auth.PayLoad)

	q := sqlc.New(h.db)

	folder, err := q.GetFolder(context.Background(), folderID)
	if err != nil {
		var pgErr *pgconn.PgError

		switch {
		case errors.As(err, &pgErr):
			log.Println(pgErr)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		case errors.Is(err, sql.ErrNoRows):
			log.Println(sql.ErrConnDone.Error())
			util.Response(w, "folder not found", http.StatusNotFound)
			return
		default:
			log.Println(err)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		}
	}

	if folder.AccountID != payload.AccountID {
		log.Println("user unauthorized for this operation")
		util.Response(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	util.JsonResponse(w, folder)
}

// MOVE FOLDERS
type moveFoldersRequest struct {
	DestinationFolderID string   `json:"destination_folder_id"`
	FolderIDs           []string `json:"folder_ids"`
}

func (m moveFoldersRequest) Validate(reqValidationChan chan error) error {
	validationErr := validation.ValidateStruct(&m,
		validation.Field(&m.FolderIDs, validation.Required.Error("Folder IDs requiured"), validation.Each(validation.Length(33, 33).Error("Folder id must be 33 characters long"))),
		validation.Field(&m.DestinationFolderID, validation.Required.Error("Destination folder id required"), validation.Length(33, 33).Error("Destination folder id must be 33 charecters long")),
	)

	reqValidationChan <- validationErr

	return validationErr
}

func (h *BaseHandler) MoveFolders(w http.ResponseWriter, r *http.Request) {
	rBody := json.NewDecoder(r.Body)
	rBody.DisallowUnknownFields()

	var req moveFoldersRequest

	err := rBody.Decode(&req)
	if err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
			util.Response(w, internalServerError, 500)
			return
		} else {
			log.Printf("error decoding request body to struct: %v", err)
			util.Response(w, internalServerError, http.StatusBadRequest)
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

	wg.Wait()

	reqValidationErr := <-reqValidationChan
	if reqValidationErr != nil {
		if e, ok := reqValidationErr.(validation.InternalError); ok {
			log.Println(e)
			util.Response(w, internalServerError, 500)
			return
		} else {
			log.Printf("folder ids validation error: %v", reqValidationErr)
			util.Response(w, reqValidationErr.Error(), 500)
			return
		}
	}

	payload := r.Context().Value("payload").(*auth.PayLoad)

	q := sqlc.New(h.db)

	destinationFolder, err := q.GetFolder(context.Background(), req.DestinationFolderID)
	if err != nil {
		var pgErr *pgconn.PgError

		switch {
		case errors.As(err, &pgErr):
			log.Println(pgErr)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		case errors.Is(err, sql.ErrNoRows):
			log.Println("folder not found")
			util.Response(w, "folder not found", http.StatusNotFound)
			return
		default:
			log.Println(err)
			util.Response(w, internalServerError, http.StatusInternalServerError)
			return
		}
	}

	if destinationFolder.AccountID != payload.AccountID {
		log.Println("unauthorized")
		util.Response(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var foldersMoved []sqlc.Folder

	for _, folder_ID := range req.FolderIDs {
		folder, err := q.GetFolder(context.Background(), folder_ID)
		if err != nil {
			var pgErr *pgconn.PgError

			switch {
			case errors.As(err, &pgErr):
				log.Println(pgErr)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			case errors.Is(err, sql.ErrNoRows):
				log.Println("folder not found")
				util.Response(w, "folder not found", http.StatusNotFound)
				return
			default:
				log.Println(err)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			}
		}

		if folder.AccountID != payload.AccountID {
			log.Println("unauthorized")
			util.Response(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		arg := sqlc.MoveFolderParams{
			Label:   destinationFolder.Label,
			Label_2: folder.Label,
			Label_3: folder.Label,
		}

		movedFolders, err := q.MoveFolder(context.Background(), arg)
		if err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) {
				log.Println(pgErr)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			} else {
				log.Println(err)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			}
		}

		arg2 := sqlc.UpdateFolderSubfolderOfParams{
			SubfolderOf: sql.NullString{String: destinationFolder.FolderID, Valid: true},
			FolderID:    folder.FolderID,
		}

		_, err = q.UpdateFolderSubfolderOf(context.Background(), arg2)
		if err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) {
				log.Println(pgErr)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			} else {
				log.Println(err)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			}
		}

		for _, movedFolder := range movedFolders {
			movedFolder, err = q.GetFolder(context.Background(), movedFolder.FolderID)
			if err != nil {
				var pgErr *pgconn.PgError

				if errors.As(err, &pgErr) {
					log.Println(pgErr)
					util.Response(w, internalServerError, http.StatusInternalServerError)
					return
				} else {
					log.Println(err)
					util.Response(w, internalServerError, http.StatusInternalServerError)
					return
				}
			}

			foldersMoved = append(foldersMoved, movedFolder)
		}
	}

	util.JsonResponse(w, foldersMoved)
}

// MOVE FOLDERS TO ROOT
type moveFoldersToRootRequest struct {
	FolderIDs []string `json:"folder_ids"`
}

func (m moveFoldersToRootRequest) Validate(reqValidationChan chan error) error {
	validationErr := validation.ValidateStruct(&m,
		validation.Field(&m.FolderIDs, validation.Required.Error("Folder IDs requiured"), validation.Each(validation.Length(33, 33).Error("Folder id must be 33 characters long"))),
	)

	reqValidationChan <- validationErr

	return validationErr
}

func (h *BaseHandler) MoveFoldersToRoot(w http.ResponseWriter, r *http.Request) {
	rBody := json.NewDecoder(r.Body)
	rBody.DisallowUnknownFields()

	var req moveFoldersToRootRequest

	err := rBody.Decode(&req)
	if err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
			util.Response(w, internalServerError, 500)
			return
		} else {
			log.Printf("error decoding request body to struct: %v", err)
			util.Response(w, internalServerError, http.StatusBadRequest)
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

	wg.Wait()

	reqValidationErr := <-reqValidationChan
	if reqValidationErr != nil {
		if e, ok := reqValidationErr.(validation.InternalError); ok {
			log.Println(e)
			util.Response(w, internalServerError, 500)
			return
		} else {
			log.Printf("folder ids validation error: %v", reqValidationErr)
			util.Response(w, reqValidationErr.Error(), 500)
			return
		}
	}

	q := sqlc.New(h.db)

	payload := r.Context().Value("payload").(*auth.PayLoad)

	var foldersMovedToRoot []sqlc.Folder

	for _, folderID := range req.FolderIDs {
		folder, err := q.GetFolder(context.Background(), folderID)
		if err != nil {
			var pgErr *pgconn.PgError

			switch {
			case errors.As(err, &pgErr):
				log.Println(pgErr)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			case errors.Is(err, sql.ErrNoRows):
				log.Println("folder not found")
				util.Response(w, "folder not found", http.StatusNotFound)
				return
			default:
				log.Println(err)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			}
		}

		if folder.AccountID != payload.AccountID {
			log.Println("unauthorized")
			util.Response(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		arg := sqlc.MoveFoldersToRootParams{
			Label:   folder.Label,
			Label_2: folder.Label,
		}

		folderMovedToRoot, err := q.MoveFoldersToRoot(context.Background(), arg)
		if err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) {
				log.Println(pgErr)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			} else {
				log.Println(err)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			}
		}

		if err := q.UpdateParentFolderToNull(context.Background(), folder.FolderID); err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) {
				log.Println(pgErr)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			} else {
				log.Println(err)
				util.Response(w, internalServerError, http.StatusInternalServerError)
				return
			}
		}

		foldersMovedToRoot = append(foldersMovedToRoot, folderMovedToRoot...)
	}

	util.JsonResponse(w, foldersMovedToRoot)
}

type restoreFoldersRequest struct {
	FolderIDS []string `json:"folder_ids"`
}

func (r restoreFoldersRequest) Validate(requestValidationChan chan error) error {
	requestValidationChan <- validation.ValidateStruct(&r,
		validation.Field(&r.FolderIDS, validation.Required.When(len(r.FolderIDS) > 0), validation.Each(validation.Length(33, 33).Error("each folder id must be 33 characters long"))),
	)
	return validation.ValidateStruct(&r,
		validation.Field(&r.FolderIDS, validation.Required.When(len(r.FolderIDS) > 0), validation.Each(validation.Length(33, 33).Error("each folder id must be 33 characters long"))),
	)
}

func (h *BaseHandler) RestoreFoldersFromTrash(w http.ResponseWriter, r *http.Request) {
	body := json.NewDecoder(r.Body)

	body.DisallowUnknownFields()

	var req restoreFoldersRequest

	if err := body.Decode(&req); err != nil {
		ErrorDecodingRequest(w, err)
		return
	}

	requestValidationChan := make(chan error, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		req.Validate(requestValidationChan)
	}()

	err := <-requestValidationChan
	if err != nil {
		log.Println(err)
		ErrorInvalidRequest(w, err)
		return
	}

	q := sqlc.New(h.db)

	var folders []sqlc.Folder

	for _, folderID := range req.FolderIDS {
		f, err := q.RestoreFolderFromTrash(context.Background(), folderID)
		if err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) {
				log.Println(err)
				util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
				return
			} else {
				log.Println(err)
				util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
				return
			}
		}

		folders = append(folders, f)
	}

	util.JsonResponse(w, folders)
}

type deleteFoldersForeverRequest struct {
	FolderIDS []string `json:"folder_ids"`
}

func (d deleteFoldersForeverRequest) Validate(requestValidationChan chan error) error {
	requestValidationChan <- validation.ValidateStruct(&d,
		validation.Field(&d.FolderIDS, validation.Required.When(len(d.FolderIDS) > 0), validation.Each(validation.Length(33, 33).Error("each folder id must be 33 characters long"))),
	)
	return validation.ValidateStruct(&d,
		validation.Field(&d.FolderIDS, validation.Required.When(len(d.FolderIDS) > 0), validation.Each(validation.Length(33, 33).Error("each folder id must be 33 characters long"))),
	)
}

func (h *BaseHandler) DeleteFoldersForever(w http.ResponseWriter, r *http.Request) {
	body := json.NewDecoder(r.Body)

	body.DisallowUnknownFields()

	var req deleteFoldersForeverRequest

	if err := body.Decode(&req); err != nil {
		ErrorDecodingRequest(w, err)
		return
	}

	requestValidationChan := make(chan error, 1)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		req.Validate(requestValidationChan)
	}()

	err := <-requestValidationChan
	if err != nil {
		log.Println(err)
		ErrorInvalidRequest(w, err)
		return
	}

	q := sqlc.New(h.db)

	var folders []sqlc.Folder

	for _, folderID := range req.FolderIDS {
		f, err := q.DeleteFolderForever(context.Background(), folderID)
		if err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) {
				log.Println(err)
				util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
				return
			} else {
				log.Println(err)
				util.Response(w, errors.New("something went wrong").Error(), http.StatusInternalServerError)
				return
			}
		}

		folders = append(folders, f...)
	}

	util.JsonResponse(w, folders)
}
