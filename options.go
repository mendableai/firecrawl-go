package firecrawl

// requestOptions represents options for making requests.
type requestOptions struct {
	retries int
	backoff int
}

// requestOption is a functional option type for requestOptions.
type requestOption func(*requestOptions)

// newRequestOptions creates a new requestOptions instance with the provided options.
//
// Parameters:
//   - opts: Optional request options.
//
// Returns:
//   - *requestOptions: A new instance of requestOptions with the provided options.
func newRequestOptions(opts ...requestOption) *requestOptions {
	options := &requestOptions{retries: 1}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// withRetries sets the number of retries for a request.
//
// Parameters:
//   - retries: The number of retries to be performed.
//
// Returns:
//   - requestOption: A functional option that sets the number of retries for a request.
func withRetries(retries int) requestOption {
	return func(opts *requestOptions) {
		opts.retries = retries
	}
}

// withBackoff sets the backoff interval for a request.
//
// Parameters:
//   - backoff: The backoff interval (in milliseconds) to be used for retries.
//
// Returns:
//   - requestOption: A functional option that sets the backoff interval for a request.
func withBackoff(backoff int) requestOption {
	return func(opts *requestOptions) {
		opts.backoff = backoff
	}
}
