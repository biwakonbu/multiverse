package main

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

// TestRun_InvalidYAML verifies that Run returns an error for invalid YAML input.
func TestRun_InvalidYAML(t *testing.T) {
	input := strings.NewReader("invalid: yaml: content")
	var stdout, stderr bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&stderr, nil))

	err := Run(context.Background(), input, &stdout, &stderr, logger)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

// TestRun_EmptyInput verifies that Run returns an error for empty input.
func TestRun_EmptyInput(t *testing.T) {
	input := strings.NewReader("")
	var stdout, stderr bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&stderr, nil))

	err := Run(context.Background(), input, &stdout, &stderr, logger)
	if err == nil {
		t.Error("Expected error for empty input, got nil")
	}
}
