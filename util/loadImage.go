package util

import (
	"log"
	"os"
)

// load url screenshot from os
func LoadImage(imgFileChan chan *os.File, f string) error {
	imgFile, err := os.Open(f)
	if err != nil {
		log.Printf("failed to open file: %v", err)
		return err
	}

	imgFileChan <- imgFile

	return nil
}
