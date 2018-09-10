package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/CaliDog/certstream-go"
	"github.com/santrancisco/certstream_screenshot/screenshotworker"
	"github.com/santrancisco/logutils"
)

var NEEDLES = []string{"tesla.com", "campaignmonitor.com", "tyro.com", "ebfe.pw", "cornmbank.com", "gov.au"}

func main() {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("DEBUG"),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)
	// create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log.Println("[INFO] Starting our Screenshot worker")
	wp := screenshotworker.NewWorkerPool(ctx, 6)
	wp.Start()

	log.Println("[INFO] Monitor Certstream for needles")
	stream, errStream := certstream.CertStreamEventStream(false)

MAINLOOP:
	for {
		select {
		case _ = <-sigchan:
			log.Print("[INFO] SIGTERM received, exiting app...")
			cancel()
			wp.Wg.Wait()
			return
		case jq := <-stream:
			messageType, err := jq.String("message_type")
			if err != nil {
				log.Print("Error decoding jq string")
				continue
			}
			if messageType != "certificate_update" {
				continue
			}
			all_domains_array, err := jq.Array("data", "leaf_cert", "all_domains")
			for _, element := range all_domains_array {
				domain := element.(string)
				// log.Printf("[DEBUG] Checking %s ", domain)
				for _, needle := range NEEDLES {
					if strings.HasSuffix(domain, needle) {
						log.Printf("[INFO] New Cert for %s was requested", domain)
						wp.URL <- "https://" + strings.TrimLeft(domain, "*.")
						continue MAINLOOP
					}
				}
			}
		case err := <-errStream:
			log.Printf("%v", err)
		}
	}
}
