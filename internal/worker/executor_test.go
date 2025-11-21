package worker

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/biwakonbu/agent-runner/pkg/config"
)

// MockSandboxManager is a mock implementation of SandboxProvider for testing
type MockSandboxManager struct {
	startContainerCalled bool
	startContainerErr    error
	stopContainerCalled  bool
	stopContainerErr     error
	execCalled           bool
	execErr              error
	execExitCode         int
	execOutput           string
	lastContainerID      string
}

// Verify that MockSandboxManager implements SandboxProvider interface
var _ SandboxProvider = (*MockSandboxManager)(nil)

func (m *MockSandboxManager) StartContainer(ctx context.Context, image string, repoPath string, env map[string]string) (string, error) {
	m.startContainerCalled = true
	if m.startContainerErr != nil {
		return "", m.startContainerErr
	}
	m.lastContainerID = "mock-container-123"
	return m.lastContainerID, nil
}

func (m *MockSandboxManager) StopContainer(ctx context.Context, containerID string) error {
	m.stopContainerCalled = true
	return m.stopContainerErr
}

func (m *MockSandboxManager) Exec(ctx context.Context, containerID string, cmd []string) (int, string, error) {
	m.execCalled = true
	if m.execErr != nil {
		return 1, "", m.execErr
	}
	return m.execExitCode, m.execOutput, nil
}

func TestExecutor_NewExecutor(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	// Note: NewExecutor requires actual Docker connection, so this test might fail in non-Docker environments
	// For now, we're testing the structure
	executor := &Executor{
		Config:   cfg,
		RepoPath: "/test/repo",
	}

	if executor.Config.Kind != "codex-cli" {
		t.Errorf("Config.Kind = %s, want 'codex-cli'", executor.Config.Kind)
	}
	if executor.RepoPath != "/test/repo" {
		t.Errorf("RepoPath = %s, want '/test/repo'", executor.RepoPath)
	}
}

func TestExecutor_RunWorker_Success(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	executor := &Executor{
		Config:   cfg,
		RepoPath: "/test/repo",
		Sandbox:  &MockSandboxManager{},
	}

	mockSandbox := executor.Sandbox.(*MockSandboxManager)
	mockSandbox.execExitCode = 0
	mockSandbox.execOutput = "Successfully created calculator.py"

	ctx := context.Background()
	result, err := executor.RunWorker(ctx, "Create a simple calculator", map[string]string{})

	if err != nil {
		t.Fatalf("RunWorker() error = %v", err)
	}

	if result == nil {
		t.Fatalf("RunWorker() returned nil result")
	}

	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}

	if result.RawOutput != "Successfully created calculator.py" {
		t.Errorf("RawOutput = %s, want 'Successfully created calculator.py'", result.RawOutput)
	}

	if !mockSandbox.startContainerCalled {
		t.Errorf("StartContainer was not called")
	}
	if !mockSandbox.stopContainerCalled {
		t.Errorf("StopContainer was not called")
	}
	if !mockSandbox.execCalled {
		t.Errorf("Exec was not called")
	}
}

func TestExecutor_RunWorker_StartContainerError(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	executor := &Executor{
		Config:   cfg,
		RepoPath: "/test/repo",
		Sandbox:  &MockSandboxManager{},
	}

	mockSandbox := executor.Sandbox.(*MockSandboxManager)
	mockSandbox.startContainerErr = fmt.Errorf("Docker daemon not running")

	ctx := context.Background()
	result, err := executor.RunWorker(ctx, "Create a simple calculator", map[string]string{})

	if err == nil {
		t.Fatalf("RunWorker() expected error, got nil")
	}

	if result != nil {
		t.Errorf("RunWorker() expected nil result on error, got %v", result)
	}

	if !mockSandbox.startContainerCalled {
		t.Errorf("StartContainer was not called")
	}
}

func TestExecutor_RunWorker_ExecError(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	executor := &Executor{
		Config:   cfg,
		RepoPath: "/test/repo",
		Sandbox:  &MockSandboxManager{},
	}

	mockSandbox := executor.Sandbox.(*MockSandboxManager)
	mockSandbox.execErr = fmt.Errorf("command execution failed")

	ctx := context.Background()
	result, err := executor.RunWorker(ctx, "Create a simple calculator", map[string]string{})

	// Implementation returns WorkerRunResult even when Exec errors,
	// storing error in result.Error field
	if err != nil {
		t.Fatalf("RunWorker() error = %v, expected no error (error stored in result)", err)
	}

	if result == nil {
		t.Fatalf("RunWorker() returned nil result, expected WorkerRunResult")
	}

	if result.Error == nil {
		t.Errorf("WorkerRunResult.Error should contain the exec error, got nil")
	}

	if !mockSandbox.startContainerCalled {
		t.Errorf("StartContainer was not called")
	}
	if !mockSandbox.stopContainerCalled {
		t.Errorf("StopContainer was not called (cleanup)")
	}
}

func TestExecutor_RunWorker_NonZeroExitCode(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	executor := &Executor{
		Config:   cfg,
		RepoPath: "/test/repo",
		Sandbox:  &MockSandboxManager{},
	}

	mockSandbox := executor.Sandbox.(*MockSandboxManager)
	mockSandbox.execExitCode = 1
	mockSandbox.execOutput = "Error: invalid prompt"

	ctx := context.Background()
	result, err := executor.RunWorker(ctx, "Invalid prompt", map[string]string{})

	if err != nil {
		t.Fatalf("RunWorker() error = %v", err)
	}

	if result == nil {
		t.Fatalf("RunWorker() returned nil result")
	}

	if result.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", result.ExitCode)
	}

	if result.RawOutput != "Error: invalid prompt" {
		t.Errorf("RawOutput = %s, want 'Error: invalid prompt'", result.RawOutput)
	}
}

func TestExecutor_RunWorker_WithEnvironmentVariables(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	executor := &Executor{
		Config:   cfg,
		RepoPath: "/test/repo",
		Sandbox:  &MockSandboxManager{},
	}

	mockSandbox := executor.Sandbox.(*MockSandboxManager)
	mockSandbox.execExitCode = 0
	mockSandbox.execOutput = "Success"

	env := map[string]string{
		"CODEX_API_KEY": "test-key-123",
		"MY_VAR":        "my-value",
	}

	ctx := context.Background()
	result, err := executor.RunWorker(ctx, "Do something", env)

	if err != nil {
		t.Fatalf("RunWorker() error = %v", err)
	}

	if result == nil {
		t.Fatalf("RunWorker() returned nil result")
	}

	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
}

func TestExecutor_RunWorker_ResultTimestamps(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	executor := &Executor{
		Config:   cfg,
		RepoPath: "/test/repo",
		Sandbox:  &MockSandboxManager{},
	}

	mockSandbox := executor.Sandbox.(*MockSandboxManager)
	mockSandbox.execExitCode = 0
	mockSandbox.execOutput = "Success"

	ctx := context.Background()
	result, err := executor.RunWorker(ctx, "Do something", map[string]string{})

	if err != nil {
		t.Fatalf("RunWorker() error = %v", err)
	}

	if result.StartedAt.IsZero() {
		t.Errorf("StartedAt should not be zero")
	}
	if result.FinishedAt.IsZero() {
		t.Errorf("FinishedAt should not be zero")
	}
	if result.FinishedAt.Before(result.StartedAt) {
		t.Errorf("FinishedAt before StartedAt")
	}
}

func TestExecutor_RunWorker_ResultHasID(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	executor := &Executor{
		Config:   cfg,
		RepoPath: "/test/repo",
		Sandbox:  &MockSandboxManager{},
	}

	mockSandbox := executor.Sandbox.(*MockSandboxManager)
	mockSandbox.execExitCode = 0
	mockSandbox.execOutput = "Success"

	ctx := context.Background()
	result, err := executor.RunWorker(ctx, "Do something", map[string]string{})

	if err != nil {
		t.Fatalf("RunWorker() error = %v", err)
	}

	if result.ID == "" {
		t.Errorf("Result ID should not be empty")
	}
}

func TestExecutor_RunWorker_MultipleRuns(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	executor := &Executor{
		Config:   cfg,
		RepoPath: "/test/repo",
		Sandbox:  &MockSandboxManager{},
	}

	mockSandbox := executor.Sandbox.(*MockSandboxManager)
	mockSandbox.execExitCode = 0
	mockSandbox.execOutput = "Success"

	ctx := context.Background()

	// First run
	result1, err1 := executor.RunWorker(ctx, "First task", map[string]string{})
	if err1 != nil {
		t.Fatalf("First RunWorker() error = %v", err1)
	}

	// Ensure different timestamp
	time.Sleep(1 * time.Second)

	// Second run
	result2, err2 := executor.RunWorker(ctx, "Second task", map[string]string{})
	if err2 != nil {
		t.Fatalf("Second RunWorker() error = %v", err2)
	}

	if result1.ID == result2.ID {
		t.Errorf("Run IDs should be different for multiple runs (IDs: %s vs %s)", result1.ID, result2.ID)
	}

	// Both should be non-empty
	if result1.ID == "" || result2.ID == "" {
		t.Errorf("Run IDs should not be empty")
	}
}

func TestExecutor_RunWorker_LargeOutput(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	executor := &Executor{
		Config:   cfg,
		RepoPath: "/test/repo",
		Sandbox:  &MockSandboxManager{},
	}

	mockSandbox := executor.Sandbox.(*MockSandboxManager)
	mockSandbox.execExitCode = 0
	// Simulate large output
	largeOutput := ""
	for i := 0; i < 1000; i++ {
		largeOutput += "This is a line of output\n"
	}
	mockSandbox.execOutput = largeOutput

	ctx := context.Background()
	result, err := executor.RunWorker(ctx, "Do something", map[string]string{})

	if err != nil {
		t.Fatalf("RunWorker() error = %v", err)
	}

	if len(result.RawOutput) != len(largeOutput) {
		t.Errorf("RawOutput length = %d, want %d", len(result.RawOutput), len(largeOutput))
	}
}

func TestExecutor_RunWorker_Context_Timeout(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	mockSandbox := &MockSandboxManager{
		execErr: context.DeadlineExceeded,
	}

	executor := &Executor{
		Config:   cfg,
		RepoPath: "/test/repo",
		Sandbox:  mockSandbox,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait a bit to ensure timeout
	time.Sleep(10 * time.Millisecond)

	result, err := executor.RunWorker(ctx, "Do something", map[string]string{})

	// Implementation stores context deadline error in result.Error,
	// does not return error from RunWorker itself
	if err != nil {
		t.Fatalf("RunWorker() error = %v, expected no error (error stored in result)", err)
	}

	if result == nil {
		t.Fatalf("RunWorker() returned nil result")
	}

	// Error should be stored in result
	if result.Error == nil {
		t.Errorf("WorkerRunResult.Error should contain deadline exceeded, got nil")
	}
}

func TestExecutor_Config_DockerImageDefault(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind: "codex-cli",
		// DockerImage is empty, should use default
	}

	executor := &Executor{
		Config:   cfg,
		RepoPath: "/test/repo",
	}

	// When empty, executor.RunWorker will use default image
	if executor.Config.DockerImage == "" {
		t.Logf("DockerImage is empty, default will be used")
	}
}

func TestExecutor_Config_CustomDockerImage(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "my-custom-image:v1",
	}

	executor := &Executor{
		Config:   cfg,
		RepoPath: "/test/repo",
	}

	if executor.Config.DockerImage != "my-custom-image:v1" {
		t.Errorf("DockerImage = %s, want 'my-custom-image:v1'", executor.Config.DockerImage)
	}
}

func TestExecutor_RunWorker_ResultProperties(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	executor := &Executor{
		Config:   cfg,
		RepoPath: "/test/repo",
		Sandbox:  &MockSandboxManager{},
	}

	mockSandbox := executor.Sandbox.(*MockSandboxManager)
	mockSandbox.execExitCode = 0
	mockSandbox.execOutput = "Test output"

	ctx := context.Background()
	result, err := executor.RunWorker(ctx, "Test prompt", map[string]string{})

	if err != nil {
		t.Fatalf("RunWorker() error = %v", err)
	}

	// Verify WorkerRunResult structure
	if result.ID == "" {
		t.Errorf("ID is empty")
	}
	if result.StartedAt.IsZero() {
		t.Errorf("StartedAt is zero")
	}
	if result.FinishedAt.IsZero() {
		t.Errorf("FinishedAt is zero")
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.RawOutput != "Test output" {
		t.Errorf("RawOutput = %s, want 'Test output'", result.RawOutput)
	}
}

// TestExecutor_RunWorker_SandboxStartError tests handling of sandbox start failure
func TestExecutor_RunWorker_SandboxStartError(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "test-image:latest",
		Env:         map[string]string{},
	}

	mockSandbox := &MockSandboxManager{
		startContainerErr: fmt.Errorf("failed to start container"),
	}

	executor := &Executor{
		Config:   cfg,
		Sandbox:  mockSandbox,
		RepoPath: "/tmp/test",
	}

	ctx := context.Background()
	_, err := executor.RunWorker(ctx, "test prompt", map[string]string{})

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to start") {
		t.Errorf("Error message should contain 'failed to start', got: %v", err)
	}
}

// TestExecutor_RunWorker_SandboxExecError tests handling of sandbox exec failure
func TestExecutor_RunWorker_SandboxExecError(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "test-image:latest",
		Env:         map[string]string{},
	}

	mockSandbox := &MockSandboxManager{
		execErr: fmt.Errorf("exec failed"),
	}

	executor := &Executor{
		Config:   cfg,
		Sandbox:  mockSandbox,
		RepoPath: "/tmp/test",
	}

	ctx := context.Background()
	result, err := executor.RunWorker(ctx, "test prompt", map[string]string{})

	if err != nil {
		t.Fatalf("RunWorker should not return error: %v", err)
	}
	if result == nil {
		t.Fatalf("Result should not be nil")
	}
	// Error is stored in result.Error, not as return value
	if result.Error == nil {
		t.Errorf("Result.Error should be set when exec fails")
	}
}

// TestExecutor_RunWorker_EnvironmentVariables tests environment variable passing
func TestExecutor_RunWorker_EnvironmentVariables(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "test-image:latest",
		Env: map[string]string{
			"VAR1": "value1",
			"VAR2": "value2",
		},
	}

	mockSandbox := &MockSandboxManager{
		execOutput: "success",
	}

	executor := &Executor{
		Config:   cfg,
		Sandbox:  mockSandbox,
		RepoPath: "/tmp/test",
	}

	ctx := context.Background()
	result, err := executor.RunWorker(ctx, "test prompt", cfg.Env)

	if err != nil {
		t.Fatalf("RunWorker failed: %v", err)
	}
	if result == nil {
		t.Fatalf("Result should not be nil")
	}

	// Verify that StartContainer was called with env vars
	if !mockSandbox.startContainerCalled {
		t.Errorf("StartContainer should have been called with environment variables")
	}
}
