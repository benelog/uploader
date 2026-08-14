package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/benelog/uploader/internal/uploader"
)

func main() {
	httpPort := flag.Int("httpPort", 8080, "http port to listen on")
	flag.Parse()

	handler, err := uploader.NewHandler()
	if err != nil {
		log.Fatalf("failed to initialize uploader: %v", err)
	}

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(*httpPort),
		Handler: handler.Routes(),
		// Only the headers are given a deadline: a large upload may legitimately
		// take as long as the network needs.
		ReadHeaderTimeout: 20 * time.Second,
	}

	log.Printf("uploader is listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
