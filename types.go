package firecrawl

import (
	"encoding/json"
	"fmt"
)

type StringOrStringSlice []string

func (s *StringOrStringSlice) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*s = []string{single}
		return nil
	}

	var list []string
	if err := json.Unmarshal(data, &list); err == nil {
		*s = list
		return nil
	}

	return fmt.Errorf("field is neither a string nor a list of strings")
}

// FirecrawlDocumentMetadata represents metadata for a Firecrawl document
type FirecrawlDocumentMetadata struct {
	Title             *string              `json:"title,omitempty"`
	Description       *StringOrStringSlice `json:"description,omitempty"`
	Language          *StringOrStringSlice `json:"language,omitempty"`
	Keywords          *StringOrStringSlice `json:"keywords,omitempty"`
	Robots            *StringOrStringSlice `json:"robots,omitempty"`
	OGTitle           *StringOrStringSlice `json:"ogTitle,omitempty"`
	OGDescription     *StringOrStringSlice `json:"ogDescription,omitempty"`
	OGURL             *StringOrStringSlice `json:"ogUrl,omitempty"`
	OGImage           *StringOrStringSlice `json:"ogImage,omitempty"`
	OGAudio           *StringOrStringSlice `json:"ogAudio,omitempty"`
	OGDeterminer      *StringOrStringSlice `json:"ogDeterminer,omitempty"`
	OGLocale          *StringOrStringSlice `json:"ogLocale,omitempty"`
	OGLocaleAlternate []*string            `json:"ogLocaleAlternate,omitempty"`
	OGSiteName        *StringOrStringSlice `json:"ogSiteName,omitempty"`
	OGVideo           *StringOrStringSlice `json:"ogVideo,omitempty"`
	DCTermsCreated    *StringOrStringSlice `json:"dctermsCreated,omitempty"`
	DCDateCreated     *StringOrStringSlice `json:"dcDateCreated,omitempty"`
	DCDate            *StringOrStringSlice `json:"dcDate,omitempty"`
	DCTermsType       *StringOrStringSlice `json:"dctermsType,omitempty"`
	DCType            *StringOrStringSlice `json:"dcType,omitempty"`
	DCTermsAudience   *StringOrStringSlice `json:"dctermsAudience,omitempty"`
	DCTermsSubject    *StringOrStringSlice `json:"dctermsSubject,omitempty"`
	DCSubject         *StringOrStringSlice `json:"dcSubject,omitempty"`
	DCDescription     *StringOrStringSlice `json:"dcDescription,omitempty"`
	DCTermsKeywords   *StringOrStringSlice `json:"dctermsKeywords,omitempty"`
	ModifiedTime      *StringOrStringSlice `json:"modifiedTime,omitempty"`
	PublishedTime     *StringOrStringSlice `json:"publishedTime,omitempty"`
	ArticleTag        *StringOrStringSlice `json:"articleTag,omitempty"`
	ArticleSection    *StringOrStringSlice `json:"articleSection,omitempty"`
	URL               *string              `json:"url,omitempty"`
	ScrapeID          *string              `json:"scrapeId,omitempty"`
	SourceURL         *string              `json:"sourceURL,omitempty"`
	StatusCode        *int                 `json:"statusCode,omitempty"`
	Error             *string              `json:"error,omitempty"`
}

// JsonOptions represents the options for JSON extraction
type JsonOptions struct {
	Schema       map[string]any `json:"schema,omitempty"`
	SystemPrompt *string        `json:"systemPrompt,omitempty"`
	Prompt       *string        `json:"prompt,omitempty"`
}

// FirecrawlDocument represents a document in Firecrawl
type FirecrawlDocument struct {
	Markdown   string                     `json:"markdown,omitempty"`
	HTML       string                     `json:"html,omitempty"`
	RawHTML    string                     `json:"rawHtml,omitempty"`
	Screenshot string                     `json:"screenshot,omitempty"`
	JSON       map[string]any             `json:"json,omitempty"`
	Links      []string                   `json:"links,omitempty"`
	Metadata   *FirecrawlDocumentMetadata `json:"metadata,omitempty"`
}

// ScrapeParams represents the parameters for a scrape request.
type ScrapeParams struct {
	Formats         []string           `json:"formats,omitempty"`
	Headers         *map[string]string `json:"headers,omitempty"`
	IncludeTags     []string           `json:"includeTags,omitempty"`
	ExcludeTags     []string           `json:"excludeTags,omitempty"`
	OnlyMainContent *bool              `json:"onlyMainContent,omitempty"`
	WaitFor         *int               `json:"waitFor,omitempty"`
	ParsePDF        *bool              `json:"parsePDF,omitempty"`
	Timeout         *int               `json:"timeout,omitempty"`
	MaxAge          *int               `json:"maxAge,omitempty"`
	JsonOptions     *JsonOptions       `json:"jsonOptions,omitempty"`
}

// ScrapeResponse represents the response for scraping operations
type ScrapeResponse struct {
	Success bool               `json:"success"`
	Data    *FirecrawlDocument `json:"data,omitempty"`
}

// CrawlParams represents the parameters for a crawl request.
type CrawlParams struct {
	ScrapeOptions         ScrapeParams `json:"scrapeOptions"`
	Webhook               *string      `json:"webhook,omitempty"`
	Limit                 *int         `json:"limit,omitempty"`
	IncludePaths          []string     `json:"includePaths,omitempty"`
	ExcludePaths          []string     `json:"excludePaths,omitempty"`
	MaxDepth              *int         `json:"maxDepth,omitempty"`
	AllowBackwardLinks    *bool        `json:"allowBackwardLinks,omitempty"`
	AllowExternalLinks    *bool        `json:"allowExternalLinks,omitempty"`
	IgnoreSitemap         *bool        `json:"ignoreSitemap,omitempty"`
	IgnoreQueryParameters *bool        `json:"ignoreQueryParameters,omitempty"`
}

// CrawlResponse represents the response for crawling operations
type CrawlResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	URL     string `json:"url,omitempty"`
}

// CrawlStatusResponse (old JobStatusResponse) represents the response for checking crawl job
type CrawlStatusResponse struct {
	Status      string               `json:"status"`
	Total       int                  `json:"total,omitempty"`
	Completed   int                  `json:"completed,omitempty"`
	CreditsUsed int                  `json:"creditsUsed,omitempty"`
	ExpiresAt   string               `json:"expiresAt,omitempty"`
	Next        *string              `json:"next,omitempty"`
	Data        []*FirecrawlDocument `json:"data,omitempty"`
}

// CancelCrawlJobResponse represents the response for canceling a crawl job
type CancelCrawlJobResponse struct {
	Success bool   `json:"success"`
	Status  string `json:"status"`
}

// MapParams represents the parameters for a map request.
type MapParams struct {
	IncludeSubdomains *bool   `json:"includeSubdomains,omitempty"`
	Search            *string `json:"search,omitempty"`
	IgnoreSitemap     *bool   `json:"ignoreSitemap,omitempty"`
	Limit             *int    `json:"limit,omitempty"`
}

// MapResponse represents the response for mapping operations
type MapResponse struct {
	Success bool     `json:"success"`
	Links   []string `json:"links,omitempty"`
	Error   string   `json:"error,omitempty"`
}
