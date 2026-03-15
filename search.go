package firecrawl

import (
	"context"
	"fmt"
)

// Search searches for a URL using the Firecrawl API.
//
// Parameters:
//   - ctx: Context for cancellation and deadlines.
//   - query: The search query.
//   - params: Optional parameters for the search request.
//
// Returns:
//   - any: The search results (not yet implemented).
//   - error: An error if the search request fails.
//
// Search is not implemented in API version 1.0.0.
func (app *FirecrawlApp) Search(ctx context.Context, query string, params *any) (any, error) {
	return nil, fmt.Errorf("Search is not implemented in API version 1.0.0")
}
