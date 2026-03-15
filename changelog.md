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
