package util

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
)

func DownloaFavicon(URL, fileName string) error {

	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Printf("status code: %v", response.StatusCode)
		return errors.New("received non 200 response code")
	}

	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer file.Close()

	//Write the bytes to the file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}
