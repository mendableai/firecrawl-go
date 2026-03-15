package firecrawl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// batchScrapeRequest is the internal request struct for batch scrape operations.
// It is unexported — callers use BatchScrapeParams instead.
type batchScrapeRequest struct {
	URLs              []string       `json:"urls"`
	ScrapeOptions     *ScrapeParams  `json:"scrapeOptions,omitempty"`
	MaxConcurrency    *int           `json:"maxConcurrency,omitempty"`
	IgnoreInvalidURLs *bool          `json:"ignoreInvalidURLs,omitempty"`
	Webhook           *WebhookConfig `json:"webhook,omitempty"`
}

// AsyncBatchScrapeURLs starts a batch scrape job asynchronously.
//
// Parameters:
//   - ctx: Context for cancellation and deadlines.
//   - urls: The list of URLs to scrape.
//   - params: Optional parameters for the batch scrape request.
//   - idempotencyKey: An optional idempotency key (can be nil).
//
// Returns:
//   - *BatchScrapeResponse: The response with job ID for polling.
//   - error: An error if starting the batch scrape fails.
func (app *FirecrawlApp) AsyncBatchScrapeURLs(ctx context.Context, urls []string, params *BatchScrapeParams, idempotencyKey *string) (*BatchScrapeResponse, error) {
	headers := app.prepareHeaders(idempotencyKey)

	req := batchScrapeRequest{URLs: urls}
	if params != nil {
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
		req.MaxConcurrency = params.MaxConcurrency
		req.IgnoreInvalidURLs = params.IgnoreInvalidURLs
		req.Webhook = params.Webhook
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch scrape request: %w", err)
	}

	resp, err := app.makeRequest(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/v2/batch/scrape", app.APIURL),
		body,
		headers,
		"start batch scrape job",
		withRetries(3),
		withBackoff(500),
	)
	if err != nil {
		return nil, err
	}

	var batchResponse BatchScrapeResponse
	if err := json.Unmarshal(resp, &batchResponse); err != nil {
		return nil, fmt.Errorf("failed to parse batch scrape response: %w", err)
	}

	if batchResponse.ID == "" {
		return nil, fmt.Errorf("failed to get batch scrape job ID")
	}

	return &batchResponse, nil
}

// BatchScrapeURLs starts a batch scrape job and polls until completion.
//
// Parameters:
//   - ctx: Context for cancellation and deadlines.
//   - urls: The list of URLs to scrape.
//   - params: Optional parameters for the batch scrape request.
//   - idempotencyKey: An optional idempotency key (can be nil).
//   - pollInterval: An optional interval (in seconds) at which to poll. Default is 2 seconds.
//
// Returns:
//   - *BatchScrapeStatusResponse: The batch scrape result with all scraped documents.
//   - error: An error if the batch scrape fails.
func (app *FirecrawlApp) BatchScrapeURLs(ctx context.Context, urls []string, params *BatchScrapeParams, idempotencyKey *string, pollInterval ...int) (*BatchScrapeStatusResponse, error) {
	response, err := app.AsyncBatchScrapeURLs(ctx, urls, params, idempotencyKey)
	if err != nil {
		return nil, err
	}

	actualPollInterval := 2
	if len(pollInterval) > 0 {
		actualPollInterval = pollInterval[0]
	}

	headers := app.prepareHeaders(nil)
	return app.monitorBatchScrapeStatus(ctx, response.ID, headers, actualPollInterval)
}

// CheckBatchScrapeStatus checks the status of a batch scrape job.
//
// When a PaginationConfig is provided with AutoPaginate enabled, it automatically
// follows Next URLs to collect all results, respecting MaxPages, MaxResults, and
// MaxWaitTime limits. Without PaginationConfig, only the first page is returned.
//
// Parameters:
//   - ctx: Context for cancellation and deadlines.
//   - id: The ID of the batch scrape job to check.
//   - pagination: An optional PaginationConfig to control auto-pagination behavior.
//
// Returns:
//   - *BatchScrapeStatusResponse: The current status of the batch scrape job (possibly spanning multiple pages).
//   - error: An error if the status check fails.
func (app *FirecrawlApp) CheckBatchScrapeStatus(ctx context.Context, id string, pagination ...*PaginationConfig) (*BatchScrapeStatusResponse, error) {
	if err := validateJobID(id); err != nil {
		return nil, err
	}

	headers := app.prepareHeaders(nil)

	resp, err := app.makeRequest(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/v2/batch/scrape/%s", app.APIURL, id),
		nil,
		headers,
		"check batch scrape status",
		withRetries(3),
		withBackoff(500),
	)
	if err != nil {
		return nil, err
	}

	var statusResponse BatchScrapeStatusResponse
	if err := json.Unmarshal(resp, &statusResponse); err != nil {
		return nil, fmt.Errorf("failed to parse batch scrape status response: %w", err)
	}

	// Without PaginationConfig or AutoPaginate disabled, return the single page.
	if len(pagination) == 0 || pagination[0] == nil || pagination[0].AutoPaginate == nil || !*pagination[0].AutoPaginate {
		return &statusResponse, nil
	}

	return app.autoPaginateBatchScrapeStatus(ctx, &statusResponse, headers, pagination[0])
}

// autoPaginateBatchScrapeStatus follows Next URLs collecting all data, respecting
// MaxPages, MaxResults, and MaxWaitTime limits from the provided PaginationConfig.
func (app *FirecrawlApp) autoPaginateBatchScrapeStatus(ctx context.Context, initial *BatchScrapeStatusResponse, headers map[string]string, cfg *PaginationConfig) (*BatchScrapeStatusResponse, error) {
	allData := initial.Data
	current := initial
	pagesCollected := 1
	startTime := time.Now()

	maxPages := 0
	if cfg.MaxPages != nil {
		maxPages = *cfg.MaxPages
	}
	maxResults := 0
	if cfg.MaxResults != nil {
		maxResults = *cfg.MaxResults
	}
	maxWaitSeconds := 0
	if cfg.MaxWaitTime != nil {
		maxWaitSeconds = *cfg.MaxWaitTime
	}

	for current.Next != nil {
		// Check page limit.
		if maxPages > 0 && pagesCollected >= maxPages {
			break
		}
		// Check result limit.
		if maxResults > 0 && len(allData) >= maxResults {
			allData = allData[:maxResults]
			break
		}
		// Check time limit.
		if maxWaitSeconds > 0 && int(time.Since(startTime).Seconds()) >= maxWaitSeconds {
			break
		}

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if err := validatePaginationURL(app.APIURL, *current.Next); err != nil {
			return nil, fmt.Errorf("unsafe pagination URL: %w", err)
		}

		resp, err := app.makeRequest(
			ctx,
			http.MethodGet,
			*current.Next,
			nil,
			headers,
			"fetch next page of batch scrape status",
			withRetries(3),
			withBackoff(500),
		)
		if err != nil {
			return nil, err
		}

		var pageData BatchScrapeStatusResponse
		if err := json.Unmarshal(resp, &pageData); err != nil {
			return nil, fmt.Errorf("failed to parse batch scrape status page: %w", err)
		}

		if pageData.Data != nil {
			allData = append(allData, pageData.Data...)
		}
		current = &pageData
		pagesCollected++
	}

	current.Data = allData
	return current, nil
}

// GetBatchScrapeStatusPage fetches a specific page of batch scrape status results by URL.
// Use this for manual pagination — pass the Next URL from a previous BatchScrapeStatusResponse.
//
// Parameters:
//   - ctx: Context for cancellation and deadlines.
//   - nextURL: The full URL of the next results page (from BatchScrapeStatusResponse.Next).
//
// Returns:
//   - *BatchScrapeStatusResponse: The results for this page.
//   - error: An error if the request fails or the URL is not trusted.
func (app *FirecrawlApp) GetBatchScrapeStatusPage(ctx context.Context, nextURL string) (*BatchScrapeStatusResponse, error) {
	if err := validatePaginationURL(app.APIURL, nextURL); err != nil {
		return nil, fmt.Errorf("unsafe pagination URL: %w", err)
	}

	headers := app.prepareHeaders(nil)

	resp, err := app.makeRequest(
		ctx,
		http.MethodGet,
		nextURL,
		nil,
		headers,
		"fetch batch scrape status page",
		withRetries(3),
		withBackoff(500),
	)
	if err != nil {
		return nil, err
	}

	var statusResponse BatchScrapeStatusResponse
	if err := json.Unmarshal(resp, &statusResponse); err != nil {
		return nil, fmt.Errorf("failed to parse batch scrape status page: %w", err)
	}

	return &statusResponse, nil
}

// monitorBatchScrapeStatus polls a batch scrape job until completion.
// Mirrors monitorJobStatus from helpers.go but returns BatchScrapeStatusResponse.
func (app *FirecrawlApp) monitorBatchScrapeStatus(ctx context.Context, id string, headers map[string]string, pollInterval int) (*BatchScrapeStatusResponse, error) {
	attempts := 0

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		resp, err := app.makeRequest(
			ctx,
			http.MethodGet,
			fmt.Sprintf("%s/v2/batch/scrape/%s", app.APIURL, id),
			nil,
			headers,
			"check batch scrape status",
			withRetries(3),
			withBackoff(500),
		)
		if err != nil {
			return nil, err
		}

		var statusData BatchScrapeStatusResponse
		if err := json.Unmarshal(resp, &statusData); err != nil {
			return nil, err
		}

		status := statusData.Status
		if status == "" {
			return nil, fmt.Errorf("invalid status in batch scrape response")
		}

		switch status {
		case "completed":
			if statusData.Data != nil {
				allData := statusData.Data
				for statusData.Next != nil {
					if ctx.Err() != nil {
						return nil, ctx.Err()
					}

					if err := validatePaginationURL(app.APIURL, *statusData.Next); err != nil {
						return nil, fmt.Errorf("unsafe pagination URL: %w", err)
					}

					resp, err := app.makeRequest(
						ctx,
						http.MethodGet,
						*statusData.Next,
						nil,
						headers,
						"fetch next page of batch scrape status",
						withRetries(3),
						withBackoff(500),
					)
					if err != nil {
						return nil, err
					}

					if err := json.Unmarshal(resp, &statusData); err != nil {
						return nil, err
					}

					if statusData.Data != nil {
						allData = append(allData, statusData.Data...)
					}
				}
				statusData.Data = allData
				return &statusData, nil
			}
			attempts++
			if attempts > 3 {
				return nil, fmt.Errorf("batch scrape job completed but no data was returned")
			}
		case "scraping":
			interval := max(pollInterval, 2)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(interval) * time.Second):
			}
		case "failed":
			return nil, fmt.Errorf("batch scrape job failed. Status: %s", status)
		default:
			return nil, fmt.Errorf("unknown batch scrape status: %s", status)
		}
	}
}
