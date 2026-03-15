package firecrawl

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScrapeURL_Success(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/scrape", r.URL.Path)
		assert.Equal(t, "Bearer fc-test-key", r.Header.Get("Authorization"))

		var body map[string]any
		decodeJSONBody(t, r, &body)
		assert.Equal(t, "https://example.com", body["url"])

		respondJSON(w, http.StatusOK, ScrapeResponse{
			Success: true,
			Data:    &FirecrawlDocument{Markdown: "# Hello"},
		})
	})

	result, err := app.ScrapeURL(context.Background(), "https://example.com", nil)
	require.NoError(t, err)
	assert.Equal(t, "# Hello", result.Markdown)
}

func TestScrapeURL_WithParams(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		decodeJSONBody(t, r, &body)
		assert.Equal(t, "https://example.com", body["url"])

		respondJSON(w, http.StatusOK, ScrapeResponse{
			Success: true,
			Data:    &FirecrawlDocument{Markdown: "# Hello", HTML: "<h1>Hello</h1>"},
		})
	})

	params := &ScrapeParams{
		Formats:         []string{"markdown", "html"},
		OnlyMainContent: ptr(true),
		WaitFor:         ptr(1000),
	}
	result, err := app.ScrapeURL(context.Background(), "https://example.com", params)
	require.NoError(t, err)
	assert.Equal(t, "# Hello", result.Markdown)
	assert.Equal(t, "<h1>Hello</h1>", result.HTML)
}

func TestScrapeURL_Unauthorized(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	})

	_, err := app.ScrapeURL(context.Background(), "https://example.com", nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestScrapeURL_AllParams(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		decodeJSONBody(t, r, &body)

		assert.Equal(t, "https://example.com", body["url"])
		assert.Contains(t, body["formats"], "markdown")
		assert.Contains(t, body["formats"], "html")
		assert.Equal(t, true, body["onlyMainContent"])
		assert.Equal(t, true, body["mobile"])
		assert.NotNil(t, body["waitFor"])
		assert.NotNil(t, body["timeout"])
		assert.NotNil(t, body["location"])
		assert.NotNil(t, body["actions"])

		respondJSON(w, http.StatusOK, ScrapeResponse{
			Success: true,
			Data:    &FirecrawlDocument{Markdown: "# Test"},
		})
	})

	result, err := app.ScrapeURL(context.Background(), "https://example.com", &ScrapeParams{
		Formats:         []string{"markdown", "html"},
		OnlyMainContent: ptr(true),
		Mobile:          ptr(true),
		WaitFor:         ptr(1000),
		Timeout:         ptr(30000),
		Location:        &LocationConfig{Country: "US", Languages: []string{"en"}},
		Actions: []ActionConfig{
			{Type: "wait", Milliseconds: ptr(500)},
			{Type: "click", Selector: ptr("#button")},
		},
		Proxy:              ptr("basic"),
		RemoveBase64Images: ptr(true),
		ZeroDataRetention:  ptr(true),
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestScrapeURL_ServerError(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal failure"})
	})

	_, err := app.ScrapeURL(context.Background(), "https://example.com", nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrServerError)
}

func TestScrapeURL_RateLimited(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusTooManyRequests, map[string]string{"error": "Too many requests"})
	})

	_, err := app.ScrapeURL(context.Background(), "https://example.com", nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRateLimited)
}

func TestScrapeURL_FailedResponse(t *testing.T) {
	// The server returns 200 OK but success:false
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, ScrapeResponse{
			Success: false,
			Data:    nil,
		})
	})

	_, err := app.ScrapeURL(context.Background(), "https://example.com", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to scrape URL")
}

func TestScrapeURL_InvalidJSON(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{not valid json`))
	})

	_, err := app.ScrapeURL(context.Background(), "https://example.com", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse scrape response")
}

func TestScrapeURL_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be made with cancelled context")
	})

	_, err := app.ScrapeURL(ctx, "https://example.com", nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestScrapeURL_NilParams(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		decodeJSONBody(t, r, &body)
		// Only url should be present when params is nil
		assert.Equal(t, "https://example.com", body["url"])
		assert.Nil(t, body["formats"])
		assert.Nil(t, body["mobile"])

		respondJSON(w, http.StatusOK, ScrapeResponse{
			Success: true,
			Data:    &FirecrawlDocument{Markdown: "# Hello"},
		})
	})

	result, err := app.ScrapeURL(context.Background(), "https://example.com", nil)
	require.NoError(t, err)
	assert.NotNil(t, result)
}
