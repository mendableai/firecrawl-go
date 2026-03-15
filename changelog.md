## [MIG-03: Foundation — CI/CD Pipeline Setup] - 2026-03-15

### Added
- `Makefile` — `help`, `build`, `test`, `test-integration`, `lint`, `fmt`, `vet`, `coverage`, `clean`, `check` targets; `.DEFAULT_GOAL := help`
- `.golangci.yml` — golangci-lint config enabling errcheck (with check-type-assertions), govet (enable-all), staticcheck, gosimple, unused, ineffassign, gofumpt, misspell, bodyclose, noctx, gosec (G402 excluded), prealloc; 5m timeout
- `.github/workflows/ci.yml` — Three-job CI pipeline: `lint` (Go 1.23, golangci-lint-action v6), `test` (matrix Go 1.22/1.23, race detector, 80% coverage threshold), `integration` (push to main only, needs lint+test, uses FIRECRAWL_API_KEY secret)
- `.github/dependabot.yml` — Weekly updates for gomod and github-actions ecosystems
- `.editorconfig` — Tabs for Go/Makefile, spaces for YAML, LF line endings, final newline
- `go build ./...` and `go vet ./...` both verified passing via Makefile targets

### Changed
- `.gitignore` — Added `coverage.out`, `coverage.html`, `*.test`, `*.prof`; `vendor` corrected to `vendor/`
- `.env.example` — Updated API_URL to `https://api.firecrawl.dev` (was localhost), added descriptive comment

### Fixed
- Deleted `firecrawl_test.go_V0` (dead v0 test file with no build tag; was included in `go test ./...` but all tests required an API key)

### Notes
- `make build` passes clean
- `make vet` passes clean
- CI coverage threshold (80%) will be enforced once unit tests are added in MIG-07 (IMP-06)
- Concurrency group cancels in-progress runs on same ref to avoid redundant CI runs

## [MIG-02: Foundation — File Splitting] - 2026-03-15

### Added
- `client.go` — `FirecrawlApp` struct, `NewFirecrawlApp` constructor, `prepareHeaders` method
- `types.go` — All request/response type definitions: `StringOrStringSlice`, `FirecrawlDocumentMetadata`, `JsonOptions`, `FirecrawlDocument`, `ScrapeParams`, `ScrapeResponse`, `CrawlParams`, `CrawlResponse`, `CrawlStatusResponse`, `CancelCrawlJobResponse`, `MapParams`, `MapResponse`
- `options.go` — `requestOptions` struct, `requestOption` type, `newRequestOptions`, `withRetries`, `withBackoff`
- `scrape.go` — `ScrapeURL` method
- `crawl.go` — `CrawlURL`, `AsyncCrawlURL`, `CheckCrawlStatus`, `CancelCrawlJob` methods
- `map.go` — `MapURL` method
- `search.go` — `Search` stub method
- `errors.go` — `handleError` method
- `helpers.go` — `makeRequest`, `monitorJobStatus` methods

### Changed
- `firecrawl.go` — Reduced to package doc comment only; all code moved to dedicated files above

### Notes
- Pure structural refactor — zero logic changes
- `go build ./...` passes clean
- `go vet ./...` passes clean
- All files use `package firecrawl`; each file imports only what it needs

## [MIG-01: Foundation — Bug Fixes] - 2026-03-15

### Fixed
- `monitorJobStatus`: retry counter `attempts` initialized to `0` instead of `3`; the old value caused the "completed but no data" branch to error immediately without retrying
- `makeRequest`: removed `defer resp.Body.Close()` from inside the retry loop; intermediate 502 response bodies are now closed explicitly before each retry, and the final response body is deferred after the loop — eliminates HTTP connection leaks under retry conditions
- `makeRequest`: request body (`bytes.NewBuffer(body)`) and headers are now recreated inside the retry loop for each attempt; the old code consumed the buffer on the first `Do()` call, causing all subsequent retries to send an empty body
- `ScrapeURL`: `json.Unmarshal` error is now checked before accessing `scrapeResponse.Success`; the old ordering could silently return corrupted data or swallow the unmarshal error
- `CrawlURL` / `AsyncCrawlURL`: `scrapeOptions` is now included in the request body when any field of `ScrapeOptions` is non-zero, not just when `Formats` is non-nil; the old gate dropped all other scrape options (headers, tags, timeouts, etc.) silently

### Changed
- `ScrapeURL`: removed 17 lines of commented-out extractor code (v0 legacy dead code)

### Notes
- `go build ./...` passes clean with no warnings
- No existing tests were broken; no new tests added (IMP-06/IMP-07 will cover test additions)
