package firecrawl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// scrapeRequest is the internal request struct for scrape operations.
// It is unexported — callers use ScrapeParams instead.
type scrapeRequest struct {
	URL                 string             `json:"url"`
	Formats             []string           `json:"formats,omitempty"`
	Headers             *map[string]string `json:"headers,omitempty"`
	IncludeTags         []string           `json:"includeTags,omitempty"`
	ExcludeTags         []string           `json:"excludeTags,omitempty"`
	OnlyMainContent     *bool              `json:"onlyMainContent,omitempty"`
	WaitFor             *int               `json:"waitFor,omitempty"`
	Timeout             *int               `json:"timeout,omitempty"`
	MaxAge              *int               `json:"maxAge,omitempty"`
	MinAge              *int               `json:"minAge,omitempty"`
	JsonOptions         *JsonOptions       `json:"jsonOptions,omitempty"`
	Mobile              *bool              `json:"mobile,omitempty"`
	SkipTlsVerification *bool              `json:"skipTlsVerification,omitempty"`
	BlockAds            *bool              `json:"blockAds,omitempty"`
	Proxy               *string            `json:"proxy,omitempty"`
	Location            *LocationConfig    `json:"location,omitempty"`
	Parsers             []ParserConfig     `json:"parsers,omitempty"`
	Actions             []ActionConfig     `json:"actions,omitempty"`
	RemoveBase64Images  *bool              `json:"removeBase64Images,omitempty"`
	StoreInCache        *bool              `json:"storeInCache,omitempty"`
	ZeroDataRetention   *bool              `json:"zeroDataRetention,omitempty"`
}

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

	req := scrapeRequest{URL: url}
	if params != nil {
		req.Formats = params.Formats
		req.Headers = params.Headers
		req.IncludeTags = params.IncludeTags
		req.ExcludeTags = params.ExcludeTags
		req.OnlyMainContent = params.OnlyMainContent
		req.WaitFor = params.WaitFor
		req.Timeout = params.Timeout
		req.MaxAge = params.MaxAge
		req.MinAge = params.MinAge
		req.JsonOptions = params.JsonOptions
		req.Mobile = params.Mobile
		req.SkipTlsVerification = params.SkipTlsVerification
		req.BlockAds = params.BlockAds
		req.Proxy = params.Proxy
		req.Location = params.Location
		req.Parsers = params.Parsers
		req.Actions = params.Actions
		req.RemoveBase64Images = params.RemoveBase64Images
		req.StoreInCache = params.StoreInCache
		req.ZeroDataRetention = params.ZeroDataRetention
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scrape request: %w", err)
	}

	resp, err := app.makeRequest(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/v2/scrape", app.APIURL),
		body,
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
