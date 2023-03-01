package main

import (
	"log"
	"net/http"

	"github.com/kwandapchumba/go-bookmark-manager/router"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Printf("failed to load config with error: %v", err)
		return
	}

	log.Printf("config file successfully loaded as: %v", config)

	server := &http.Server{
		Addr:    config.PORT,
		Handler: router.Router(),
	}

	log.Fatal(server.ListenAndServe())
}
