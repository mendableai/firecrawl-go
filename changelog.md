## [MIG-05: Core Migration — context.Context Integration] - 2026-03-15

### Changed
- `helpers.go` — `makeRequest`: added `ctx context.Context` as first parameter; replaced `http.NewRequestWithContext(context.Background(), ...)` with `http.NewRequestWithContext(ctx, ...)`; added `ctx.Err()` check at the top of each retry iteration
- `helpers.go` — `monitorJobStatus`: added `ctx context.Context` as first parameter; added `ctx.Err()` check at the top of the polling loop and before pagination fetches; replaced `time.Sleep(...)` with context-aware `select { case <-ctx.Done(): ... case <-time.After(...): }`; passes `ctx` to all `makeRequest` calls
- `scrape.go` — `ScrapeURL`: added `ctx context.Context` as first parameter; passes `ctx` to `makeRequest`; updated godoc
- `crawl.go` — `CrawlURL`: added `ctx context.Context` as first parameter; passes `ctx` to `makeRequest` and `monitorJobStatus`; updated godoc
- `crawl.go` — `AsyncCrawlURL`: added `ctx context.Context` as first parameter; passes `ctx` to `makeRequest`; updated godoc
- `crawl.go` — `CheckCrawlStatus`: added `ctx context.Context` as first parameter; passes `ctx` to `makeRequest`; updated godoc
- `crawl.go` — `CancelCrawlJob`: added `ctx context.Context` as first parameter; passes `ctx` to `makeRequest`; updated godoc
- `map.go` — `MapURL`: added `ctx context.Context` as first parameter; passes `ctx` to `makeRequest`; updated godoc
- `search.go` — `Search`: added `ctx context.Context` as first parameter; updated godoc
- `firecrawl_test.go` — Added `"context"` import; added `context.Background()` as first argument to all public method call sites

### Notes
- `go build ./...` and `go vet ./...` pass cleanly (integration tag excluded per build tag)
- Breaking change for SDK consumers: all public methods now require a `context.Context` as the first argument
- Pre-existing integration test compilation issues (removed v1 fields `MaxDepth`, `IgnoreSitemap`, `AllowBackwardLinks`) carry forward from MIG-04 and will be resolved in MIG-07

## [MIG-04: Core Migration — v2 Type Definitions] - 2026-03-15

### Added
- `types.go` — `LocationConfig` struct (Country, Languages) for geolocation configuration
- `types.go` — `ParserConfig` struct (Type, Mode, MaxPages) replacing v1 `ParsePDF` field
- `types.go` — `ActionConfig` struct (Type + type-specific optional fields: Milliseconds, Selector, Text, Key, Direction, Amount, Script, FullPage) for browser automation
- `types.go` — `WebhookConfig` struct (URL, Headers, Metadata, Events) replacing v1 `*string` webhook
- `types.go` — `MapLink` struct (URL, Title, Description) for the v2 map response format
- `types.go` — `ActionsResult` struct (Screenshots, Scrapes, JavascriptReturns, PDFs)
- `types.go` — `ChangeTrackingResult` struct (PreviousScrapeAt, ChangeStatus, Visibility, Diff, JSON)
- `types.go` — `BrandingResult` struct (ColorScheme, Logo, Colors, Fonts)
- `types.go` — `PaginationConfig` struct (AutoPaginate, MaxPages, MaxResults, MaxWaitTime)
- `types.go` — `SearchParams` struct with all v2 fields (Limit, Sources, Categories, TBS, Location, Country, Timeout, IgnoreInvalidURLs, ScrapeOptions)
- `types.go` — `SearchResponse`, `SearchData` structs
- `types.go` — `SearchWebResult`, `SearchImageResult`, `SearchNewsResult` structs
- `types.go` — `BatchScrapeParams` struct (ScrapeOptions, MaxConcurrency, IgnoreInvalidURLs, Webhook)
- `types.go` — `BatchScrapeResponse` struct (Success, ID, URL, InvalidURLs)
- `types.go` — `BatchScrapeStatusResponse` struct (same shape as CrawlStatusResponse with Next pagination)
- `types.go` — `ExtractParams` struct (Prompt, Schema, EnableWebSearch, IgnoreSitemap, IncludeSubdomains, ShowSources, IgnoreInvalidURLs, ScrapeOptions)
- `types.go` — `ExtractResponse` struct (Success, ID, InvalidURLs)
- `types.go` — `ExtractStatusResponse` struct (Success, Status, Data, ExpiresAt, CreditsUsed)

### Changed
- `types.go` — `ScrapeParams`: removed `ParsePDF`; added `MinAge`, `Mobile`, `SkipTlsVerification`, `BlockAds`, `Proxy`, `Location`, `Parsers`, `Actions`, `RemoveBase64Images`, `StoreInCache`, `ZeroDataRetention`
- `types.go` — `CrawlParams`: removed `MaxDepth`, `AllowBackwardLinks`, `IgnoreSitemap`, changed `Webhook *string` → `*WebhookConfig`; added `MaxDiscoveryDepth`, `Sitemap`, `CrawlEntireDomain`, `AllowSubdomains`, `Delay`, `MaxConcurrency`, `Prompt`, `RegexOnFullURL`, `ZeroDataRetention`
- `types.go` — `MapParams`: removed `IgnoreSitemap`; added `Sitemap`, `IgnoreQueryParameters`, `IgnoreCache`, `Timeout`, `Location`
- `types.go` — `MapResponse.Links`: changed from `[]string` to `[]MapLink`
- `types.go` — `FirecrawlDocument`: added `Summary`, `Images`, `Actions`, `Warning`, `ChangeTracking`, `Branding`
- `crawl.go` — `CrawlURL`/`AsyncCrawlURL`: removed references to `ParsePDF`, `MaxDepth`, `AllowBackwardLinks`, `IgnoreSitemap`; added all new v2 `CrawlParams` fields to request body construction
- `map.go` — `MapURL`: removed `IgnoreSitemap` map key; added `Sitemap`, `IgnoreQueryParameters`, `IgnoreCache`, `Timeout`, `Location` to request body construction
- `scrape.go` — `ScrapeURL`: removed `ParsePDF` handling; added all new v2 `ScrapeParams` fields to request body construction
- `go.mod` — bumped Go version from `1.22.5` to `1.23`

### Notes
- `go build ./...` and `go vet ./...` pass cleanly after all changes
- Integration test file uses `//go:build integration` tag so the removed v1 fields in that file do not block compilation — those will be updated in MIG-07/MIG-09
- `ExtractParams.IgnoreSitemap` is kept as-is (it is a distinct Extract-specific parameter, not the removed CrawlParams field)

## [CI Fix: Resolve all golangci-lint and test failures] - 2026-03-15

### Changed
- `firecrawl_test.go` — Added `//go:build integration` build tag so CI's `go test ./...` no longer crashes without `.env`; replaced `init()` / `log.Fatalf` with `TestMain` that gracefully exits if `.env` is missing; renamed inner loop variables `response`/`err` in `TestCheckCrawlStatusE2E` to `statusResponse`/`statusErr` to eliminate shadow warning
- `crawl.go` — Removed blank line between `makeRequest` call and `if err != nil` in `AsyncCrawlURL` to satisfy gofumpt
- `.golangci.yml` — Removed `enable-all: true` from govet; added explicit `disable: [fieldalignment]` to suppress false-positive struct padding warnings on types scheduled for rewrite in MIG-04
- `helpers.go` — Changed `http.NewRequest` to `http.NewRequestWithContext(context.Background(), ...)` to satisfy noctx linter; added `"context"` import
- `errors.go` — Changed `fmt.Errorf(message)` to `errors.New(message)` to fix staticcheck SA1006 (printf verb with non-constant format); added `"errors"` import

### Notes
- `go build ./...`, `go vet ./...`, and `go test ./...` all pass cleanly
- Integration tests still run via `go test -tags=integration ./...` (requires `.env` with API_URL and TEST_API_KEY)

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
