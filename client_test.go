package firecrawl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFirecrawlApp_ValidKey(t *testing.T) {
	app, err := NewFirecrawlApp("fc-test-key", "https://api.example.com")
	require.NoError(t, err)
	assert.Equal(t, "fc-test-key", app.APIKey())
	assert.Equal(t, "https://api.example.com", app.APIURL)
}

func TestNewFirecrawlApp_EmptyKey(t *testing.T) {
	// Unset env var to ensure no fallback
	t.Setenv("FIRECRAWL_API_KEY", "")
	_, err := NewFirecrawlApp("", "https://api.example.com")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNoAPIKey)
}

func TestNewFirecrawlApp_DefaultURL(t *testing.T) {
	t.Setenv("FIRECRAWL_API_URL", "")
	app, err := NewFirecrawlApp("fc-test-key", "")
	require.NoError(t, err)
	assert.Equal(t, "https://api.firecrawl.dev", app.APIURL)
}

func TestNewFirecrawlApp_EnvFallback(t *testing.T) {
	t.Setenv("FIRECRAWL_API_KEY", "fc-env-key")
	app, err := NewFirecrawlApp("", "https://api.example.com")
	require.NoError(t, err)
	assert.Equal(t, "fc-env-key", app.APIKey())
}

func TestNewFirecrawlApp_EnvURLFallback(t *testing.T) {
	t.Setenv("FIRECRAWL_API_URL", "https://custom.api.example.com")
	app, err := NewFirecrawlApp("fc-test-key", "")
	require.NoError(t, err)
	assert.Equal(t, "https://custom.api.example.com", app.APIURL)
}

func TestNewFirecrawlApp_ClientNotNil(t *testing.T) {
	// Verify the HTTP client is properly initialized
	app, err := NewFirecrawlApp("fc-test-key", "https://api.example.com")
	require.NoError(t, err)
	assert.NotNil(t, app.Client)
}

func TestPrepareHeaders_WithIdempotencyKey(t *testing.T) {
	app, err := NewFirecrawlApp("fc-test-key", "https://api.firecrawl.dev")
	require.NoError(t, err)

	key := "my-idempotency-key"
	headers := app.prepareHeaders(&key)
	assert.Equal(t, "Bearer fc-test-key", headers["Authorization"])
	assert.Equal(t, "application/json", headers["Content-Type"])
	assert.Equal(t, "my-idempotency-key", headers["x-idempotency-key"])
}

func TestPrepareHeaders_WithEmptyIdempotencyKey(t *testing.T) {
	app, err := NewFirecrawlApp("fc-test-key", "https://api.firecrawl.dev")
	require.NoError(t, err)

	key := ""
	headers := app.prepareHeaders(&key)
	assert.Equal(t, "Bearer fc-test-key", headers["Authorization"])
	// Empty key pointer — the key is still included (empty string)
	assert.Equal(t, "", headers["x-idempotency-key"])
}

func TestPrepareHeaders_NilIdempotencyKey(t *testing.T) {
	app, err := NewFirecrawlApp("fc-test-key", "https://api.firecrawl.dev")
	require.NoError(t, err)

	headers := app.prepareHeaders(nil)
	assert.Equal(t, "Bearer fc-test-key", headers["Authorization"])
	assert.Equal(t, "application/json", headers["Content-Type"])
	_, hasKey := headers["x-idempotency-key"]
	assert.False(t, hasKey, "nil idempotency key should not set the header")
}

func TestPrepareHeaders_AuthorizationFormat(t *testing.T) {
	app, err := NewFirecrawlApp("fc-my-secret-key", "https://api.firecrawl.dev")
	require.NoError(t, err)

	headers := app.prepareHeaders(nil)
	assert.Equal(t, "Bearer fc-my-secret-key", headers["Authorization"])
}
