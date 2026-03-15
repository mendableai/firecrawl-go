package firecrawl

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validExtractID is a valid UUID used across extract tests.
const validExtractID = "660e8400-e29b-41d4-a716-446655440002"

// ---- AsyncExtract ----

func TestAsyncExtract_Success(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/extract", r.URL.Path)

		var body map[string]any
		decodeJSONBody(t, r, &body)
		urls, ok := body["urls"].([]any)
		require.True(t, ok)
		assert.Equal(t, "https://example.com", urls[0])

		respondJSON(w, http.StatusOK, ExtractResponse{
			Success: true,
			ID:      validExtractID,
		})
	})

	result, err := app.AsyncExtract(context.Background(), []string{"https://example.com"}, nil)
	require.NoError(t, err)
	assert.Equal(t, validExtractID, result.ID)
	assert.True(t, result.Success)
}

func TestAsyncExtract_WithParams(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		decodeJSONBody(t, r, &body)

		urls, ok := body["urls"].([]any)
		require.True(t, ok)
		assert.Len(t, urls, 2)

		assert.NotNil(t, body["prompt"])
		assert.NotNil(t, body["schema"])
		assert.NotNil(t, body["enableWebSearch"])
		assert.NotNil(t, body["ignoreSitemap"])
		assert.NotNil(t, body["includeSubdomains"])
		assert.NotNil(t, body["showSources"])
		assert.NotNil(t, body["ignoreInvalidURLs"])
		assert.NotNil(t, body["scrapeOptions"])

		respondJSON(w, http.StatusOK, ExtractResponse{
			Success: true,
			ID:      validExtractID,
		})
	})

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
	}
	params := &ExtractParams{
		Prompt:            ptr("Extract the company name"),
		Schema:            schema,
		EnableWebSearch:   ptr(true),
		IgnoreSitemap:     ptr(false),
		IncludeSubdomains: ptr(true),
		ShowSources:       ptr(true),
		IgnoreInvalidURLs: ptr(true),
		ScrapeOptions: &ScrapeParams{
			Formats: []string{"markdown"},
		},
	}

	result, err := app.AsyncExtract(
		context.Background(),
		[]string{"https://example.com", "https://example.org"},
		params,
	)
	require.NoError(t, err)
	assert.Equal(t, validExtractID, result.ID)
}

func TestAsyncExtract_MissingID(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, ExtractResponse{
			Success: true,
			ID:      "", // Missing ID
		})
	})

	_, err := app.AsyncExtract(context.Background(), []string{"https://example.com"}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job ID")
}

func TestAsyncExtract_Unauthorized(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	})

	_, err := app.AsyncExtract(context.Background(), []string{"https://example.com"}, nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

// ---- Extract ----

func TestExtract_PollsUntilComplete(t *testing.T) {
	var requestCount atomic.Int32
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		if r.Method == http.MethodPost && r.URL.Path == "/v2/extract" {
			respondJSON(w, http.StatusOK, ExtractResponse{
				Success: true,
				ID:      validExtractID,
			})
			return
		}
		// First GET returns "processing", subsequent returns "completed".
		if count == 2 {
			respondJSON(w, http.StatusOK, ExtractStatusResponse{
				Status: "processing",
			})
			return
		}
		respondJSON(w, http.StatusOK, ExtractStatusResponse{
			Status:      "completed",
			Success:     true,
			CreditsUsed: 2,
			Data: map[string]any{
				"name": "Acme Corp",
			},
		})
	})

	result, err := app.Extract(context.Background(), []string{"https://example.com"}, nil)
	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	assert.NotNil(t, result.Data)
	assert.Equal(t, "Acme Corp", result.Data["name"])
}

func TestExtract_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately before any request

	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be made with cancelled context")
	})

	_, err := app.Extract(ctx, []string{"https://example.com"}, nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestExtract_Failed(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			respondJSON(w, http.StatusOK, ExtractResponse{Success: true, ID: validExtractID})
			return
		}
		respondJSON(w, http.StatusOK, ExtractStatusResponse{Status: "failed"})
	})

	_, err := app.Extract(context.Background(), []string{"https://example.com"}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed")
}

// ---- CheckExtractStatus ----

func TestCheckExtractStatus_Success(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/extract/"+validExtractID, r.URL.Path)

		respondJSON(w, http.StatusOK, ExtractStatusResponse{
			Success:     true,
			Status:      "completed",
			CreditsUsed: 5,
			ExpiresAt:   "2026-04-15T00:00:00Z",
			Data: map[string]any{
				"company": "Acme Corp",
				"founded": float64(1990),
			},
		})
	})

	result, err := app.CheckExtractStatus(context.Background(), validExtractID)
	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	assert.True(t, result.Success)
	assert.Equal(t, 5, result.CreditsUsed)
	assert.Equal(t, "2026-04-15T00:00:00Z", result.ExpiresAt)
	assert.Equal(t, "Acme Corp", result.Data["company"])
}

func TestCheckExtractStatus_Processing(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, ExtractStatusResponse{
			Status: "processing",
		})
	})

	result, err := app.CheckExtractStatus(context.Background(), validExtractID)
	require.NoError(t, err)
	assert.Equal(t, "processing", result.Status)
}

func TestCheckExtractStatus_InvalidID(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be made for invalid ID")
	})

	_, err := app.CheckExtractStatus(context.Background(), "not-a-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UUID")
}

func TestCheckExtractStatus_PathTraversalID(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be made for path traversal ID")
	})

	_, err := app.CheckExtractStatus(context.Background(), "../../etc/passwd")
	assert.Error(t, err)
}

func TestCheckExtractStatus_Unauthorized(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	})

	_, err := app.CheckExtractStatus(context.Background(), validExtractID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

// ---- monitorExtractStatus ----

func TestMonitorExtractStatus_CompletedImmediately(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		respondJSON(w, http.StatusOK, ExtractStatusResponse{
			Status:  "completed",
			Success: true,
			Data:    map[string]any{"result": "value"},
		})
	})

	headers := app.prepareHeaders(nil)
	result, err := app.monitorExtractStatus(context.Background(), validExtractID, headers)
	require.NoError(t, err)
	assert.Equal(t, "completed", result.Status)
	assert.Equal(t, "value", result.Data["result"])
}

func TestMonitorExtractStatus_Failed(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, ExtractStatusResponse{Status: "failed"})
	})

	headers := app.prepareHeaders(nil)
	_, err := app.monitorExtractStatus(context.Background(), validExtractID, headers)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed")
}

func TestMonitorExtractStatus_UnknownStatus(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, ExtractStatusResponse{Status: "pending"})
	})

	headers := app.prepareHeaders(nil)
	_, err := app.monitorExtractStatus(context.Background(), validExtractID, headers)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown extract status")
}

func TestMonitorExtractStatus_EmptyStatus(t *testing.T) {
	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, ExtractStatusResponse{Status: ""})
	})

	headers := app.prepareHeaders(nil)
	_, err := app.monitorExtractStatus(context.Background(), validExtractID, headers)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
}

func TestMonitorExtractStatus_ContextCancelledBeforeRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	app, _ := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not be made with cancelled context")
	})

	headers := app.prepareHeaders(nil)
	_, err := app.monitorExtractStatus(ctx, validExtractID, headers)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}
