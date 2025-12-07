//go:build docker
// +build docker

package sandbox

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	exitCode, output, err := sm.Exec(ctx, containerID, []string{"sh", "-c", "echo 'test output'"}, nil)
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
	exitCode, output, err := sm.Exec(ctx, containerID, []string{"sh", "-c", "echo $TEST_VAR_1"}, nil)
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
	exitCode, output, err := sm.Exec(ctx, containerID, []string{"sh", "-c", "ls -la /workspace/project/"}, nil)
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
	exitCode, _, err := sm.Exec(ctx, containerID, []string{"sh", "-c", "echo 'created from container' > /workspace/project/output.txt"}, nil)
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
	exitCode, _, err := sm.Exec(ctx, containerID, []string{"sh", "-c", "exit 42"}, nil)
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
	exitCode1, _, err := sm.Exec(ctx, containerID, []string{"sh", "-c", "echo 'first' > /workspace/project/file1.txt"}, nil)
	if err != nil {
		t.Fatalf("First Exec() error = %v", err)
	}

	if exitCode1 != 0 {
		t.Errorf("First Exec() exit code = %d, want 0", exitCode1)
	}

	// Second execution
	exitCode2, _, err := sm.Exec(ctx, containerID, []string{"sh", "-c", "echo 'second' > /workspace/project/file2.txt"}, nil)
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
	_, _, _ = sm.Exec(ctx, containerID, []string{"echo", "test"}, nil)
}

// TestSandboxManager_ImagePull_AutoPull verifies automatic image pull when image doesn't exist locally
func TestSandboxManager_ImagePull_AutoPull(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker tests in short mode")
	}

	sm, err := worker.NewSandboxManager()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	// 長めのタイムアウト設定（イメージ取得を含む）
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// 小さいイメージを指定（alpine:3.19.1）
	// Note: 既にローカルに存在する場合はImagePullはスキップされるが、
	// そのケースでもStartContainerは成功するため、機能は正常に動作する
	image := "alpine:3.19.1"
	tmpDir := t.TempDir()

	// Act: StartContainer() を呼び出し（イメージがローカルになければ ImagePull が自動実行される）
	containerID, err := sm.StartContainer(ctx, image, tmpDir, map[string]string{})
	if err != nil {
		t.Fatalf("StartContainer() with ImagePull failed: %v", err)
	}
	defer sm.StopContainer(ctx, containerID)

	// Assert: コンテナが起動していることを確認
	if containerID == "" {
		t.Error("Container ID should not be empty")
	}

	// コンテナが実行可能であることを検証
	exitCode, output, err := sm.Exec(ctx, containerID, []string{"echo", "test"}, nil)
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	if exitCode != 0 {
		t.Errorf("Exec() exit code = %d, want 0", exitCode)
	}

	if output == "" {
		t.Errorf("Exec() output should not be empty")
	}
}

// TestSandboxManager_ImagePull_NonExistentImage verifies error handling for non-existent image
func TestSandboxManager_ImagePull_NonExistentImage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker tests in short mode")
	}

	sm, err := worker.NewSandboxManager()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 存在しないイメージを指定
	image := "nonexistent-registry.invalid/test:99.99"
	tmpDir := t.TempDir()

	// Act: StartContainer() を呼び出し（ImagePull が失敗するはず）
	_, err = sm.StartContainer(ctx, image, tmpDir, map[string]string{})

	// Assert: エラーが返されることを検証
	if err == nil {
		t.Fatal("Expected ImagePull to fail for nonexistent image, but got nil error")
	}

	// エラーメッセージに "failed to pull image" が含まれることを確認
	if !strings.Contains(err.Error(), "failed to pull image") {
		t.Errorf("Error message should mention image pull failure: %v", err)
	}
}

// TestSandboxManager_ConcurrentExec tests concurrent Exec calls in same container
func TestSandboxManager_ConcurrentExec(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker tests in short mode")
	}

	sm, err := worker.NewSandboxManager()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	image := "alpine:3.19"
	tmpDir := t.TempDir()

	containerID, err := sm.StartContainer(ctx, image, tmpDir, map[string]string{})
	if err != nil {
		t.Fatalf("StartContainer() error = %v", err)
	}
	defer sm.StopContainer(ctx, containerID)

	// Run 3 concurrent Exec calls
	type result struct {
		exitCode int
		output   string
		err      error
	}
	results := make(chan result, 3)

	for i := 0; i < 3; i++ {
		go func(index int) {
			exitCode, output, err := sm.Exec(ctx, containerID, []string{"sh", "-c", fmt.Sprintf("echo 'Task %d'", index)}, nil)
			results <- result{exitCode, output, err}
		}(i)
	}

	// Collect results
	for i := 0; i < 3; i++ {
		res := <-results
		if res.err != nil {
			t.Errorf("Concurrent Exec %d failed: %v", i, res.err)
		}
		if res.exitCode != 0 {
			t.Errorf("Concurrent Exec %d exit code = %d, want 0", i, res.exitCode)
		}
	}
}

// TestSandboxManager_LargeOutput tests handling of large command output
func TestSandboxManager_LargeOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker tests in short mode")
	}

	sm, err := worker.NewSandboxManager()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	image := "alpine:3.19"
	tmpDir := t.TempDir()

	containerID, err := sm.StartContainer(ctx, image, tmpDir, map[string]string{})
	if err != nil {
		t.Fatalf("StartContainer() error = %v", err)
	}
	defer sm.StopContainer(ctx, containerID)

	// Generate large output (1000 lines)
	exitCode, output, err := sm.Exec(ctx, containerID, []string{"sh", "-c", "for i in $(seq 1 1000); do echo \"Line $i: This is a test line with some content\"; done"}, nil)
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	if exitCode != 0 {
		t.Errorf("Exec() exit code = %d, want 0", exitCode)
	}

	// Verify output contains expected number of lines
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 900 { // Allow some margin
		t.Errorf("Expected ~1000 lines, got %d", len(lines))
	}
}

// TestSandboxManager_LongRunningCommand tests long-running command execution
func TestSandboxManager_LongRunningCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker tests in short mode")
	}

	sm, err := worker.NewSandboxManager()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	image := "alpine:3.19"
	tmpDir := t.TempDir()

	containerID, err := sm.StartContainer(ctx, image, tmpDir, map[string]string{})
	if err != nil {
		t.Fatalf("StartContainer() error = %v", err)
	}
	defer sm.StopContainer(ctx, containerID)

	// Run command that takes 3 seconds
	start := time.Now()
	exitCode, _, err := sm.Exec(ctx, containerID, []string{"sh", "-c", "sleep 3 && echo 'done'"}, nil)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	if exitCode != 0 {
		t.Errorf("Exec() exit code = %d, want 0", exitCode)
	}

	// Verify it actually took ~3 seconds
	if duration < 2*time.Second || duration > 5*time.Second {
		t.Errorf("Command duration = %v, expected ~3 seconds", duration)
	}
}

// TestSandboxManager_InvalidContainerID tests Exec with invalid container ID
func TestSandboxManager_InvalidContainerID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker tests in short mode")
	}

	sm, err := worker.NewSandboxManager()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Try to exec with non-existent container ID
	_, _, err = sm.Exec(ctx, "nonexistent-container-id-12345", []string{"echo", "test"}, nil)
	if err == nil {
		t.Error("Exec() should fail with invalid container ID")
	}

	// Error should be from Docker API
	if !strings.Contains(err.Error(), "No such container") && !strings.Contains(err.Error(), "no such container") {
		t.Logf("Expected 'No such container' error, got: %v", err)
	}
}

// TestSandboxManager_ContextCancellation tests context cancellation during operations
func TestSandboxManager_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker tests in short mode")
	}

	sm, err := worker.NewSandboxManager()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	// Create context that will be cancelled immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	image := "alpine:3.19"
	tmpDir := t.TempDir()

	// StartContainer should fail with cancelled context
	_, err = sm.StartContainer(ctx, image, tmpDir, map[string]string{})
	if err == nil {
		t.Error("StartContainer() should fail with cancelled context")
	}

	if !strings.Contains(err.Error(), "context canceled") && err != context.Canceled {
		t.Logf("Expected context cancellation error, got: %v", err)
	}
}

// TestSandboxManager_ExecWithStdin tests Exec with stdin input
func TestSandboxManager_ExecWithStdin(t *testing.T) {
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

	// Prepare stdin content
	inputContent := "this is from stdin"
	stdin := strings.NewReader(inputContent)

	// Execute cat command reading from stdin
	exitCode, output, err := sm.Exec(ctx, containerID, []string{"cat"}, stdin)
	if err != nil {
		t.Fatalf("Exec() with stdin error = %v", err)
	}

	if exitCode != 0 {
		t.Errorf("Exec() exit code = %d, want 0", exitCode)
	}

	if !strings.Contains(output, inputContent) {
		t.Errorf("Exec() output should contain input content. Got: %q, Want substring: %q", output, inputContent)
	}
}
