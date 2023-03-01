package main

import (
	"log"
	"net/http"
	"os"

	"github.com/kwandapchumba/go-bookmark-manager/router"
)

func main() {
	// config, err := util.LoadConfig(".")
	// if err != nil {
	// 	log.Fatal("cannot load config:", err)
	// }

	// log.Printf("config file successfully loaded as: %v", config)

	setEnvs()

	log.Println(os.Getenv("port"))

	server := &http.Server{
		// Addr: config.PORT,
		Addr:    os.Getenv("port"),
		Handler: router.Router(),
	}

	log.Fatal(server.ListenAndServe())
}

func setEnvs() {
	os.Setenv("port", ":5000")
	os.Setenv("accessTokenDuration", "24h")
	os.Setenv("dbString", "postgresql://kibet:535169003@localhost:5432/saasita?sslmode=disable")
	os.Setenv("doSecret", "MNTa0Ht5xsZSZNr/cdG+B/Z1Nuk3On6i0pTUBdVeKvc")
	os.Setenv("doSpaces", "DO00PBMJQ2HD2X6MZPH4")
	os.Setenv("mailJetApiKey", "349b55371fe86fa09ac6901c1be30ebe")
	os.Setenv("mailJetSecretKey", "d27b1aa47c7c006e7c6d76098437be1b")
	os.Setenv("mailgunApiKey", "305de2dae3e05b5a2887c00038ce0512-381f2624-94263a4b")
	os.Setenv("mailgunDomain", "sandbox915a87e05b764cf6be37483020e8cfa0.mailgun.org")
	os.Setenv("publicKeyHex", "1eb9dbbbbc047c03fd70604e0071f0987e16b28b757225c11f00415d0e20b1a2")
	os.Setenv("refreshTokenDuration", "168h")
	os.Setenv("secretKeyHex", "b4cbfb43df4ce210727d953e4a713307fa19bb7d9f85041438d9e11b942a37741eb9dbbbbc047c03fd70604e0071f0987e16b28b757225c11f00415d0e20b1a2")
}
