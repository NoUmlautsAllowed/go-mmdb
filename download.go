package mmdb

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	MaxmindAccountId  = "MAXMIND_ACCOUNT_ID"
	MaxmindLicenseKey = "MAXMIND_LICENSE_KEY"
	MaxmindEditionIds = "MAXMIND_EDITION_IDS"
	MaxmindBasePath   = "MAXMIND_BASE_PATH"
)

// Downloader holds configuration & HTTP client for fetching MMDBs.
type Downloader struct {
	AccountID  string
	LicenseKey string
	BasePath   string
	Client     *http.Client
}

// NewDownloader reads env vars and returns a configured Downloader.
func NewDownloader() (*Downloader, error) {
	account := os.Getenv(MaxmindAccountId)
	license := os.Getenv(MaxmindLicenseKey)
	base := os.Getenv(MaxmindBasePath)
	if base == "" {
		base = "."
	}
	if account == "" || license == "" {
		return nil, fmt.Errorf("mmdb: missing %s or %s",
			MaxmindAccountId, MaxmindLicenseKey)
	}

	return &Downloader{
		AccountID:  account,
		LicenseKey: license,
		BasePath:   base,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// DownloadDatabases downloads (or skips) each requested DB.
func (d *Downloader) DownloadDatabases(ctx context.Context, dbs ...string) error {
	if err := os.MkdirAll(d.BasePath, 0o755); err != nil {
		return fmt.Errorf("mkdir %q: %w", d.BasePath, err)
	}

	for _, db := range dbs {
		if err := d.downloadOne(ctx, db); err != nil {
			log.Printf("mmdb [%s]: %v", db, err)
		}
	}
	return nil
}

func (d *Downloader) downloadOne(ctx context.Context, db string) error {
	url := fmt.Sprintf(
		"https://download.maxmind.com/geoip/databases/%s/download?suffix=tar.gz",
		db,
	)

	// 1) fetch remote build time
	remoteTime, err := d.fetchRemoteTime(ctx, url)
	if err != nil {
		return fmt.Errorf("fetch remote time: %w", err)
	}

	localFile := dbPath(d.BasePath, db)
	if !needsDownload(localFile, remoteTime) {
		info, _ := os.Stat(localFile)
		log.Printf("mmdb [%s] up to date (%s)", db, info.ModTime().UTC())
		return nil
	}

	tmp := localFile + ".tmp"
	if err := d.fetchAndExtract(ctx, url, tmp); err != nil {
		return fmt.Errorf("download+extract: %w", err)
	}

	// preserve build timestamp
	if err := os.Chtimes(tmp, time.Now(), remoteTime); err != nil {
		log.Printf("mmdb [%s] chtimes: %v", db, err)
	}

	// backup old file
	if err := backupOld(localFile); err != nil {
		log.Printf("mmdb [%s] backup: %v", db, err)
	}

	// atomically replace
	if err := os.Rename(tmp, localFile); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("final rename: %w", err)
	}

	log.Printf("mmdb [%s] updated → %s (build %s)", db, localFile, remoteTime.UTC())
	return nil
}

func (d *Downloader) fetchRemoteTime(ctx context.Context, url string) (time.Time, error) {
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return time.Time{}, err
	}
	req.SetBasicAuth(d.AccountID, d.LicenseKey)

	resp, err := d.Client.Do(req)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return time.Time{}, fmt.Errorf("HEAD status %d", resp.StatusCode)
	}

	lm := resp.Header.Get("Last-Modified")
	if lm == "" {
		return time.Time{}, fmt.Errorf("missing Last-Modified header")
	}
	t, err := time.Parse(http.TimeFormat, lm)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse Last-Modified: %w", err)
	}
	return t, nil
}

func needsDownload(path string, remote time.Time) bool {
	info, err := os.Stat(path)
	if err != nil {
		return true
	}
	return remote.After(info.ModTime())
}

func (d *Downloader) fetchAndExtract(ctx context.Context, url, tmpFile string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(d.AccountID, d.LicenseKey)

	resp, err := d.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("GET status %d", resp.StatusCode)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	out, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	defer func() {
		out.Close()
		if err != nil {
			os.Remove(tmpFile)
		}
	}()

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read: %w", err)
		}
		if hdr.Typeflag == tar.TypeReg && strings.HasSuffix(hdr.Name, ".mmdb") {
			if _, err := io.Copy(out, tr); err != nil {
				return fmt.Errorf("extract mmdb: %w", err)
			}
			return nil
		}
	}
	return fmt.Errorf("no .mmdb entry found in archive")
}

func backupOld(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	bak := path + ".old"
	os.Remove(bak)
	return os.Rename(path, bak)
}
