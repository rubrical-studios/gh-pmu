package api

import (
	"errors"
	"testing"
	"time"
)

func TestWithRetryDelays_SuccessOnFirstAttempt(t *testing.T) {
	callCount := 0
	err := WithRetryDelays(func() error {
		callCount++
		return nil
	}, 3, []time.Duration{1 * time.Millisecond})

	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestWithRetryDelays_SuccessAfterRetry(t *testing.T) {
	callCount := 0
	err := WithRetryDelays(func() error {
		callCount++
		if callCount < 3 {
			return ErrRateLimited
		}
		return nil
	}, 3, []time.Duration{1 * time.Millisecond, 1 * time.Millisecond})

	if err != nil {
		t.Errorf("Expected nil error after retry, got: %v", err)
	}
	if callCount != 3 {
		t.Errorf("Expected 3 calls (2 failures + 1 success), got %d", callCount)
	}
}

func TestWithRetryDelays_GivesUpAfterMaxRetries(t *testing.T) {
	callCount := 0
	err := WithRetryDelays(func() error {
		callCount++
		return ErrRateLimited
	}, 3, []time.Duration{1 * time.Millisecond})

	if err == nil {
		t.Fatal("Expected error after max retries, got nil")
	}
	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("Expected ErrRateLimited, got: %v", err)
	}
	// 4 attempts: initial + 3 retries
	if callCount != 4 {
		t.Errorf("Expected 4 calls (1 initial + 3 retries), got %d", callCount)
	}
}

func TestWithRetryDelays_NonRateLimitedErrorReturnsImmediately(t *testing.T) {
	callCount := 0
	otherErr := errors.New("some other error")
	err := WithRetryDelays(func() error {
		callCount++
		return otherErr
	}, 3, []time.Duration{1 * time.Millisecond})

	if err != otherErr {
		t.Errorf("Expected original error, got: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call (no retries for non-rate-limit errors), got %d", callCount)
	}
}

func TestWithRetryDelays_UsesLastDelayWhenAttemptsExceedDelayLength(t *testing.T) {
	callCount := 0
	delays := []time.Duration{1 * time.Millisecond, 2 * time.Millisecond}

	start := time.Now()
	_ = WithRetryDelays(func() error {
		callCount++
		return ErrRateLimited
	}, 3, delays)
	elapsed := time.Since(start)

	// Should have delays: 1ms, 2ms, 2ms (uses last delay when index exceeds length)
	// Total: 5ms minimum
	if elapsed < 5*time.Millisecond {
		t.Errorf("Expected at least 5ms total delay, got %v", elapsed)
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{3, 3, 3},
		{0, 5, 0},
		{-1, 1, -1},
	}

	for _, tt := range tests {
		result := min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("min(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestWithRetry_CallsWithRetryDelays(t *testing.T) {
	callCount := 0
	err := WithRetry(func() error {
		callCount++
		return nil
	}, 0)

	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}
