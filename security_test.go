package firecrawl

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

// ---- validatePaginationURL ----

func TestValidatePaginationURL_SameHost(t *testing.T) {
	err := validatePaginationURL(
		"https://api.firecrawl.dev",
		"https://api.firecrawl.dev/v2/crawl/abc123?cursor=2",
	)
	if err != nil {
		t.Fatalf("expected no error for matching hosts, got: %v", err)
	}
}

func TestValidatePaginationURL_DifferentHost(t *testing.T) {
	err := validatePaginationURL(
		"https://api.firecrawl.dev",
		"https://attacker.example.com/steal-token",
	)
	if err == nil {
		t.Fatal("expected error for mismatched hosts, got nil")
	}
	if !strings.Contains(err.Error(), "does not match API host") {
		t.Fatalf("expected host mismatch error, got: %v", err)
	}
}

func TestValidatePaginationURL_EmptyNextURL(t *testing.T) {
	// An empty string parses to a URL with no host — should fail since base has a host.
	err := validatePaginationURL(
		"https://api.firecrawl.dev",
		"",
	)
	if err == nil {
		t.Fatal("expected error for empty next URL, got nil")
	}
}

func TestValidatePaginationURL_RelativeURL(t *testing.T) {
	// A relative URL has no host — should fail host comparison.
	err := validatePaginationURL(
		"https://api.firecrawl.dev",
		"/v2/crawl/abc123?cursor=2",
	)
	if err == nil {
		t.Fatal("expected error for relative URL (no host), got nil")
	}
	if !strings.Contains(err.Error(), "does not match API host") {
		t.Fatalf("expected host mismatch error, got: %v", err)
	}
}

// ---- validateJobID ----

func TestValidateJobID_ValidUUID(t *testing.T) {
	err := validateJobID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("expected no error for valid UUID, got: %v", err)
	}
}

func TestValidateJobID_InvalidString(t *testing.T) {
	err := validateJobID("not-a-uuid")
	if err == nil {
		t.Fatal("expected error for non-UUID string, got nil")
	}
	if !strings.Contains(err.Error(), "must be a valid UUID") {
		t.Fatalf("expected UUID error message, got: %v", err)
	}
}

func TestValidateJobID_PathTraversal(t *testing.T) {
	err := validateJobID("../../admin")
	if err == nil {
		t.Fatal("expected error for path traversal string, got nil")
	}
	if !strings.Contains(err.Error(), "must be a valid UUID") {
		t.Fatalf("expected UUID error message, got: %v", err)
	}
}

func TestValidateJobID_EmptyString(t *testing.T) {
	err := validateJobID("")
	if err == nil {
		t.Fatal("expected error for empty string, got nil")
	}
	if !strings.Contains(err.Error(), "must be a valid UUID") {
		t.Fatalf("expected UUID error message, got: %v", err)
	}
}

// ---- FirecrawlApp.String() redaction ----

func TestFirecrawlApp_String_Redaction(t *testing.T) {
	app := &FirecrawlApp{
		apiKey: "fc-abcdefghijklmnop",
		APIURL: "https://api.firecrawl.dev",
	}
	s := app.String()
	if strings.Contains(s, "fc-abcdefghijklmnop") {
		t.Fatalf("String() should redact the API key, but found full key in: %s", s)
	}
	// Should show first 3 chars and last 4 chars.
	if !strings.Contains(s, "fc-") {
		t.Fatalf("String() should show first 3 chars, got: %s", s)
	}
	if !strings.Contains(s, "mnop") {
		t.Fatalf("String() should show last 4 chars, got: %s", s)
	}
	if !strings.Contains(s, "...") {
		t.Fatalf("String() should contain '...', got: %s", s)
	}
}

func TestFirecrawlApp_String_ShortKey(t *testing.T) {
	app := &FirecrawlApp{
		apiKey: "mykey",
		APIURL: "https://api.firecrawl.dev",
	}
	s := app.String()
	// Short keys (<=7 chars) get fully replaced with "***".
	if strings.Contains(s, "mykey") {
		t.Fatalf("String() should redact short keys, but found full key in: %s", s)
	}
	if !strings.Contains(s, "***") {
		t.Fatalf("String() should use '***' for short keys, got: %s", s)
	}
}

// ---- APIKey() accessor ----

func TestFirecrawlApp_APIKey_Accessor(t *testing.T) {
	app := &FirecrawlApp{
		apiKey: "fc-test-key-1234",
		APIURL: "https://api.firecrawl.dev",
	}
	if app.APIKey() != "fc-test-key-1234" {
		t.Fatalf("APIKey() returned %q, want %q", app.APIKey(), "fc-test-key-1234")
	}
}

// ---- HTTPS warning in NewFirecrawlApp ----

func TestNewFirecrawlApp_HTTPWarning(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil) // restore default output after test

	_, err := NewFirecrawlApp("test-key", "http://remote.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "WARNING") {
		t.Fatalf("expected WARNING log for non-localhost HTTP URL, got: %q", logOutput)
	}
	if !strings.Contains(logOutput, "cleartext") {
		t.Fatalf("expected cleartext warning in log, got: %q", logOutput)
	}
}

func TestNewFirecrawlApp_HTTPSNoWarning(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	_, err := NewFirecrawlApp("test-key", "https://api.firecrawl.dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logOutput := buf.String()
	if strings.Contains(logOutput, "WARNING") {
		t.Fatalf("expected no WARNING for HTTPS URL, got: %q", logOutput)
	}
}

func TestNewFirecrawlApp_HTTPLocalhostNoWarning(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	for _, host := range []string{
		"http://localhost:8080",
		"http://127.0.0.1:3000",
	} {
		buf.Reset()
		_, err := NewFirecrawlApp("test-key", host)
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", host, err)
		}
		logOutput := buf.String()
		if strings.Contains(logOutput, "WARNING") {
			t.Fatalf("expected no WARNING for localhost URL %s, got: %q", host, logOutput)
		}
	}
}
