package firecrawl

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ScrapeURL scrapes the content of the specified URL using the Firecrawl API.
//
// Parameters:
//   - url: The URL to be scraped.
//   - params: Optional parameters for the scrape request, including extractor options for LLM extraction.
//
// Returns:
//   - *FirecrawlDocument or *FirecrawlDocumentV0: The scraped document data depending on the API version.
//   - error: An error if the scrape request fails.
func (app *FirecrawlApp) ScrapeURL(url string, params *ScrapeParams) (*FirecrawlDocument, error) {
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
		if params.ParsePDF != nil {
			scrapeBody["parsePDF"] = params.ParsePDF
		}
		if params.Timeout != nil {
			scrapeBody["timeout"] = params.Timeout
		}
		if params.MaxAge != nil {
			scrapeBody["maxAge"] = params.MaxAge
		}
		if params.JsonOptions != nil {
			scrapeBody["jsonOptions"] = params.JsonOptions
		}
	}

	resp, err := app.makeRequest(
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
