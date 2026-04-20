// Package firecrawl provides a client for interacting with the Firecrawl API.
package firecrawl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"time"
)

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

// FirecrawlDocumentMetadata represents metadata for a Firecrawl document
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

// JsonOptions represents the options for JSON extraction
type JsonOptions struct {
	Schema       map[string]any `json:"schema,omitempty"`
	SystemPrompt *string        `json:"systemPrompt,omitempty"`
	Prompt       *string        `json:"prompt,omitempty"`
}

// FirecrawlDocument represents a document in Firecrawl
type FirecrawlDocument struct {
	Markdown   string                     `json:"markdown,omitempty"`
	HTML       string                     `json:"html,omitempty"`
	RawHTML    string                     `json:"rawHtml,omitempty"`
	Screenshot string                     `json:"screenshot,omitempty"`
	JSON       map[string]any             `json:"json,omitempty"`
	Links      []string                   `json:"links,omitempty"`
	Metadata   *FirecrawlDocumentMetadata `json:"metadata,omitempty"`
}

// ScrapeParams represents the parameters for a scrape request.
type ScrapeParams struct {
	Formats         []string           `json:"formats,omitempty"`
	Headers         *map[string]string `json:"headers,omitempty"`
	IncludeTags     []string           `json:"includeTags,omitempty"`
	ExcludeTags     []string           `json:"excludeTags,omitempty"`
	OnlyMainContent *bool              `json:"onlyMainContent,omitempty"`
	WaitFor         *int               `json:"waitFor,omitempty"`
	ParsePDF        *bool              `json:"parsePDF,omitempty"`
	Timeout         *int               `json:"timeout,omitempty"`
	MaxAge          *int               `json:"maxAge,omitempty"`
	JsonOptions     *JsonOptions       `json:"jsonOptions,omitempty"`
}

// ScrapeResponse represents the response for scraping operations
type ScrapeResponse struct {
	Success bool               `json:"success"`
	Data    *FirecrawlDocument `json:"data,omitempty"`
}

// CrawlParams represents the parameters for a crawl request.
type CrawlParams struct {
	ScrapeOptions         ScrapeParams `json:"scrapeOptions"`
	Webhook               *string      `json:"webhook,omitempty"`
	Limit                 *int         `json:"limit,omitempty"`
	IncludePaths          []string     `json:"includePaths,omitempty"`
	ExcludePaths          []string     `json:"excludePaths,omitempty"`
	MaxDepth              *int         `json:"maxDepth,omitempty"`
	AllowBackwardLinks    *bool        `json:"allowBackwardLinks,omitempty"`
	AllowExternalLinks    *bool        `json:"allowExternalLinks,omitempty"`
	IgnoreSitemap         *bool        `json:"ignoreSitemap,omitempty"`
	IgnoreQueryParameters *bool        `json:"ignoreQueryParameters,omitempty"`
}

// CrawlResponse represents the response for crawling operations
type CrawlResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	URL     string `json:"url,omitempty"`
}

// CrawlStatusResponse (old JobStatusResponse) represents the response for checking crawl job
type CrawlStatusResponse struct {
	Status      string               `json:"status"`
	Total       int                  `json:"total,omitempty"`
	Completed   int                  `json:"completed,omitempty"`
	CreditsUsed int                  `json:"creditsUsed,omitempty"`
	ExpiresAt   string               `json:"expiresAt,omitempty"`
	Next        *string              `json:"next,omitempty"`
	Data        []*FirecrawlDocument `json:"data,omitempty"`
}

// CancelCrawlJobResponse represents the response for canceling a crawl job
type CancelCrawlJobResponse struct {
	Success bool   `json:"success"`
	Status  string `json:"status"`
}

// WebhookSpec represents a webhook specification
type WebhookSpec struct {
	URL      string            `json:"url"`
	Headers  map[string]string `json:"headers,omitempty"`
	Metadata map[string]any    `json:"metadata,omitempty"`
	Events   []string          `json:"events,omitempty"`
}

// FormatSpec represents a format specification
type FormatSpec struct {
	Type     string         `json:"type"`
	FullPage *bool          `json:"fullPage,omitempty"`
	Quality  *int           `json:"quality,omitempty"`
	Viewport *Viewport      `json:"viewport,omitempty"`
	Schema   map[string]any `json:"schema,omitempty"`
	Prompt   *string        `json:"prompt,omitempty"`
	Modes    []string       `json:"modes,omitempty"`
	Tag      *string        `json:"tag,omitempty"`
}

// Viewport represents viewport dimensions
type Viewport struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ParserSpec represents a parser specification
type ParserSpec struct {
	Type     string `json:"type"`
	MaxPages *int   `json:"maxPages,omitempty"`
}

// ActionSpec represents an action specification
type ActionSpec struct {
	Type         string    `json:"type"`
	Milliseconds *int      `json:"milliseconds,omitempty"`
	Selector     *string   `json:"selector,omitempty"`
	FullPage     *bool     `json:"fullPage,omitempty"`
	Quality      *int      `json:"quality,omitempty"`
	Viewport     *Viewport `json:"viewport,omitempty"`
	All          *bool     `json:"all,omitempty"`
	Text         *string   `json:"text,omitempty"`
	Key          *string   `json:"key,omitempty"`
	Direction    *string   `json:"direction,omitempty"`
	Script       *string   `json:"script,omitempty"`
	Format       *string   `json:"format,omitempty"`
	Landscape    *bool     `json:"landscape,omitempty"`
	Scale        *float64  `json:"scale,omitempty"`
}

// LocationSpec represents location settings
type LocationSpec struct {
	Country   *string  `json:"country,omitempty"`
	Languages []string `json:"languages,omitempty"`
}

// BatchScrapeParams represents the parameters for a batch scrape request
type BatchScrapeParams struct {
	URLs                []string          `json:"urls"`
	Webhook             *WebhookSpec      `json:"webhook,omitempty"`
	MaxConcurrency      *int              `json:"maxConcurrency,omitempty"`
	IgnoreInvalidURLs   *bool             `json:"ignoreInvalidURLs,omitempty"`
	Formats             []interface{}     `json:"formats,omitempty"`
	OnlyMainContent     *bool             `json:"onlyMainContent,omitempty"`
	IncludeTags         []string          `json:"includeTags,omitempty"`
	ExcludeTags         []string          `json:"excludeTags,omitempty"`
	MaxAge              *int              `json:"maxAge,omitempty"`
	Headers             map[string]string `json:"headers,omitempty"`
	WaitFor             *int              `json:"waitFor,omitempty"`
	Mobile              *bool             `json:"mobile,omitempty"`
	SkipTlsVerification *bool             `json:"skipTlsVerification,omitempty"`
	Timeout             *int              `json:"timeout,omitempty"`
	Parsers             []interface{}     `json:"parsers,omitempty"`
	Actions             []ActionSpec      `json:"actions,omitempty"`
	Location            *LocationSpec     `json:"location,omitempty"`
	RemoveBase64Images  *bool             `json:"removeBase64Images,omitempty"`
	BlockAds            *bool             `json:"blockAds,omitempty"`
	Proxy               *string           `json:"proxy,omitempty"`
	StoreInCache        *bool             `json:"storeInCache,omitempty"`
	ZeroDataRetention   *bool             `json:"zeroDataRetention,omitempty"`
}

// BatchScrapeResponse represents the response for batch scrape operations
type BatchScrapeResponse struct {
	Success     bool     `json:"success"`
	ID          string   `json:"id"`
	URL         string   `json:"url"`
	InvalidURLs []string `json:"invalidURLs,omitempty"`
}

// MapParams represents the parameters for a map request.
type MapParams struct {
	IncludeSubdomains *bool   `json:"includeSubdomains,omitempty"`
	Search            *string `json:"search,omitempty"`
	IgnoreSitemap     *bool   `json:"ignoreSitemap,omitempty"`
	Limit             *int    `json:"limit,omitempty"`
}

// MapResponse represents the response for mapping operations
type MapResponse struct {
	Success bool     `json:"success"`
	Links   []string `json:"links,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// requestOptions represents options for making requests.
type requestOptions struct {
	retries int
	backoff int
}

// requestOption is a functional option type for requestOptions.
type requestOption func(*requestOptions)

// newRequestOptions creates a new requestOptions instance with the provided options.
//
// Parameters:
//   - opts: Optional request options.
//
// Returns:
//   - *requestOptions: A new instance of requestOptions with the provided options.
func newRequestOptions(opts ...requestOption) *requestOptions {
	options := &requestOptions{retries: 1}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// withRetries sets the number of retries for a request.
//
// Parameters:
//   - retries: The number of retries to be performed.
//
// Returns:
//   - requestOption: A functional option that sets the number of retries for a request.
func withRetries(retries int) requestOption {
	return func(opts *requestOptions) {
		opts.retries = retries
	}
}

// withBackoff sets the backoff interval for a request.
//
// Parameters:
//   - backoff: The backoff interval (in milliseconds) to be used for retries.
//
// Returns:
//   - requestOption: A functional option that sets the backoff interval for a request.
func withBackoff(backoff int) requestOption {
	return func(opts *requestOptions) {
		opts.backoff = backoff
	}
}

// FirecrawlApp represents a client for the Firecrawl API.
type FirecrawlApp struct {
	APIKey  string
	APIURL  string
	Client  *http.Client
	Version string
}

// NewFirecrawlApp creates a new instance of FirecrawlApp with the provided API key and API URL.
// If the API key or API URL is not provided, it attempts to retrieve them from environment variables.
// If the API key is still not found, it returns an error.
//
// Parameters:
//   - apiKey: The API key for authenticating with the Firecrawl API. If empty, it will be retrieved from the FIRECRAWL_API_KEY environment variable.
//   - apiURL: The base URL for the Firecrawl API. If empty, it will be retrieved from the FIRECRAWL_API_URL environment variable, defaulting to "https://api.firecrawl.dev".
//   - timeout: The timeout for the HTTP client. If not provided, it will default to 60 seconds.
//
// Returns:
//   - *FirecrawlApp: A new instance of FirecrawlApp configured with the provided or retrieved API key and API URL.
//   - error: An error if the API key is not provided or retrieved.
func NewFirecrawlApp(apiKey, apiURL string, timeout ...time.Duration) (*FirecrawlApp, error) {
	if apiKey == "" {
		apiKey = os.Getenv("FIRECRAWL_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("no API key provided")
		}
	}

	if apiURL == "" {
		apiURL = os.Getenv("FIRECRAWL_API_URL")
		if apiURL == "" {
			apiURL = "https://api.firecrawl.dev"
		}
	}

	t := 120 * time.Second // default
	if len(timeout) > 0 {
		t = timeout[0]
	}

	client := &http.Client{
		Timeout:   t,
		Transport: http.DefaultTransport,
	}

	return &FirecrawlApp{
		APIKey: apiKey,
		APIURL: apiURL,
		Client: client,
	}, nil
}

// ScrapeURL scrapes the content of the specified URL using the Firecrawl API.
//
// Parameters:
//   - url: The URL to be scraped.
//   - params: Optional parameters for the scrape request, including extractor options for LLM extraction.
//
// Returns:
//   - *FirecrawlDocument or *FirecrawlDocumentV0: The scraped document data depending on the API version.
//   - error: An error if the scrape request fails.
func (app *FirecrawlApp) ScrapeURL(url string, params *ScrapeParams) (*FirecrawlDocument, error) {
	headers := app.prepareHeaders(nil)
	scrapeBody := map[string]any{"url": url}

	// if params != nil {
	// 	if extractorOptions, ok := params["extractorOptions"].(ExtractorOptions); ok {
	// 		if schema, ok := extractorOptions.ExtractionSchema.(interface{ schema() any }); ok {
	// 			extractorOptions.ExtractionSchema = schema.schema()
	// 		}
	// 		if extractorOptions.Mode == "" {
	// 			extractorOptions.Mode = "llm-extraction"
	// 		}
	// 		scrapeBody["extractorOptions"] = extractorOptions
	// 	}

	// 	for key, value := range params {
	// 		if key != "extractorOptions" {
	// 			scrapeBody[key] = value
	// 		}
	// 	}
	// }

	if params != nil {
		if params.Formats != nil {
			scrapeBody["formats"] = params.Formats
		}
		if params.Headers != nil {
			scrapeBody["headers"] = params.Headers
		}
		if params.IncludeTags != nil {
			scrapeBody["includeTags"] = params.IncludeTags
		}
		if params.ExcludeTags != nil {
			scrapeBody["excludeTags"] = params.ExcludeTags
		}
		if params.OnlyMainContent != nil {
			scrapeBody["onlyMainContent"] = params.OnlyMainContent
		}
		if params.WaitFor != nil {
			scrapeBody["waitFor"] = params.WaitFor
		}
		if params.ParsePDF != nil {
			scrapeBody["parsePDF"] = params.ParsePDF
		}
		if params.Timeout != nil {
			scrapeBody["timeout"] = params.Timeout
		}
		if params.MaxAge != nil {
			scrapeBody["maxAge"] = params.MaxAge
		}
		if params.JsonOptions != nil {
			scrapeBody["jsonOptions"] = params.JsonOptions
		}
	}

	resp, err := app.makeRequest(
		http.MethodPost,
		fmt.Sprintf("%s/v1/scrape", app.APIURL),
		scrapeBody,
		headers,
		"scrape URL",
	)
	if err != nil {
		return nil, err
	}

	var scrapeResponse ScrapeResponse
	err = json.Unmarshal(resp, &scrapeResponse)

	if scrapeResponse.Success {
		return scrapeResponse.Data, nil
	}

	if err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("failed to scrape URL")
}

// CrawlURL starts a crawl job for the specified URL using the Firecrawl API.
//
// Parameters:
//   - url: The URL to crawl.
//   - params: Optional parameters for the crawl request.
//   - idempotencyKey: An optional idempotency key to ensure the request is idempotent (can be nil).
//   - pollInterval: An optional interval (in seconds) at which to poll the job status. Default is 2 seconds.
//
// Returns:
//   - CrawlStatusResponse: The crawl result if the job is completed.
//   - error: An error if the crawl request fails.
func (app *FirecrawlApp) CrawlURL(url string, params *CrawlParams, idempotencyKey *string, pollInterval ...int) (*CrawlStatusResponse, error) {
	var key string
	if idempotencyKey != nil {
		key = *idempotencyKey
	}

	headers := app.prepareHeaders(&key)
	crawlBody := map[string]any{"url": url}

	if params != nil {
		if params.ScrapeOptions.Formats != nil {
			crawlBody["scrapeOptions"] = params.ScrapeOptions
		}
		if params.Webhook != nil {
			crawlBody["webhook"] = params.Webhook
		}
		if params.Limit != nil {
			crawlBody["limit"] = params.Limit
		}
		if params.IncludePaths != nil {
			crawlBody["includePaths"] = params.IncludePaths
		}
		if params.ExcludePaths != nil {
			crawlBody["excludePaths"] = params.ExcludePaths
		}
		if params.MaxDepth != nil {
			crawlBody["maxDepth"] = params.MaxDepth
		}
		if params.AllowBackwardLinks != nil {
			crawlBody["allowBackwardLinks"] = params.AllowBackwardLinks
		}
		if params.AllowExternalLinks != nil {
			crawlBody["allowExternalLinks"] = params.AllowExternalLinks
		}
		if params.IgnoreSitemap != nil {
			crawlBody["ignoreSitemap"] = params.IgnoreSitemap
		}
		if params.IgnoreQueryParameters != nil {
			crawlBody["ignoreQueryParameters"] = params.IgnoreQueryParameters
		}
	}

	actualPollInterval := 2
	if len(pollInterval) > 0 {
		actualPollInterval = pollInterval[0]
	}

	resp, err := app.makeRequest(
		http.MethodPost,
		fmt.Sprintf("%s/v1/crawl", app.APIURL),
		crawlBody,
		headers,
		"start crawl job",
		withRetries(3),
		withBackoff(500),
	)
	if err != nil {
		return nil, err
	}

	var crawlResponse CrawlResponse
	err = json.Unmarshal(resp, &crawlResponse)
	if err != nil {
		return nil, err
	}

	return app.monitorJobStatus(crawlResponse.ID, headers, actualPollInterval)
}

// CrawlURL starts a crawl job for the specified URL using the Firecrawl API.
//
// Parameters:
//   - url: The URL to crawl.
//   - params: Optional parameters for the crawl request.
//   - idempotencyKey: An optional idempotency key to ensure the request is idempotent.
//
// Returns:
//   - *CrawlResponse: The crawl response with id.
//   - error: An error if the crawl request fails.
func (app *FirecrawlApp) AsyncCrawlURL(url string, params *CrawlParams, idempotencyKey *string) (*CrawlResponse, error) {
	var key string
	if idempotencyKey != nil {
		key = *idempotencyKey
	}

	headers := app.prepareHeaders(&key)
	crawlBody := map[string]any{"url": url}

	if params != nil {
		if params.ScrapeOptions.Formats != nil {
			crawlBody["scrapeOptions"] = params.ScrapeOptions
		}
		if params.Webhook != nil {
			crawlBody["webhook"] = params.Webhook
		}
		if params.Limit != nil {
			crawlBody["limit"] = params.Limit
		}
		if params.IncludePaths != nil {
			crawlBody["includePaths"] = params.IncludePaths
		}
		if params.ExcludePaths != nil {
			crawlBody["excludePaths"] = params.ExcludePaths
		}
		if params.MaxDepth != nil {
			crawlBody["maxDepth"] = params.MaxDepth
		}
		if params.AllowBackwardLinks != nil {
			crawlBody["allowBackwardLinks"] = params.AllowBackwardLinks
		}
		if params.AllowExternalLinks != nil {
			crawlBody["allowExternalLinks"] = params.AllowExternalLinks
		}
		if params.IgnoreSitemap != nil {
			crawlBody["ignoreSitemap"] = params.IgnoreSitemap
		}
		if params.IgnoreQueryParameters != nil {
			crawlBody["ignoreQueryParameters"] = params.IgnoreQueryParameters
		}
	}

	resp, err := app.makeRequest(
		http.MethodPost,
		fmt.Sprintf("%s/v1/crawl", app.APIURL),
		crawlBody,
		headers,
		"start crawl job",
		withRetries(3),
		withBackoff(500),
	)

	if err != nil {
		return nil, err
	}

	var crawlResponse CrawlResponse
	err = json.Unmarshal(resp, &crawlResponse)
	if err != nil {
		return nil, err
	}

	if crawlResponse.ID == "" {
		return nil, fmt.Errorf("failed to get job ID")
	}

	return &crawlResponse, nil
}

// CheckCrawlStatus checks the status of a crawl job using the Firecrawl API.
//
// Parameters:
//   - ID: The ID of the crawl job to check.
//
// Returns:
//   - *CrawlStatusResponse: The status of the crawl job.
//   - error: An error if the crawl status check request fails.
func (app *FirecrawlApp) CheckCrawlStatus(ID string) (*CrawlStatusResponse, error) {
	headers := app.prepareHeaders(nil)
	apiURL := fmt.Sprintf("%s/v1/crawl/%s", app.APIURL, ID)

	resp, err := app.makeRequest(
		http.MethodGet,
		apiURL,
		nil,
		headers,
		"check crawl status",
		withRetries(3),
		withBackoff(500),
	)
	if err != nil {
		return nil, err
	}

	var jobStatusResponse CrawlStatusResponse
	err = json.Unmarshal(resp, &jobStatusResponse)
	if err != nil {
		return nil, err
	}

	return &jobStatusResponse, nil
}

// CancelCrawlJob cancels a crawl job using the Firecrawl API.
//
// Parameters:
//   - ID: The ID of the crawl job to cancel.
//
// Returns:
//   - string: The status of the crawl job after cancellation.
//   - error: An error if the crawl job cancellation request fails.
func (app *FirecrawlApp) CancelCrawlJob(ID string) (string, error) {
	headers := app.prepareHeaders(nil)
	apiURL := fmt.Sprintf("%s/v1/crawl/%s", app.APIURL, ID)
	resp, err := app.makeRequest(
		http.MethodDelete,
		apiURL,
		nil,
		headers,
		"cancel crawl job",
	)
	if err != nil {
		return "", err
	}

	var cancelCrawlJobResponse CancelCrawlJobResponse
	err = json.Unmarshal(resp, &cancelCrawlJobResponse)
	if err != nil {
		return "", err
	}

	return cancelCrawlJobResponse.Status, nil
}

// MapURL initiates a mapping operation for a URL using the Firecrawl API.
//
// Parameters:
//   - url: The URL to map.
//   - params: Optional parameters for the mapping request.
//
// Returns:
//   - *MapResponse: The response from the mapping operation.
//   - error: An error if the mapping request fails.
func (app *FirecrawlApp) MapURL(url string, params *MapParams) (*MapResponse, error) {
	headers := app.prepareHeaders(nil)
	jsonData := map[string]any{"url": url}

	if params != nil {
		if params.IncludeSubdomains != nil {
			jsonData["includeSubdomains"] = params.IncludeSubdomains
		}
		if params.Search != nil {
			jsonData["search"] = params.Search
		}
		if params.IgnoreSitemap != nil {
			jsonData["ignoreSitemap"] = params.IgnoreSitemap
		}
		if params.Limit != nil {
			jsonData["limit"] = params.Limit
		}
	}

	resp, err := app.makeRequest(
		http.MethodPost,
		fmt.Sprintf("%s/v1/map", app.APIURL),
		jsonData,
		headers,
		"map",
	)
	if err != nil {
		return nil, err
	}

	var mapResponse MapResponse
	err = json.Unmarshal(resp, &mapResponse)
	if err != nil {
		return nil, err
	}

	if mapResponse.Success {
		return &mapResponse, nil
	} else {
		return nil, fmt.Errorf("map operation failed: %s", mapResponse.Error)
	}
}

// BatchScrape starts a batch scrape job for the specified URLs using the Firecrawl API.
//
// Parameters:
//   - params: Parameters for the batch scrape request, including URLs and optional configuration.
//
// Returns:
//   - *BatchScrapeResponse: The batch scrape response with job ID and URL.
//   - error: An error if the batch scrape request fails.
func (app *FirecrawlApp) BatchScrape(params *BatchScrapeParams) (*BatchScrapeResponse, error) {
	if params == nil || len(params.URLs) == 0 {
		return nil, fmt.Errorf("urls are required")
	}

	headers := app.prepareHeaders(nil)
	batchBody := map[string]any{
		"urls": params.URLs,
	}

	if params.Webhook != nil {
		batchBody["webhook"] = params.Webhook
	}
	if params.MaxConcurrency != nil {
		batchBody["maxConcurrency"] = params.MaxConcurrency
	}
	if params.IgnoreInvalidURLs != nil {
		batchBody["ignoreInvalidURLs"] = params.IgnoreInvalidURLs
	}
	if params.Formats != nil {
		batchBody["formats"] = params.Formats
	}
	if params.OnlyMainContent != nil {
		batchBody["onlyMainContent"] = params.OnlyMainContent
	}
	if params.IncludeTags != nil {
		batchBody["includeTags"] = params.IncludeTags
	}
	if params.ExcludeTags != nil {
		batchBody["excludeTags"] = params.ExcludeTags
	}
	if params.MaxAge != nil {
		batchBody["maxAge"] = params.MaxAge
	}
	if params.Headers != nil {
		batchBody["headers"] = params.Headers
	}
	if params.WaitFor != nil {
		batchBody["waitFor"] = params.WaitFor
	}
	if params.Mobile != nil {
		batchBody["mobile"] = params.Mobile
	}
	if params.SkipTlsVerification != nil {
		batchBody["skipTlsVerification"] = params.SkipTlsVerification
	}
	if params.Timeout != nil {
		batchBody["timeout"] = params.Timeout
	}
	if params.Parsers != nil {
		batchBody["parsers"] = params.Parsers
	}
	if params.Actions != nil {
		batchBody["actions"] = params.Actions
	}
	if params.Location != nil {
		batchBody["location"] = params.Location
	}
	if params.RemoveBase64Images != nil {
		batchBody["removeBase64Images"] = params.RemoveBase64Images
	}
	if params.BlockAds != nil {
		batchBody["blockAds"] = params.BlockAds
	}
	if params.Proxy != nil {
		batchBody["proxy"] = params.Proxy
	}
	if params.StoreInCache != nil {
		batchBody["storeInCache"] = params.StoreInCache
	}
	if params.ZeroDataRetention != nil {
		batchBody["zeroDataRetention"] = params.ZeroDataRetention
	}

	resp, err := app.makeRequest(
		http.MethodPost,
		fmt.Sprintf("%s/v2/batch/scrape", app.APIURL),
		batchBody,
		headers,
		"start batch scrape",
		withRetries(3),
		withBackoff(500),
	)
	if err != nil {
		return nil, err
	}

	var batchResponse BatchScrapeResponse
	err = json.Unmarshal(resp, &batchResponse)
	if err != nil {
		return nil, err
	}

	if !batchResponse.Success {
		return nil, fmt.Errorf("failed to start batch scrape")
	}

	return &batchResponse, nil
}

// SearchURL searches for a URL using the Firecrawl API.
//
// Parameters:
//   - url: The URL to search for.
//   - params: Optional parameters for the search request.
//   - error: An error if the search request fails.
//
// Search is not implemented in API version 1.0.0.
func (app *FirecrawlApp) Search(query string, params *any) (any, error) {
	return nil, fmt.Errorf("Search is not implemented in API version 1.0.0")
}

// prepareHeaders prepares the headers for an HTTP request.
//
// Parameters:
//   - idempotencyKey: A string representing the idempotency key to be included in the headers.
//     If the idempotency key is an empty string, it will not be included in the headers.
//
// Returns:
//   - map[string]string: A map containing the headers for the HTTP request.
func (app *FirecrawlApp) prepareHeaders(idempotencyKey *string) map[string]string {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", app.APIKey),
	}
	if idempotencyKey != nil {
		headers["x-idempotency-key"] = *idempotencyKey
	}
	return headers
}

// makeRequest makes a request to the specified URL with the provided method, data, headers, and options.
//
// Parameters:
//   - method: The HTTP method to use for the request (e.g., "GET", "POST", "DELETE").
//   - url: The URL to send the request to.
//   - data: The data to be sent in the request body.
//   - headers: The headers to be included in the request.
//   - action: A string describing the action being performed.
//   - opts: Optional request options.
//
// Returns:
//   - []byte: The response body from the request.
//   - error: An error if the request fails.
func (app *FirecrawlApp) makeRequest(method, url string, data map[string]any, headers map[string]string, action string, opts ...requestOption) ([]byte, error) {
	var body []byte
	var err error
	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	var resp *http.Response
	options := newRequestOptions(opts...)
	for i := 0; i < options.retries; i++ {
		resp, err = app.Client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 502 && resp.StatusCode != 503 {
			break
		}

		time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Duration(options.backoff) * time.Millisecond)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	statusCode := resp.StatusCode
	if statusCode != 200 {
		return nil, app.handleError(statusCode, respBody, action)
	}

	return respBody, nil
}

// monitorJobStatus monitors the status of a crawl job using the Firecrawl API.
//
// Parameters:
//   - ID: The ID of the crawl job to monitor.
//   - headers: The headers to be included in the request.
//   - pollInterval: The interval (in seconds) at which to poll the job status.
//
// Returns:
//   - *CrawlStatusResponse: The crawl result if the job is completed.
//   - error: An error if the crawl status check request fails.
func (app *FirecrawlApp) monitorJobStatus(ID string, headers map[string]string, pollInterval int) (*CrawlStatusResponse, error) {
	attempts := 3

	for {
		resp, err := app.makeRequest(
			http.MethodGet,
			fmt.Sprintf("%s/v1/crawl/%s", app.APIURL, ID),
			nil,
			headers,
			"check crawl status",
			withRetries(3),
			withBackoff(500),
		)
		if err != nil {
			return nil, err
		}

		var statusData CrawlStatusResponse
		err = json.Unmarshal(resp, &statusData)
		if err != nil {
			return nil, err
		}

		status := statusData.Status
		if status == "" {
			return nil, fmt.Errorf("invalid status in response")
		}
		if status == "completed" {
			if statusData.Data != nil {
				allData := statusData.Data
				for statusData.Next != nil {
					resp, err := app.makeRequest(
						http.MethodGet,
						*statusData.Next,
						nil,
						headers,
						"fetch next page of crawl status",
						withRetries(3),
						withBackoff(500),
					)
					if err != nil {
						return nil, err
					}

					err = json.Unmarshal(resp, &statusData)
					if err != nil {
						return nil, err
					}

					if statusData.Data != nil {
						allData = append(allData, statusData.Data...)
					}
				}
				statusData.Data = allData
				return &statusData, nil
			} else {
				attempts++
				if attempts > 3 {
					return nil, fmt.Errorf("crawl job completed but no data was returned")
				}
			}
		} else if status == "active" || status == "paused" || status == "pending" || status == "queued" || status == "waiting" || status == "scraping" {
			pollInterval = max(pollInterval, 2)
			time.Sleep(time.Duration(pollInterval) * time.Second)
		} else {
			return nil, fmt.Errorf("crawl job failed or was stopped. Status: %s", status)
		}
	}
}

// handleError handles errors returned by the Firecrawl API.
//
// Parameters:
//   - resp: The HTTP response object.
//   - body: The response body from the HTTP response.
//   - action: A string describing the action being performed.
//
// Returns:
//   - error: An error describing the failure reason.
func (app *FirecrawlApp) handleError(statusCode int, body []byte, action string) error {
	var errorData map[string]any
	err := json.Unmarshal(body, &errorData)
	if err != nil {
		return fmt.Errorf("failed to parse error response: %v", err)
	}

	errorMessage, _ := errorData["error"].(string)
	if errorMessage == "" {
		errorMessage = "No additional error details provided."
	}

	var message string
	switch statusCode {
	case 401:
		message = fmt.Sprintf("Unauthorized: Failed to %s. %s", action, errorMessage)
	case 402:
		message = fmt.Sprintf("Payment Required: Failed to %s. %s", action, errorMessage)
	case 403:
		message = fmt.Sprintf("Forbidden: Failed to %s. %s", action, errorMessage)
	case 408:
		message = fmt.Sprintf("Request Timeout: Failed to %s as the request timed out. %s", action, errorMessage)
	case 409:
		message = fmt.Sprintf("Conflict: Failed to %s due to a conflict. %s", action, errorMessage)
	case 429:
		message = fmt.Sprintf("Too Many Requests: Failed to %s. %s", action, errorMessage)
	case 500:
		message = fmt.Sprintf("Internal Server Error: Failed to %s. %s", action, errorMessage)
	default:
		message = fmt.Sprintf("Unexpected error during %s: Status code %d. %s", action, statusCode, errorMessage)
	}

	return fmt.Errorf(message)
}
