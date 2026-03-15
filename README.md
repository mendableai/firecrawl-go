# Firecrawl Go SDK

Go client library for the [Firecrawl API v2](https://docs.firecrawl.dev/api-reference/v2-introduction). Scrape, crawl, and map websites with output formatted for LLMs.

> **Fork of [firecrawl/firecrawl-go](https://github.com/firecrawl/firecrawl-go)** — migrated to Firecrawl API v2 with expanded parameters, typed request structs, `context.Context` support, and a modern CI pipeline.

## Quick Start

```bash
go get github.com/firecrawl/firecrawl-go/v2
```

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/firecrawl/firecrawl-go/v2"
)

func main() {
	app, err := firecrawl.NewFirecrawlApp("YOUR_API_KEY", "")
	if err != nil {
		log.Fatal(err)
	}

	// Scrape a URL
	doc, err := app.ScrapeURL(context.Background(), "https://example.com", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(doc.Markdown)
}
```

## Tech Stack

| Technology | Version | Purpose |
|-----------|---------|---------|
| Go | 1.23+ | Language runtime |
| golangci-lint | v2.x | Linting (errcheck, govet, staticcheck, gosec, etc.) |
| gofumpt | latest | Code formatting |
| GitHub Actions | CI | Lint + test matrix (Go 1.23/1.24/1.25) |
| testify | v1.10 | Test assertions (integration tests) |

## Project Structure

```
firecrawl-go/
├── client.go          # FirecrawlApp struct, NewFirecrawlApp(), prepareHeaders()
├── types.go           # All request/response type definitions (31 v2 types)
├── scrape.go          # ScrapeURL — POST /v2/scrape
├── crawl.go           # CrawlURL, AsyncCrawlURL, CheckCrawlStatus, CancelCrawlJob
├── map.go             # MapURL — POST /v2/map
├── search.go          # Search — stub (v2 implementation pending)
├── errors.go          # handleError — HTTP error mapping
├── helpers.go         # makeRequest, monitorJobStatus — internal HTTP + polling
├── options.go         # requestOptions, withRetries(), withBackoff()
├── firecrawl.go       # Package doc comment
├── firecrawl_test.go  # Integration tests (gated: //go:build integration)
├── Makefile           # Build, test, lint, coverage targets
├── .golangci.yml      # golangci-lint v2 configuration
├── .github/
│   ├── workflows/ci.yml   # CI pipeline (lint + test matrix + integration)
│   └── dependabot.yml     # Automated dependency updates
├── .editorconfig      # Editor settings
├── .env.example       # Environment template for integration tests
├── go.mod / go.sum    # Module: github.com/firecrawl/firecrawl-go/v2
├── changelog.md       # Migration changelog
└── LICENSE            # MIT
```

## API Methods

All methods accept `context.Context` as the first parameter for cancellation and deadlines.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `ScrapeURL(ctx, url, params)` | `POST /v2/scrape` | Scrape a single URL, returns markdown/HTML/JSON |
| `CrawlURL(ctx, url, params, key, pollInterval)` | `POST /v2/crawl` | Synchronous crawl with polling until complete |
| `AsyncCrawlURL(ctx, url, params, key)` | `POST /v2/crawl` | Start async crawl, returns job ID |
| `CheckCrawlStatus(ctx, id)` | `GET /v2/crawl/{id}` | Check crawl job status and retrieve results |
| `CancelCrawlJob(ctx, id)` | `DELETE /v2/crawl/{id}` | Cancel a running crawl job |
| `MapURL(ctx, url, params)` | `POST /v2/map` | Discover URLs on a site (returns MapLink objects) |
| `Search(ctx, query, params)` | — | Not yet implemented (pending IMP-01) |

## Usage Examples

### Scrape with Options

```go
ctx := context.Background()

doc, err := app.ScrapeURL(ctx, "https://example.com", &firecrawl.ScrapeParams{
	Formats:         []string{"markdown", "html"},
	OnlyMainContent: ptr(true),
	Mobile:          ptr(true),
	BlockAds:        ptr(true),
	Location:        &firecrawl.LocationConfig{Country: "US", Languages: []string{"en"}},
})
```

### Crawl a Website

```go
ctx := context.Background()

result, err := app.CrawlURL(ctx, "https://example.com", &firecrawl.CrawlParams{
	Limit:             ptr(100),
	MaxDiscoveryDepth: ptr(3),
	CrawlEntireDomain: ptr(true),
	Sitemap:           ptr("include"),
	ExcludePaths:      []string{"blog/*"},
}, nil) // no idempotency key
```

### Async Crawl with Context Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

crawlResp, err := app.AsyncCrawlURL(ctx, "https://example.com", nil, nil)
if err != nil {
	log.Fatal(err)
}

// Poll for status
status, err := app.CheckCrawlStatus(ctx, crawlResp.ID)
```

### Map a Website

```go
ctx := context.Background()

mapResp, err := app.MapURL(ctx, "https://example.com", &firecrawl.MapParams{
	Limit:   ptr(5000),
	Sitemap: ptr("include"),
})
// mapResp.Links is []MapLink with URL, Title, Description
for _, link := range mapResp.Links {
	fmt.Printf("%s — %s\n", link.URL, *link.Title)
}
```

## Available Commands

| Command | Description |
|---------|-------------|
| `make help` | Show all available targets |
| `make build` | Compile the library |
| `make test` | Run unit tests (no API key needed) |
| `make test-integration` | Run integration tests (requires `.env`) |
| `make lint` | Run golangci-lint |
| `make fmt` | Format code with gofumpt |
| `make vet` | Run go vet |
| `make coverage` | Generate HTML coverage report |
| `make clean` | Remove generated files |
| `make check` | Run lint + vet + test (full pre-commit check) |

## Configuration

### Environment Variables

| Variable | Used By | Required For |
|----------|---------|-------------|
| `FIRECRAWL_API_KEY` | SDK runtime | Production (fallback if not passed to constructor) |
| `FIRECRAWL_API_URL` | SDK runtime | Custom API URL (defaults to `https://api.firecrawl.dev`) |
| `TEST_API_KEY` | Integration tests | `make test-integration` |
| `API_URL` | Integration tests | `make test-integration` |

### Config Files

| File | Purpose |
|------|---------|
| `.env.example` | Template for integration test credentials |
| `.golangci.yml` | Linter configuration (golangci-lint v2) |
| `.editorconfig` | Editor settings (tabs for Go, spaces for YAML) |
| `.github/workflows/ci.yml` | CI pipeline definition |
| `.github/dependabot.yml` | Dependency update schedule |

## Development

### Prerequisites

- Go 1.23+
- golangci-lint v2 (`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`)
- gofumpt (`go install mvdan.cc/gofumpt@latest`)

### Setup

```bash
git clone git@github.com:ArmandoHerra/firecrawl-go.git
cd firecrawl-go
go mod download
make check  # lint + vet + test
```

### Development Loop

```bash
# Edit code...
make fmt      # Format
make check    # Lint + vet + test
# Commit (pre-commit hook runs make check automatically)
```

## Testing

### Unit Tests

```bash
make test  # No API key needed
```

Unit tests use `httptest.NewServer` for mock-based testing (pending implementation via IMP-06/07).

### Integration Tests

```bash
cp .env.example .env
# Edit .env with your API key
make test-integration  # Hits live Firecrawl API
```

Integration tests are gated behind `//go:build integration` and will not run with `make test`.

## License

MIT License. See [LICENSE](LICENSE) for details.

This SDK is a fork of [firecrawl/firecrawl-go](https://github.com/firecrawl/firecrawl-go). The upstream project may have different licensing terms.
