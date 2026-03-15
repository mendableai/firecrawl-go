package firecrawl

import "fmt"

// SearchURL searches for a URL using the Firecrawl API.
//
// Parameters:
//   - url: The URL to search for.
//   - params: Optional parameters for the search request.
//   - error: An error if the search request fails.
//
// Search is not implemented in API version 1.0.0.
func (app *FirecrawlApp) Search(query string, params *any) (any, error) {
	return nil, fmt.Errorf("Search is not implemented in API version 1.0.0")
}
