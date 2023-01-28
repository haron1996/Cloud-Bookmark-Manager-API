package util

import (
	"github.com/go-rod/rod"
)

func GetUrlHeading(page *rod.Page, urlHeadingChan chan string) {

	pageHasHeading, h1, _ := page.Has("h1")

	if pageHasHeading {
		urlHeadingChan <- h1.MustText()

	} else {
		urlHeadingChan <- ""
	}
}
