package firecrawl

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var API_URL string
var TEST_API_KEY string

func init() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	API_URL = os.Getenv("API_URL")
	TEST_API_KEY = os.Getenv("TEST_API_KEY")
}

func TestNoAPIKey(t *testing.T) {
	_, err := NewFirecrawlApp("", API_URL, "v1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no API key provided")
}

func TestScrapeURLInvalidAPIKey(t *testing.T) {
	app, err := NewFirecrawlApp("invalid_api_key", API_URL, "v1")
	require.NoError(t, err)

	_, err = app.ScrapeURL("https://firecrawl.dev", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Unexpected error during scrape URL: Status code 401. Unauthorized: Invalid token")
}

func TestBlocklistedURL(t *testing.T) {
	app, err := NewFirecrawlApp(TEST_API_KEY, API_URL, "v1")
	require.NoError(t, err)

	_, err = app.ScrapeURL("https://facebook.com/fake-test", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "URL is blocked. Firecrawl currently does not support social media scraping due to policy restrictions.")
}

func TestSuccessfulResponseWithValidPreviewToken(t *testing.T) {
	app, err := NewFirecrawlApp("this_is_just_a_preview_token", API_URL, "v1")
	require.NoError(t, err)

	response, err := app.ScrapeURL("https://roastmywebsite.ai", nil)
	require.NoError(t, err)
	assert.NotNil(t, response)

	scrapeResponse := response.(*FirecrawlDocument)
	assert.Contains(t, scrapeResponse.Markdown, "_Roast_")
}

func TestScrapeURLE2E(t *testing.T) {
	app, err := NewFirecrawlApp(TEST_API_KEY, API_URL, "v1")
	require.NoError(t, err)

	response, err := app.ScrapeURL("https://roastmywebsite.ai", nil)
	require.NoError(t, err)
	assert.NotNil(t, response)

	scrapeResponse := response.(*FirecrawlDocument)
	assert.Contains(t, scrapeResponse.Markdown, "_Roast_")
	assert.NotEqual(t, scrapeResponse.Markdown, "")
	assert.NotNil(t, scrapeResponse.Metadata)
	assert.Equal(t, scrapeResponse.HTML, "")
}

func TestSuccessfulResponseWithValidAPIKeyAndIncludeHTML(t *testing.T) {
	app, err := NewFirecrawlApp(TEST_API_KEY, API_URL, "v1")
	require.NoError(t, err)

	params := map[string]any{
		"formats":         []string{"markdown", "html", "rawHtml", "screenshot", "links"},
		"headers":         map[string]string{"x-key": "test"},
		"includeTags":     []string{"h1"},
		"excludeTags":     []string{"h2"},
		"onlyMainContent": true,
		"timeout":         30000,
		"waitFor":         1000,
	}

	response, err := app.ScrapeURL("https://roastmywebsite.ai", params)
	require.NoError(t, err)
	assert.NotNil(t, response)

	scrapeResponse := response.(*FirecrawlDocument)
	assert.NotNil(t, scrapeResponse)
	assert.Contains(t, scrapeResponse.Markdown, "_Roast_")
	assert.Contains(t, scrapeResponse.HTML, "<h1")
	assert.Contains(t, scrapeResponse.RawHTML, "<h1")
	assert.NotNil(t, scrapeResponse.Screenshot)
	assert.NotEmpty(t, scrapeResponse.Screenshot)
	assert.Contains(t, scrapeResponse.Screenshot, "https://")
	assert.NotNil(t, scrapeResponse.Links)
	assert.Greater(t, len(scrapeResponse.Links), 0)
	assert.Contains(t, scrapeResponse.Links[0], "https://")
	assert.NotNil(t, scrapeResponse.Metadata)
}

func TestSuccessfulResponseForValidScrapeWithPDFFile(t *testing.T) {
	app, err := NewFirecrawlApp(TEST_API_KEY, API_URL, "v1")
	require.NoError(t, err)

	response, err := app.ScrapeURL("https://arxiv.org/pdf/astro-ph/9301001.pdf", nil)
	require.NoError(t, err)
	assert.NotNil(t, response)

	scrapeResponse := response.(*FirecrawlDocument)

	assert.Contains(t, scrapeResponse.Markdown, "We present spectrophotometric observations of the Broad Line Radio Galaxy")
	assert.NotNil(t, scrapeResponse.Metadata)
}

func TestSuccessfulResponseForValidScrapeWithPDFFileWithoutExplicitExtension(t *testing.T) {
	app, err := NewFirecrawlApp(TEST_API_KEY, API_URL, "v1")
	require.NoError(t, err)

	response, err := app.ScrapeURL("https://arxiv.org/pdf/astro-ph/9301001", nil)
	time.Sleep(6 * time.Second) // wait for 6 seconds
	require.NoError(t, err)
	assert.NotNil(t, response)

	scrapeResponse := response.(*FirecrawlDocument)

	assert.Contains(t, scrapeResponse.Markdown, "We present spectrophotometric observations of the Broad Line Radio Galaxy")
	assert.NotNil(t, scrapeResponse.Metadata)
}

func TestCrawlURLInvalidAPIKey(t *testing.T) {
	app, err := NewFirecrawlApp("invalid_api_key", API_URL, "v1")
	require.NoError(t, err)

	_, err = app.CrawlURL("https://firecrawl.dev", nil, false, 2, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Unexpected error during start crawl job: Status code 401. Unauthorized: Invalid token")
}

func TestShouldReturnErrorForBlocklistedURL(t *testing.T) {
	app, err := NewFirecrawlApp(TEST_API_KEY, API_URL, "v1")
	require.NoError(t, err)

	_, err = app.CrawlURL("https://twitter.com/fake-test", nil, false, 2, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "URL is blocked. Firecrawl currently does not support social media scraping due to policy restrictions.")
}

func TestCrawlURLWaitForCompletionE2E(t *testing.T) {
	app, err := NewFirecrawlApp(TEST_API_KEY, API_URL, "v1")
	require.NoError(t, err)

	response, err := app.CrawlURL("https://roastmywebsite.ai", nil, true, 2, "")
	require.NoError(t, err)
	assert.NotNil(t, response)

	crawlResponse := response.(CrawlStatusResponse)

	assert.Greater(t, crawlResponse.TotalCount, 0)
	assert.Greater(t, crawlResponse.CreditsUsed, 0)
	assert.NotEmpty(t, crawlResponse.ExpiresAt)
	assert.Equal(t, crawlResponse.Status, "completed")

	data := crawlResponse.Data
	assert.IsType(t, []*FirecrawlDocument{}, data)

	assert.Greater(t, len(data), 0)
	assert.Contains(t, data[0].Markdown, "_Roast_")
	assert.NotNil(t, data[0].Metadata)
}

func TestCrawlURLWaitForCompletionWithOptionsE2E(t *testing.T) {
	app, err := NewFirecrawlApp(TEST_API_KEY, API_URL, "v1")
	require.NoError(t, err)

	response, err := app.CrawlURL("https://roastmywebsite.ai",
		map[string]any{
			"excludePaths":       []string{"blog/*"},
			"includePaths":       []string{"/"},
			"maxDepth":           2,
			"ignoreSitemap":      true,
			"limit":              10,
			"allowBackwardLinks": true,
			"allowExternalLinks": true,
			"scrapeOptions": map[string]any{
				"formats":         []string{"markdown", "html", "rawHtml", "screenshot", "links"},
				"headers":         map[string]string{"x-key": "test"},
				"includeTags":     []string{"h1"},
				"excludeTags":     []string{"h2"},
				"onlyMainContent": true,
				"waitFor":         1000,
			}}, true, 2, "")
	require.NoError(t, err)
	assert.NotNil(t, response)

	crawlResponse := response.(CrawlStatusResponse)

	assert.Greater(t, crawlResponse.TotalCount, 0)
	assert.Greater(t, crawlResponse.CreditsUsed, 0)
	assert.NotEmpty(t, crawlResponse.ExpiresAt)
	assert.Equal(t, crawlResponse.Status, "completed")

	data := crawlResponse.Data
	assert.IsType(t, []*FirecrawlDocument{}, data)

	assert.Greater(t, len(data), 0)
	assert.Contains(t, data[0].Markdown, "_Roast_")
	assert.NotNil(t, data[0].Metadata)
	assert.Contains(t, data[0].HTML, "<h1")
	assert.Contains(t, data[0].RawHTML, "<h1")
	assert.Contains(t, data[0].Screenshot, "https://")
	assert.NotNil(t, data[0].Links)
	assert.Greater(t, len(data[0].Links), 0)
	assert.NotNil(t, data[0].Metadata.Title)
	assert.NotNil(t, data[0].Metadata.Description)
	assert.NotNil(t, data[0].Metadata.Language)
	assert.NotNil(t, data[0].Metadata.SourceURL)
	assert.NotNil(t, data[0].Metadata.StatusCode)
	assert.Equal(t, 200, data[0].Metadata.StatusCode)
	assert.Empty(t, data[0].Metadata.Error)
}

func TestCrawlURLWithIdempotencyKeyE2E(t *testing.T) {
	app, err := NewFirecrawlApp(TEST_API_KEY, API_URL, "v1")
	require.NoError(t, err)

	uniqueIdempotencyKey := uuid.New().String()
	params := map[string]any{
		"excludePaths": []string{"blog/*"},
	}
	response, err := app.CrawlURL("https://roastmywebsite.ai", params, true, 2, uniqueIdempotencyKey)
	require.NoError(t, err)
	assert.NotNil(t, response)

	crawlResponse := response.(CrawlStatusResponse)

	data := crawlResponse.Data
	require.Greater(t, len(data), 0)
	require.IsType(t, []*FirecrawlDocument{}, data)
	assert.Contains(t, data[0].Markdown, "_Roast_")

	_, err = app.CrawlURL("https://firecrawl.dev", params, true, 2, uniqueIdempotencyKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Conflict: Failed to start crawl job due to a conflict. Idempotency key already used")
}

func TestCheckCrawlStatusE2E(t *testing.T) {
	app, err := NewFirecrawlApp(TEST_API_KEY, API_URL, "v1")
	require.NoError(t, err)

	params := map[string]any{
		"scrapeOptions": map[string]any{
			"formats": []string{"markdown", "html", "rawHtml", "screenshot", "links"},
		},
	}
	response, err := app.CrawlURL("https://firecrawl.dev", params, false, 2, "")
	require.NoError(t, err)
	assert.NotNil(t, response)

	jobID, ok := response.(string)
	assert.True(t, ok)
	assert.NotEqual(t, "", jobID)

	const maxChecks = 15
	checks := 0

	for {
		if checks >= maxChecks {
			break
		}

		time.Sleep(5 * time.Second) // wait for 5 seconds

		statusResponse, err := app.CheckCrawlStatus(jobID)
		require.NoError(t, err)
		assert.NotNil(t, statusResponse)

		checkCrawlStatusResponse := statusResponse.(*CrawlStatusResponse)
		assert.Greater(t, len(checkCrawlStatusResponse.Data), 0)
		assert.GreaterOrEqual(t, checkCrawlStatusResponse.TotalCount, 0)
		assert.GreaterOrEqual(t, checkCrawlStatusResponse.CreditsUsed, 0)

		if checkCrawlStatusResponse.Status == "completed" {
			break
		}

		checks++
	}

	// Final check after loop or if completed
	statusResponse, err := app.CheckCrawlStatus(jobID)
	require.NoError(t, err)
	assert.NotNil(t, statusResponse)

	finalStatusResponse := statusResponse.(*CrawlStatusResponse)
	assert.Equal(t, "completed", finalStatusResponse.Status)
	assert.Greater(t, len(finalStatusResponse.Data), 0)
	assert.Greater(t, finalStatusResponse.TotalCount, 0)
	assert.Greater(t, finalStatusResponse.CreditsUsed, 0)
	assert.NotNil(t, finalStatusResponse.Data[0].Markdown)
	assert.Contains(t, finalStatusResponse.Data[0].HTML, "<div")
	assert.Contains(t, finalStatusResponse.Data[0].RawHTML, "<div")
	assert.Contains(t, finalStatusResponse.Data[0].Screenshot, "https://")
	assert.NotNil(t, finalStatusResponse.Data[0].Links)
	assert.Greater(t, len(finalStatusResponse.Data[0].Links), 0)
	assert.NotNil(t, finalStatusResponse.Data[0].Metadata.Title)
	assert.NotNil(t, finalStatusResponse.Data[0].Metadata.Description)
	assert.NotNil(t, finalStatusResponse.Data[0].Metadata.Language)
	assert.NotNil(t, finalStatusResponse.Data[0].Metadata.SourceURL)
	assert.NotNil(t, finalStatusResponse.Data[0].Metadata.StatusCode)
	assert.Empty(t, finalStatusResponse.Data[0].Metadata.Error)
}

func TestMapURLInvalidAPIKey(t *testing.T) {
	invalidApp, err := NewFirecrawlApp("invalid_api_key", API_URL, "v1")
	require.NoError(t, err)
	_, err = invalidApp.MapURL("https://roastmywebsite.ai", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Unexpected error during map: Status code 401. Unauthorized: Invalid token")
}

func TestMapURLBlocklistedURL(t *testing.T) {
	app, err := NewFirecrawlApp(TEST_API_KEY, API_URL, "v1")
	require.NoError(t, err)
	blocklistedUrl := "https://facebook.com/fake-test"
	_, err = app.MapURL(blocklistedUrl, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Unexpected error during map: Status code 403. URL is blocked. Firecrawl currently does not support social media scraping due to policy restrictions.")
}

func TestMapURLValidPreviewToken(t *testing.T) {
	app, err := NewFirecrawlApp("this_is_just_a_preview_token", API_URL, "v1")
	require.NoError(t, err)
	response, err := app.MapURL("https://roastmywebsite.ai", nil)
	require.NoError(t, err)

	assert.NotNil(t, response)
	assert.IsType(t, &MapResponse{}, response)
	assert.Greater(t, len(response.Links), 0)
	assert.Contains(t, response.Links[0], "https://")
	assert.Contains(t, response.Links[0], "roastmywebsite.ai")
}

func TestMapURLValidMap(t *testing.T) {
	app, err := NewFirecrawlApp(TEST_API_KEY, API_URL, "v1")
	require.NoError(t, err)

	response, err := app.MapURL("https://roastmywebsite.ai", nil)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.IsType(t, &MapResponse{}, response)
	assert.Greater(t, len(response.Links), 0)
	assert.Contains(t, response.Links[0], "https://")
	assert.Contains(t, response.Links[0], "roastmywebsite.ai")
}

func TestSearchNotImplementedError(t *testing.T) {
	app, err := NewFirecrawlApp(TEST_API_KEY, API_URL, "v1")
	require.NoError(t, err)
	_, err = app.Search("test query", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Search is not supported in v1")
}
