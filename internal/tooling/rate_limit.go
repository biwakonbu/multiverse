package tooling

import "strings"

// IsRateLimitError detects common rate-limit errors across providers.
func IsRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "rate limit") {
		return true
	}
	if strings.Contains(msg, "too many requests") {
		return true
	}
	if strings.Contains(msg, "429") {
		return true
	}
	if strings.Contains(msg, "resource exhausted") {
		return true
	}
	return false
}
