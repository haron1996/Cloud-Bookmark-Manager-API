package util

import (
	"log"

	"github.com/go-rod/rod"
)

func GetUrlParagraph(url string, urlParagraphChan chan string) {
	page := rod.New().MustConnect().MustPage(url)

	page = page.MustWaitLoad()

	pageHasParagraph, p, _ := page.Has("p")

	log.Println("passed here!")
	log.Println(pageHasParagraph, p.MustText())

	if pageHasParagraph {

		urlParagraphChan <- p.MustText()
	} else {
		urlParagraphChan <- ""
	}
}
