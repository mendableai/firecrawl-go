package firecrawl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// searchRequest is the internal request struct for search operations.
// It is unexported — callers use SearchParams instead.
type searchRequest struct {
	Query             string        `json:"query"`
	Limit             *int          `json:"limit,omitempty"`
	Sources           []string      `json:"sources,omitempty"`
	Categories        []string      `json:"categories,omitempty"`
	TBS               *string       `json:"tbs,omitempty"`
	Location          *string       `json:"location,omitempty"`
	Country           *string       `json:"country,omitempty"`
	Timeout           *int          `json:"timeout,omitempty"`
	IgnoreInvalidURLs *bool         `json:"ignoreInvalidURLs,omitempty"`
	ScrapeOptions     *ScrapeParams `json:"scrapeOptions,omitempty"`
}

// Search performs a web search using the Firecrawl API.
//
// Parameters:
//   - ctx: Context for cancellation and deadlines.
//   - query: The search query string.
//   - params: Optional search parameters. If nil, defaults are used.
//
// Returns:
//   - *SearchResponse: The search results containing web, image, and news results.
//   - error: An error if the search request fails.
func (app *FirecrawlApp) Search(ctx context.Context, query string, params *SearchParams) (*SearchResponse, error) {
	headers := app.prepareHeaders(nil)

	req := searchRequest{Query: query}
	if params != nil {
		req.Limit = params.Limit
		req.Sources = params.Sources
		req.Categories = params.Categories
		req.TBS = params.TBS
		req.Location = params.Location
		req.Country = params.Country
		req.Timeout = params.Timeout
		req.IgnoreInvalidURLs = params.IgnoreInvalidURLs
		req.ScrapeOptions = params.ScrapeOptions
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	resp, err := app.makeRequest(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/v2/search", app.APIURL),
		body,
		headers,
		"search",
	)
	if err != nil {
		return nil, err
	}

	var searchResponse SearchResponse
	if err := json.Unmarshal(resp, &searchResponse); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	if !searchResponse.Success {
		return nil, fmt.Errorf("search operation failed")
	}

	return &searchResponse, nil
}
