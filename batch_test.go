package firecrawl

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validBatchID is a valid UUID used across batch scrape tests.
const validBatchID = "660e8400-e29b-41d4-a716-446655440001"

// ---- AsyncBatchScrapeURLs ----

func TestAsyncBatchScrapeURLs_Success(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/batch/scrape", r.URL.Path)

		var body map[string]any
		decodeJSONBody(t, r, &body)
		urls, ok := body["urls"].([]any)
		require.True(t, ok)
		assert.Equal(t, "https://example.com", urls[0])

		respondJSON(w, http.StatusOK, BatchScrapeResponse{
			Success: true,
			ID:      validBatchID,
		})
	})

	result, err := app.AsyncBatchScrapeURLs(context.Background(), []string{"https://example.com"}, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, validBatchID, result.ID)
	assert.True(t, result.Success)
}

func TestAsyncBatchScrapeURLs_WithParams(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		decodeJSONBody(t, r, &body)

		urls, ok := body["urls"].([]any)
		require.True(t, ok)
		assert.Len(t, urls, 2)

		assert.NotNil(t, body["maxConcurrency"])
		assert.NotNil(t, body["ignoreInvalidURLs"])
		assert.NotNil(t, body["webhook"])
		assert.NotNil(t, body["scrapeOptions"])

		respondJSON(w, http.StatusOK, BatchScrapeResponse{
			Success: true,
			ID:      validBatchID,
		})
	})

	params := &BatchScrapeParams{
		ScrapeOptions: ScrapeParams{
			Formats:         []string{"markdown"},
			OnlyMainContent: ptr(true),
		},
		MaxConcurrency:    ptr(5),
		IgnoreInvalidURLs: ptr(true),
		Webhook: &WebhookConfig{
			URL:    "https://webhook.example.com/callback",
			Events: []string{"completed", "failed"},
		},
	}

	result, err := app.AsyncBatchScrapeURLs(
		context.Background(),
		[]string{"https://example.com", "https://example.org"},
		params,
		nil,
	)
	require.NoError(t, err)
	assert.Equal(t, validBatchID, result.ID)
}

func TestAsyncBatchScrapeURLs_WithIdempotencyKey(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-idem-key", r.Header.Get("x-idempotency-key"))
		respondJSON(w, http.StatusOK, BatchScrapeResponse{
			Success: true,
			ID:      validBatchID,
		})
	})

	result, err := app.AsyncBatchScrapeURLs(context.Background(), []string{"https://example.com"}, nil, ptr("test-idem-key"))
	require.NoError(t, err)
	assert.Equal(t, validBatchID, result.ID)
}

func TestAsyncBatchScrapeURLs_MissingID(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, BatchScrapeResponse{
			Success: true,
			ID:      "", // Missing ID
		})
	})

	_, err := app.AsyncBatchScrapeURLs(context.Background(), []string{"https://example.com"}, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job ID")
}

func TestAsyncBatchScrapeURLs_Unauthorized(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	})

	_, err := app.AsyncBatchScrapeURLs(context.Background(), []string{"https://example.com"}, nil, nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

// ---- BatchScrapeURLs ----

func TestBatchScrapeURLs_PollsUntilComplete(t *testing.T) {
	var requestCount atomic.Int32
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		if r.Method == http.MethodPost && r.URL.Path == "/v2/batch/scrape" {
			respondJSON(w, http.StatusOK, BatchScrapeResponse{
				Success: true,
				ID:      validBatchID,
			})
			return
		}
		// First GET returns "scraping", subsequent returns "completed".
		if count == 2 {
			respondJSON(w, http.StatusOK, BatchScrapeStatusResponse{
				Status:    "scraping",
				Total:     2,
				Completed: 0,
			})
			return
		}
		respondJSON(w, http.StatusOK, BatchScrapeStatusResponse{
			Status:    "completed",
			Total:     2,
			Completed: 2,
			Data:      []*FirecrawlDocument{{Markdown: "# Page 1"}, {Markdown: "# Page 2"}},
		})
	})

	// Use pollInterval of 0 to skip the 2-second minimum enforcement
	// (the select fires immediately when interval is 0 — this is tested via timeout context).
	// To avoid the min(pollInterval,2) clamp blocking the test, we pass a cancelled-friendly path:
	// the mock returns completed on the third request so no actual sleep occurs.
	result, err := app.BatchScrapeURLs(context.Background(), []string{"https://a.com", "https://b.com"}, nil, nil, 0)
	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	assert.Len(t, result.Data, 2)
}

func TestBatchScrapeURLs_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately before any request

	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be made with cancelled context")
	})

	_, err := app.BatchScrapeURLs(ctx, []string{"https://example.com"}, nil, nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestBatchScrapeURLs_Failed(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			respondJSON(w, http.StatusOK, BatchScrapeResponse{Success: true, ID: validBatchID})
			return
		}
		respondJSON(w, http.StatusOK, BatchScrapeStatusResponse{Status: "failed"})
	})

	_, err := app.BatchScrapeURLs(context.Background(), []string{"https://example.com"}, nil, nil, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed")
}

func TestBatchScrapeURLs_DefaultPollInterval(t *testing.T) {
	// Verify that omitting pollInterval uses the default (no panic, correct code path).
	// The mock returns "completed" immediately so the polling sleep never fires.
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			respondJSON(w, http.StatusOK, BatchScrapeResponse{Success: true, ID: validBatchID})
			return
		}
		respondJSON(w, http.StatusOK, BatchScrapeStatusResponse{
			Status:    "completed",
			Total:     1,
			Completed: 1,
			Data:      []*FirecrawlDocument{{Markdown: "# Page"}},
		})
	})

	result, err := app.BatchScrapeURLs(context.Background(), []string{"https://example.com"}, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
}

// ---- CheckBatchScrapeStatus ----

func TestCheckBatchScrapeStatus_Success(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/batch/scrape/"+validBatchID, r.URL.Path)

		respondJSON(w, http.StatusOK, BatchScrapeStatusResponse{
			Status:      "completed",
			Total:       3,
			Completed:   3,
			CreditsUsed: 3,
			Data:        []*FirecrawlDocument{{Markdown: "# Doc"}},
		})
	})

	result, err := app.CheckBatchScrapeStatus(context.Background(), validBatchID)
	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	assert.Equal(t, 3, result.Total)
	assert.Equal(t, 3, result.Completed)
	assert.Equal(t, 3, result.CreditsUsed)
	assert.Len(t, result.Data, 1)
}

func TestCheckBatchScrapeStatus_Scraping(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, BatchScrapeStatusResponse{
			Status:    "scraping",
			Total:     5,
			Completed: 2,
		})
	})

	result, err := app.CheckBatchScrapeStatus(context.Background(), validBatchID)
	require.NoError(t, err)
	assert.Equal(t, "scraping", result.Status)
	assert.Equal(t, 5, result.Total)
	assert.Equal(t, 2, result.Completed)
}

func TestCheckBatchScrapeStatus_InvalidID(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be made for invalid ID")
	})

	_, err := app.CheckBatchScrapeStatus(context.Background(), "not-a-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UUID")
}

func TestCheckBatchScrapeStatus_PathTraversalID(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be made for path traversal ID")
	})

	_, err := app.CheckBatchScrapeStatus(context.Background(), "../../etc/passwd")
	assert.Error(t, err)
}

func TestCheckBatchScrapeStatus_Unauthorized(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	})

	_, err := app.CheckBatchScrapeStatus(context.Background(), validBatchID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

// ---- monitorBatchScrapeStatus ----

func TestMonitorBatchScrapeStatus_CompletedImmediately(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		respondJSON(w, http.StatusOK, BatchScrapeStatusResponse{
			Status:    "completed",
			Total:     2,
			Completed: 2,
			Data:      []*FirecrawlDocument{{Markdown: "# A"}, {Markdown: "# B"}},
		})
	})

	headers := app.prepareHeaders(nil)
	result, err := app.monitorBatchScrapeStatus(context.Background(), validBatchID, headers, 0)
	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	assert.Len(t, result.Data, 2)
}

func TestMonitorBatchScrapeStatus_Failed(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, BatchScrapeStatusResponse{Status: "failed"})
	})

	headers := app.prepareHeaders(nil)
	_, err := app.monitorBatchScrapeStatus(context.Background(), validBatchID, headers, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed")
}

func TestMonitorBatchScrapeStatus_UnknownStatus(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, BatchScrapeStatusResponse{Status: "pending"})
	})

	headers := app.prepareHeaders(nil)
	_, err := app.monitorBatchScrapeStatus(context.Background(), validBatchID, headers, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown batch scrape status")
}

func TestMonitorBatchScrapeStatus_EmptyStatus(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, BatchScrapeStatusResponse{Status: ""})
	})

	headers := app.prepareHeaders(nil)
	_, err := app.monitorBatchScrapeStatus(context.Background(), validBatchID, headers, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
}

func TestMonitorBatchScrapeStatus_ContextCancelledBeforeRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be made with cancelled context")
	})

	headers := app.prepareHeaders(nil)
	_, err := app.monitorBatchScrapeStatus(ctx, validBatchID, headers, 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestMonitorBatchScrapeStatus_CompletedNoData(t *testing.T) {
	requestCount := 0
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		respondJSON(w, http.StatusOK, BatchScrapeStatusResponse{
			Status:    "completed",
			Total:     1,
			Completed: 1,
			Data:      nil,
		})
	})

	headers := app.prepareHeaders(nil)
	_, err := app.monitorBatchScrapeStatus(context.Background(), validBatchID, headers, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no data was returned")
	assert.GreaterOrEqual(t, requestCount, 3)
}

func TestMonitorBatchScrapeStatus_PaginationUnsafeURL(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		next := "https://attacker.example.com/steal?cursor=2"
		respondJSON(w, http.StatusOK, BatchScrapeStatusResponse{
			Status:    "completed",
			Total:     2,
			Completed: 2,
			Data:      []*FirecrawlDocument{{Markdown: "# Page 1"}},
			Next:      &next,
		})
	})

	headers := app.prepareHeaders(nil)
	_, err := app.monitorBatchScrapeStatus(context.Background(), validBatchID, headers, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsafe pagination URL")
}
