package firecrawl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// crawlRequest is the internal request struct for crawl operations.
// It is unexported — callers use CrawlParams instead.
type crawlRequest struct {
	URL                   string         `json:"url"`
	ScrapeOptions         *ScrapeParams  `json:"scrapeOptions,omitempty"`
	Webhook               *WebhookConfig `json:"webhook,omitempty"`
	Limit                 *int           `json:"limit,omitempty"`
	IncludePaths          []string       `json:"includePaths,omitempty"`
	ExcludePaths          []string       `json:"excludePaths,omitempty"`
	MaxDiscoveryDepth     *int           `json:"maxDiscoveryDepth,omitempty"`
	AllowExternalLinks    *bool          `json:"allowExternalLinks,omitempty"`
	IgnoreQueryParameters *bool          `json:"ignoreQueryParameters,omitempty"`
	Sitemap               *string        `json:"sitemap,omitempty"`
	CrawlEntireDomain     *bool          `json:"crawlEntireDomain,omitempty"`
	AllowSubdomains       *bool          `json:"allowSubdomains,omitempty"`
	Delay                 *float64       `json:"delay,omitempty"`
	MaxConcurrency        *int           `json:"maxConcurrency,omitempty"`
	Prompt                *string        `json:"prompt,omitempty"`
	RegexOnFullURL        *bool          `json:"regexOnFullURL,omitempty"`
	ZeroDataRetention     *bool          `json:"zeroDataRetention,omitempty"`
}

// buildCrawlRequest creates a crawlRequest from URL and CrawlParams.
// Shared by CrawlURL and AsyncCrawlURL to eliminate duplicated body construction.
func buildCrawlRequest(url string, params *CrawlParams) (*crawlRequest, error) {
	req := &crawlRequest{URL: url}
	if params == nil {
		return req, nil
	}

	// Only include ScrapeOptions if at least one field is set.
	scrapeOpts := params.ScrapeOptions
	if scrapeOpts.Formats != nil || scrapeOpts.Headers != nil || scrapeOpts.IncludeTags != nil ||
		scrapeOpts.ExcludeTags != nil || scrapeOpts.OnlyMainContent != nil || scrapeOpts.WaitFor != nil ||
		scrapeOpts.Timeout != nil || scrapeOpts.MaxAge != nil || scrapeOpts.MinAge != nil ||
		scrapeOpts.JsonOptions != nil || scrapeOpts.Mobile != nil || scrapeOpts.SkipTlsVerification != nil ||
		scrapeOpts.BlockAds != nil || scrapeOpts.Proxy != nil || scrapeOpts.Location != nil ||
		scrapeOpts.Parsers != nil || scrapeOpts.Actions != nil || scrapeOpts.RemoveBase64Images != nil ||
		scrapeOpts.StoreInCache != nil || scrapeOpts.ZeroDataRetention != nil {
		req.ScrapeOptions = &scrapeOpts
	}

	req.Webhook = params.Webhook
	req.Limit = params.Limit
	req.IncludePaths = params.IncludePaths
	req.ExcludePaths = params.ExcludePaths
	req.MaxDiscoveryDepth = params.MaxDiscoveryDepth
	req.AllowExternalLinks = params.AllowExternalLinks
	req.IgnoreQueryParameters = params.IgnoreQueryParameters
	req.Sitemap = params.Sitemap
	req.CrawlEntireDomain = params.CrawlEntireDomain
	req.AllowSubdomains = params.AllowSubdomains
	req.Delay = params.Delay
	req.MaxConcurrency = params.MaxConcurrency
	req.Prompt = params.Prompt
	req.RegexOnFullURL = params.RegexOnFullURL
	req.ZeroDataRetention = params.ZeroDataRetention

	return req, nil
}

// CrawlURL starts a crawl job for the specified URL using the Firecrawl API.
//
// Parameters:
//   - ctx: Context for cancellation and deadlines.
//   - url: The URL to crawl.
//   - params: Optional parameters for the crawl request.
//   - idempotencyKey: An optional idempotency key to ensure the request is idempotent (can be nil).
//   - pollInterval: An optional interval (in seconds) at which to poll the job status. Default is 2 seconds.
//
// Returns:
//   - CrawlStatusResponse: The crawl result if the job is completed.
//   - error: An error if the crawl request fails.
func (app *FirecrawlApp) CrawlURL(ctx context.Context, url string, params *CrawlParams, idempotencyKey *string, pollInterval ...int) (*CrawlStatusResponse, error) {
	var key string
	if idempotencyKey != nil {
		key = *idempotencyKey
	}

	headers := app.prepareHeaders(&key)

	req, err := buildCrawlRequest(url, params)
	if err != nil {
		return nil, fmt.Errorf("failed to build crawl request: %w", err)
	}

	crawlBodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal crawl request: %w", err)
	}

	actualPollInterval := 2
	if len(pollInterval) > 0 {
		actualPollInterval = pollInterval[0]
	}

	resp, err := app.makeRequest(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/v2/crawl", app.APIURL),
		crawlBodyBytes,
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

	return app.monitorJobStatus(ctx, crawlResponse.ID, headers, actualPollInterval)
}

// AsyncCrawlURL starts a crawl job for the specified URL using the Firecrawl API.
//
// Parameters:
//   - ctx: Context for cancellation and deadlines.
//   - url: The URL to crawl.
//   - params: Optional parameters for the crawl request.
//   - idempotencyKey: An optional idempotency key to ensure the request is idempotent.
//
// Returns:
//   - *CrawlResponse: The crawl response with id.
//   - error: An error if the crawl request fails.
func (app *FirecrawlApp) AsyncCrawlURL(ctx context.Context, url string, params *CrawlParams, idempotencyKey *string) (*CrawlResponse, error) {
	var key string
	if idempotencyKey != nil {
		key = *idempotencyKey
	}

	headers := app.prepareHeaders(&key)

	req, err := buildCrawlRequest(url, params)
	if err != nil {
		return nil, fmt.Errorf("failed to build crawl request: %w", err)
	}

	crawlBodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal crawl request: %w", err)
	}

	resp, err := app.makeRequest(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/v2/crawl", app.APIURL),
		crawlBodyBytes,
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
//   - ctx: Context for cancellation and deadlines.
//   - ID: The ID of the crawl job to check.
//
// Returns:
//   - *CrawlStatusResponse: The status of the crawl job.
//   - error: An error if the crawl status check request fails.
func (app *FirecrawlApp) CheckCrawlStatus(ctx context.Context, ID string) (*CrawlStatusResponse, error) {
	headers := app.prepareHeaders(nil)
	apiURL := fmt.Sprintf("%s/v2/crawl/%s", app.APIURL, ID)

	resp, err := app.makeRequest(
		ctx,
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
//   - ctx: Context for cancellation and deadlines.
//   - ID: The ID of the crawl job to cancel.
//
// Returns:
//   - string: The status of the crawl job after cancellation.
//   - error: An error if the crawl job cancellation request fails.
func (app *FirecrawlApp) CancelCrawlJob(ctx context.Context, ID string) (string, error) {
	headers := app.prepareHeaders(nil)
	apiURL := fmt.Sprintf("%s/v2/crawl/%s", app.APIURL, ID)
	resp, err := app.makeRequest(
		ctx,
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
