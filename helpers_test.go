package firecrawl

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- makeRequest ----

func TestMakeRequest_Success(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/test", r.URL.Path)
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	headers := app.prepareHeaders(nil)
	resp, err := app.makeRequest(
		context.Background(),
		http.MethodGet,
		app.APIURL+"/test",
		nil,
		headers,
		"test request",
	)
	require.NoError(t, err)
	assert.Contains(t, string(resp), "ok")
}

func TestMakeRequest_PostWithBody(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body map[string]any
		decodeJSONBody(t, r, &body)
		assert.Equal(t, "https://example.com", body["url"])

		respondJSON(w, http.StatusOK, map[string]string{"result": "success"})
	})

	headers := app.prepareHeaders(nil)
	reqBody := []byte(`{"url":"https://example.com"}`)
	resp, err := app.makeRequest(
		context.Background(),
		http.MethodPost,
		app.APIURL+"/test",
		reqBody,
		headers,
		"test post",
	)
	require.NoError(t, err)
	assert.Contains(t, string(resp), "success")
}

func TestMakeRequest_RetryOn502(t *testing.T) {
	attempts := 0
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	headers := app.prepareHeaders(nil)
	resp, err := app.makeRequest(
		context.Background(),
		http.MethodGet,
		app.APIURL+"/test",
		nil,
		headers,
		"test retry",
		withRetries(3),
		withBackoff(0), // 0ms backoff for fast tests
	)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 3, attempts)
}

func TestMakeRequest_NoRetryOn4xx(t *testing.T) {
	attempts := 0
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Bad request"})
	})

	headers := app.prepareHeaders(nil)
	_, err := app.makeRequest(
		context.Background(),
		http.MethodGet,
		app.APIURL+"/test",
		nil,
		headers,
		"test no retry",
		withRetries(3),
		withBackoff(0),
	)
	// Should fail immediately, not retry
	assert.Error(t, err)
	assert.Equal(t, 1, attempts, "4xx errors should not be retried")
}

func TestMakeRequest_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be made with cancelled context")
	})

	headers := app.prepareHeaders(nil)
	_, err := app.makeRequest(
		ctx,
		http.MethodGet,
		app.APIURL+"/test",
		nil,
		headers,
		"test cancelled",
	)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestMakeRequest_NonJSONErrorBody(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error\n"))
	})

	headers := app.prepareHeaders(nil)
	_, err := app.makeRequest(
		context.Background(),
		http.MethodGet,
		app.APIURL+"/test",
		nil,
		headers,
		"test non-json error",
	)
	assert.Error(t, err)
	// Should still produce an error with status code info
	var apiErr *APIError
	assert.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 500, apiErr.StatusCode)
}

func TestMakeRequest_AuthorizationHeader(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer fc-test-key", r.Header.Get("Authorization"))
		respondJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	})

	headers := app.prepareHeaders(nil)
	_, err := app.makeRequest(
		context.Background(),
		http.MethodGet,
		app.APIURL+"/test",
		nil,
		headers,
		"test auth header",
	)
	require.NoError(t, err)
}

// ---- monitorJobStatus ----

func TestMonitorJobStatus_CompletedImmediately(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		respondJSON(w, http.StatusOK, CrawlStatusResponse{
			Status:    "completed",
			Total:     3,
			Completed: 3,
			Data:      []*FirecrawlDocument{{Markdown: "# Doc 1"}, {Markdown: "# Doc 2"}, {Markdown: "# Doc 3"}},
		})
	})

	headers := app.prepareHeaders(nil)
	result, err := app.monitorJobStatus(context.Background(), validCrawlID, headers, 0)
	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	assert.Len(t, result.Data, 3)
}

func TestMonitorJobStatus_Failed(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, CrawlStatusResponse{Status: "failed"})
	})

	headers := app.prepareHeaders(nil)
	_, err := app.monitorJobStatus(context.Background(), validCrawlID, headers, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed")
}

func TestMonitorJobStatus_UnknownStatus(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, CrawlStatusResponse{Status: "unknown_status"})
	})

	headers := app.prepareHeaders(nil)
	_, err := app.monitorJobStatus(context.Background(), validCrawlID, headers, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown crawl status")
}

func TestMonitorJobStatus_EmptyStatus(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, CrawlStatusResponse{Status: ""})
	})

	headers := app.prepareHeaders(nil)
	_, err := app.monitorJobStatus(context.Background(), validCrawlID, headers, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
}

func TestMonitorJobStatus_ContextCancelledBeforeRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel before any request

	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be made with cancelled context")
	})

	headers := app.prepareHeaders(nil)
	_, err := app.monitorJobStatus(ctx, validCrawlID, headers, 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestMonitorJobStatus_CompletedNoData(t *testing.T) {
	// When status is "completed" but Data is nil, it retries up to 3 times then errors.
	requestCount := 0
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		respondJSON(w, http.StatusOK, CrawlStatusResponse{
			Status:    "completed",
			Total:     1,
			Completed: 1,
			Data:      nil, // No data
		})
	})

	headers := app.prepareHeaders(nil)
	_, err := app.monitorJobStatus(context.Background(), validCrawlID, headers, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no data was returned")
	// Should have retried 3+ times before giving up
	assert.GreaterOrEqual(t, requestCount, 3)
}

func TestMonitorJobStatus_PaginationUnsafeURL(t *testing.T) {
	// Verify that monitorJobStatus rejects pagination Next URLs pointing to a different host (SSRF prevention).
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Return a completed status with a Next URL pointing to a different (attacker-controlled) host
		next := "https://attacker.example.com/steal-token?cursor=2"
		respondJSON(w, http.StatusOK, CrawlStatusResponse{
			Status:    "completed",
			Total:     2,
			Completed: 2,
			Data:      []*FirecrawlDocument{{Markdown: "# Page 1"}},
			Next:      &next,
		})
	})

	headers := app.prepareHeaders(nil)
	_, err := app.monitorJobStatus(context.Background(), validCrawlID, headers, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsafe pagination URL")
}
