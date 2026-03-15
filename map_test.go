package firecrawl

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapURL_Success(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/map", r.URL.Path)
		assert.Equal(t, "Bearer fc-test-key", r.Header.Get("Authorization"))

		var body map[string]any
		decodeJSONBody(t, r, &body)
		assert.Equal(t, "https://example.com", body["url"])

		respondJSON(w, http.StatusOK, MapResponse{
			Success: true,
			Links: []MapLink{
				{URL: "https://example.com/page1", Title: ptr("Page 1")},
				{URL: "https://example.com/page2", Title: ptr("Page 2")},
			},
		})
	})

	result, err := app.MapURL(context.Background(), "https://example.com", nil)
	require.NoError(t, err)
	assert.Len(t, result.Links, 2)
	assert.Equal(t, "https://example.com/page1", result.Links[0].URL)
	assert.Equal(t, "Page 1", *result.Links[0].Title)
}

func TestMapURL_AllParams(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		decodeJSONBody(t, r, &body)

		assert.Equal(t, "https://example.com", body["url"])
		assert.Equal(t, true, body["includeSubdomains"])
		assert.Equal(t, "blog", body["search"])
		assert.NotNil(t, body["limit"])
		assert.Equal(t, "include", body["sitemap"])
		assert.Equal(t, true, body["ignoreQueryParameters"])
		assert.Equal(t, true, body["ignoreCache"])
		assert.NotNil(t, body["timeout"])
		assert.NotNil(t, body["location"])

		respondJSON(w, http.StatusOK, MapResponse{
			Success: true,
			Links:   []MapLink{{URL: "https://example.com/blog/post-1"}},
		})
	})

	result, err := app.MapURL(context.Background(), "https://example.com", &MapParams{
		IncludeSubdomains:     ptr(true),
		Search:                ptr("blog"),
		Limit:                 ptr(1000),
		Sitemap:               ptr("include"),
		IgnoreQueryParameters: ptr(true),
		IgnoreCache:           ptr(true),
		Timeout:               ptr(30000),
		Location:              &LocationConfig{Country: "US", Languages: []string{"en"}},
	})
	require.NoError(t, err)
	assert.Len(t, result.Links, 1)
}

func TestMapURL_NilParams(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		decodeJSONBody(t, r, &body)
		// Only url should be present, no optional params
		assert.Equal(t, "https://example.com", body["url"])
		assert.Nil(t, body["includeSubdomains"])
		assert.Nil(t, body["search"])

		respondJSON(w, http.StatusOK, MapResponse{
			Success: true,
			Links:   []MapLink{{URL: "https://example.com"}},
		})
	})

	result, err := app.MapURL(context.Background(), "https://example.com", nil)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestMapURL_EmptyLinks(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, MapResponse{
			Success: true,
			Links:   []MapLink{},
		})
	})

	result, err := app.MapURL(context.Background(), "https://example.com", nil)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Links)
}

func TestMapURL_FailedResponse(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, MapResponse{
			Success: false,
			Error:   "map operation failed: site not reachable",
		})
	})

	_, err := app.MapURL(context.Background(), "https://example.com", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "map operation failed")
}

func TestMapURL_Unauthorized(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	})

	_, err := app.MapURL(context.Background(), "https://example.com", nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestMapURL_ServerError(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal failure"})
	})

	_, err := app.MapURL(context.Background(), "https://example.com", nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrServerError)
}
