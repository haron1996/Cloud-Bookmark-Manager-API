package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/kwandapchumba/go-bookmark-manager/api"
	"github.com/kwandapchumba/go-bookmark-manager/db/connection"
	cm "github.com/kwandapchumba/go-bookmark-manager/middleware"
)

func Router() *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.AllowContentEncoding("application/json", "application/x-www-form-urlencoded"))
	r.Use(middleware.CleanPath)
	r.Use(middleware.RedirectSlashes)

	h := api.NewBaseHandler(connection.ConnectDB())

	// public routes go here
	r.Route("/public", func(r chi.Router) {
		r.Post("/checkIfIsAuthenticated", h.CheckIfIsAuthenticated)

		r.Post("/refreshToken", h.RefreshToken)

		r.Post("/sendOTP", h.SendOTP)

		r.Post("/verifyOTP", h.VerifyOTP)

		r.Route("/account", func(r chi.Router) {
			r.Post("/", h.ContinueWithGoogle)
			r.Post("/create", h.NewAccount)
			r.Get("/getAllAccounts", h.GetAllAccounts)
			r.Post("/signin", h.SignIn)
		})
	})

	// private routes ie they require authenticated calls
	r.Route("/private", func(r chi.Router) {
		r.Use(cm.ReturnVerifiedUserToken())

		r.Get("/getLinksAndFolders/{accountID}/{folderID}", h.GetLinksAndFolders)

		r.Get("/getFoldersAndLinksMovedToTrash/{accountID}", h.GetFoldersAndLinksMovedToTrash)

		r.Route("/folder", func(r chi.Router) {
			r.Post("/create", h.CreateFolder)
			r.Post("/new-child-folder", h.CreateChildFolder)
			r.Patch("/star", h.StarFolders)
			r.Patch("/unstar", h.UnstarFolders)
			r.Patch("/rename", h.RenameFolder)
			r.Patch("/moveFoldersToTrash", h.MoveFoldersToTrash)
			r.Patch("/moveFolders", h.MoveFolders)
			r.Patch("/moveFoldersToRoot", h.MoveFoldersToRoot)
			r.Patch("/toggle-folder-starred", h.ToggleFolderStarred)
			r.Patch("/restoreFoldersFromTrash", h.RestoreFoldersFromTrash)
			r.Delete("/deleteFoldersForever", h.DeleteFoldersForever)
			r.Get("/{folderID}", h.GetFolder)
			r.Get("/get-folders/{accountID}", h.GetRootFolders)
			r.Get("/getFolderChildren/{folderID}/{accountID}", h.GetFolderChildren)
			r.Get("/getFolderAncestors/{folderID}", h.GetFolderAncestors)
			r.Get("/searchFolders/{query}", h.SearchFolders)
		})

		r.Route("/link", func(r chi.Router) {
			r.Post("/add", h.AddLink)
			r.Patch("/rename", h.RenameLink)
			r.Patch("/move", h.MoveLinks)
			r.Patch("/moveLinksToTrash", h.MoveLinksToTrash)
			r.Patch("/restoreLinksFromTrash", h.RestoreLinksFromTrash)
			r.Delete("/deleteLinksForever", h.DeleteLinksForever)
			r.Get("/getRootLinks/{accountID}", h.GetRootLinks)
			r.Get("/get_folder_links/{accountID}/{folderID}", h.GetFolderLinks)
			r.Get("/searchLinks/{query}", h.SearchLinks)
		})

		r.Post("/contactSupport", h.ContactSupport)
	})

	return r
}
