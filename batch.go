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
// Parameters:
//   - ctx: Context for cancellation and deadlines.
//   - id: The ID of the batch scrape job to check.
//
// Returns:
//   - *BatchScrapeStatusResponse: The current status of the batch scrape job.
//   - error: An error if the status check fails.
func (app *FirecrawlApp) CheckBatchScrapeStatus(ctx context.Context, id string) (*BatchScrapeStatusResponse, error) {
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
