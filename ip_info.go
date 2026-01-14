package mmdb

import (
	"net"
	"net/http"
	"strings"
)

type IpInfo struct {
	IP          string
	CountryCode string
	ASN         string
	City        string
}

func (c *Client) GetIpInfo(r *http.Request) IpInfo {
	var info IpInfo

	// 1) determine client IP
	ip := clientIP(r)
	if ip == "" {
		return info
	}
	info.IP = ip

	// 2) parse it
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return info
	}

	// 3) lookup City
	if cityDB := c.CityDB(); cityDB != nil {
		if rec, err := cityDB.City(parsedIP); err == nil {
			// is country code
			info.CountryCode = rec.Country.IsoCode

			// city name (English)
			if name, ok := rec.City.Names["en"]; ok {
				info.City = name
			}
		}
	}

	// 4) lookup ASN
	if asnDB := c.AsnDB(); asnDB != nil {
		if rec, err := asnDB.ASN(parsedIP); err == nil {
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
