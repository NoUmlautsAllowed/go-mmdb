# Project Guidelines

## Project Overview
`go-mmdb` is a Go library designed to facilitate working with MaxMind GeoIP2 databases (MMDB). It provides a high-level client for querying City, Country, and ASN databases, an automated downloader to keep these databases up-to-date, and utilities for extracting IP information from HTTP requests.

## Project Structure
- `client.go`: Contains the `Client` struct which manages multiple GeoIP2 database readers and handles periodic reloading of database files from disk.
- `download.go`: Implements the `Downloader` which handles fetching and extracting MaxMind databases.
- `ip_info.go`: Provides the `IPInfo` struct and methods to lookup IP information from `net.IP` or `http.Request`.
- `metrics.go`: Defines Prometheus metrics for monitoring HTTP requests, lookups, and downloads.
- `go.mod`: Project dependencies, notably `github.com/oschwald/geoip2-golang` and `github.com/prometheus/client_golang`.

## Core Components
### Client
The `Client` (in `client.go`) is the main entry point. It opens City, Country, and ASN databases and starts a background goroutine to periodically reload them (defaulting to every 2 hours).

### Downloader
The `Downloader` (in `download.go`) manages the downloading of `.mmdb` files. It checks if a local file needs updating based on the remote's last modified time.

### Concurrency & Updates
The `Downloader` and `Client` are designed to run concurrently. When updating a database:
- The `Downloader` downloads the new file to a temporary location.
- It moves the existing database file to a `.old` backup (e.g., `GeoLite2-City.mmdb` → `GeoLite2-City.mmdb.old`).
- It then renames the new file to the target filename.
- Renaming preserves the inode of the open file, allowing the `Client` to continue querying the old database until its next scheduled reload, at which point it opens the new file and closes the old reader. This design ensures zero-downtime updates and must be preserved.

### IPInfo
The `IPInfo` functionality (in `ip_info.go`) simplifies looking up details about an IP address, combining data from City and ASN databases into a single struct.

### Metrics
The project uses Prometheus for monitoring. Metrics are defined in `metrics.go` and the server (in `cmd/server/main.go`) exposes them on a configurable port (defaulting to `:9090`).

## Development Guidelines
- **Go Version**: The project uses Go 1.24.4 or higher.
- **Dependencies**: Managed via Go modules. Major dependencies include `github.com/oschwald/geoip2-golang` and `github.com/prometheus/client_golang`.
- **Concurrency**: The `Client` is thread-safe, using `sync.RWMutex` to manage access to database readers during reloads.

## Testing
- Use standard Go testing: `go test ./...`.
- Ensure any new functionality is covered by tests.
- When fixing bugs, provide a reproduction test case.

## Code Style
- Follow standard Go formatting (`gofmt`).
- Use descriptive variable names and maintain the existing naming conventions (e.g., camelCase for internal use, PascalCase for exported symbols).
- Comments should be used where necessary to explain complex logic, following the existing style in the codebase.

## Task Completion Checklist
Always check after completing the task, if the following files need to be updated:
- AGENTS.md
- README.md
- .example.env
