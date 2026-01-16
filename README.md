# go-mmdb

`go-mmdb` is a Go library and server designed to simplify working with MaxMind GeoIP2 databases (MMDB). It provides a high-level client for querying City, Country, and ASN databases, an automated downloader to keep these databases up-to-date, and an embedded HTTP server for IP information lookups.

## Features

- **Automated Downloads**: Periodically fetches and extracts the latest MaxMind databases using your license key.
- **Zero-Downtime Updates**: Uses atomic renames and periodic reloads to update databases without interrupting active queries.
- **Unified IP Lookups**: Combines data from City and ASN databases into a single, easy-to-use `IPInfo` struct.
- **Prometheus Metrics**: Built-in instrumentation for monitoring HTTP requests, lookups, and database downloads.
- **HTTP Server**: Built-in server providing both HTML and JSON interfaces for IP lookups.
- **Thread-Safe**: Designed for concurrent use in high-traffic applications.

## Installation

```bash
go get gitlab.w1lhelm.de/swilhelm/go-mmdb
```

## Configuration

The following environment variables are used to configure the downloader and client:

| Variable              | Description                                        | Default          |
|-----------------------|----------------------------------------------------|------------------|
| `MAXMIND_ACCOUNT_ID`  | Your MaxMind Account ID (Required for downloader)  | -                |
| `MAXMIND_LICENSE_KEY` | Your MaxMind License Key (Required for downloader) | -                |
| `MAXMIND_BASE_PATH`   | Directory where `.mmdb` files are stored           | `.`              |
| `BIND_ADDR`           | Address for the built-in HTTP server               | `localhost:8080` |
| `METRICS_ADDR`        | Address for the Prometheus metrics server           | `:9090`          |

## Metrics

When running the server, Prometheus metrics are exposed at `http://<METRICS_ADDR>/metrics`.

Available metrics:

- `mmdb_http_requests_total`: Total number of HTTP requests (labels: `path`, `method`, `status`).
- `mmdb_http_request_duration_seconds`: Duration of HTTP requests (labels: `path`, `method`).
- `mmdb_lookup_total`: Total number of IP lookups (labels: `type`: `city`, `asn`).
- `mmdb_download_total`: Total number of database downloads (labels: `database`, `status`: `success`, `failure`, `skipped`).

## Usage

### As a Library

```go
import (
    "net"
    "gitlab.w1lhelm.de/swilhelm/go-mmdb"
)

// Initialize client
client, err := mmdb.NewClient()
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Lookup IP info
ip := net.ParseIP("8.8.8.8")
info := client.IPInfo(ip)

fmt.Printf("City: %s, Country: %s, ASN: %s\n", info.City, info.CountryCode, info.ASN)
```

### Running the Server

You can run the included server which provides a web interface and a JSON API:

```bash
export MAXMIND_ACCOUNT_ID=your_id
export MAXMIND_LICENSE_KEY=your_key
go run cmd/server/main.go
```

Access the server at `http://localhost:8080`.
- Append `?format=json` or set the `Accept: application/json` header to receive JSON responses.
- Append `?format=text` or set the `Accept: text/plain` header to receive only the IP address as plain text.

## How it Works

### Updates and Reloads

The `Downloader` and `Client` work together to ensure your application always has the latest data without downtime:

1.  **Downloader**: Downloads new `.mmdb` files to a `.tmp` location.
2.  **Backup**: Moves the current database to a `.old` file.
3.  **Atomic Swap**: Renames the `.tmp` file to the target filename. Since the inode of the open file is preserved, the `Client` continues to work with the old data.
4.  **Reload**: The `Client` periodically (every 2 hours) checks for updated files on disk. When it detects a change, it opens the new file, swaps the internal reader, and closes the old one.

## License

This project is licensed under the MIT License - see the LICENSE file for details (if applicable).
MaxMind GeoLite2 databases are subject to the [MaxMind EULA](https://www.maxmind.com/en/geolite2/eula).
