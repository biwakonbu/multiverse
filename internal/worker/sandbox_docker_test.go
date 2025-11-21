//go:build docker
// +build docker

package worker

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker/client"
)

// TestSandboxManager_Integration tests the SandboxManager with a real Docker daemon.
// Requires Docker to be running.
func TestSandboxManager_Integration(t *testing.T) {
	// Ensure Docker client can be created
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Skipf("Skipping Docker integration test: failed to create client: %v", err)
	}
	ctx := context.Background()
	if _, err := cli.Ping(ctx); err != nil {
		t.Skipf("Skipping Docker integration test: Docker daemon not available: %v", err)
	}

	// Create a temporary directory for the workspace
	tmpDir, err := os.MkdirTemp("", "agent-runner-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a dummy file in the workspace
	testFile := filepath.Join(tmpDir, "hello.txt")
	if err := os.WriteFile(testFile, []byte("Hello from Host"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Initialize SandboxManager
	manager, err := NewSandboxManager()
	if err != nil {
		t.Fatalf("Failed to create SandboxManager: %v", err)
	}

	// Use a lightweight image
	image := "alpine:latest"

	// Test StartContainer
	env := map[string]string{"TEST_VAR": "test_value"}
	containerID, err := manager.StartContainer(ctx, image, tmpDir, env)
	if err != nil {
		t.Fatalf("StartContainer failed: %v", err)
	}
	defer func() {
		// Cleanup
		if err := manager.StopContainer(ctx, containerID); err != nil {
			t.Logf("Failed to stop container %s: %v", containerID, err)
		}
	}()

	// Test Exec: Check file existence and content
	// cat /workspace/project/hello.txt
	exitCode, output, err := manager.Exec(ctx, containerID, []string{"cat", "/workspace/project/hello.txt"})
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Exec exit code = %d, want 0", exitCode)
	}
	if !strings.Contains(output, "Hello from Host") {
		t.Errorf("Exec output = %q, want to contain 'Hello from Host'", output)
	}

	// Test Exec: Check environment variable
	// env
	exitCode, output, err = manager.Exec(ctx, containerID, []string{"env"})
	if err != nil {
		t.Fatalf("Exec env failed: %v", err)
	}
	if !strings.Contains(output, "TEST_VAR=test_value") {
		t.Errorf("Exec env output missing TEST_VAR. Output: %s", output)
	}

	// Test StopContainer is handled by defer, but let's verify it's running first
	// No easy way to check "running" state via manager interface without adding method,
	// but Exec success implies running.
}
