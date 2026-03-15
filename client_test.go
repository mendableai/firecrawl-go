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
