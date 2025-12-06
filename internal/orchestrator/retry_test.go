package orchestrator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultRetryPolicy(t *testing.T) {
	policy := DefaultRetryPolicy()

	assert.Equal(t, 3, policy.MaxAttempts)
	assert.Equal(t, 5*time.Second, policy.BackoffBase)
	assert.Equal(t, 5*time.Minute, policy.BackoffMax)
	assert.Equal(t, 2.0, policy.BackoffFactor)
	assert.True(t, policy.RequireHuman)
}

func TestRetryPolicy_CalculateBackoff(t *testing.T) {
	policy := &RetryPolicy{
		BackoffBase:   5 * time.Second,
		BackoffMax:    5 * time.Minute,
		BackoffFactor: 2.0,
	}

	tests := []struct {
		name          string
		attemptNumber int
		expected      time.Duration
	}{
		{
			name:          "first attempt",
			attemptNumber: 1,
			expected:      5 * time.Second, // 5 * 2^0 = 5s
		},
		{
			name:          "second attempt",
			attemptNumber: 2,
			expected:      10 * time.Second, // 5 * 2^1 = 10s
		},
		{
			name:          "third attempt",
			attemptNumber: 3,
			expected:      20 * time.Second, // 5 * 2^2 = 20s
		},
		{
			name:          "fourth attempt",
			attemptNumber: 4,
			expected:      40 * time.Second, // 5 * 2^3 = 40s
		},
		{
			name:          "zero attempt treated as first",
			attemptNumber: 0,
			expected:      5 * time.Second,
		},
		{
			name:          "negative attempt treated as first",
			attemptNumber: -1,
			expected:      5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := policy.CalculateBackoff(tt.attemptNumber)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRetryPolicy_CalculateBackoff_CapsAtMax(t *testing.T) {
	policy := &RetryPolicy{
		BackoffBase:   1 * time.Minute,
		BackoffMax:    5 * time.Minute,
		BackoffFactor: 2.0,
	}

	// attempt 4: 1min * 2^3 = 8min, but capped at 5min
	result := policy.CalculateBackoff(4)
	assert.Equal(t, 5*time.Minute, result)

	// attempt 10: would be huge, but capped at 5min
	result = policy.CalculateBackoff(10)
	assert.Equal(t, 5*time.Minute, result)
}

func TestRetryPolicy_ShouldRetry(t *testing.T) {
	policy := &RetryPolicy{
		MaxAttempts: 3,
	}

	tests := []struct {
		name          string
		attemptNumber int
		expected      bool
	}{
		{
			name:          "first attempt - should retry",
			attemptNumber: 1,
			expected:      true,
		},
		{
			name:          "second attempt - should retry",
			attemptNumber: 2,
			expected:      true,
		},
		{
			name:          "third attempt - should not retry (at max)",
			attemptNumber: 3,
			expected:      false,
		},
		{
			name:          "fourth attempt - should not retry (over max)",
			attemptNumber: 4,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := policy.ShouldRetry(tt.attemptNumber)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRetryPolicy_DetermineNextAction(t *testing.T) {
	t.Run("should retry when under max attempts", func(t *testing.T) {
		policy := &RetryPolicy{MaxAttempts: 3, RequireHuman: true}

		assert.Equal(t, NextActionRetry, policy.DetermineNextAction(1))
		assert.Equal(t, NextActionRetry, policy.DetermineNextAction(2))
	})

	t.Run("should add to backlog when at max and require human", func(t *testing.T) {
		policy := &RetryPolicy{MaxAttempts: 3, RequireHuman: true}

		assert.Equal(t, NextActionBacklog, policy.DetermineNextAction(3))
		assert.Equal(t, NextActionBacklog, policy.DetermineNextAction(4))
	})

	t.Run("should fail when at max and not require human", func(t *testing.T) {
		policy := &RetryPolicy{MaxAttempts: 3, RequireHuman: false}

		assert.Equal(t, NextActionFail, policy.DetermineNextAction(3))
		assert.Equal(t, NextActionFail, policy.DetermineNextAction(4))
	})
}
