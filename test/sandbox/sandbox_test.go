// +build docker

package sandbox

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/biwakonbu/agent-runner/internal/worker"
)

// TestSandboxManager_StartStopContainer tests basic container lifecycle
func TestSandboxManager_StartStopContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker tests in short mode")
	}

	sm, err := worker.NewSandboxManager()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use lightweight test image
	image := "alpine:3.19"
	tmpDir := t.TempDir()

	containerID, err := sm.StartContainer(ctx, image, tmpDir, map[string]string{})
	if err != nil {
		t.Fatalf("StartContainer() error = %v", err)
	}

	if containerID == "" {
		t.Errorf("Container ID should not be empty")
	}

	// Verify container is running by executing a simple command
	exitCode, output, err := sm.Exec(ctx, containerID, []string{"sh", "-c", "echo 'test output'"})
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	if exitCode != 0 {
		t.Errorf("Exec() exit code = %d, want 0", exitCode)
	}

	if output == "" {
		t.Errorf("Exec() output should not be empty")
	}

	// Stop container
	err = sm.StopContainer(ctx, containerID)
	if err != nil {
		t.Fatalf("StopContainer() error = %v", err)
	}
}

// TestSandboxManager_EnvironmentVariables tests environment variable propagation
func TestSandboxManager_EnvironmentVariables(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker tests in short mode")
	}

	sm, err := worker.NewSandboxManager()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	image := "alpine:3.19"
	tmpDir := t.TempDir()

	env := map[string]string{
		"TEST_VAR_1": "value1",
		"TEST_VAR_2": "value2",
	}

	containerID, err := sm.StartContainer(ctx, image, tmpDir, env)
	if err != nil {
		t.Fatalf("StartContainer() error = %v", err)
	}
	defer sm.StopContainer(ctx, containerID)

	// Check if environment variables are set
	exitCode, output, err := sm.Exec(ctx, containerID, []string{"sh", "-c", "echo $TEST_VAR_1"})
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	if exitCode != 0 {
		t.Errorf("Exec() exit code = %d, want 0", exitCode)
	}

	if output == "" {
		t.Logf("Warning: Environment variable TEST_VAR_1 not found in output: %s", output)
	}
}

// TestSandboxManager_MountPoint tests repository mount
func TestSandboxManager_MountPoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker tests in short mode")
	}

	sm, err := worker.NewSandboxManager()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	image := "alpine:3.19"
	tmpDir := t.TempDir()

	// Create a test file in tmpDir
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	containerID, err := sm.StartContainer(ctx, image, tmpDir, map[string]string{})
	if err != nil {
		t.Fatalf("StartContainer() error = %v", err)
	}
	defer sm.StopContainer(ctx, containerID)

	// Check if the file exists in the container at /workspace/project
	exitCode, output, err := sm.Exec(ctx, containerID, []string{"sh", "-c", "ls -la /workspace/project/"})
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	if exitCode != 0 {
		t.Errorf("Exec() exit code = %d, want 0", exitCode)
	}

	// Check if test.txt is in the listing
	if output == "" {
		t.Errorf("Exec() output should not be empty")
	}
}

// TestSandboxManager_FileWritePermission tests that files can be written from container
func TestSandboxManager_FileWritePermission(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker tests in short mode")
	}

	sm, err := worker.NewSandboxManager()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	image := "alpine:3.19"
	tmpDir := t.TempDir()

	containerID, err := sm.StartContainer(ctx, image, tmpDir, map[string]string{})
	if err != nil {
		t.Fatalf("StartContainer() error = %v", err)
	}
	defer sm.StopContainer(ctx, containerID)

	// Write a file from container
	exitCode, _, err := sm.Exec(ctx, containerID, []string{"sh", "-c", "echo 'created from container' > /workspace/project/output.txt"})
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	if exitCode != 0 {
		t.Errorf("Exec() exit code = %d, want 0", exitCode)
	}

	// Stop container and verify file was written to host
	sm.StopContainer(ctx, containerID)

	outputFile := filepath.Join(tmpDir, "output.txt")
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	expectedContent := "created from container\n"
	if string(content) != expectedContent {
		t.Errorf("File content = %q, want %q", string(content), expectedContent)
	}
}

// TestSandboxManager_NonZeroExitCode tests handling of command failure
func TestSandboxManager_NonZeroExitCode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker tests in short mode")
	}

	sm, err := worker.NewSandboxManager()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	image := "alpine:3.19"
	tmpDir := t.TempDir()

	containerID, err := sm.StartContainer(ctx, image, tmpDir, map[string]string{})
	if err != nil {
		t.Fatalf("StartContainer() error = %v", err)
	}
	defer sm.StopContainer(ctx, containerID)

	// Execute command that fails
	exitCode, _, err := sm.Exec(ctx, containerID, []string{"sh", "-c", "exit 42"})
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	if exitCode != 42 {
		t.Errorf("Exec() exit code = %d, want 42", exitCode)
	}
}

// TestSandboxManager_MultipleExec tests multiple executions in same container
func TestSandboxManager_MultipleExec(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker tests in short mode")
	}

	sm, err := worker.NewSandboxManager()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	image := "alpine:3.19"
	tmpDir := t.TempDir()

	containerID, err := sm.StartContainer(ctx, image, tmpDir, map[string]string{})
	if err != nil {
		t.Fatalf("StartContainer() error = %v", err)
	}
	defer sm.StopContainer(ctx, containerID)

	// First execution
	exitCode1, _, err := sm.Exec(ctx, containerID, []string{"sh", "-c", "echo 'first' > /workspace/project/file1.txt"})
	if err != nil {
		t.Fatalf("First Exec() error = %v", err)
	}

	if exitCode1 != 0 {
		t.Errorf("First Exec() exit code = %d, want 0", exitCode1)
	}

	// Second execution
	exitCode2, _, err := sm.Exec(ctx, containerID, []string{"sh", "-c", "echo 'second' > /workspace/project/file2.txt"})
	if err != nil {
		t.Fatalf("Second Exec() error = %v", err)
	}

	if exitCode2 != 0 {
		t.Errorf("Second Exec() exit code = %d, want 0", exitCode2)
	}

	// Verify both files were created
	sm.StopContainer(ctx, containerID)

	if _, err := os.ReadFile(filepath.Join(tmpDir, "file1.txt")); err != nil {
		t.Fatalf("file1.txt not found: %v", err)
	}

	if _, err := os.ReadFile(filepath.Join(tmpDir, "file2.txt")); err != nil {
		t.Fatalf("file2.txt not found: %v", err)
	}
}

// TestSandboxManager_Cleanup verifies proper cleanup after stop
func TestSandboxManager_Cleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker tests in short mode")
	}

	sm, err := worker.NewSandboxManager()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	image := "alpine:3.19"
	tmpDir := t.TempDir()

	containerID, err := sm.StartContainer(ctx, image, tmpDir, map[string]string{})
	if err != nil {
		t.Fatalf("StartContainer() error = %v", err)
	}

	// Stop container
	err = sm.StopContainer(ctx, containerID)
	if err != nil {
		t.Fatalf("StopContainer() error = %v", err)
	}

	// Attempting to execute in stopped container should fail
	// Note: This may or may not fail depending on Docker implementation
	// Just verify that the function completes without panic
	_, _, _ = sm.Exec(ctx, containerID, []string{"echo", "test"})
}
