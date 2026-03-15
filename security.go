package firecrawl

import (
	"fmt"
	"net/url"

	"github.com/google/uuid"
)

// validatePaginationURL ensures a Next pagination URL points to the same host
// as the SDK's configured API URL, preventing SSRF attacks via malicious Next
// URLs in API responses.
func validatePaginationURL(baseURL, nextURL string) error {
	base, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	next, err := url.Parse(nextURL)
	if err != nil {
		return fmt.Errorf("invalid pagination URL: %w", err)
	}

	if next.Host != base.Host {
		return fmt.Errorf("pagination URL host %q does not match API host %q", next.Host, base.Host)
	}

	return nil
}

// validateJobID ensures a job ID is a valid UUID, preventing path injection
// attacks via crafted IDs like "../../admin".
func validateJobID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid job ID %q: must be a valid UUID: %w", id, err)
	}
	return nil
}
