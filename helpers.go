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

	var resp *http.Response
	options := newRequestOptions(opts...)
	for i := 0; i < options.retries; i++ {
		var req *http.Request
		req, err = http.NewRequestWithContext(context.Background(), method, url, bytes.NewBuffer(body))
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
		resp.Body.Close()
		time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Duration(options.backoff) * time.Millisecond)
	}
	defer resp.Body.Close() // Defer close of the final response only

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
	attempts := 0

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
