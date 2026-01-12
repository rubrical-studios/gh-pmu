package api

import (
	"fmt"
	"os"
	"time"
)

// DefaultRetryDelays defines the exponential backoff delays for retry attempts
var DefaultRetryDelays = []time.Duration{
	1 * time.Second,
	2 * time.Second,
	4 * time.Second,
	8 * time.Second,
}

// WithRetry executes a function with automatic retry on rate limit errors.
// Uses exponential backoff: 1s, 2s, 4s, 8s delays.
// Returns the function's error if not rate limited, or after max retries exceeded.
func WithRetry(fn func() error, maxRetries int) error {
	return WithRetryDelays(fn, maxRetries, DefaultRetryDelays)
}

// WithRetryDelays executes a function with automatic retry using custom delays.
// This variant is primarily for testing to allow faster tests.
func WithRetryDelays(fn func() error, maxRetries int, delays []time.Duration) error {
	var err error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err = fn()
		if err == nil || !IsRateLimited(err) {
			return err
		}
		if attempt < maxRetries {
			delay := delays[min(attempt, len(delays)-1)]
			fmt.Fprintf(os.Stderr, "Warning: rate limited, retrying in %v...\n", delay)
			time.Sleep(delay)
		}
	}
	return err
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
