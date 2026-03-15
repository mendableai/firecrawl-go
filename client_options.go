package firecrawl

import (
	"net/http"
	"time"
)

// SDKVersion is the current version of the firecrawl-go SDK.
const SDKVersion = "2.0.0"

// clientConfig holds the configuration for building the FirecrawlApp HTTP client.
type clientConfig struct {
	timeout             time.Duration
	transport           *http.Transport
	userAgent           string
	maxIdleConns        int
	maxIdleConnsPerHost int
}

// defaultClientConfig returns sensible defaults for the HTTP client configuration.
func defaultClientConfig() *clientConfig {
	return &clientConfig{
		timeout:             120 * time.Second,
		userAgent:           "firecrawl-go/" + SDKVersion,
		maxIdleConns:        100,
		maxIdleConnsPerHost: 10,
	}
}

// ClientOption configures the FirecrawlApp HTTP client.
type ClientOption func(*clientConfig)

// WithTimeout sets the HTTP client timeout.
//
// This is the recommended alternative to passing a time.Duration as the variadic
// argument to NewFirecrawlApp. Default: 120 seconds.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *clientConfig) {
		c.timeout = d
	}
}

// WithTransport sets a custom http.Transport for the HTTP client.
// When set, WithMaxIdleConns and WithMaxIdleConnsPerHost are ignored —
// configure those directly on the transport you provide.
func WithTransport(t *http.Transport) ClientOption {
	return func(c *clientConfig) {
		c.transport = t
	}
}

// WithUserAgent sets a custom User-Agent header sent with every request.
// Default: "firecrawl-go/{version}".
func WithUserAgent(ua string) ClientOption {
	return func(c *clientConfig) {
		c.userAgent = ua
	}
}

// WithMaxIdleConns sets the maximum number of idle (keep-alive) connections
// across all hosts. Only applies when no custom Transport is provided. Default: 100.
func WithMaxIdleConns(n int) ClientOption {
	return func(c *clientConfig) {
		c.maxIdleConns = n
	}
}

// WithMaxIdleConnsPerHost sets the maximum number of idle (keep-alive) connections
// per host. Only applies when no custom Transport is provided. Default: 10.
func WithMaxIdleConnsPerHost(n int) ClientOption {
	return func(c *clientConfig) {
		c.maxIdleConnsPerHost = n
	}
}
