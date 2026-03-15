package firecrawl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ScrapeURL scrapes the content of the specified URL using the Firecrawl API.
//
// Parameters:
//   - ctx: Context for cancellation and deadlines.
//   - url: The URL to be scraped.
//   - params: Optional parameters for the scrape request, including formats, actions, location, and LLM extraction options.
//
// Returns:
//   - *FirecrawlDocument: The scraped document data.
//   - error: An error if the scrape request fails.
func (app *FirecrawlApp) ScrapeURL(ctx context.Context, url string, params *ScrapeParams) (*FirecrawlDocument, error) {
	headers := app.prepareHeaders(nil)
	scrapeBody := map[string]any{"url": url}

	if params != nil {
		if params.Formats != nil {
			scrapeBody["formats"] = params.Formats
		}
		if params.Headers != nil {
			scrapeBody["headers"] = params.Headers
		}
		if params.IncludeTags != nil {
			scrapeBody["includeTags"] = params.IncludeTags
		}
		if params.ExcludeTags != nil {
			scrapeBody["excludeTags"] = params.ExcludeTags
		}
		if params.OnlyMainContent != nil {
			scrapeBody["onlyMainContent"] = params.OnlyMainContent
		}
		if params.WaitFor != nil {
			scrapeBody["waitFor"] = params.WaitFor
		}
		if params.Timeout != nil {
			scrapeBody["timeout"] = params.Timeout
		}
		if params.MaxAge != nil {
			scrapeBody["maxAge"] = params.MaxAge
		}
		if params.MinAge != nil {
			scrapeBody["minAge"] = params.MinAge
		}
		if params.JsonOptions != nil {
			scrapeBody["jsonOptions"] = params.JsonOptions
		}
		if params.Mobile != nil {
			scrapeBody["mobile"] = params.Mobile
		}
		if params.SkipTlsVerification != nil {
			scrapeBody["skipTlsVerification"] = params.SkipTlsVerification
		}
		if params.BlockAds != nil {
			scrapeBody["blockAds"] = params.BlockAds
		}
		if params.Proxy != nil {
			scrapeBody["proxy"] = params.Proxy
		}
		if params.Location != nil {
			scrapeBody["location"] = params.Location
		}
		if params.Parsers != nil {
			scrapeBody["parsers"] = params.Parsers
		}
		if params.Actions != nil {
			scrapeBody["actions"] = params.Actions
		}
		if params.RemoveBase64Images != nil {
			scrapeBody["removeBase64Images"] = params.RemoveBase64Images
		}
		if params.StoreInCache != nil {
			scrapeBody["storeInCache"] = params.StoreInCache
		}
		if params.ZeroDataRetention != nil {
			scrapeBody["zeroDataRetention"] = params.ZeroDataRetention
		}
	}

	resp, err := app.makeRequest(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/v1/scrape", app.APIURL),
		scrapeBody,
		headers,
		"scrape URL",
	)
	if err != nil {
		return nil, err
	}

	var scrapeResponse ScrapeResponse
	if err := json.Unmarshal(resp, &scrapeResponse); err != nil {
		return nil, fmt.Errorf("failed to parse scrape response: %w", err)
	}

	if !scrapeResponse.Success {
		return nil, fmt.Errorf("failed to scrape URL")
	}

	return scrapeResponse.Data, nil
}
