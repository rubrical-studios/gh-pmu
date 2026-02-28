package api

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestIsNotFound_WithNotFoundError(t *testing.T) {
	err := ErrNotFound
	if !IsNotFound(err) {
		t.Error("Expected IsNotFound to return true for ErrNotFound")
	}
}

func TestIsNotFound_WithGraphQLNotFound(t *testing.T) {
	err := errors.New("GraphQL: Could not resolve to a User")
	if !IsNotFound(err) {
		t.Error("Expected IsNotFound to return true for GraphQL not found")
	}
}

func TestIsNotFound_WithOtherError(t *testing.T) {
	err := errors.New("some other error")
	if IsNotFound(err) {
		t.Error("Expected IsNotFound to return false for other errors")
	}
}

func TestIsNotFound_WithNil(t *testing.T) {
	if IsNotFound(nil) {
		t.Error("Expected IsNotFound to return false for nil")
	}
}

func TestIsRateLimited_WithRateLimitError(t *testing.T) {
	err := ErrRateLimited
	if !IsRateLimited(err) {
		t.Error("Expected IsRateLimited to return true for ErrRateLimited")
	}
}

func TestIsRateLimited_WithRateLimitMessage(t *testing.T) {
	err := errors.New("API rate limit exceeded")
	if !IsRateLimited(err) {
		t.Error("Expected IsRateLimited to return true for rate limit message")
	}
}

func TestIsAuthError_WithAuthError(t *testing.T) {
	err := ErrNotAuthenticated
	if !IsAuthError(err) {
		t.Error("Expected IsAuthError to return true for ErrNotAuthenticated")
	}
}

func TestIsAuthError_With401(t *testing.T) {
	err := errors.New("401 Unauthorized")
	if !IsAuthError(err) {
		t.Error("Expected IsAuthError to return true for 401")
	}
}

func TestWrapError_WithNil(t *testing.T) {
	err := WrapError("get", "project", nil)
	if err != nil {
		t.Error("Expected WrapError to return nil for nil error")
	}
}

func TestWrapError_WithRateLimitError(t *testing.T) {
	original := errors.New("rate limit exceeded")
	err := WrapError("get", "project", original)

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatal("Expected wrapped error to be APIError")
	}

	if apiErr.Operation != "get" {
		t.Errorf("Expected operation 'get', got '%s'", apiErr.Operation)
	}

	if !errors.Is(apiErr.Err, ErrRateLimited) {
		t.Error("Expected wrapped error to contain ErrRateLimited")
	}
}

func TestWrapError_WithNotFoundError(t *testing.T) {
	original := errors.New("Could not resolve to a User")
	err := WrapError("get", "user", original)

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatal("Expected wrapped error to be APIError")
	}

	if !errors.Is(apiErr.Err, ErrNotFound) {
		t.Error("Expected wrapped error to contain ErrNotFound")
	}
}

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		Operation: "get",
		Resource:  "project",
		Err:       ErrNotFound,
	}

	expected := "get project: resource not found"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestAPIError_Unwrap(t *testing.T) {
	inner := ErrNotFound
	err := &APIError{
		Operation: "get",
		Resource:  "project",
		Err:       inner,
	}

	if !errors.Is(err, ErrNotFound) {
		t.Error("Expected errors.Is to find ErrNotFound")
	}
}

func TestIsRateLimited_With403RateLimitHTTPError(t *testing.T) {
	err := &ghHTTPError{statusCode: 403, message: "API rate limit exceeded"}
	if !IsRateLimited(err) {
		t.Error("Expected IsRateLimited to return true for 403 with rate limit message")
	}
}

func TestIsRateLimited_With429HTTPError(t *testing.T) {
	err := &ghHTTPError{statusCode: 429, message: "Too Many Requests"}
	if !IsRateLimited(err) {
		t.Error("Expected IsRateLimited to return true for 429")
	}
}

func TestIsRateLimited_With403PermissionDenied(t *testing.T) {
	err := &ghHTTPError{statusCode: 403, message: "Resource not accessible by integration"}
	if IsRateLimited(err) {
		t.Error("Expected IsRateLimited to return false for 403 permission denied")
	}
}

func TestGetRetryAfter_FromHTTPError(t *testing.T) {
	err := &ghHTTPError{
		statusCode: 429,
		message:    "Too Many Requests",
		retryAfter: "5",
	}
	d := GetRetryAfter(err)
	if d != 5*time.Second {
		t.Errorf("Expected 5s retry-after, got %v", d)
	}
}

func TestGetRetryAfter_NoHeader(t *testing.T) {
	err := &ghHTTPError{statusCode: 429, message: "Too Many Requests"}
	d := GetRetryAfter(err)
	if d != 0 {
		t.Errorf("Expected 0 (no Retry-After), got %v", d)
	}
}

func TestGetRetryAfter_NonHTTPError(t *testing.T) {
	err := errors.New("some error")
	d := GetRetryAfter(err)
	if d != 0 {
		t.Errorf("Expected 0 for non-HTTP error, got %v", d)
	}
}

// ghHTTPError simulates go-gh's api.HTTPError for testing within this package.
// We can't import github.com/cli/go-gh/v2/pkg/api here (same package name),
// so we use a local type and test that IsRateLimited handles the StatusCode interface.
type ghHTTPError struct {
	statusCode int
	message    string
	retryAfter string
}

func (e *ghHTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.statusCode, e.message)
}

func (e *ghHTTPError) HTTPStatusCode() int {
	return e.statusCode
}

func (e *ghHTTPError) RetryAfterSeconds() string {
	return e.retryAfter
}
