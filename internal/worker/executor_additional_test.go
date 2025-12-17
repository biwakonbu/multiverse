package worker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/biwakonbu/agent-runner/pkg/config"
)

// TestNewExecutor_Success tests successful Executor creation
func TestNewExecutor_Success(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	// Use mock sandbox to avoid Docker dependency
	mockSandbox := &MockSandboxManager{}
	executor := &Executor{
		Config:      cfg,
		Sandbox:     mockSandbox,
		RepoPath:    "/test/repo",
		containerID: "",
	}

	if executor.Config.Kind != "codex-cli" {
		t.Errorf("Config.Kind = %s, want 'codex-cli'", executor.Config.Kind)
	}
	if executor.RepoPath != "/test/repo" {
		t.Errorf("RepoPath = %s, want '/test/repo'", executor.RepoPath)
	}
	if executor.Sandbox == nil {
		t.Errorf("Sandbox should not be nil")
	}
}

// TestStart_DefaultImage tests that Start uses default image when not specified
func TestStart_DefaultImage(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind: "codex-cli",
		// DockerImage not specified
	}

	mockSandbox := &MockSandboxManager{}
	executor := &Executor{
		Config:      cfg,
		Sandbox:     mockSandbox,
		RepoPath:    "/test/repo",
		containerID: "",
	}

	ctx := context.Background()
	err := executor.Start(ctx)

	if err != nil {
		t.Fatalf("Start() error = %v, want nil", err)
	}

	if executor.containerID == "" {
		t.Errorf("containerID should be set after Start()")
	}

	if !mockSandbox.startContainerCalled {
		t.Errorf("StartContainer should have been called")
	}
}

// TestStart_AbsolutePathResolution tests that relative paths are resolved to absolute
func TestStart_AbsolutePathResolution(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "test-image:latest",
	}

	mockSandbox := &MockSandboxManager{}
	executor := &Executor{
		Config:      cfg,
		Sandbox:     mockSandbox,
		RepoPath:    "", // Empty repo path should be resolved
		containerID: "",
	}

	ctx := context.Background()
	err := executor.Start(ctx)

	if err != nil {
		t.Fatalf("Start() error = %v, want nil", err)
	}

	// Should have called StartContainer with an absolute path
	if !mockSandbox.startContainerCalled {
		t.Errorf("StartContainer should have been called")
	}
}

// TestStart_RelativePathHandling tests handling of relative paths
func TestStart_RelativePathHandling(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "test-image:latest",
	}

	// Create a temporary directory to test with
	tmpDir, err := os.MkdirTemp("", "agent-runner-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Get relative path from current directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	relPath, err := filepath.Rel(cwd, tmpDir)
	if err != nil {
		t.Fatalf("Failed to get relative path: %v", err)
	}

	mockSandbox := &MockSandboxManager{}
	executor := &Executor{
		Config:      cfg,
		Sandbox:     mockSandbox,
		RepoPath:    relPath, // Use relative path
		containerID: "",
	}

	ctx := context.Background()
	err = executor.Start(ctx)

	if err != nil {
		t.Fatalf("Start() with relative path error = %v, want nil", err)
	}

	if !mockSandbox.startContainerCalled {
		t.Errorf("StartContainer should have been called")
	}

	// Verify that the path passed to StartContainer is absolute
	if !filepath.IsAbs(mockSandbox.lastRepoPath) {
		t.Errorf("StartContainer should receive absolute path, got: %s", mockSandbox.lastRepoPath)
	}
}

// TestStart_ErrorPropagation tests that StartContainer errors are propagated
func TestStart_ErrorPropagation(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "test-image:latest",
	}

	mockSandbox := &MockSandboxManager{
		startContainerErr: fmt.Errorf("Docker daemon not available"),
	}
	executor := &Executor{
		Config:      cfg,
		Sandbox:     mockSandbox,
		RepoPath:    "/test/repo",
		containerID: "",
	}

	ctx := context.Background()
	err := executor.Start(ctx)

	if err == nil {
		t.Fatalf("Start() expected error, got nil")
	}

	// containerID should remain empty on error
	if executor.containerID != "" {
		t.Errorf("containerID should be empty on error, got: %s", executor.containerID)
	}
}

func TestIsClaudeWorkerKind(t *testing.T) {
	t.Run("canonical", func(t *testing.T) {
		if !isClaudeWorkerKind("claude-code") {
			t.Fatalf("expected true for claude-code")
		}
	})
	t.Run("alias", func(t *testing.T) {
		if !isClaudeWorkerKind("claude-code-cli") {
			t.Fatalf("expected true for claude-code-cli")
		}
	})
	t.Run("non-claude", func(t *testing.T) {
		if isClaudeWorkerKind("codex-cli") {
			t.Fatalf("expected false for codex-cli")
		}
	})
}
