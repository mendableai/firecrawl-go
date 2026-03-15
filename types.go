package firecrawl

import (
	"encoding/json"
	"fmt"
)

// StringOrStringSlice is a type that can unmarshal either a JSON string or a JSON array of strings.
type StringOrStringSlice []string

func (s *StringOrStringSlice) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*s = []string{single}
		return nil
	}

	var list []string
	if err := json.Unmarshal(data, &list); err == nil {
		*s = list
		return nil
	}

	return fmt.Errorf("field is neither a string nor a list of strings")
}

// FirecrawlDocumentMetadata represents metadata for a Firecrawl document.
type FirecrawlDocumentMetadata struct {
	Title             *string              `json:"title,omitempty"`
	Description       *StringOrStringSlice `json:"description,omitempty"`
	Language          *StringOrStringSlice `json:"language,omitempty"`
	Keywords          *StringOrStringSlice `json:"keywords,omitempty"`
	Robots            *StringOrStringSlice `json:"robots,omitempty"`
	OGTitle           *StringOrStringSlice `json:"ogTitle,omitempty"`
	OGDescription     *StringOrStringSlice `json:"ogDescription,omitempty"`
	OGURL             *StringOrStringSlice `json:"ogUrl,omitempty"`
	OGImage           *StringOrStringSlice `json:"ogImage,omitempty"`
	OGAudio           *StringOrStringSlice `json:"ogAudio,omitempty"`
	OGDeterminer      *StringOrStringSlice `json:"ogDeterminer,omitempty"`
	OGLocale          *StringOrStringSlice `json:"ogLocale,omitempty"`
	OGLocaleAlternate []*string            `json:"ogLocaleAlternate,omitempty"`
	OGSiteName        *StringOrStringSlice `json:"ogSiteName,omitempty"`
	OGVideo           *StringOrStringSlice `json:"ogVideo,omitempty"`
	DCTermsCreated    *StringOrStringSlice `json:"dctermsCreated,omitempty"`
	DCDateCreated     *StringOrStringSlice `json:"dcDateCreated,omitempty"`
	DCDate            *StringOrStringSlice `json:"dcDate,omitempty"`
	DCTermsType       *StringOrStringSlice `json:"dctermsType,omitempty"`
	DCType            *StringOrStringSlice `json:"dcType,omitempty"`
	DCTermsAudience   *StringOrStringSlice `json:"dctermsAudience,omitempty"`
	DCTermsSubject    *StringOrStringSlice `json:"dctermsSubject,omitempty"`
	DCSubject         *StringOrStringSlice `json:"dcSubject,omitempty"`
	DCDescription     *StringOrStringSlice `json:"dcDescription,omitempty"`
	DCTermsKeywords   *StringOrStringSlice `json:"dctermsKeywords,omitempty"`
	ModifiedTime      *StringOrStringSlice `json:"modifiedTime,omitempty"`
	PublishedTime     *StringOrStringSlice `json:"publishedTime,omitempty"`
	ArticleTag        *StringOrStringSlice `json:"articleTag,omitempty"`
	ArticleSection    *StringOrStringSlice `json:"articleSection,omitempty"`
	URL               *string              `json:"url,omitempty"`
	ScrapeID          *string              `json:"scrapeId,omitempty"`
	SourceURL         *string              `json:"sourceURL,omitempty"`
	StatusCode        *int                 `json:"statusCode,omitempty"`
	Error             *string              `json:"error,omitempty"`
}

// JsonOptions represents the options for JSON extraction.
type JsonOptions struct {
	// Schema is an optional JSON schema for structured data extraction.
	Schema map[string]any `json:"schema,omitempty"`
	// SystemPrompt is an optional system-level prompt for the LLM.
	SystemPrompt *string `json:"systemPrompt,omitempty"`
	// Prompt is an optional user-level prompt for the LLM.
	Prompt *string `json:"prompt,omitempty"`
}

// LocationConfig represents geolocation settings for requests.
type LocationConfig struct {
	// Country is the ISO 3166-1 alpha-2 country code (e.g., "US", "GB").
	Country string `json:"country,omitempty"`
	// Languages is the list of BCP-47 language codes to prefer (e.g., ["en", "en-US"]).
	Languages []string `json:"languages,omitempty"`
}

// ParserConfig represents parser configuration for document parsing.
// It replaces the v1 ParsePDF field. Use Type "pdf" to parse PDF documents.
type ParserConfig struct {
	// Type is the parser type (e.g., "pdf").
	Type string `json:"type"`
	// Mode is the optional parsing mode (e.g., "auto", "ocr").
	Mode *string `json:"mode,omitempty"`
	// MaxPages is the optional maximum number of pages to parse.
	MaxPages *int `json:"maxPages,omitempty"`
}

// ActionConfig represents a browser action to execute during scraping.
// The Type field is a discriminator: "wait", "click", "write", "press",
// "scroll", "screenshot", "scrape", "executeJavascript", "pdf".
// Type-specific fields are optional and only apply to relevant action types.
type ActionConfig struct {
	// Type is the action discriminator (required).
	Type string `json:"type"`
	// Milliseconds is the duration for "wait" actions.
	Milliseconds *int `json:"milliseconds,omitempty"`
	// Selector is the CSS selector for "click" and "write" actions.
	Selector *string `json:"selector,omitempty"`
	// Text is the text to write for "write" actions.
	Text *string `json:"text,omitempty"`
	// Key is the key to press for "press" actions (e.g., "Enter").
	Key *string `json:"key,omitempty"`
	// Direction is the scroll direction for "scroll" actions ("up" or "down").
	Direction *string `json:"direction,omitempty"`
	// Amount is the scroll amount in pixels for "scroll" actions.
	Amount *int `json:"amount,omitempty"`
	// Script is the JavaScript source code for "executeJavascript" actions.
	Script *string `json:"script,omitempty"`
	// FullPage captures the full page for "screenshot" actions.
	FullPage *bool `json:"fullPage,omitempty"`
}

// WebhookConfig represents webhook configuration for async operations.
type WebhookConfig struct {
	// URL is the webhook endpoint URL (required).
	URL string `json:"url"`
	// Headers are optional custom HTTP headers to send with webhook requests.
	Headers map[string]string `json:"headers,omitempty"`
	// Metadata is optional arbitrary metadata to include in the webhook payload.
	Metadata map[string]any `json:"metadata,omitempty"`
	// Events is the list of event types to subscribe to (e.g., "completed", "page", "failed").
	Events []string `json:"events,omitempty"`
}

// ActionsResult contains the results of browser actions executed during scraping.
type ActionsResult struct {
	// Screenshots contains base64-encoded screenshots from "screenshot" actions.
	Screenshots []string `json:"screenshots,omitempty"`
	// Scrapes contains scraped documents from "scrape" actions.
	Scrapes []FirecrawlDocument `json:"scrapes,omitempty"`
	// JavascriptReturns contains return values from "executeJavascript" actions.
	JavascriptReturns []any `json:"javascriptReturns,omitempty"`
	// PDFs contains base64-encoded PDF data from "pdf" actions.
	PDFs []string `json:"pdfs,omitempty"`
}

// ChangeTrackingResult contains change tracking information between consecutive scrapes.
type ChangeTrackingResult struct {
	// PreviousScrapeAt is the RFC3339 timestamp of the previous scrape used for comparison.
	PreviousScrapeAt *string `json:"previousScrapeAt,omitempty"`
	// ChangeStatus indicates whether the page changed ("changed", "unchanged", "new").
	ChangeStatus *string `json:"changeStatus,omitempty"`
	// Visibility indicates the page visibility status.
	Visibility *string `json:"visibility,omitempty"`
	// Diff is the text diff between the current and previous scrape.
	Diff *string `json:"diff,omitempty"`
	// JSON contains the structured diff data.
	JSON map[string]any `json:"json,omitempty"`
}

// BrandingResult contains extracted branding information from a page.
type BrandingResult struct {
	// ColorScheme is the detected color scheme ("light" or "dark").
	ColorScheme *string `json:"colorScheme,omitempty"`
	// Logo is the URL of the detected logo image.
	Logo *string `json:"logo,omitempty"`
	// Colors contains extracted color values keyed by role (e.g., "primary", "background").
	Colors map[string]any `json:"colors,omitempty"`
	// Fonts contains extracted font information keyed by role (e.g., "heading", "body").
	Fonts map[string]any `json:"fonts,omitempty"`
}

// FirecrawlDocument represents a scraped document returned by the Firecrawl API.
type FirecrawlDocument struct {
	// Markdown is the page content rendered as Markdown.
	Markdown string `json:"markdown,omitempty"`
	// HTML is the page content as cleaned HTML.
	HTML string `json:"html,omitempty"`
	// RawHTML is the raw, unprocessed HTML of the page.
	RawHTML string `json:"rawHtml,omitempty"`
	// Screenshot is the base64-encoded screenshot of the page.
	Screenshot string `json:"screenshot,omitempty"`
	// JSON contains structured data extracted according to JsonOptions.
	JSON map[string]any `json:"json,omitempty"`
	// Links is a list of URLs found on the page.
	Links []string `json:"links,omitempty"`
	// Metadata contains page metadata (title, OG tags, HTTP status, etc.).
	Metadata *FirecrawlDocumentMetadata `json:"metadata,omitempty"`
	// Summary is a generated summary of the page content.
	Summary *string `json:"summary,omitempty"`
	// Images is a list of image URLs found on the page.
	Images []string `json:"images,omitempty"`
	// Actions contains the results of browser actions executed during scraping.
	Actions *ActionsResult `json:"actions,omitempty"`
	// Warning is a non-fatal warning message from the scrape operation.
	Warning *string `json:"warning,omitempty"`
	// ChangeTracking contains change tracking information if the "changeTracking" format was requested.
	ChangeTracking *ChangeTrackingResult `json:"changeTracking,omitempty"`
	// Branding contains extracted branding information if the "branding" format was requested.
	Branding *BrandingResult `json:"branding,omitempty"`
}

// ScrapeParams represents the parameters for a scrape request.
type ScrapeParams struct {
	// Formats specifies which output formats to return (e.g., "markdown", "html", "rawHtml",
	// "screenshot", "json", "links", "summary", "images", "changeTracking", "branding").
	Formats []string `json:"formats,omitempty"`
	// Headers are custom HTTP headers to send with the request.
	Headers *map[string]string `json:"headers,omitempty"`
	// IncludeTags limits HTML parsing to only these CSS selectors.
	IncludeTags []string `json:"includeTags,omitempty"`
	// ExcludeTags removes these CSS selectors from the parsed output.
	ExcludeTags []string `json:"excludeTags,omitempty"`
	// OnlyMainContent strips navigation, footers, and sidebars when true.
	OnlyMainContent *bool `json:"onlyMainContent,omitempty"`
	// WaitFor is the number of milliseconds to wait after page load before scraping.
	WaitFor *int `json:"waitFor,omitempty"`
	// Timeout is the maximum time in milliseconds to wait for the page to load.
	Timeout *int `json:"timeout,omitempty"`
	// MaxAge is the maximum age in milliseconds of a cached result to accept.
	MaxAge *int `json:"maxAge,omitempty"`
	// MinAge is the minimum age in milliseconds of a cached result to accept.
	MinAge *int `json:"minAge,omitempty"`
	// JsonOptions configures LLM-based JSON extraction.
	JsonOptions *JsonOptions `json:"jsonOptions,omitempty"`
	// Mobile emulates a mobile browser when true.
	Mobile *bool `json:"mobile,omitempty"`
	// SkipTlsVerification skips TLS certificate verification when true.
	SkipTlsVerification *bool `json:"skipTlsVerification,omitempty"`
	// BlockAds blocks ads and tracking scripts when true.
	BlockAds *bool `json:"blockAds,omitempty"`
	// Proxy selects the proxy tier to use ("basic", "enhanced", or "auto").
	Proxy *string `json:"proxy,omitempty"`
	// Location configures the geolocation for the request.
	Location *LocationConfig `json:"location,omitempty"`
	// Parsers configures document parsers. Use Type "pdf" to replace the v1 ParsePDF flag.
	Parsers []ParserConfig `json:"parsers,omitempty"`
	// Actions is a list of browser actions to execute before scraping.
	Actions []ActionConfig `json:"actions,omitempty"`
	// RemoveBase64Images strips base64-encoded inline images from the output when true.
	RemoveBase64Images *bool `json:"removeBase64Images,omitempty"`
	// StoreInCache stores the scrape result in the Firecrawl cache when true.
	StoreInCache *bool `json:"storeInCache,omitempty"`
	// ZeroDataRetention prevents Firecrawl from retaining scraped data when true.
	ZeroDataRetention *bool `json:"zeroDataRetention,omitempty"`
	// ParsePDF is removed in v2 — use Parsers: []ParserConfig{{Type: "pdf"}} instead.
	// Deprecated: removed in v2.
}

// ScrapeResponse represents the response for a scrape operation.
type ScrapeResponse struct {
	Success bool               `json:"success"`
	Data    *FirecrawlDocument `json:"data,omitempty"`
}

// CrawlParams represents the parameters for a crawl request.
type CrawlParams struct {
	// ScrapeOptions configures how each page is scraped during the crawl.
	ScrapeOptions ScrapeParams `json:"scrapeOptions,omitempty"`
	// Webhook configures the webhook endpoint to receive crawl events.
	Webhook *WebhookConfig `json:"webhook,omitempty"`
	// Limit is the maximum number of pages to crawl (default 10000).
	Limit *int `json:"limit,omitempty"`
	// IncludePaths restricts crawling to URLs matching these path patterns.
	IncludePaths []string `json:"includePaths,omitempty"`
	// ExcludePaths skips URLs matching these path patterns.
	ExcludePaths []string `json:"excludePaths,omitempty"`
	// AllowExternalLinks allows following links to external domains when true.
	AllowExternalLinks *bool `json:"allowExternalLinks,omitempty"`
	// IgnoreQueryParameters ignores URL query parameters when deduplicating pages.
	IgnoreQueryParameters *bool `json:"ignoreQueryParameters,omitempty"`
	// MaxDiscoveryDepth is the maximum link depth from the seed URL to follow. Replaces v1 MaxDepth.
	MaxDiscoveryDepth *int `json:"maxDiscoveryDepth,omitempty"`
	// Sitemap controls sitemap behavior: "skip", "include", or "only". Replaces v1 IgnoreSitemap.
	Sitemap *string `json:"sitemap,omitempty"`
	// CrawlEntireDomain follows links back to previously visited pages when true. Replaces v1 AllowBackwardLinks.
	CrawlEntireDomain *bool `json:"crawlEntireDomain,omitempty"`
	// AllowSubdomains allows crawling subdomains of the seed URL when true.
	AllowSubdomains *bool `json:"allowSubdomains,omitempty"`
	// Delay is the number of seconds to wait between page scrapes.
	Delay *float64 `json:"delay,omitempty"`
	// MaxConcurrency is the maximum number of pages to scrape concurrently.
	MaxConcurrency *int `json:"maxConcurrency,omitempty"`
	// Prompt is a natural language description of which pages to crawl.
	Prompt *string `json:"prompt,omitempty"`
	// RegexOnFullURL applies include/exclude path patterns to the full URL when true.
	RegexOnFullURL *bool `json:"regexOnFullURL,omitempty"`
	// ZeroDataRetention prevents Firecrawl from retaining crawled data when true.
	ZeroDataRetention *bool `json:"zeroDataRetention,omitempty"`
}

// CrawlResponse represents the initial response when starting a crawl job.
type CrawlResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	URL     string `json:"url,omitempty"`
}

// CrawlStatusResponse represents the status of an in-progress or completed crawl job.
// v2 status values are: "scraping", "completed", "failed".
type CrawlStatusResponse struct {
	// Status is the current crawl status ("scraping", "completed", "failed").
	Status string `json:"status"`
	// Total is the total number of pages discovered.
	Total int `json:"total,omitempty"`
	// Completed is the number of pages scraped so far.
	Completed int `json:"completed,omitempty"`
	// CreditsUsed is the number of API credits consumed by the crawl.
	CreditsUsed int `json:"creditsUsed,omitempty"`
	// ExpiresAt is the RFC3339 timestamp when the crawl result expires.
	ExpiresAt string `json:"expiresAt,omitempty"`
	// Next is the URL of the next results page for paginated crawl status responses.
	Next *string `json:"next,omitempty"`
	// Data contains the scraped documents for the current results page.
	Data []*FirecrawlDocument `json:"data,omitempty"`
}

// CancelCrawlJobResponse represents the response for canceling a crawl job.
type CancelCrawlJobResponse struct {
	Success bool   `json:"success"`
	Status  string `json:"status"`
}

// MapLink represents a link object in the v2 Map response.
// v2 returns rich link objects instead of plain strings.
type MapLink struct {
	// URL is the absolute URL of the discovered link.
	URL string `json:"url"`
	// Title is the optional page title of the linked page.
	Title *string `json:"title,omitempty"`
	// Description is the optional meta description of the linked page.
	Description *string `json:"description,omitempty"`
}

// MapParams represents the parameters for a map request.
type MapParams struct {
	// IncludeSubdomains includes links to subdomains of the target URL when true.
	IncludeSubdomains *bool `json:"includeSubdomains,omitempty"`
	// Search filters the map results to URLs containing this search term.
	Search *string `json:"search,omitempty"`
	// Limit is the maximum number of links to return (default 5000, max 100000).
	Limit *int `json:"limit,omitempty"`
	// Sitemap controls sitemap behavior: "skip", "include", or "only". Replaces v1 IgnoreSitemap.
	Sitemap *string `json:"sitemap,omitempty"`
	// IgnoreQueryParameters ignores URL query parameters when deduplicating links.
	IgnoreQueryParameters *bool `json:"ignoreQueryParameters,omitempty"`
	// IgnoreCache bypasses the Firecrawl cache and re-fetches the sitemap/pages.
	IgnoreCache *bool `json:"ignoreCache,omitempty"`
	// Timeout is the maximum time in milliseconds for the map operation.
	Timeout *int `json:"timeout,omitempty"`
	// Location configures the geolocation for the map request.
	Location *LocationConfig `json:"location,omitempty"`
}

// MapResponse represents the response for a map operation.
type MapResponse struct {
	Success bool      `json:"success"`
	Links   []MapLink `json:"links,omitempty"`
	Error   string    `json:"error,omitempty"`
}

// PaginationConfig controls pagination behavior for status-checking methods.
type PaginationConfig struct {
	// AutoPaginate automatically follows "next" URLs and aggregates results when true.
	AutoPaginate *bool `json:"autoPaginate,omitempty"`
	// MaxPages is the maximum number of result pages to fetch during auto-pagination.
	MaxPages *int `json:"maxPages,omitempty"`
	// MaxResults is the maximum total number of results to collect during auto-pagination.
	MaxResults *int `json:"maxResults,omitempty"`
	// MaxWaitTime is the maximum number of seconds to spend polling before giving up.
	MaxWaitTime *int `json:"maxWaitTime,omitempty"`
}

// SearchParams represents the parameters for a search request.
type SearchParams struct {
	// Limit is the maximum number of results to return.
	Limit *int `json:"limit,omitempty"`
	// Sources specifies which result types to include ("web", "images", "news").
	Sources []string `json:"sources,omitempty"`
	// Categories restricts results to specific content categories (e.g., "github", "research", "pdf").
	Categories []string `json:"categories,omitempty"`
	// TBS is the time-based search filter (e.g., "qdr:d" for past day, "qdr:w" for past week).
	TBS *string `json:"tbs,omitempty"`
	// Location is the geographic location to use for localized search results.
	Location *string `json:"location,omitempty"`
	// Country is the ISO 3166-1 alpha-2 country code for the search context.
	Country *string `json:"country,omitempty"`
	// Timeout is the maximum time in milliseconds for the search operation.
	Timeout *int `json:"timeout,omitempty"`
	// IgnoreInvalidURLs skips invalid URLs in results rather than failing.
	IgnoreInvalidURLs *bool `json:"ignoreInvalidURLs,omitempty"`
	// ScrapeOptions configures how result pages are scraped when content is requested.
	ScrapeOptions *ScrapeParams `json:"scrapeOptions,omitempty"`
}

// SearchWebResult represents a single web search result.
type SearchWebResult struct {
	// Title is the page title of the result.
	Title string `json:"title"`
	// Description is the snippet or meta description of the result.
	Description string `json:"description"`
	// URL is the URL of the result.
	URL string `json:"url"`
	// Markdown is the scraped Markdown content (present when scrapeOptions includes "markdown").
	Markdown *string `json:"markdown,omitempty"`
	// HTML is the scraped HTML content (present when scrapeOptions includes "html").
	HTML *string `json:"html,omitempty"`
	// Metadata contains page metadata for the result.
	Metadata *FirecrawlDocumentMetadata `json:"metadata,omitempty"`
}

// SearchImageResult represents a single image search result.
type SearchImageResult struct {
	// Title is the title or alt text of the image.
	Title string `json:"title"`
	// ImageURL is the direct URL of the image.
	ImageURL string `json:"imageUrl"`
	// ImageWidth is the width of the image in pixels.
	ImageWidth int `json:"imageWidth"`
	// ImageHeight is the height of the image in pixels.
	ImageHeight int `json:"imageHeight"`
	// URL is the URL of the page containing the image.
	URL string `json:"url"`
	// Position is the 1-based rank of this result in the search response.
	Position int `json:"position"`
}

// SearchNewsResult represents a single news search result.
type SearchNewsResult struct {
	// Title is the headline of the news article.
	Title string `json:"title"`
	// Snippet is a short excerpt from the news article.
	Snippet string `json:"snippet"`
	// URL is the URL of the news article.
	URL string `json:"url"`
	// Date is the publication date of the news article.
	Date string `json:"date"`
	// ImageURL is the optional URL of the article's featured image.
	ImageURL *string `json:"imageUrl,omitempty"`
	// Position is the 1-based rank of this result in the search response.
	Position int `json:"position"`
}

// SearchData contains categorized search results.
type SearchData struct {
	// Web contains web search results.
	Web []SearchWebResult `json:"web,omitempty"`
	// Images contains image search results.
	Images []SearchImageResult `json:"images,omitempty"`
	// News contains news search results.
	News []SearchNewsResult `json:"news,omitempty"`
}

// SearchResponse represents the response for a search operation.
type SearchResponse struct {
	// Success indicates whether the search request succeeded.
	Success bool `json:"success"`
	// Data contains the categorized search results.
	Data SearchData `json:"data"`
	// Warning is a non-fatal warning message from the search operation.
	Warning *string `json:"warning,omitempty"`
	// ID is the unique identifier for this search request.
	ID string `json:"id,omitempty"`
	// CreditsUsed is the number of API credits consumed by this search.
	CreditsUsed int `json:"creditsUsed,omitempty"`
}

// BatchScrapeParams represents the parameters for a batch scrape request.
type BatchScrapeParams struct {
	// ScrapeOptions configures how each URL is scraped.
	ScrapeOptions ScrapeParams `json:"scrapeOptions,omitempty"`
	// MaxConcurrency is the maximum number of URLs to scrape concurrently.
	MaxConcurrency *int `json:"maxConcurrency,omitempty"`
	// IgnoreInvalidURLs skips invalid URLs rather than failing the entire batch.
	IgnoreInvalidURLs *bool `json:"ignoreInvalidURLs,omitempty"`
	// Webhook configures the webhook endpoint to receive batch scrape events.
	Webhook *WebhookConfig `json:"webhook,omitempty"`
}

// BatchScrapeResponse represents the initial response when starting a batch scrape job.
type BatchScrapeResponse struct {
	// Success indicates whether the batch scrape job was started successfully.
	Success bool `json:"success"`
	// ID is the job identifier for polling status.
	ID string `json:"id,omitempty"`
	// URL is the polling URL for checking job status.
	URL string `json:"url,omitempty"`
	// InvalidURLs lists any URLs that were rejected before the job started.
	InvalidURLs []string `json:"invalidURLs,omitempty"`
}

// BatchScrapeStatusResponse represents the status of an in-progress or completed batch scrape job.
type BatchScrapeStatusResponse struct {
	// Status is the current job status ("scraping", "completed", "failed").
	Status string `json:"status"`
	// Total is the total number of URLs in the batch.
	Total int `json:"total,omitempty"`
	// Completed is the number of URLs scraped so far.
	Completed int `json:"completed,omitempty"`
	// CreditsUsed is the number of API credits consumed by the batch.
	CreditsUsed int `json:"creditsUsed,omitempty"`
	// ExpiresAt is the RFC3339 timestamp when the batch result expires.
	ExpiresAt string `json:"expiresAt,omitempty"`
	// Next is the URL of the next results page for paginated status responses.
	Next *string `json:"next,omitempty"`
	// Data contains the scraped documents for the current results page.
	Data []*FirecrawlDocument `json:"data,omitempty"`
}

// ExtractParams represents the parameters for an extract request.
// Extract performs LLM-based structured data extraction from one or more URLs.
type ExtractParams struct {
	// Prompt is a natural language description of the data to extract.
	Prompt *string `json:"prompt,omitempty"`
	// Schema is a JSON Schema definition for the structured output.
	Schema map[string]any `json:"schema,omitempty"`
	// EnableWebSearch augments extraction with web search when true.
	EnableWebSearch *bool `json:"enableWebSearch,omitempty"`
	// IgnoreSitemap skips sitemap discovery and only processes the provided URLs.
	IgnoreSitemap *bool `json:"ignoreSitemap,omitempty"`
	// IncludeSubdomains includes subdomains of the provided URLs in extraction.
	IncludeSubdomains *bool `json:"includeSubdomains,omitempty"`
	// ShowSources includes source attribution in the extraction result.
	ShowSources *bool `json:"showSources,omitempty"`
	// IgnoreInvalidURLs skips invalid URLs rather than failing the extraction.
	IgnoreInvalidURLs *bool `json:"ignoreInvalidURLs,omitempty"`
	// ScrapeOptions configures how pages are scraped before extraction.
	ScrapeOptions *ScrapeParams `json:"scrapeOptions,omitempty"`
}

// ExtractResponse represents the initial response when starting an extract job.
type ExtractResponse struct {
	// Success indicates whether the extract job was started successfully.
	Success bool `json:"success"`
	// ID is the job identifier for polling status.
	ID string `json:"id,omitempty"`
	// InvalidURLs lists any URLs that were rejected before the job started.
	InvalidURLs []string `json:"invalidURLs,omitempty"`
}

// ExtractStatusResponse represents the status of an in-progress or completed extract job.
type ExtractStatusResponse struct {
	// Success indicates whether the extraction succeeded.
	Success bool `json:"success"`
	// Status is the current job status ("processing", "completed", "failed").
	Status string `json:"status"`
	// Data contains the extracted structured data upon completion.
	Data map[string]any `json:"data,omitempty"`
	// ExpiresAt is the RFC3339 timestamp when the extract result expires.
	ExpiresAt string `json:"expiresAt,omitempty"`
	// CreditsUsed is the number of API credits consumed by the extraction.
	CreditsUsed int `json:"creditsUsed,omitempty"`
}
