# Firecrawl Go SDK

The Firecrawl Go SDK is a library that allows you to easily scrape and crawl websites, and output the data in a format ready for use with language models (LLMs). It provides a simple and intuitive interface for interacting with the Firecrawl API.

## Installation

To install the Firecrawl Go SDK, you can

```bash
go get github.com/mendableai/firecrawl-go
```

## Usage

1. Get an API key from [firecrawl.dev](https://firecrawl.dev)
2. Set the API key as an environment variable named `FIRECRAWL_API_KEY` or pass it as a parameter to the `FirecrawlApp` class.


Here's an example of how to use the SDK with error handling:

```go
import (
	"fmt"
	"log"

	"github.com/mendableai/firecrawl-go"
)

func main() {
	// Initialize the FirecrawlApp with your API key
	app, err := firecrawl.NewFirecrawlApp("YOUR_API_KEY")
	if err != nil {
		log.Fatalf("Failed to initialize FirecrawlApp: %v", err)
	}

	// Scrape a single URL
	scrapeResult, err := app.ScrapeURL("example.com", nil)
	if err != nil {
		log.Fatalf("Failed to scrape URL: %v", err)
	}
	fmt.Println(scrapeResult.Markdown)

	// Crawl a website
	idempotencyKey := uuid.New().String() // optional idempotency key
	crawlParams := &firecrawl.CrawlParams{
		ExcludePaths: []string{"blog/*"},
		MaxDepth:     prt(2),
	}
	crawlResult, err := app.CrawlURL("example.com", crawlParams, &idempotencyKey)
	if err != nil {
		log.Fatalf("Failed to crawl URL: %v", err)
	}
	jsonCrawlResult, err := json.MarshalIndent(crawlResult, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal crawl result: %v", err)
	}
	fmt.Println(string(jsonCrawlResult))
}
```

### Scraping a URL

To scrape a single URL with error handling, use the `ScrapeURL` method. It takes the URL as a parameter and returns the scraped data as a dictionary.

```go
url := "https://example.com"
scrapedData, err := app.ScrapeURL(url, nil)
if err != nil {
	log.Fatalf("Failed to scrape URL: %v", err)
}
fmt.Println(scrapedData)
```

### Extracting structured data from a URL

With LLM extraction, you can easily extract structured data from any URL. Here is how you to use it:

```go
jsonSchema := map[string]any{
	"type": "object",
	"properties": map[string]any{
		"top": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"title":       map[string]string{"type": "string"},
					"points":      map[string]string{"type": "number"},
					"by":          map[string]string{"type": "string"},
					"commentsURL": map[string]string{"type": "string"},
				},
				"required": []string{"title", "points", "by", "commentsURL"},
			},
			"minItems":    5,
			"maxItems":    5,
			"description": "Top 5 stories on Hacker News",
		},
	},
	"required": []string{"top"},
}

llmExtractionParams := map[string]any{
	"extractorOptions": firecrawl.ExtractorOptions{
		ExtractionSchema: jsonSchema,
	},
}

scrapeResult, err := app.ScrapeURL("https://news.ycombinator.com", llmExtractionParams)
if err != nil {
	log.Fatalf("Failed to perform LLM extraction: %v", err)
}
fmt.Println(scrapeResult)
```

### Crawling a Website

To crawl a website, use the `CrawlUrl` method. It takes the starting URL and optional parameters as arguments. The `params` argument allows you to specify additional options for the crawl job, such as the maximum number of pages to crawl, allowed domains, and the output format.

```go
response, err := app.CrawlURL("https://roastmywebsite.ai", nil,nil)

if err != nil {
 log.Fatalf("Failed to crawl URL: %v", err)
}

fmt.Println(response)
```

### Asynchronous Crawl

To initiate an asynchronous crawl of a website, utilize the `AsyncCrawlURL` method. This method requires the starting URL and optional parameters as inputs. The `params` argument enables you to define various settings for the asynchronous crawl, such as the maximum number of pages to crawl, permitted domains, and the output format. Upon successful initiation, this method returns an ID, which is essential for subsequently checking the status of the crawl.

```go
response, err := app.AsyncCrawlURL("https://roastmywebsite.ai", nil, nil)

if err != nil {
  log.Fatalf("Failed to crawl URL: %v", err)
}

fmt.Println(response) 
```


### Checking Crawl Status

To check the status of a crawl job, use the `CheckCrawlStatus` method. It takes the crawl ID as a parameter and returns the current status of the crawl job.

```go
status, err := app.CheckCrawlStatus(id)
if err != nil {
	log.Fatalf("Failed to check crawl status: %v", err)
}
fmt.Println(status)
```

### Canceling a Crawl Job
To cancel a crawl job, use the `CancelCrawlJob` method. It takes the job ID as a parameter and returns the cancellation status of the crawl job.

```go
canceled, err := app.CancelCrawlJob(jobId)
if err != nil {
	log.Fatalf("Failed to cancel crawl job: %v", err)
}
fmt.Println(canceled)
```

## Error Handling

The SDK handles errors returned by the Firecrawl API and raises appropriate exceptions. If an error occurs during a request, an exception will be raised with a descriptive error message.

## Contributing

Contributions to the Firecrawl Go SDK are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request on the GitHub repository.

## License

The Firecrawl Go SDK is licensed under the MIT License. This means you are free to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the SDK, subject to the following conditions:

- The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

Please note that while this SDK is MIT licensed, it is part of a larger project which may be under different licensing terms. Always refer to the license information in the root directory of the main project for overall licensing details.
