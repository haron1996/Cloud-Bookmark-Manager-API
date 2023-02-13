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
		log.Println(err)
		return
	}

	server := &http.Server{
		Addr:    config.PORT,
		Handler: router.Router(),
	}

	log.Fatal(server.ListenAndServe())
}
