package util

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

func GetUrlScreenshot(url string) {
	err := validation.Validate(url,
		validation.Required,
		is.URL,
	)

	if err != nil {
		log.Fatalf("url is not a valid URL: %v", err)
	}

	// create context
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		// chromedp.WithDebugf(log.Printf),
	)

	defer cancel()

	var buf []byte

	// capture entire browser viewport, returning png with quality=90
	if err := chromedp.Run(ctx, screenshot(url, 90, &buf)); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("screenshot.png", buf, 0o644); err != nil {
		log.Fatal(err)
	}

	log.Printf("wrote elementScreenshot.png and fullScreenshot.png")
}

// Note: chromedp.FullScreenshot overrides the device's emulation settings. Use
// device.Reset to reset the emulation and viewport settings.
func screenshot(urlstr string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.Sleep(time.Second * 3),
		chromedp.EmulateViewport(1200, 688, chromedp.EmulateLandscape),
		chromedp.CaptureScreenshot(res),
		//chromedp.FullScreenshot(res, quality),
	}
}
