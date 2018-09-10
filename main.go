// Command screenshot is a chromedp example demonstrating how to take a
// screenshot of a specific element.
package main

import (
	"context"
	"time"

	"github.com/santrancisco/certstream_screenshot/screenshotworker"
)

func main() {

	// create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wp := screenshotworker.NewWorkerPool(ctx, 4)
	wp.Start()
	time.Sleep(2 * time.Second)
	wp.URL <- "https://ebfe.pw"
	wp.URL <- "https://cornmbank.com"
	wp.URL <- "https://cba.com"
	wp.URL <- "https://dta.gov.au"
	wp.URL <- "https://cyber.gov.au"
	wp.URL <- "https://tesla.com"
	time.Sleep(10 * time.Second)
	cancel()
	wp.Wg.Wait()
}
