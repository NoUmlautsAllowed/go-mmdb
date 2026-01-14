package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"gitlab.w1lhelm.de/swilhelm/go-mmdb"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Downloader setup
	dl, err := mmdb.NewDownloader()
	if err != nil {
		log.Printf("Downloader not configured: %v. Continuing without downloader.", err)
	} else {
		dbs := []string{mmdb.CityDatabase, mmdb.CountryDatabase, mmdb.ASNDatabase}
		// Initial download
		log.Printf("Running initial MMDB download...")
		if err := dl.DownloadDatabases(ctx, dbs...); err != nil {
			log.Printf("Initial MMDB download failed: %v", err)
		}

		// Background downloader
		go func() {
			ticker := time.NewTicker(mmdb.DefaultReloadInterval)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					log.Printf("Running periodic MMDB download...")
					if err := dl.DownloadDatabases(ctx, dbs...); err != nil {
						log.Printf("Periodic MMDB download failed: %v", err)
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	client, err := mmdb.NewClient()
	if err != nil {
		log.Fatalf("Failed to initialize MMDB client: %v", err)
	}
	defer client.Close()

	srv, err := mmdb.NewServer(client)
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	addr := os.Getenv("BIND_ADDR")
	if addr == "" {
		addr = "localhost:8080"
	}

	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
