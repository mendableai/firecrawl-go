package firecrawl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// extractRequest is the internal request struct for extract operations.
// It is unexported — callers use ExtractParams instead.
type extractRequest struct {
	URLs              []string       `json:"urls"`
	Prompt            *string        `json:"prompt,omitempty"`
	Schema            map[string]any `json:"schema,omitempty"`
	EnableWebSearch   *bool          `json:"enableWebSearch,omitempty"`
	IgnoreSitemap     *bool          `json:"ignoreSitemap,omitempty"`
	IncludeSubdomains *bool          `json:"includeSubdomains,omitempty"`
	ShowSources       *bool          `json:"showSources,omitempty"`
	IgnoreInvalidURLs *bool          `json:"ignoreInvalidURLs,omitempty"`
	ScrapeOptions     *ScrapeParams  `json:"scrapeOptions,omitempty"`
}

// Extract performs LLM-based structured data extraction and polls until completion.
//
// Parameters:
//   - ctx: Context for cancellation and deadlines.
//   - urls: The list of URLs to extract data from.
//   - params: Optional parameters for the extraction request.
//
// Returns:
//   - *ExtractStatusResponse: The extraction result with structured data.
//   - error: An error if the extraction fails.
func (app *FirecrawlApp) Extract(ctx context.Context, urls []string, params *ExtractParams) (*ExtractStatusResponse, error) {
	response, err := app.AsyncExtract(ctx, urls, params)
	if err != nil {
		return nil, err
	}

	headers := app.prepareHeaders(nil)
	return app.monitorExtractStatus(ctx, response.ID, headers)
}

// AsyncExtract starts an extraction job asynchronously.
//
// Parameters:
//   - ctx: Context for cancellation and deadlines.
//   - urls: The list of URLs to extract data from.
//   - params: Optional parameters for the extraction request.
//
// Returns:
//   - *ExtractResponse: The response with job ID for polling.
//   - error: An error if starting the extraction fails.
func (app *FirecrawlApp) AsyncExtract(ctx context.Context, urls []string, params *ExtractParams) (*ExtractResponse, error) {
	headers := app.prepareHeaders(nil)

	req := extractRequest{URLs: urls}
	if params != nil {
		req.Prompt = params.Prompt
		req.Schema = params.Schema
		req.EnableWebSearch = params.EnableWebSearch
		req.IgnoreSitemap = params.IgnoreSitemap
		req.IncludeSubdomains = params.IncludeSubdomains
		req.ShowSources = params.ShowSources
		req.IgnoreInvalidURLs = params.IgnoreInvalidURLs
		req.ScrapeOptions = params.ScrapeOptions
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal extract request: %w", err)
	}

	resp, err := app.makeRequest(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/v2/extract", app.APIURL),
		body,
		headers,
		"start extract job",
		withRetries(3),
		withBackoff(500),
	)
	if err != nil {
		return nil, err
	}

	var extractResponse ExtractResponse
	if err := json.Unmarshal(resp, &extractResponse); err != nil {
		return nil, fmt.Errorf("failed to parse extract response: %w", err)
	}

	if extractResponse.ID == "" {
		return nil, fmt.Errorf("failed to get extract job ID")
	}

	return &extractResponse, nil
}

// CheckExtractStatus checks the status of an extraction job.
//
// Parameters:
//   - ctx: Context for cancellation and deadlines.
//   - id: The ID of the extraction job to check.
//
// Returns:
//   - *ExtractStatusResponse: The current status of the extraction job.
//   - error: An error if the status check fails.
func (app *FirecrawlApp) CheckExtractStatus(ctx context.Context, id string) (*ExtractStatusResponse, error) {
	if err := validateJobID(id); err != nil {
		return nil, err
	}

	headers := app.prepareHeaders(nil)

	resp, err := app.makeRequest(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/v2/extract/%s", app.APIURL, id),
		nil,
		headers,
		"check extract status",
		withRetries(3),
		withBackoff(500),
	)
	if err != nil {
		return nil, err
	}

	var statusResponse ExtractStatusResponse
	if err := json.Unmarshal(resp, &statusResponse); err != nil {
		return nil, fmt.Errorf("failed to parse extract status response: %w", err)
	}

	return &statusResponse, nil
}

// monitorExtractStatus polls an extraction job until completion.
// Unlike crawl/batch, extract uses "processing" status and has no pagination.
func (app *FirecrawlApp) monitorExtractStatus(ctx context.Context, id string, headers map[string]string) (*ExtractStatusResponse, error) {
	pollInterval := 2

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		resp, err := app.makeRequest(
			ctx,
			http.MethodGet,
			fmt.Sprintf("%s/v2/extract/%s", app.APIURL, id),
			nil,
			headers,
			"check extract status",
			withRetries(3),
			withBackoff(500),
		)
		if err != nil {
			return nil, err
		}

		var statusData ExtractStatusResponse
		if err := json.Unmarshal(resp, &statusData); err != nil {
			return nil, err
		}

		status := statusData.Status
		if status == "" {
			return nil, fmt.Errorf("invalid status in extract response")
		}

		switch status {
		case "completed":
			return &statusData, nil
		case "processing":
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(pollInterval) * time.Second):
			}
		case "failed":
			return nil, fmt.Errorf("extract job failed. Status: %s", status)
		default:
			return nil, fmt.Errorf("unknown extract status: %s", status)
		}
	}
}
