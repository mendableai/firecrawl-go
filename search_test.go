package firecrawl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearch_ReturnsNotImplemented(t *testing.T) {
	app, err := NewFirecrawlApp("fc-test-key", "https://api.firecrawl.dev")
	if err != nil {
		t.Fatalf("unexpected error creating app: %v", err)
	}

	_, err = app.Search(context.Background(), "test query", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}
