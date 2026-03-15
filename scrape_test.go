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
