# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Search endpoint (`POST /v2/search`) with typed `SearchResponse` (IMP-01)
- Batch Scrape endpoints: `BatchScrapeURLs`, `AsyncBatchScrapeURLs`, `CheckBatchScrapeStatus` (IMP-02)
- Extract endpoints: `Extract`, `AsyncExtract`, `CheckExtractStatus` (IMP-03)
- Typed error system: `APIError` struct with 8 sentinel errors (`ErrUnauthorized`, `ErrRateLimited`, `ErrNoAPIKey`, `ErrPaymentRequired`, `ErrNotFound`, `ErrTimeout`, `ErrConflict`, `ErrServerError`) (IMP-04)
- Security hardening: pagination URL validation against API host, UUID job ID validation, HTTPS warning on non-localhost HTTP (IMP-05)
- Unit test foundation with `httptest.NewServer` mock server helpers (IMP-06)
- 160+ unit tests covering all methods, error paths, and security behaviors (IMP-07, IMP-08)
- HTTP client options: `NewFirecrawlAppWithOptions`, `WithTimeout`, `WithTransport`, `WithUserAgent`, `WithMaxIdleConns`, `WithMaxIdleConnsPerHost` (IMP-15)
- `PaginationConfig` support for `CheckCrawlStatus` and `CheckBatchScrapeStatus` (IMP-10)
- `GetCrawlStatusPage` and `GetBatchScrapeStatusPage` public methods for manual pagination (IMP-10)
- `SDKVersion` constant (`"2.0.0"`) and `User-Agent` header on all requests (IMP-15)
- `CONTRIBUTING.md` with development workflow, code style, and endpoint addition guide (IMP-11)
- Integration tests for Search, Batch Scrape, Extract, and PaginationConfig (IMP-09)

### Changed

- **BREAKING:** All public methods now require `context.Context` as first parameter (MIG-05)
- **BREAKING:** `CrawlParams.MaxDepth` renamed to `MaxDiscoveryDepth` (MIG-04)
- **BREAKING:** `CrawlParams.AllowBackwardLinks` renamed to `CrawlEntireDomain` (MIG-04)
- **BREAKING:** `CrawlParams.IgnoreSitemap` replaced by `Sitemap` string enum (`"include"`, `"skip"`, `"only"`) (MIG-04)
- **BREAKING:** `CrawlParams.Webhook` changed from `*string` to `*WebhookConfig` (MIG-04)
- **BREAKING:** `MapResponse.Links` changed from `[]string` to `[]MapLink` (MIG-04)
- **BREAKING:** `ScrapeParams.ParsePDF` removed, replaced by `Parsers []ParserConfig` (MIG-04)
- **BREAKING:** `FirecrawlApp.APIKey` field unexported — use `APIKey()` accessor method (IMP-05)
- **BREAKING:** `Search` method signature changed from `(ctx, query, *any) (any, error)` to `(ctx, query, *SearchParams) (*SearchResponse, error)` (IMP-01)
- All endpoints migrated from `/v1/*` to `/v2/*` (MIG-06 through MIG-09)
- `makeRequest` accepts `[]byte` body instead of `map[string]any`; callers marshal before passing (MIG-06)
- `monitorJobStatus` uses v2 status values: `"scraping"` (poll), `"completed"`, `"failed"` (MIG-08)
- Minimum Go version bumped from 1.22 to 1.23 (MIG-04)
- Split monolithic `firecrawl.go` into 16 modular files (MIG-02)
- `http.DefaultTransport` is cloned instead of referenced directly (IMP-15)

### Fixed

- Retry counter in `monitorJobStatus` was initialized at retry threshold — now starts at 0 so retries actually occur (MIG-01)
- `defer resp.Body.Close()` inside retry loop leaked HTTP connections; intermediate bodies now closed explicitly (MIG-01)
- Request body (`bytes.NewBuffer`) consumed on first attempt, all retries sent empty body; body now recreated per attempt (MIG-01)
- `ScrapeURL` checked response `Success` before checking unmarshal error — order corrected (MIG-01)
- `ScrapeOptions` gate only checked `Formats` field — gate now checks any non-zero field (MIG-01)

### Removed

- Commented-out v0 extractor code (MIG-01)
- Legacy `firecrawl_test.go_V0` test file (MIG-03)
- v1 API paths (`/v1/*`) — all replaced by `/v2/*`

## [2.0.0] — 2026-03-15

### Added

- `context.Context` on all public methods and internal helpers (MIG-05)
- 31+ v2 type definitions: `LocationConfig`, `WebhookConfig`, `ActionConfig`, `ParserConfig`, `MapLink`, `PaginationConfig`, `SearchParams`, `SearchResponse`, `BatchScrapeParams`, `BatchScrapeResponse`, `ExtractParams`, `ExtractResponse`, and more (MIG-04)
- CI/CD pipeline: `Makefile` with 9 targets, `golangci-lint` v2 config, GitHub Actions with lint + test matrix (Go 1.23/1.24/1.25) (MIG-03)
- Modular file structure: 16 Go source files split by concern (MIG-02)
- `.editorconfig` and `dependabot.yml` (MIG-03)

### Changed

- All endpoints migrated to `/v2/*` paths (MIG-06 through MIG-09)
- Request bodies use typed struct marshaling instead of `map[string]any` (MIG-11)
- `monitorJobStatus` updated for v2 status values: `"scraping"`, `"completed"`, `"failed"` (MIG-08)
- Crawl parameters updated: `MaxDepth` → `MaxDiscoveryDepth`, `IgnoreSitemap` → `Sitemap`, `AllowBackwardLinks` → `CrawlEntireDomain` (MIG-07)
- `MapResponse.Links` changed from `[]string` to `[]MapLink` (MIG-09)
- `.env.example` updated to use live API URL (MIG-03)

### Fixed

- Retry counter starting at threshold instead of 0 (MIG-01)
- `defer resp.Body.Close()` connection leak in retry loop (MIG-01)
- Request body reuse across retries sending empty body (MIG-01)
- Error handling order in `ScrapeURL` — unmarshal error checked before `Success` (MIG-01)
- `ScrapeOptions` gate missing nil check on non-Formats fields (MIG-01)

### Removed

- v1 field names: `MaxDepth`, `AllowBackwardLinks`, `IgnoreSitemap` from `CrawlParams` (MIG-07)
- Dead v0 extractor code and legacy test file (MIG-01, MIG-03)

[Unreleased]: https://github.com/firecrawl/firecrawl-go/compare/v2.0.0...HEAD
[2.0.0]: https://github.com/firecrawl/firecrawl-go/releases/tag/v2.0.0
