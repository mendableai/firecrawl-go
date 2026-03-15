package firecrawl

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

// FirecrawlApp represents a client for the Firecrawl API.
type FirecrawlApp struct {
	apiKey    string // unexported — use APIKey() accessor
	APIURL    string
	Client    *http.Client
	Version   string
	userAgent string // set by constructor; sent as User-Agent header on every request
}

// APIKey returns the configured API key.
func (app *FirecrawlApp) APIKey() string {
	return app.apiKey
}

// String returns a human-readable representation with the API key redacted.
func (app *FirecrawlApp) String() string {
	redacted := "***"
	if len(app.apiKey) > 7 {
		redacted = app.apiKey[:3] + "..." + app.apiKey[len(app.apiKey)-4:]
	}
	return fmt.Sprintf("FirecrawlApp{url: %s, key: %s}", app.APIURL, redacted)
}

// NewFirecrawlApp creates a new instance of FirecrawlApp with the provided API key and API URL.
// If the API key or API URL is not provided, it attempts to retrieve them from environment variables.
// If the API key is still not found, it returns an error.
//
// Parameters:
//   - apiKey: The API key for authenticating with the Firecrawl API. If empty, it will be retrieved from the FIRECRAWL_API_KEY environment variable.
//   - apiURL: The base URL for the Firecrawl API. If empty, it will be retrieved from the FIRECRAWL_API_URL environment variable, defaulting to "https://api.firecrawl.dev".
//   - timeout: The timeout for the HTTP client. If not provided, it will default to 60 seconds.
//
// Returns:
//   - *FirecrawlApp: A new instance of FirecrawlApp configured with the provided or retrieved API key and API URL.
//   - error: An error if the API key is not provided or retrieved.
func NewFirecrawlApp(apiKey, apiURL string, timeout ...time.Duration) (*FirecrawlApp, error) {
	if apiKey == "" {
		apiKey = os.Getenv("FIRECRAWL_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("%w", ErrNoAPIKey)
		}
	}

	if apiURL == "" {
		apiURL = os.Getenv("FIRECRAWL_API_URL")
		if apiURL == "" {
			apiURL = "https://api.firecrawl.dev"
		}
	}

	// Warn when a non-localhost HTTP URL is used — API key will be sent in cleartext.
	parsedURL, err := url.Parse(apiURL)
	if err == nil && parsedURL.Scheme == "http" {
		host := parsedURL.Hostname()
		if host != "localhost" && host != "127.0.0.1" && host != "::1" {
			log.Println("WARNING: firecrawl-go: API URL uses HTTP. API key will be transmitted in cleartext. Use HTTPS in production.")
		}
	}

	cfg := defaultClientConfig()
	if len(timeout) > 0 {
		cfg.timeout = timeout[0]
	}

	return newFirecrawlAppFromConfig(apiKey, apiURL, cfg)
}

// NewFirecrawlAppWithOptions creates a new instance of FirecrawlApp using
// functional options for configuration.
//
// Parameters:
//   - apiKey: The API key for authenticating with the Firecrawl API. If empty, it will be retrieved from the FIRECRAWL_API_KEY environment variable.
//   - apiURL: The base URL for the Firecrawl API. If empty, it will be retrieved from the FIRECRAWL_API_URL environment variable, defaulting to "https://api.firecrawl.dev".
//   - opts: Functional options (WithTimeout, WithTransport, WithUserAgent, WithMaxIdleConns, WithMaxIdleConnsPerHost).
//
// Returns:
//   - *FirecrawlApp: A new instance of FirecrawlApp configured with the provided or retrieved API key, API URL, and options.
//   - error: An error if the API key is not provided or retrieved.
func NewFirecrawlAppWithOptions(apiKey, apiURL string, opts ...ClientOption) (*FirecrawlApp, error) {
	if apiKey == "" {
		apiKey = os.Getenv("FIRECRAWL_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("%w", ErrNoAPIKey)
		}
	}

	if apiURL == "" {
		apiURL = os.Getenv("FIRECRAWL_API_URL")
		if apiURL == "" {
			apiURL = "https://api.firecrawl.dev"
		}
	}

	// Warn when a non-localhost HTTP URL is used — API key will be sent in cleartext.
	parsedURL, err := url.Parse(apiURL)
	if err == nil && parsedURL.Scheme == "http" {
		host := parsedURL.Hostname()
		if host != "localhost" && host != "127.0.0.1" && host != "::1" {
			log.Println("WARNING: firecrawl-go: API URL uses HTTP. API key will be transmitted in cleartext. Use HTTPS in production.")
		}
	}

	cfg := defaultClientConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	return newFirecrawlAppFromConfig(apiKey, apiURL, cfg)
}

// newFirecrawlAppFromConfig builds a FirecrawlApp from a resolved clientConfig.
// apiKey and apiURL must already be validated and resolved before calling this.
func newFirecrawlAppFromConfig(apiKey, apiURL string, cfg *clientConfig) (*FirecrawlApp, error) {
	var transport http.RoundTripper
	if cfg.transport != nil {
		transport = cfg.transport
	} else {
		defaultT, ok := http.DefaultTransport.(*http.Transport)
		if !ok {
			return nil, fmt.Errorf("firecrawl-go: http.DefaultTransport is not *http.Transport; use WithTransport to supply a custom transport")
		}
		cloned := defaultT.Clone()
		cloned.MaxIdleConns = cfg.maxIdleConns
		cloned.MaxIdleConnsPerHost = cfg.maxIdleConnsPerHost
		transport = cloned
	}

	client := &http.Client{
		Timeout:   cfg.timeout,
		Transport: transport,
	}

	return &FirecrawlApp{
		apiKey:    apiKey,
		APIURL:    apiURL,
		Client:    client,
		Version:   SDKVersion,
		userAgent: cfg.userAgent,
	}, nil
}

// prepareHeaders prepares the headers for an HTTP request.
//
// Parameters:
//   - idempotencyKey: A string representing the idempotency key to be included in the headers.
//     If the idempotency key is an empty string, it will not be included in the headers.
//
// Returns:
//   - map[string]string: A map containing the headers for the HTTP request.
func (app *FirecrawlApp) prepareHeaders(idempotencyKey *string) map[string]string {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", app.apiKey),
		"User-Agent":    app.userAgent,
	}
	if idempotencyKey != nil {
		headers["x-idempotency-key"] = *idempotencyKey
	}
	return headers
}
