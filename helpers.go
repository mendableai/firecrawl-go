package firecrawl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

// makeRequest makes a request to the specified URL with the provided method, body, headers, and options.
//
// Parameters:
//   - ctx: Context for cancellation and deadlines.
//   - method: The HTTP method to use for the request (e.g., "GET", "POST", "DELETE").
//   - url: The URL to send the request to.
//   - body: The pre-marshaled JSON body to send in the request. Pass nil for requests with no body.
//   - headers: The headers to be included in the request.
//   - action: A string describing the action being performed.
//   - opts: Optional request options.
//
// Returns:
//   - []byte: The response body from the request.
//   - error: An error if the request fails.
func (app *FirecrawlApp) makeRequest(ctx context.Context, method, url string, body []byte, headers map[string]string, action string, opts ...requestOption) ([]byte, error) {
	var resp *http.Response
	var err error
	options := newRequestOptions(opts...)
	for i := 0; i < options.retries; i++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		var req *http.Request
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
		if err != nil {
			return nil, err
		}

		for key, value := range headers {
			req.Header.Set(key, value)
		}

		resp, err = app.Client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != 502 {
			break
		}

		// Close body before retry — do NOT defer in loop
		_ = resp.Body.Close()
		time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Duration(options.backoff) * time.Millisecond)
	}
	defer func() { _ = resp.Body.Close() }()

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
//   - ctx: Context for cancellation and deadlines.
//   - ID: The ID of the crawl job to monitor.
//   - headers: The headers to be included in the request.
//   - pollInterval: The interval (in seconds) at which to poll the job status.
//
// Returns:
//   - *CrawlStatusResponse: The crawl result if the job is completed.
//   - error: An error if the crawl status check request fails.
func (app *FirecrawlApp) monitorJobStatus(ctx context.Context, ID string, headers map[string]string, pollInterval int) (*CrawlStatusResponse, error) {
	attempts := 0

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		resp, err := app.makeRequest(
			ctx,
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
		switch status {
		case "completed":
			if statusData.Data != nil {
				allData := statusData.Data
				for statusData.Next != nil {
					if ctx.Err() != nil {
						return nil, ctx.Err()
					}

					resp, err := app.makeRequest(
						ctx,
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
			}
			attempts++
			if attempts > 3 {
				return nil, fmt.Errorf("crawl job completed but no data was returned")
			}
		case "active", "paused", "pending", "queued", "waiting", "scraping":
			pollInterval = max(pollInterval, 2)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(pollInterval) * time.Second):
			}
		default:
			return nil, fmt.Errorf("crawl job failed or was stopped. Status: %s", status)
		}
	}
}
