# firecrawl-go v2

Go SDK for the [Firecrawl](https://firecrawl.dev) v2 API. Scrape, crawl, map, search, batch-scrape, and extract structured data from websites — with output formatted for LLMs.

> **Fork of [firecrawl/firecrawl-go](https://github.com/ArmandoHerra/firecrawl-go)** — migrated to Firecrawl API v2 with typed request structs, `context.Context` on every method, typed errors, security hardening, functional client options, and a modern CI pipeline.

## Installation

```bash
go get github.com/firecrawl/firecrawl-go/v2
```

Requires Go 1.23+.

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	firecrawl "github.com/firecrawl/firecrawl-go/v2"
)

func main() {
	app, err := firecrawl.NewFirecrawlApp("YOUR_API_KEY", "")
	if err != nil {
		log.Fatal(err)
	}

	doc, err := app.ScrapeURL(context.Background(), "https://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(doc.Markdown)
}
```

## API Methods

All methods accept `context.Context` as the first parameter for cancellation and deadlines.

### Scrape

| Method | Endpoint | Description |
|--------|----------|-------------|
| `ScrapeURL(ctx, url, params)` | `POST /v2/scrape` | Scrape a single URL, returns markdown/HTML/JSON/screenshot |

### Crawl

| Method | Endpoint | Description |
|--------|----------|-------------|
| `CrawlURL(ctx, url, params, key, pollInterval...)` | `POST /v2/crawl` | Start a crawl and poll until complete |
| `AsyncCrawlURL(ctx, url, params, key)` | `POST /v2/crawl` | Start an async crawl, returns job ID |
| `CheckCrawlStatus(ctx, id, pagination...)` | `GET /v2/crawl/{id}` | Check status; optional auto-pagination |
| `GetCrawlStatusPage(ctx, nextURL)` | `GET /v2/crawl/{id}?cursor=...` | Fetch one page manually (for manual pagination) |
| `CancelCrawlJob(ctx, id)` | `DELETE /v2/crawl/{id}` | Cancel a running crawl |

### Map

| Method | Endpoint | Description |
|--------|----------|-------------|
| `MapURL(ctx, url, params)` | `POST /v2/map` | Discover all URLs on a site, returns `[]MapLink` |

### Search

| Method | Endpoint | Description |
|--------|----------|-------------|
| `Search(ctx, query, params)` | `POST /v2/search` | Web/image/news search with optional content scraping |

### Batch Scrape

| Method | Endpoint | Description |
|--------|----------|-------------|
| `BatchScrapeURLs(ctx, urls, params, key, pollInterval...)` | `POST /v2/batch/scrape` | Scrape multiple URLs, poll until complete |
| `AsyncBatchScrapeURLs(ctx, urls, params, key)` | `POST /v2/batch/scrape` | Start batch scrape async, returns job ID |
| `CheckBatchScrapeStatus(ctx, id, pagination...)` | `GET /v2/batch/scrape/{id}` | Check status; optional auto-pagination |
| `GetBatchScrapeStatusPage(ctx, nextURL)` | `GET /v2/batch/scrape/{id}?cursor=...` | Fetch one page manually |

### Extract

| Method | Endpoint | Description |
|--------|----------|-------------|
| `Extract(ctx, urls, params)` | `POST /v2/extract` | LLM-based structured extraction, poll until complete |
| `AsyncExtract(ctx, urls, params)` | `POST /v2/extract` | Start extraction async, returns job ID |
| `CheckExtractStatus(ctx, id)` | `GET /v2/extract/{id}` | Check extraction job status |

## Usage Examples

### Scrape with Options

```go
func ptr[T any](v T) *T { return &v }

doc, err := app.ScrapeURL(ctx, "https://example.com", &firecrawl.ScrapeParams{
	Formats:         []string{"markdown", "html"},
	OnlyMainContent: ptr(true),
	Mobile:          ptr(true),
	BlockAds:        ptr(true),
	Location:        &firecrawl.LocationConfig{Country: "US", Languages: []string{"en"}},
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(doc.Markdown)
```

### Crawl a Website (Synchronous)

```go
result, err := app.CrawlURL(ctx, "https://example.com", &firecrawl.CrawlParams{
	Limit:             ptr(100),
	MaxDiscoveryDepth: ptr(3),
	CrawlEntireDomain: ptr(true),
	Sitemap:           ptr("include"),
	ExcludePaths:      []string{"blog/*"},
}, nil) // nil idempotency key
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Scraped %d pages\n", len(result.Data))
```

### Async Crawl with Context Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

crawlResp, err := app.AsyncCrawlURL(ctx, "https://example.com", nil, nil)
if err != nil {
	log.Fatal(err)
}

// Check status (single page)
status, err := app.CheckCrawlStatus(ctx, crawlResp.ID)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Status: %s, Pages: %d/%d\n", status.Status, status.Completed, status.Total)
```

### Search

```go
results, err := app.Search(ctx, "go generics tutorial", &firecrawl.SearchParams{
	Limit:   ptr(5),
	Country: ptr("US"),
	Sources: []string{"web", "news"},
})
if err != nil {
	log.Fatal(err)
}
for _, r := range results.Data.Web {
	fmt.Printf("%s — %s\n", r.Title, r.URL)
}
```

### Batch Scrape (Synchronous)

```go
urls := []string{
	"https://example.com",
	"https://example.org",
	"https://example.net",
}

result, err := app.BatchScrapeURLs(ctx, urls, &firecrawl.BatchScrapeParams{
	ScrapeOptions: firecrawl.ScrapeParams{
		Formats:         []string{"markdown"},
		OnlyMainContent: ptr(true),
	},
	MaxConcurrency: ptr(5),
}, nil) // nil idempotency key
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Scraped %d URLs\n", len(result.Data))
```

### Async Batch Scrape with Manual Pagination

```go
batchResp, err := app.AsyncBatchScrapeURLs(ctx, urls, nil, nil)
if err != nil {
	log.Fatal(err)
}

// Check status — get first page
status, err := app.CheckBatchScrapeStatus(ctx, batchResp.ID)
if err != nil {
	log.Fatal(err)
}

// Manually iterate pages
for status.Next != nil {
	status, err = app.GetBatchScrapeStatusPage(ctx, *status.Next)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Page data: %d results\n", len(status.Data))
}
```

### Extract Structured Data

```go
schema := map[string]any{
	"type": "object",
	"properties": map[string]any{
		"company_name": map[string]any{"type": "string"},
		"founded":      map[string]any{"type": "integer"},
		"employees":    map[string]any{"type": "integer"},
	},
}

result, err := app.Extract(ctx, []string{"https://example.com/about"}, &firecrawl.ExtractParams{
	Prompt: ptr("Extract company information including name, founding year, and employee count."),
	Schema: schema,
})
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Extracted: %v\n", result.Data)
```

### Map a Website

```go
mapResp, err := app.MapURL(ctx, "https://example.com", &firecrawl.MapParams{
	Limit:   ptr(5000),
	Sitemap: ptr("include"),
})
if err != nil {
	log.Fatal(err)
}
for _, link := range mapResp.Links {
	fmt.Printf("%s\n", link.URL)
}
```

## Pagination

For large crawls and batch scrapes, the API returns paginated results with a `Next` URL.

### Auto-Pagination (Recommended)

Pass a `PaginationConfig` to `CheckCrawlStatus` or `CheckBatchScrapeStatus` to automatically collect all pages:

```go
result, err := app.CheckCrawlStatus(ctx, crawlID, &firecrawl.PaginationConfig{
	AutoPaginate: ptr(true),
	MaxPages:     ptr(10),    // stop after 10 pages
	MaxResults:   ptr(1000),  // stop after 1000 total results
	MaxWaitTime:  ptr(60),    // stop after 60 seconds
})
```

### Manual Pagination

Use `GetCrawlStatusPage` / `GetBatchScrapeStatusPage` to fetch one page at a time:

```go
status, err := app.CheckCrawlStatus(ctx, crawlID)
for status.Next != nil {
	status, err = app.GetCrawlStatusPage(ctx, *status.Next)
	if err != nil {
		break
	}
	// process status.Data for this page
}
```

## Error Handling

The SDK uses typed errors enabling `errors.Is` and `errors.As` for programmatic handling.

### Sentinel Errors

| Sentinel | HTTP Status | Meaning |
|----------|-------------|---------|
| `ErrNoAPIKey` | — | No API key provided to constructor |
| `ErrUnauthorized` | 401 | Invalid or expired API key |
| `ErrPaymentRequired` | 402 | Account credit limit reached |
| `ErrNotFound` | 404 | Resource not found |
| `ErrTimeout` | 408 | Request timed out |
| `ErrConflict` | 409 | Conflicting operation (e.g., duplicate idempotency key) |
| `ErrRateLimited` | 429 | Rate limit exceeded |
| `ErrServerError` | 500 | Internal server error |

### errors.Is — Check Error Type

```go
_, err := app.ScrapeURL(ctx, url, nil)
if errors.Is(err, firecrawl.ErrRateLimited) {
	time.Sleep(5 * time.Second)
	// retry...
}
if errors.Is(err, firecrawl.ErrUnauthorized) {
	log.Fatal("Check your API key")
}
```

### errors.As — Access Full Error Details

```go
var apiErr *firecrawl.APIError
if errors.As(err, &apiErr) {
	log.Printf("HTTP %d during %s: %s", apiErr.StatusCode, apiErr.Action, apiErr.Message)
}
```

## Configuration

### Default Constructor

```go
app, err := firecrawl.NewFirecrawlApp("fc-your-api-key", "")
// API URL defaults to https://api.firecrawl.dev
// Timeout defaults to 120 seconds
```

Falls back to environment variables if arguments are empty:
- `FIRECRAWL_API_KEY` — API key
- `FIRECRAWL_API_URL` — API base URL

### Functional Options Constructor

```go
app, err := firecrawl.NewFirecrawlAppWithOptions(
	"fc-your-api-key",
	"",
	firecrawl.WithTimeout(30*time.Second),
	firecrawl.WithUserAgent("my-app/1.0"),
)
```

Available options:

| Option | Default | Description |
|--------|---------|-------------|
| `WithTimeout(d)` | 120s | HTTP client timeout |
| `WithTransport(t)` | `http.DefaultTransport` clone | Custom HTTP transport |
| `WithUserAgent(ua)` | `firecrawl-go/2.0.0` | User-Agent header |
| `WithMaxIdleConns(n)` | 100 | Max idle keep-alive connections |
| `WithMaxIdleConnsPerHost(n)` | 10 | Max idle connections per host |

### API Key Access

The `apiKey` field is unexported. Use the `APIKey()` accessor method:

```go
fmt.Println(app.APIKey())  // "fc-abc...xyz"
fmt.Println(app.String())  // "FirecrawlApp{url: ..., key: fc-a...xyz}" (redacted)
```

## Security

- **API key unexported** — the `apiKey` field is unexported; use `APIKey()` to read it. `String()` returns a redacted representation.
- **URL validation** — all job IDs are validated as UUIDs before being interpolated into request paths, preventing path injection attacks.
- **Pagination SSRF prevention** — `Next` URLs from the API are validated to share the same host as the configured `APIURL` before any request is made.
- **HTTP warning** — a log warning is emitted when a non-localhost HTTP (non-TLS) URL is used, because the API key would be sent in cleartext.

## Tech Stack

| Technology | Version | Purpose |
|-----------|---------|---------|
| Go | 1.23+ | Language runtime |
| golangci-lint | v2.x | Linting (errcheck, govet, staticcheck, gosec, etc.) |
| gofumpt | latest | Code formatting |
| GitHub Actions | — | CI: lint + test matrix (Go 1.23/1.24/1.25) |
| testify | v1.10 | Test assertions (integration tests) |

## Project Structure

```
firecrawl-go/
├── client.go              # FirecrawlApp struct, NewFirecrawlApp, NewFirecrawlAppWithOptions
├── client_options.go      # ClientOption type, WithTimeout, WithTransport, WithUserAgent, etc.
├── types.go               # All request/response type definitions (35+ v2 types)
├── scrape.go              # ScrapeURL — POST /v2/scrape
├── crawl.go               # CrawlURL, AsyncCrawlURL, CheckCrawlStatus, GetCrawlStatusPage, CancelCrawlJob
├── map.go                 # MapURL — POST /v2/map
├── search.go              # Search — POST /v2/search
├── batch.go               # BatchScrapeURLs, AsyncBatchScrapeURLs, CheckBatchScrapeStatus, GetBatchScrapeStatusPage
├── extract.go             # Extract, AsyncExtract, CheckExtractStatus
├── errors.go              # APIError, sentinel errors (ErrUnauthorized, ErrRateLimited, etc.)
├── security.go            # validateJobID, validatePaginationURL
├── helpers.go             # makeRequest, monitorJobStatus — internal HTTP + polling
├── options.go             # requestOptions, withRetries, withBackoff — internal retry config
├── firecrawl.go           # Package doc comment
├── client_test.go         # Unit tests: constructor, options, security
├── scrape_test.go         # Unit tests: ScrapeURL
├── crawl_test.go          # Unit tests: CrawlURL, CheckCrawlStatus, pagination
├── map_test.go            # Unit tests: MapURL
├── search_test.go         # Unit tests: Search
├── batch_test.go          # Unit tests: BatchScrapeURLs, CheckBatchScrapeStatus, pagination
├── extract_test.go        # Unit tests: Extract, CheckExtractStatus
├── errors_test.go         # Unit tests: APIError, sentinel errors, Unwrap
├── helpers_test.go        # Unit tests: makeRequest, retry logic
├── security_test.go       # Unit tests: validateJobID, validatePaginationURL
├── types_test.go          # Unit tests: StringOrStringSlice JSON unmarshaling
├── testhelpers_test.go    # Shared test helpers: ptr[T](), test server setup
├── firecrawl_test.go      # Integration/E2E tests (//go:build integration, 32 tests)
├── Makefile               # build, test, test-integration, lint, fmt, vet, coverage, check
├── .golangci.yml          # golangci-lint v2 configuration
├── .github/
│   ├── workflows/ci.yml   # CI pipeline (lint + test matrix + integration)
│   └── dependabot.yml     # Automated dependency updates
├── .editorconfig          # Editor settings
├── .env.example           # Environment template for integration tests
├── go.mod / go.sum        # Module: github.com/firecrawl/firecrawl-go/v2
├── CHANGELOG.md           # Keep a Changelog format — all migration and improvement changes
└── LICENSE                # MIT
```

## Available Commands

| Command | Description |
|---------|-------------|
| `make help` | Show all available targets |
| `make build` | Compile the library |
| `make test` | Run unit tests (no API key needed, 160 tests) |
| `make test-integration` | Run integration/E2E tests (requires `.env`) |
| `make lint` | Run golangci-lint |
| `make fmt` | Format code with gofumpt |
| `make vet` | Run go vet |
| `make coverage` | Generate HTML coverage report |
| `make clean` | Remove generated files |
| `make check` | Run lint + vet + test (full pre-commit check) |

## Testing

### Unit Tests (no API key needed)

```bash
make test
# or: go test -race -v -count=1 ./...
```

160 unit tests run using `httptest.NewServer` mock servers. No `.env` or API key required.

### Integration Tests (live API)

```bash
cp .env.example .env
# Edit .env:
#   API_URL=https://api.firecrawl.dev
#   TEST_API_KEY=fc-your-api-key
make test-integration
# or: go test -race -v -count=1 -tags=integration ./...
```

32 E2E tests hit the live Firecrawl v2 API. These consume API credits.

### Environment Variables

| Variable | Used By | Required For |
|----------|---------|-------------|
| `FIRECRAWL_API_KEY` | SDK runtime | Production use (constructor fallback) |
| `FIRECRAWL_API_URL` | SDK runtime | Custom API URL (defaults to `https://api.firecrawl.dev`) |
| `TEST_API_KEY` | Integration tests | `make test-integration` |
| `API_URL` | Integration tests | `make test-integration` |

## Development

### Prerequisites

- Go 1.23+
- golangci-lint v2 (`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`)
- gofumpt (`go install mvdan.cc/gofumpt@latest`)

### Setup

```bash
git clone git@github.com:firecrawl/firecrawl-go.git
cd firecrawl-go
go mod download
make check  # lint + vet + test
```

### Development Loop

```bash
make fmt      # Format with gofumpt
make check    # Lint + vet + all unit tests
# Commit — pre-commit hook runs make check automatically
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for setup, code style, and pull request guidelines.

## License

MIT License. See [LICENSE](LICENSE) for details.
