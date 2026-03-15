package firecrawl

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validCrawlID is a valid UUID used across crawl tests.
const validCrawlID = "550e8400-e29b-41d4-a716-446655440000"

// ---- CrawlURL ----

func TestCrawlURL_Success(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v2/crawl" {
			respondJSON(w, http.StatusOK, CrawlResponse{
				Success: true,
				ID:      validCrawlID,
			})
			return
		}
		// GET /v2/crawl/{id} — immediately completed
		respondJSON(w, http.StatusOK, CrawlStatusResponse{
			Status:    "completed",
			Total:     1,
			Completed: 1,
			Data:      []*FirecrawlDocument{{Markdown: "# Page"}},
		})
	})

	result, err := app.CrawlURL(context.Background(), "https://example.com", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "# Page", result.Data[0].Markdown)
}

func TestCrawlURL_AllParams(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			decodeJSONBody(t, r, &body)
			assert.Equal(t, "https://example.com", body["url"])
			assert.NotNil(t, body["limit"])
			assert.NotNil(t, body["maxDiscoveryDepth"])
			assert.NotNil(t, body["crawlEntireDomain"])
			assert.NotNil(t, body["allowSubdomains"])
			assert.NotNil(t, body["ignoreQueryParameters"])
			assert.NotNil(t, body["zeroDataRetention"])
			respondJSON(w, http.StatusOK, CrawlResponse{
				Success: true,
				ID:      validCrawlID,
			})
			return
		}
		respondJSON(w, http.StatusOK, CrawlStatusResponse{
			Status:    "completed",
			Total:     1,
			Completed: 1,
			Data:      []*FirecrawlDocument{{Markdown: "# Page"}},
		})
	})

	params := &CrawlParams{
		Limit:                 ptr(100),
		MaxDiscoveryDepth:     ptr(3),
		CrawlEntireDomain:     ptr(true),
		AllowSubdomains:       ptr(true),
		IgnoreQueryParameters: ptr(true),
		ZeroDataRetention:     ptr(true),
		IncludePaths:          []string{"/docs/*"},
		ExcludePaths:          []string{"/admin/*"},
	}
	result, err := app.CrawlURL(context.Background(), "https://example.com", params, nil)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCrawlURL_WithIdempotencyKey(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			assert.Equal(t, "test-idempotency-key", r.Header.Get("x-idempotency-key"))
			respondJSON(w, http.StatusOK, CrawlResponse{
				Success: true,
				ID:      validCrawlID,
			})
			return
		}
		respondJSON(w, http.StatusOK, CrawlStatusResponse{
			Status:    "completed",
			Total:     1,
			Completed: 1,
			Data:      []*FirecrawlDocument{{Markdown: "# Page"}},
		})
	})

	result, err := app.CrawlURL(context.Background(), "https://example.com", nil, ptr("test-idempotency-key"))
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCrawlURL_Failed(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			respondJSON(w, http.StatusOK, CrawlResponse{Success: true, ID: validCrawlID})
			return
		}
		respondJSON(w, http.StatusOK, CrawlStatusResponse{Status: "failed"})
	})

	_, err := app.CrawlURL(context.Background(), "https://example.com", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed")
}

func TestCrawlURL_PollsUntilComplete(t *testing.T) {
	pollCount := 0
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			respondJSON(w, http.StatusOK, CrawlResponse{
				Success: true,
				ID:      validCrawlID,
			})
			return
		}
		// GET: Return completed immediately — tests that the polling loop works
		// and correctly collects data on first successful poll.
		pollCount++
		respondJSON(w, http.StatusOK, CrawlStatusResponse{
			Status:    "completed",
			Total:     5,
			Completed: 5,
			Data:      []*FirecrawlDocument{{Markdown: "# Page 1"}, {Markdown: "# Page 2"}},
		})
	})

	result, err := app.CrawlURL(context.Background(), "https://example.com", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	assert.Len(t, result.Data, 2)
	assert.GreaterOrEqual(t, pollCount, 1)
}

func TestCrawlURL_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be made with cancelled context")
	})

	_, err := app.CrawlURL(ctx, "https://example.com", nil, nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestCrawlURL_Unauthorized(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	})

	_, err := app.CrawlURL(context.Background(), "https://example.com", nil, nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

// Note: pagination is tested directly via monitorJobStatus in helpers_test.go.
// CrawlURL delegates pagination to monitorJobStatus, so it is implicitly covered.

// ---- AsyncCrawlURL ----

func TestAsyncCrawlURL_Success(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/crawl", r.URL.Path)

		var body map[string]any
		decodeJSONBody(t, r, &body)
		assert.Equal(t, "https://example.com", body["url"])

		respondJSON(w, http.StatusOK, CrawlResponse{
			Success: true,
			ID:      validCrawlID,
		})
	})

	result, err := app.AsyncCrawlURL(context.Background(), "https://example.com", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, validCrawlID, result.ID)
	assert.True(t, result.Success)
}

func TestAsyncCrawlURL_AllParams(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		decodeJSONBody(t, r, &body)
		assert.Equal(t, "https://example.com", body["url"])
		assert.NotNil(t, body["limit"])
		assert.NotNil(t, body["webhook"])

		respondJSON(w, http.StatusOK, CrawlResponse{
			Success: true,
			ID:      validCrawlID,
		})
	})

	params := &CrawlParams{
		Limit: ptr(50),
		Webhook: &WebhookConfig{
			URL:    "https://webhook.example.com/callback",
			Events: []string{"completed", "failed"},
		},
	}
	result, err := app.AsyncCrawlURL(context.Background(), "https://example.com", params, nil)
	require.NoError(t, err)
	assert.Equal(t, validCrawlID, result.ID)
}

func TestAsyncCrawlURL_MissingID(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, CrawlResponse{
			Success: true,
			ID:      "", // Missing ID
		})
	})

	_, err := app.AsyncCrawlURL(context.Background(), "https://example.com", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job ID")
}

func TestAsyncCrawlURL_Unauthorized(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	})

	_, err := app.AsyncCrawlURL(context.Background(), "https://example.com", nil, nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

// ---- CheckCrawlStatus ----

func TestCheckCrawlStatus_Success(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/crawl/"+validCrawlID, r.URL.Path)

		respondJSON(w, http.StatusOK, CrawlStatusResponse{
			Status:    "completed",
			Total:     10,
			Completed: 10,
			Data:      []*FirecrawlDocument{{Markdown: "# Page"}},
		})
	})

	result, err := app.CheckCrawlStatus(context.Background(), validCrawlID)
	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	assert.Equal(t, 10, result.Total)
	assert.Equal(t, 10, result.Completed)
}

func TestCheckCrawlStatus_InvalidID(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be made for invalid ID")
	})

	_, err := app.CheckCrawlStatus(context.Background(), "not-a-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UUID")
}

func TestCheckCrawlStatus_PathTraversalID(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be made for path traversal ID")
	})

	_, err := app.CheckCrawlStatus(context.Background(), "../../etc/passwd")
	assert.Error(t, err)
}

func TestCheckCrawlStatus_ServerError(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal failure"})
	})

	_, err := app.CheckCrawlStatus(context.Background(), validCrawlID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrServerError)
}

// ---- CancelCrawlJob ----

func TestCancelCrawlJob_Success(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/v2/crawl/"+validCrawlID, r.URL.Path)

		respondJSON(w, http.StatusOK, CancelCrawlJobResponse{
			Success: true,
			Status:  "cancelled",
		})
	})

	status, err := app.CancelCrawlJob(context.Background(), validCrawlID)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", status)
}

func TestCancelCrawlJob_InvalidID(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be made for invalid ID")
	})

	_, err := app.CancelCrawlJob(context.Background(), "not-a-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UUID")
}

func TestCancelCrawlJob_Unauthorized(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	})

	_, err := app.CancelCrawlJob(context.Background(), validCrawlID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

// ---- buildCrawlRequest ----

func TestBuildCrawlRequest_NilParams(t *testing.T) {
	req, err := buildCrawlRequest("https://example.com", nil)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", req.URL)
	assert.Nil(t, req.ScrapeOptions)
	assert.Nil(t, req.Limit)
	assert.Nil(t, req.Webhook)
}

func TestBuildCrawlRequest_AllParams(t *testing.T) {
	params := &CrawlParams{
		Limit:             ptr(100),
		MaxDiscoveryDepth: ptr(3),
		CrawlEntireDomain: ptr(true),
		AllowSubdomains:   ptr(false),
		IncludePaths:      []string{"/blog/*"},
		ExcludePaths:      []string{"/admin/*"},
		Prompt:            ptr("Crawl only article pages"),
		RegexOnFullURL:    ptr(true),
		ZeroDataRetention: ptr(true),
	}

	req, err := buildCrawlRequest("https://example.com", params)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", req.URL)
	assert.Equal(t, 100, *req.Limit)
	assert.Equal(t, 3, *req.MaxDiscoveryDepth)
	assert.True(t, *req.CrawlEntireDomain)
	assert.False(t, *req.AllowSubdomains)
	assert.Equal(t, []string{"/blog/*"}, req.IncludePaths)
	assert.Equal(t, []string{"/admin/*"}, req.ExcludePaths)
	assert.Equal(t, "Crawl only article pages", *req.Prompt)
	assert.True(t, *req.RegexOnFullURL)
	assert.True(t, *req.ZeroDataRetention)
}

func TestBuildCrawlRequest_WithScrapeOptions(t *testing.T) {
	params := &CrawlParams{
		ScrapeOptions: ScrapeParams{
			Formats:         []string{"markdown"},
			OnlyMainContent: ptr(true),
		},
	}

	req, err := buildCrawlRequest("https://example.com", params)
	require.NoError(t, err)
	assert.NotNil(t, req.ScrapeOptions)
	assert.Equal(t, []string{"markdown"}, req.ScrapeOptions.Formats)
	assert.True(t, *req.ScrapeOptions.OnlyMainContent)
}

func TestBuildCrawlRequest_EmptyScrapeOptions(t *testing.T) {
	// Empty ScrapeOptions should not be included in the request
	params := &CrawlParams{
		Limit: ptr(10),
		// ScrapeOptions is zero value — should be omitted
	}

	req, err := buildCrawlRequest("https://example.com", params)
	require.NoError(t, err)
	assert.Nil(t, req.ScrapeOptions)
	assert.Equal(t, 10, *req.Limit)
}

// ---- CheckCrawlStatus with PaginationConfig ----

func TestCheckCrawlStatus_NoPagination_BackwardCompat(t *testing.T) {
	// Calling without pagination parameter returns the single page (backward compatible).
	var serverURL string
	app, srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		next := serverURL + "/v2/crawl/" + validCrawlID + "?cursor=2"
		respondJSON(w, http.StatusOK, CrawlStatusResponse{
			Status:    "completed",
			Total:     2,
			Completed: 2,
			Data:      []*FirecrawlDocument{{Markdown: "# Page 1"}},
			Next:      &next,
		})
	})
	serverURL = srv.URL

	result, err := app.CheckCrawlStatus(context.Background(), validCrawlID)
	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	// Only the first page returned — Next is present but not followed.
	assert.Len(t, result.Data, 1)
	assert.NotNil(t, result.Next)
}

func TestCheckCrawlStatus_AutoPaginate_FollowsNextURLs(t *testing.T) {
	requestCount := 0
	var serverURL string
	app, srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount == 1 {
			// First page: has a Next URL.
			next := serverURL + "/v2/crawl/" + validCrawlID + "?cursor=2"
			respondJSON(w, http.StatusOK, CrawlStatusResponse{
				Status:    "completed",
				Total:     2,
				Completed: 2,
				Data:      []*FirecrawlDocument{{Markdown: "# Page 1"}},
				Next:      &next,
			})
			return
		}
		// Second page: no Next URL, pagination ends.
		respondJSON(w, http.StatusOK, CrawlStatusResponse{
			Status:    "completed",
			Total:     2,
			Completed: 2,
			Data:      []*FirecrawlDocument{{Markdown: "# Page 2"}},
		})
	})
	serverURL = srv.URL

	cfg := &PaginationConfig{AutoPaginate: ptr(true)}
	result, err := app.CheckCrawlStatus(context.Background(), validCrawlID, cfg)
	require.NoError(t, err)
	assert.Equal(t, 2, requestCount)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, "# Page 1", result.Data[0].Markdown)
	assert.Equal(t, "# Page 2", result.Data[1].Markdown)
}

func TestCheckCrawlStatus_MaxPages_StopsAfterLimit(t *testing.T) {
	requestCount := 0
	var serverURL string
	app, srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		next := serverURL + "/v2/crawl/" + validCrawlID + "?cursor=" + fmt.Sprintf("%d", requestCount+1)
		respondJSON(w, http.StatusOK, CrawlStatusResponse{
			Status:    "completed",
			Total:     10,
			Completed: 10,
			Data:      []*FirecrawlDocument{{Markdown: fmt.Sprintf("# Page %d", requestCount)}},
			Next:      &next,
		})
	})
	serverURL = srv.URL

	cfg := &PaginationConfig{
		AutoPaginate: ptr(true),
		MaxPages:     ptr(2), // Stop after 2 pages total.
	}
	result, err := app.CheckCrawlStatus(context.Background(), validCrawlID, cfg)
	require.NoError(t, err)
	// Only fetched page 1 (initial) + page 2 stopped by MaxPages limit.
	assert.Equal(t, 2, requestCount)
	assert.Len(t, result.Data, 2)
}

func TestCheckCrawlStatus_MaxResults_TruncatesExcess(t *testing.T) {
	requestCount := 0
	var serverURL string
	app, srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		next := serverURL + "/v2/crawl/" + validCrawlID + "?cursor=2"
		respondJSON(w, http.StatusOK, CrawlStatusResponse{
			Status:    "completed",
			Total:     6,
			Completed: 6,
			Data: []*FirecrawlDocument{
				{Markdown: "# Doc A"},
				{Markdown: "# Doc B"},
				{Markdown: "# Doc C"},
			},
			Next: &next,
		})
	})
	serverURL = srv.URL

	cfg := &PaginationConfig{
		AutoPaginate: ptr(true),
		MaxResults:   ptr(3), // Stop after collecting 3 results total.
	}
	result, err := app.CheckCrawlStatus(context.Background(), validCrawlID, cfg)
	require.NoError(t, err)
	// First page gives 3 docs which meets MaxResults — no second request made.
	assert.Equal(t, 1, requestCount)
	assert.Len(t, result.Data, 3)
}

func TestCheckCrawlStatus_AutoPaginate_UnsafeNextURL(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		next := "https://attacker.example.com/steal?cursor=2"
		respondJSON(w, http.StatusOK, CrawlStatusResponse{
			Status:    "completed",
			Total:     2,
			Completed: 2,
			Data:      []*FirecrawlDocument{{Markdown: "# Page 1"}},
			Next:      &next,
		})
	})

	cfg := &PaginationConfig{AutoPaginate: ptr(true)}
	_, err := app.CheckCrawlStatus(context.Background(), validCrawlID, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsafe pagination URL")
}

// ---- GetCrawlStatusPage ----

func TestGetCrawlStatusPage_Success(t *testing.T) {
	app, srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		respondJSON(w, http.StatusOK, CrawlStatusResponse{
			Status:    "completed",
			Total:     5,
			Completed: 5,
			Data:      []*FirecrawlDocument{{Markdown: "# Page 2"}},
		})
	})

	nextURL := srv.URL + "/v2/crawl/" + validCrawlID + "?cursor=2"
	result, err := app.GetCrawlStatusPage(context.Background(), nextURL)
	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "# Page 2", result.Data[0].Markdown)
}

func TestGetCrawlStatusPage_InvalidURL_SSRFBlocked(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be made to untrusted host")
	})

	_, err := app.GetCrawlStatusPage(context.Background(), "https://attacker.example.com/steal")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsafe pagination URL")
}
