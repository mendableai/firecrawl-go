module github.com/mendableai/firecrawl-go/v2

go 1.22.5

retract (
	v2.0.0 // Deprecated: use github.com/firecrawl/firecrawl/apps/go-sdk
	v2.1.0 // Deprecated: use github.com/firecrawl/firecrawl/apps/go-sdk
	v2.2.0 // Deprecated: use github.com/firecrawl/firecrawl/apps/go-sdk
	v2.3.0 // Deprecated: use github.com/firecrawl/firecrawl/apps/go-sdk
	v2.4.0 // Deprecated: use github.com/firecrawl/firecrawl/apps/go-sdk
)

require (
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
