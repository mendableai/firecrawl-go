package firecrawl

import (
	"encoding/json"
	"errors"
	"fmt"
)

// handleError handles errors returned by the Firecrawl API.
//
// Parameters:
//   - resp: The HTTP response object.
//   - body: The response body from the HTTP response.
//   - action: A string describing the action being performed.
//
// Returns:
//   - error: An error describing the failure reason.
func (app *FirecrawlApp) handleError(statusCode int, body []byte, action string) error {
	var errorData map[string]any
	err := json.Unmarshal(body, &errorData)
	if err != nil {
		return fmt.Errorf("failed to parse error response: %v", err)
	}

	errorMessage, _ := errorData["error"].(string)
	if errorMessage == "" {
		errorMessage = "No additional error details provided."
	}

	var message string
	switch statusCode {
	case 402:
		message = fmt.Sprintf("Payment Required: Failed to %s. %s", action, errorMessage)
	case 408:
		message = fmt.Sprintf("Request Timeout: Failed to %s as the request timed out. %s", action, errorMessage)
	case 409:
		message = fmt.Sprintf("Conflict: Failed to %s due to a conflict. %s", action, errorMessage)
	case 500:
		message = fmt.Sprintf("Internal Server Error: Failed to %s. %s", action, errorMessage)
	default:
		message = fmt.Sprintf("Unexpected error during %s: Status code %d. %s", action, statusCode, errorMessage)
	}

	return errors.New(message)
}
