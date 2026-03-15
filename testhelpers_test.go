package firecrawl

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// newMockServer creates a test HTTP server and a FirecrawlApp configured to use it.
// The server is automatically cleaned up when the test completes.
func newMockServer(t *testing.T, handler http.HandlerFunc) (*FirecrawlApp, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	app, err := NewFirecrawlApp("fc-test-key", server.URL)
	require.NoError(t, err)
	return app, server
}

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(v) //nolint:gosec
}

// decodeJSONBody decodes the request body into the given pointer.
func decodeJSONBody(t *testing.T, r *http.Request, v any) {
	t.Helper()
	err := json.NewDecoder(r.Body).Decode(v)
	require.NoError(t, err, "failed to decode request body")
}

// ptr returns a pointer to the given value. Useful for constructing test params.
func ptr[T any](v T) *T {
	return &v
}
