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
	apiKey  string // unexported — use APIKey() accessor
	APIURL  string
	Client  *http.Client
	Version string
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

	t := 120 * time.Second // default
	if len(timeout) > 0 {
		t = timeout[0]
	}

	client := &http.Client{
		Timeout:   t,
		Transport: http.DefaultTransport,
	}

	return &FirecrawlApp{
		apiKey: apiKey,
		APIURL: apiURL,
		Client: client,
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
	}
	if idempotencyKey != nil {
		headers["x-idempotency-key"] = *idempotencyKey
	}
	return headers
}
