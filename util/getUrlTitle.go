package util

import (
	"log"

	"github.com/go-rod/rod"
)

func GetUrlTitle(page *rod.Page, urlTitleChan chan string) {
	log.Println(page.Has("title"))
	pageHasTitle, title, _ := page.Has("title")

	if pageHasTitle {
		urlTitleChan <- title.MustText()
	} else {
		urlTitleChan <- ""
	}
}
