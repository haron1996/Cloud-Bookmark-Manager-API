package util

import (
	"github.com/go-rod/rod"
)

func GetUrlTitle(page *rod.Page, urlTitleChan chan string) {

	pageHasTitle, title, _ := page.Has("title")

	if pageHasTitle {
		urlTitleChan <- title.MustText()
	} else {
		urlTitleChan <- ""
	}
}
