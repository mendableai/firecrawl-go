package firecrawl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

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
	crawlBody := map[string]any{"url": url}

	if params != nil {
		scrapeOpts := params.ScrapeOptions
		if scrapeOpts.Formats != nil || scrapeOpts.Headers != nil || scrapeOpts.IncludeTags != nil ||
			scrapeOpts.ExcludeTags != nil || scrapeOpts.OnlyMainContent != nil || scrapeOpts.WaitFor != nil ||
			scrapeOpts.Timeout != nil || scrapeOpts.MaxAge != nil || scrapeOpts.MinAge != nil ||
			scrapeOpts.JsonOptions != nil || scrapeOpts.Mobile != nil || scrapeOpts.SkipTlsVerification != nil ||
			scrapeOpts.BlockAds != nil || scrapeOpts.Proxy != nil || scrapeOpts.Location != nil ||
			scrapeOpts.Parsers != nil || scrapeOpts.Actions != nil || scrapeOpts.RemoveBase64Images != nil ||
			scrapeOpts.StoreInCache != nil || scrapeOpts.ZeroDataRetention != nil {
			crawlBody["scrapeOptions"] = scrapeOpts
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
		if params.MaxDiscoveryDepth != nil {
			crawlBody["maxDiscoveryDepth"] = params.MaxDiscoveryDepth
		}
		if params.CrawlEntireDomain != nil {
			crawlBody["crawlEntireDomain"] = params.CrawlEntireDomain
		}
		if params.AllowExternalLinks != nil {
			crawlBody["allowExternalLinks"] = params.AllowExternalLinks
		}
		if params.Sitemap != nil {
			crawlBody["sitemap"] = params.Sitemap
		}
		if params.IgnoreQueryParameters != nil {
			crawlBody["ignoreQueryParameters"] = params.IgnoreQueryParameters
		}
		if params.AllowSubdomains != nil {
			crawlBody["allowSubdomains"] = params.AllowSubdomains
		}
		if params.Delay != nil {
			crawlBody["delay"] = params.Delay
		}
		if params.MaxConcurrency != nil {
			crawlBody["maxConcurrency"] = params.MaxConcurrency
		}
		if params.Prompt != nil {
			crawlBody["prompt"] = params.Prompt
		}
		if params.RegexOnFullURL != nil {
			crawlBody["regexOnFullURL"] = params.RegexOnFullURL
		}
		if params.ZeroDataRetention != nil {
			crawlBody["zeroDataRetention"] = params.ZeroDataRetention
		}
	}

	actualPollInterval := 2
	if len(pollInterval) > 0 {
		actualPollInterval = pollInterval[0]
	}

	resp, err := app.makeRequest(
		ctx,
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
	crawlBody := map[string]any{"url": url}

	if params != nil {
		scrapeOpts := params.ScrapeOptions
		if scrapeOpts.Formats != nil || scrapeOpts.Headers != nil || scrapeOpts.IncludeTags != nil ||
			scrapeOpts.ExcludeTags != nil || scrapeOpts.OnlyMainContent != nil || scrapeOpts.WaitFor != nil ||
			scrapeOpts.Timeout != nil || scrapeOpts.MaxAge != nil || scrapeOpts.MinAge != nil ||
			scrapeOpts.JsonOptions != nil || scrapeOpts.Mobile != nil || scrapeOpts.SkipTlsVerification != nil ||
			scrapeOpts.BlockAds != nil || scrapeOpts.Proxy != nil || scrapeOpts.Location != nil ||
			scrapeOpts.Parsers != nil || scrapeOpts.Actions != nil || scrapeOpts.RemoveBase64Images != nil ||
			scrapeOpts.StoreInCache != nil || scrapeOpts.ZeroDataRetention != nil {
			crawlBody["scrapeOptions"] = scrapeOpts
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
		if params.MaxDiscoveryDepth != nil {
			crawlBody["maxDiscoveryDepth"] = params.MaxDiscoveryDepth
		}
		if params.CrawlEntireDomain != nil {
			crawlBody["crawlEntireDomain"] = params.CrawlEntireDomain
		}
		if params.AllowExternalLinks != nil {
			crawlBody["allowExternalLinks"] = params.AllowExternalLinks
		}
		if params.Sitemap != nil {
			crawlBody["sitemap"] = params.Sitemap
		}
		if params.IgnoreQueryParameters != nil {
			crawlBody["ignoreQueryParameters"] = params.IgnoreQueryParameters
		}
		if params.AllowSubdomains != nil {
			crawlBody["allowSubdomains"] = params.AllowSubdomains
		}
		if params.Delay != nil {
			crawlBody["delay"] = params.Delay
		}
		if params.MaxConcurrency != nil {
			crawlBody["maxConcurrency"] = params.MaxConcurrency
		}
		if params.Prompt != nil {
			crawlBody["prompt"] = params.Prompt
		}
		if params.RegexOnFullURL != nil {
			crawlBody["regexOnFullURL"] = params.RegexOnFullURL
		}
		if params.ZeroDataRetention != nil {
			crawlBody["zeroDataRetention"] = params.ZeroDataRetention
		}
	}

	resp, err := app.makeRequest(
		ctx,
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
//   - ctx: Context for cancellation and deadlines.
//   - ID: The ID of the crawl job to check.
//
// Returns:
//   - *CrawlStatusResponse: The status of the crawl job.
//   - error: An error if the crawl status check request fails.
func (app *FirecrawlApp) CheckCrawlStatus(ctx context.Context, ID string) (*CrawlStatusResponse, error) {
	headers := app.prepareHeaders(nil)
	apiURL := fmt.Sprintf("%s/v1/crawl/%s", app.APIURL, ID)

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
	apiURL := fmt.Sprintf("%s/v1/crawl/%s", app.APIURL, ID)
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
