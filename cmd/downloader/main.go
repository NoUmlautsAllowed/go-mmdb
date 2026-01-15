package main

import (
	"context"

	"github.com/joho/godotenv"
	"gitlab.w1lhelm.de/swilhelm/go-mmdb"

	"log"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	d, err := mmdb.NewDownloader()
	if err != nil {
		log.Fatalf("error creating downloader: %v", err)
	}

	dbs := []string{mmdb.CityDatabase, mmdb.CountryDatabase, mmdb.ASNDatabase}
	err = d.DownloadDatabases(context.Background(), dbs...)
	if err != nil {
		log.Fatalf("error downloading databases: %v", err)
	}
}
