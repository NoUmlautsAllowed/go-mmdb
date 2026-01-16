package mmdb

import (
	"net"
	"net/http"
	"strings"
)

type IPInfo struct {
	IP            net.IP `json:"ip"`
	CountryCode   string `json:"country_code"`
	ASN           string `json:"asn"`
	City          string `json:"city"`
	CityBuildDate uint   `json:"city_build_date"`
	ASNBuildDate  uint   `json:"asn_build_date"`
}

func (c *Client) IPInfoFromRequest(r *http.Request) IPInfo {
	ip := net.ParseIP(clientIP(r))

	return c.IPInfo(ip)
}

func (c *Client) IPInfo(ip net.IP) IPInfo {
	var info IPInfo

	if ip == nil {
		return info
	}

	info.IP = ip

	if cityDB := c.CityDB(); cityDB != nil {
		LookupTotal.WithLabelValues("city").Inc()
		info.CityBuildDate = cityDB.Metadata().BuildEpoch
		if rec, err := cityDB.City(ip); err == nil {
			// is country code
			info.CountryCode = rec.Country.IsoCode

			// city name (English)
			if name, ok := rec.City.Names["en"]; ok {
				info.City = name
			}
		}
	}

	if asnDB := c.AsnDB(); asnDB != nil {
		LookupTotal.WithLabelValues("asn").Inc()
		info.ASNBuildDate = asnDB.Metadata().BuildEpoch
		if rec, err := asnDB.ASN(ip); err == nil {
			info.ASN = rec.AutonomousSystemOrganization
		}
	}

	return info
}

// clientIP tries X-Forwarded-For, then falls back to RemoteAddr.
func clientIP(r *http.Request) string {
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		// may be a comma-separated list; take first
		parts := strings.Split(xf, ",")
		return strings.TrimSpace(parts[0])
	}
	host := r.RemoteAddr
	if host == "" {
		return ""
	}
	// strip port if present
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
}
