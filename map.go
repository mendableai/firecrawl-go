package firecrawl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// mapRequest is the internal request body for the v2 /map endpoint.
type mapRequest struct {
	URL                   string          `json:"url"`
	IncludeSubdomains     *bool           `json:"includeSubdomains,omitempty"`
	Search                *string         `json:"search,omitempty"`
	Limit                 *int            `json:"limit,omitempty"`
	Sitemap               *string         `json:"sitemap,omitempty"`
	IgnoreQueryParameters *bool           `json:"ignoreQueryParameters,omitempty"`
	IgnoreCache           *bool           `json:"ignoreCache,omitempty"`
	Timeout               *int            `json:"timeout,omitempty"`
	Location              *LocationConfig `json:"location,omitempty"`
}

// MapURL initiates a mapping operation for a URL using the Firecrawl API.
//
// Parameters:
//   - ctx: Context for cancellation and deadlines.
//   - url: The URL to map.
//   - params: Optional parameters for the mapping request.
//
// Returns:
//   - *MapResponse: The response from the mapping operation, with Links as []MapLink.
//   - error: An error if the mapping request fails.
func (app *FirecrawlApp) MapURL(ctx context.Context, url string, params *MapParams) (*MapResponse, error) {
	headers := app.prepareHeaders(nil)

	req := mapRequest{URL: url}
	if params != nil {
		req.IncludeSubdomains = params.IncludeSubdomains
		req.Search = params.Search
		req.Limit = params.Limit
		req.Sitemap = params.Sitemap
		req.IgnoreQueryParameters = params.IgnoreQueryParameters
		req.IgnoreCache = params.IgnoreCache
		req.Timeout = params.Timeout
		req.Location = params.Location
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal map request: %w", err)
	}

	resp, err := app.makeRequest(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/v2/map", app.APIURL),
		body,
		headers,
		"map",
	)
	if err != nil {
		return nil, err
	}

	var mapResponse MapResponse
	err = json.Unmarshal(resp, &mapResponse)
	if err != nil {
		return nil, err
	}

	if mapResponse.Success {
		return &mapResponse, nil
	}
	return nil, fmt.Errorf("map operation failed: %s", mapResponse.Error)
}
