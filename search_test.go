package firecrawl

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearch_Success(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/search", r.URL.Path)
		assert.Equal(t, "Bearer fc-test-key", r.Header.Get("Authorization"))

		var body map[string]any
		decodeJSONBody(t, r, &body)
		assert.Equal(t, "golang tutorials", body["query"])

		respondJSON(w, http.StatusOK, SearchResponse{
			Success: true,
			Data: SearchData{
				Web: []SearchWebResult{
					{Title: "Go Tour", Description: "A tour of Go", URL: "https://go.dev/tour"},
					{Title: "Go Docs", Description: "Go documentation", URL: "https://pkg.go.dev"},
				},
			},
			CreditsUsed: 1,
		})
	})

	result, err := app.Search(context.Background(), "golang tutorials", nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Len(t, result.Data.Web, 2)
	assert.Equal(t, "Go Tour", result.Data.Web[0].Title)
	assert.Equal(t, "https://go.dev/tour", result.Data.Web[0].URL)
	assert.Equal(t, 1, result.CreditsUsed)
}

func TestSearch_WithParams(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		decodeJSONBody(t, r, &body)

		assert.Equal(t, "go programming", body["query"])
		assert.NotNil(t, body["limit"])
		assert.Equal(t, float64(5), body["limit"])
		assert.Contains(t, body["sources"], "web")
		assert.Contains(t, body["sources"], "news")
		assert.Contains(t, body["categories"], "github")
		assert.Equal(t, "qdr:d", body["tbs"])
		assert.Equal(t, "New York", body["location"])
		assert.Equal(t, "US", body["country"])
		assert.NotNil(t, body["timeout"])
		assert.Equal(t, true, body["ignoreInvalidURLs"])
		assert.NotNil(t, body["scrapeOptions"])

		respondJSON(w, http.StatusOK, SearchResponse{
			Success: true,
			Data: SearchData{
				Web:  []SearchWebResult{{Title: "Result", Description: "Desc", URL: "https://example.com"}},
				News: []SearchNewsResult{{Title: "News", Snippet: "Snippet", URL: "https://news.example.com", Date: "2026-03-15", Position: 1}},
			},
		})
	})

	params := &SearchParams{
		Limit:             ptr(5),
		Sources:           []string{"web", "news"},
		Categories:        []string{"github"},
		TBS:               ptr("qdr:d"),
		Location:          ptr("New York"),
		Country:           ptr("US"),
		Timeout:           ptr(10000),
		IgnoreInvalidURLs: ptr(true),
		ScrapeOptions:     &ScrapeParams{Formats: []string{"markdown"}},
	}

	result, err := app.Search(context.Background(), "go programming", params)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Data.Web, 1)
	assert.Len(t, result.Data.News, 1)
}

func TestSearch_EmptyQuery(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		decodeJSONBody(t, r, &body)
		// Empty string query is sent — the API decides whether to accept it
		assert.Equal(t, "", body["query"])

		respondJSON(w, http.StatusOK, SearchResponse{
			Success: true,
			Data:    SearchData{},
		})
	})

	result, err := app.Search(context.Background(), "", nil)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSearch_Unauthorized(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid API key"})
	})

	_, err := app.Search(context.Background(), "test query", nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestSearch_ServerError(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	})

	_, err := app.Search(context.Background(), "test query", nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrServerError)
}

func TestSearch_FailedResponse(t *testing.T) {
	// Server returns 200 OK but success:false
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, SearchResponse{
			Success: false,
		})
	})

	_, err := app.Search(context.Background(), "test query", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "search operation failed")
}

func TestSearch_RateLimited(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
	})

	_, err := app.Search(context.Background(), "test query", nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRateLimited)
}

func TestSearch_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel before making any request

	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be made with cancelled context")
	})

	_, err := app.Search(ctx, "test query", nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}
