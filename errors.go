package firecrawl

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Sentinel errors for programmatic error handling via errors.Is().
var (
	// ErrNoAPIKey is returned when no API key is provided to the constructor.
	ErrNoAPIKey = errors.New("no API key provided")

	// ErrUnauthorized is returned for HTTP 401 responses.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrPaymentRequired is returned for HTTP 402 responses.
	ErrPaymentRequired = errors.New("payment required")

	// ErrNotFound is returned for HTTP 404 responses.
	ErrNotFound = errors.New("not found")

	// ErrTimeout is returned for HTTP 408 responses.
	ErrTimeout = errors.New("request timeout")

	// ErrConflict is returned for HTTP 409 responses.
	ErrConflict = errors.New("conflict")

	// ErrRateLimited is returned for HTTP 429 responses.
	ErrRateLimited = errors.New("rate limit exceeded")

	// ErrServerError is returned for HTTP 500 responses.
	ErrServerError = errors.New("internal server error")
)

// APIError represents a structured error from the Firecrawl API.
// It wraps a sentinel error based on the HTTP status code, enabling
// programmatic error handling via errors.Is() and errors.As().
//
// Example usage:
//
//	_, err := app.ScrapeURL(ctx, url, nil)
//	if errors.Is(err, firecrawl.ErrRateLimited) {
//	    // back off and retry
//	}
//
//	var apiErr *firecrawl.APIError
//	if errors.As(err, &apiErr) {
//	    log.Printf("API error %d during %s: %s", apiErr.StatusCode, apiErr.Action, apiErr.Message)
//	}
type APIError struct {
	// StatusCode is the HTTP status code from the API response.
	StatusCode int
	// Message is the error message from the API response body.
	Message string
	// Action is the SDK operation that triggered the error (e.g., "scrape URL", "start crawl job").
	Action string
}

// Error returns a human-readable error string.
func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d during %s: %s", e.StatusCode, e.Action, e.Message)
}

// Unwrap returns the sentinel error corresponding to the HTTP status code.
// This enables errors.Is(err, firecrawl.ErrRateLimited) and similar checks.
func (e *APIError) Unwrap() error {
	switch e.StatusCode {
	case 401:
		return ErrUnauthorized
	case 402:
		return ErrPaymentRequired
	case 404:
		return ErrNotFound
	case 408:
		return ErrTimeout
	case 409:
		return ErrConflict
	case 429:
		return ErrRateLimited
	case 500:
		return ErrServerError
	default:
		return nil
	}
}

// handleError constructs an *APIError from an HTTP status code and response body.
//
// Parameters:
//   - statusCode: The HTTP status code from the response.
//   - body: The raw response body bytes.
//   - action: A string describing the SDK operation being performed.
//
// Returns:
//   - error: An *APIError wrapping the appropriate sentinel for the status code.
func (app *FirecrawlApp) handleError(statusCode int, body []byte, action string) error {
	var errorData map[string]any
	err := json.Unmarshal(body, &errorData)
	if err != nil {
		return &APIError{
			StatusCode: statusCode,
			Message:    fmt.Sprintf("failed to parse error response: %v", err),
			Action:     action,
		}
	}

	errorMessage, _ := errorData["error"].(string)
	if errorMessage == "" {
		errorMessage = "No additional error details provided."
	}

	return &APIError{
		StatusCode: statusCode,
		Message:    errorMessage,
		Action:     action,
	}
}
