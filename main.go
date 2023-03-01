package main

import (
	"log"
	"net/http"

	"github.com/kwandapchumba/go-bookmark-manager/router"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

func main() {
	// viper.SetConfigFile(".env")
	// viper.ReadInConfig()
	// viper.AutomaticEnv()

	// port := fmt.Sprint(viper.Get("PORT"))

	// log.Println(port)

	config, err := util.LoadConfig("/production/")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	log.Printf("config file successfully loaded as: %v", config)

	server := &http.Server{
		Addr: config.PORT,
		// Addr:    port,
		Handler: router.Router(),
	}

	log.Fatal(server.ListenAndServe())
}
