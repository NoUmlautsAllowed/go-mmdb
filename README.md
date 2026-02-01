# go-mmdb

`go-mmdb` is a comprehensive Go library and server for MaxMind GeoIP2 databases (MMDB). It provides high-level tools for querying City, Country, and ASN data with a focus on ease of use, performance, and zero-downtime updates.

## ✨ Features

- **Automated Downloads**: Periodically fetches and extracts the latest MaxMind databases using your license key.
- **Zero-Downtime Updates**: Uses atomic renames and periodic reloads to update databases without interrupting active queries.
- **Unified IP Lookups**: Combines data from City and ASN databases into a single, easy-to-use `IPInfo` struct.
- **Prometheus Metrics**: Built-in instrumentation for monitoring HTTP requests, lookups, and database downloads.
- **Embedded HTTP Server**: Ready-to-use server providing HTML, JSON, and Plain Text interfaces.
- **Thread-Safe**: Designed for high-concurrency environments.

## 🚀 Getting Started

### Installation

```bash
go get github.com/noumlautsallowed/go-mmdb
```

### Quick Start (Server)

1. Set your MaxMind credentials:
   ```bash
   export MAXMIND_ACCOUNT_ID=your_id
   export MAXMIND_LICENSE_KEY=your_key
   ```
2. Run the server:
   ```bash
   go run github.com/NoUmlautsAllowed/go-mmdb/cmd/server
   ```
3. Access the interface:
   - **Web UI**: `http://localhost:8080/`
   - **JSON API**: `http://localhost:8080/?format=json` or `Accept: application/json`
   - **Plain Text**: `http://localhost:8080/?format=text` or `Accept: text/plain`

## ⚙️ Configuration

The application can be configured using environment variables or a `.env` file:

| Variable              | Description                                        | Default          |
|:----------------------|:---------------------------------------------------|:-----------------|
| `MAXMIND_ACCOUNT_ID`  | Your MaxMind Account ID (Required for downloader)  | -                |
| `MAXMIND_LICENSE_KEY` | Your MaxMind License Key (Required for downloader) | -                |
| `MAXMIND_BASE_PATH`   | Directory where `.mmdb` files are stored           | `.`              |
| `BIND_ADDR`           | Address for the built-in HTTP server               | `localhost:8080` |
| `METRICS_ADDR`        | Address for the Prometheus metrics server          | `localhost:9090` |

## 📊 Metrics

Prometheus metrics are exposed at `http://<METRICS_ADDR>/metrics`.

Key metrics include:
- `mmdb_http_requests_total`: HTTP request counter.
- `mmdb_http_request_duration_seconds`: HTTP request latency histogram.
- `mmdb_lookup_total`: IP lookup counter (labels: `type`).
- `mmdb_download_total`: Database download status tracker (labels: `database`, `status`).

## 🛠️ Usage as a Library

```go
package main

import (
    "fmt"
    "log"
    "net"
    "github.com/noumlautsallowed/go-mmdb"
)

func main() {
    // Initialize client (automatically manages DB reloads)
    client, err := mmdb.NewClient()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Lookup IP info
    ip := net.ParseIP("8.8.8.8")
    info := client.IPInfo(ip)

    fmt.Printf("City: %s, Country: %s, ASN: %s\n", info.City, info.CountryCode, info.ASN)
}
```

## 🔍 How it Works

`go-mmdb` ensures your application always uses the latest GeoIP data without restart:

1. **Downloader**: Fetches new `.mmdb` files to a temporary location.
2. **Atomic Swap**: Replaces the active database file using an atomic rename.
3. **Transparent Reload**: The `Client` detects the file change (every 2 hours), opens the new reader, and gracefully closes the old one. Existing queries are not affected as they continue to use the open file handle (inode) until completion.

## 📄 License

This project is licensed under the MIT License. MaxMind GeoLite2 databases are subject to the [MaxMind EULA](https://www.maxmind.com/en/geolite2/eula).
