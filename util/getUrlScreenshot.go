package util

import (
	"github.com/go-rod/rod"
)

func RodGetUrlScreenshot(page *rod.Page) {

	page.MustScreenshot("a.png")
}
