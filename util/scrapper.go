package util

import (
	"log"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/gocolly/colly"
)

func Scrapper(link string) {
	err := validation.Validate(link,
		validation.Required, // not empty
		is.URL,              // is a valid URL
	)

	if err != nil {
		log.Println(err)
		return
	}

	c := colly.NewCollector()

	c.OnHTML("html", func(e *colly.HTMLElement) {
		e.ForEach("img", func(i int, h *colly.HTMLElement) {
			log.Println(h.Attr("src"))
		})
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	c.OnResponse(func(r *colly.Response) {
		log.Println("Visited", r.Request.URL)
	})

	c.OnScraped(func(r *colly.Response) {
		log.Println("Finished", r.Request.URL)
	})

	c.Visit(link)
}
