package firecrawl

import (
	"fmt"
	"net/http"
	"testing"
	"time"

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

// IMP-15: HTTP Client Improvements tests

func TestSDKVersion_NotEmpty(t *testing.T) {
	assert.NotEmpty(t, SDKVersion, "SDKVersion constant must not be empty")
}

func TestDefaultUserAgent(t *testing.T) {
	app, err := NewFirecrawlApp("fc-test-key", "https://api.example.com")
	require.NoError(t, err)

	headers := app.prepareHeaders(nil)
	expectedUA := fmt.Sprintf("firecrawl-go/%s", SDKVersion)
	assert.Equal(t, expectedUA, headers["User-Agent"], "default User-Agent should be firecrawl-go/{version}")
}

func TestDefaultUserAgent_WithOptions(t *testing.T) {
	app, err := NewFirecrawlAppWithOptions("fc-test-key", "https://api.example.com")
	require.NoError(t, err)

	headers := app.prepareHeaders(nil)
	expectedUA := fmt.Sprintf("firecrawl-go/%s", SDKVersion)
	assert.Equal(t, expectedUA, headers["User-Agent"], "default User-Agent via WithOptions should be firecrawl-go/{version}")
}

func TestCustomUserAgent(t *testing.T) {
	app, err := NewFirecrawlAppWithOptions(
		"fc-test-key",
		"https://api.example.com",
		WithUserAgent("my-custom-agent/1.0"),
	)
	require.NoError(t, err)

	headers := app.prepareHeaders(nil)
	assert.Equal(t, "my-custom-agent/1.0", headers["User-Agent"])
}

func TestWithTimeout(t *testing.T) {
	wantTimeout := 30 * time.Second
	app, err := NewFirecrawlAppWithOptions(
		"fc-test-key",
		"https://api.example.com",
		WithTimeout(wantTimeout),
	)
	require.NoError(t, err)
	assert.Equal(t, wantTimeout, app.Client.Timeout)
}

func TestWithTransport(t *testing.T) {
	customTransport := &http.Transport{
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 5,
	}

	app, err := NewFirecrawlAppWithOptions(
		"fc-test-key",
		"https://api.example.com",
		WithTransport(customTransport),
	)
	require.NoError(t, err)
	assert.Equal(t, customTransport, app.Client.Transport, "custom transport should be used")
}

func TestWithMaxIdleConns(t *testing.T) {
	app, err := NewFirecrawlAppWithOptions(
		"fc-test-key",
		"https://api.example.com",
		WithMaxIdleConns(200),
		WithMaxIdleConnsPerHost(20),
	)
	require.NoError(t, err)

	transport, ok := app.Client.Transport.(*http.Transport)
	require.True(t, ok, "transport should be *http.Transport when no custom transport is set")
	assert.Equal(t, 200, transport.MaxIdleConns)
	assert.Equal(t, 20, transport.MaxIdleConnsPerHost)
}

func TestDefaultTransportCloned(t *testing.T) {
	app, err := NewFirecrawlApp("fc-test-key", "https://api.example.com")
	require.NoError(t, err)

	// The transport must NOT be the same pointer as http.DefaultTransport —
	// it should be a cloned copy so SDK settings don't bleed into the process.
	assert.NotEqual(t, http.DefaultTransport, app.Client.Transport,
		"transport should be a clone of http.DefaultTransport, not the same pointer")
}

func TestBackwardCompatibility_NoTimeout(t *testing.T) {
	app, err := NewFirecrawlApp("fc-test-key", "https://api.example.com")
	require.NoError(t, err)
	assert.Equal(t, 120*time.Second, app.Client.Timeout, "default timeout should be 120s")
}

func TestBackwardCompatibility_WithTimeout(t *testing.T) {
	app, err := NewFirecrawlApp("fc-test-key", "https://api.example.com", 30*time.Second)
	require.NoError(t, err)
	assert.Equal(t, 30*time.Second, app.Client.Timeout, "variadic timeout parameter should still work")
}

func TestNewFirecrawlAppWithOptions_EmptyKey(t *testing.T) {
	t.Setenv("FIRECRAWL_API_KEY", "")
	_, err := NewFirecrawlAppWithOptions("", "https://api.example.com")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNoAPIKey)
}

func TestNewFirecrawlAppWithOptions_DefaultURL(t *testing.T) {
	t.Setenv("FIRECRAWL_API_URL", "")
	app, err := NewFirecrawlAppWithOptions("fc-test-key", "")
	require.NoError(t, err)
	assert.Equal(t, "https://api.firecrawl.dev", app.APIURL)
}

func TestVersionFieldSet(t *testing.T) {
	app, err := NewFirecrawlApp("fc-test-key", "https://api.example.com")
	require.NoError(t, err)
	assert.Equal(t, SDKVersion, app.Version, "Version field should be set to SDKVersion")
}
