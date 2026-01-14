package mmdb

import (
	"log"
	"path"
	"sync"
	"time"

	"github.com/oschwald/geoip2-golang"
)

const (
	CityDatabase    = "GeoLite2-City.mmdb"
	CountryDatabase = "GeoLite2-Country.mmdb"
	ASNDatabase     = "GeoLite2-ASN.mmdb"
	// DefaultReloadInterval is how often we check for a new DB file.
	DefaultReloadInterval = 2 * time.Hour
)

type Client struct {
	DataDirectory string

	muCountry sync.RWMutex
	country   *geoip2.Reader

	muCity sync.RWMutex
	city   *geoip2.Reader

	muASN sync.RWMutex
	asn   *geoip2.Reader

	ticker *time.Ticker
	done   chan struct{}
}

// NewClient creates a Client and opens all three GeoIP2 databases.
// If dataDirectory is empty, it defaults to the working directory
// On any error it closes any readers it already opened.
func NewClient(dataDirectory string) (*Client, error) {
	if dataDirectory == "" {
		dataDirectory = "."
	}

	var (
		err     error
		country *geoip2.Reader
		city    *geoip2.Reader
		asn     *geoip2.Reader
	)

	country, err = geoip2.Open(path.Join(dataDirectory, CountryDatabase))
	if err != nil {
		return nil, err
	}

	city, err = geoip2.Open(path.Join(dataDirectory, CityDatabase))
	if err != nil {
		_ = country.Close()
		return nil, err
	}

	asn, err = geoip2.Open(path.Join(dataDirectory, ASNDatabase))
	if err != nil {
		_ = country.Close()
		_ = city.Close()
		return nil, err
	}

	c := &Client{
		DataDirectory: dataDirectory,
		country:       country,
		city:          city,
		asn:           asn,
		ticker:        time.NewTicker(DefaultReloadInterval),
		done:          make(chan struct{}),
	}

	go c.startReload()
	return c, nil
}

// startReload runs until Close() is called.
// On panic it logs and restarts one more time.
func (c *Client) startReload() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("mmdb reload panic recovered: %v", r)
			// restart in a fresh goroutine
			go c.startReload()
		}
	}()

	for {
		select {
		case <-c.ticker.C:
			c.reloadAll()
		case <-c.done:
			return
		}
	}
}

// reloadAll reloads each DB file in turn.
func (c *Client) reloadAll() {
	c.reloadDB(&c.muCountry, &c.country, CountryDatabase)
	c.reloadDB(&c.muCity, &c.city, CityDatabase)
	c.reloadDB(&c.muASN, &c.asn, ASNDatabase)
}

// reloadDB opens the filename, swaps it in under mu, closes the old reader.
func (c *Client) reloadDB(mu *sync.RWMutex, ptr **geoip2.Reader, filename string) {
	newPath := path.Join(c.DataDirectory, filename)
	newDB, err := geoip2.Open(newPath)
	if err != nil {
		log.Printf("Failed to open %s: %v", filename, err)
		return
	}

	mu.Lock()
	old := *ptr
	*ptr = newDB
	mu.Unlock()

	if err := old.Close(); err != nil {
		log.Printf("Failed to close old %s: %v", filename, err)
	}
}

// CityDB returns the current city database reader.
func (c *Client) CityDB() *geoip2.Reader {
	c.muCity.RLock()
	defer c.muCity.RUnlock()
	return c.city
}

// CountryDB returns the current country database reader.
func (c *Client) CountryDB() *geoip2.Reader {
	c.muCountry.RLock()
	defer c.muCountry.RUnlock()
	return c.country
}

// AsnDB returns the current ASN database reader.
func (c *Client) AsnDB() *geoip2.Reader {
	c.muASN.RLock()
	defer c.muASN.RUnlock()
	return c.asn
}

// Close stops the reload loop and closes all readers.
func (c *Client) Close() error {
	c.ticker.Stop()
	close(c.done)

	c.muCountry.Lock()
	_ = c.country.Close()
	c.muCountry.Unlock()

	c.muCity.Lock()
	_ = c.city.Close()
	c.muCity.Unlock()

	c.muASN.Lock()
	_ = c.asn.Close()
	c.muASN.Unlock()

	return nil
}
