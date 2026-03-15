package firecrawl

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// MapURL initiates a mapping operation for a URL using the Firecrawl API.
//
// Parameters:
//   - url: The URL to map.
//   - params: Optional parameters for the mapping request.
//
// Returns:
//   - *MapResponse: The response from the mapping operation.
//   - error: An error if the mapping request fails.
func (app *FirecrawlApp) MapURL(url string, params *MapParams) (*MapResponse, error) {
	headers := app.prepareHeaders(nil)
	jsonData := map[string]any{"url": url}

	if params != nil {
		if params.IncludeSubdomains != nil {
			jsonData["includeSubdomains"] = params.IncludeSubdomains
		}
		if params.Search != nil {
			jsonData["search"] = params.Search
		}
		if params.IgnoreSitemap != nil {
			jsonData["ignoreSitemap"] = params.IgnoreSitemap
		}
		if params.Limit != nil {
			jsonData["limit"] = params.Limit
		}
	}

	resp, err := app.makeRequest(
		http.MethodPost,
		fmt.Sprintf("%s/v1/map", app.APIURL),
		jsonData,
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
	} else {
		return nil, fmt.Errorf("map operation failed: %s", mapResponse.Error)
	}
}
