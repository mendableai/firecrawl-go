package firecrawl

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleError_StatusCodes(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		body         string
		wantSentinel error
	}{
		{"401 Unauthorized", 401, `{"error": "Invalid token"}`, ErrUnauthorized},
		{"402 Payment Required", 402, `{"error": "Insufficient credits"}`, ErrPaymentRequired},
		{"404 Not Found", 404, `{"error": "Resource not found"}`, ErrNotFound},
		{"408 Timeout", 408, `{"error": "Timed out"}`, ErrTimeout},
		{"409 Conflict", 409, `{"error": "Duplicate request"}`, ErrConflict},
		{"429 Rate Limited", 429, `{"error": "Too many requests"}`, ErrRateLimited},
		{"500 Server Error", 500, `{"error": "Internal error"}`, ErrServerError},
	}

	app := &FirecrawlApp{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := app.handleError(tt.statusCode, []byte(tt.body), "test action")
			assert.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantSentinel), "expected errors.Is to match %v", tt.wantSentinel)

			var apiErr *APIError
			assert.True(t, errors.As(err, &apiErr))
			assert.Equal(t, tt.statusCode, apiErr.StatusCode)
			assert.Equal(t, "test action", apiErr.Action)
		})
	}
}

func TestHandleError_InvalidJSON(t *testing.T) {
	app := &FirecrawlApp{}
	err := app.handleError(500, []byte("not json"), "test action")
	assert.Error(t, err)

	var apiErr *APIError
	assert.True(t, errors.As(err, &apiErr))
	assert.Equal(t, 500, apiErr.StatusCode)
	assert.Contains(t, apiErr.Message, "failed to parse")
}

func TestHandleError_UnknownStatusCode(t *testing.T) {
	app := &FirecrawlApp{}
	err := app.handleError(418, []byte(`{"error": "I am a teapot"}`), "brew coffee")
	assert.Error(t, err)

	var apiErr *APIError
	assert.True(t, errors.As(err, &apiErr))
	assert.Equal(t, 418, apiErr.StatusCode)
	// Unknown status should have nil Unwrap (no sentinel)
	assert.Nil(t, apiErr.Unwrap())
}

func TestAPIError_ErrorMessage(t *testing.T) {
	err := &APIError{StatusCode: 401, Message: "Invalid token", Action: "scrape URL"}
	assert.Contains(t, err.Error(), "scrape URL")
	assert.Contains(t, err.Error(), "401")
	assert.Contains(t, err.Error(), "Invalid token")
}
