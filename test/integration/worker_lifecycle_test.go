package integration_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"io"

	"github.com/biwakonbu/agent-runner/internal/meta"
	"github.com/biwakonbu/agent-runner/internal/worker"
	"github.com/biwakonbu/agent-runner/pkg/config"
)

// MockSandboxForLifecycle is a mock implementation of SandboxProvider for lifecycle testing
type MockSandboxForLifecycle struct {
	startCalled   bool
	stopCalled    bool
	execCalled    bool
	startErr      error
	stopErr       error
	execErr       error
	execExitCode  int
	execOutput    string
	containerID   string
	execCallCount int
}

var _ worker.SandboxProvider = (*MockSandboxForLifecycle)(nil)

func (m *MockSandboxForLifecycle) StartContainer(ctx context.Context, image string, repoPath string, env map[string]string) (string, error) {
	m.startCalled = true
	if m.startErr != nil {
		return "", m.startErr
	}
	m.containerID = "mock-container-lifecycle-123"
	return m.containerID, nil
}

func (m *MockSandboxForLifecycle) StopContainer(ctx context.Context, containerID string) error {
	m.stopCalled = true
	return m.stopErr
}

func (m *MockSandboxForLifecycle) Exec(ctx context.Context, containerID string, cmd []string, stdin io.Reader) (int, string, error) {
	m.execCalled = true
	m.execCallCount++
	if m.execErr != nil {
		return 1, "", m.execErr
	}
	return m.execExitCode, m.execOutput, nil
}

// TestWorkerLifecycle_StartStopSuccess tests normal Start/Stop lifecycle
func TestWorkerLifecycle_StartStopSuccess(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	mockSandbox := &MockSandboxForLifecycle{}
	executor := &worker.Executor{
		Config:   cfg,
		Sandbox:  mockSandbox,
		RepoPath: "/test/repo",
	}

	ctx := context.Background()

	// Start container
	err := executor.Start(ctx)
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	if !mockSandbox.startCalled {
		t.Error("StartContainer should have been called")
	}

	// Stop container
	err = executor.Stop(ctx)
	if err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}

	if !mockSandbox.stopCalled {
		t.Error("StopContainer should have been called")
	}
}

// TestWorkerLifecycle_MultipleRunsInSameContainer tests multiple RunWorker calls in same container
func TestWorkerLifecycle_MultipleRunsInSameContainer(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	mockSandbox := &MockSandboxForLifecycle{
		execExitCode: 0,
		execOutput:   "Success",
	}
	executor := &worker.Executor{
		Config:   cfg,
		Sandbox:  mockSandbox,
		RepoPath: "/test/repo",
	}

	ctx := context.Background()

	// Start container
	err := executor.Start(ctx)
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// First RunWorker
	result1, err := executor.RunWorker(ctx, meta.WorkerCall{WorkerType: "codex-cli", Mode: "exec", Prompt: "Task 1"}, map[string]string{})
	if err != nil {
		t.Fatalf("First RunWorker() failed: %v", err)
	}
	if result1.ExitCode != 0 {
		t.Errorf("First RunWorker() ExitCode = %d, want 0", result1.ExitCode)
	}

	// Second RunWorker (same container)
	mockSandbox.execOutput = "Second success"
	result2, err := executor.RunWorker(ctx, meta.WorkerCall{WorkerType: "codex-cli", Mode: "exec", Prompt: "Task 2"}, map[string]string{})
	if err != nil {
		t.Fatalf("Second RunWorker() failed: %v", err)
	}
	if result2.ExitCode != 0 {
		t.Errorf("Second RunWorker() ExitCode = %d, want 0", result2.ExitCode)
	}
	if result2.RawOutput != "Second success" {
		t.Errorf("Second RunWorker() RawOutput = %s, want 'Second success'", result2.RawOutput)
	}

	// Verify Exec was called twice
	if mockSandbox.execCallCount != 2 {
		t.Errorf("Exec should have been called 2 times, got %d", mockSandbox.execCallCount)
	}

	// Stop container
	err = executor.Stop(ctx)
	if err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}
}

// TestWorkerLifecycle_StopWithoutStart tests Stop() error when container not started
func TestWorkerLifecycle_StopWithoutStart(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	mockSandbox := &MockSandboxForLifecycle{}
	executor := &worker.Executor{
		Config:   cfg,
		Sandbox:  mockSandbox,
		RepoPath: "/test/repo",
	}

	ctx := context.Background()

	// Try to stop without starting
	err := executor.Stop(ctx)
	if err == nil {
		t.Fatal("Stop() should fail when container not started")
	}

	if !strings.Contains(err.Error(), "no container") {
		t.Errorf("Error should mention 'no container', got: %v", err)
	}

	if mockSandbox.stopCalled {
		t.Error("StopContainer should NOT have been called")
	}
}

// TestWorkerLifecycle_StartTwice tests Start() error when already started
func TestWorkerLifecycle_StartTwice(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	mockSandbox := &MockSandboxForLifecycle{}
	executor := &worker.Executor{
		Config:   cfg,
		Sandbox:  mockSandbox,
		RepoPath: "/test/repo",
	}

	ctx := context.Background()

	// First Start
	err := executor.Start(ctx)
	if err != nil {
		t.Fatalf("First Start() failed: %v", err)
	}

	// Second Start (should fail)
	err = executor.Start(ctx)
	if err == nil {
		t.Fatal("Second Start() should fail when already started")
	}

	if !strings.Contains(err.Error(), "already started") {
		t.Errorf("Error should mention 'already started', got: %v", err)
	}
}

// TestWorkerLifecycle_RunWorkerWithoutStart tests RunWorker() error when container not started
func TestWorkerLifecycle_RunWorkerWithoutStart(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	mockSandbox := &MockSandboxForLifecycle{}
	executor := &worker.Executor{
		Config:   cfg,
		Sandbox:  mockSandbox,
		RepoPath: "/test/repo",
	}

	ctx := context.Background()

	// Try to run worker without starting container
	result, err := executor.RunWorker(ctx, meta.WorkerCall{WorkerType: "codex-cli", Mode: "exec", Prompt: "Task"}, map[string]string{})
	if err == nil {
		t.Fatal("RunWorker() should fail when container not started")
	}

	if result != nil {
		t.Errorf("RunWorker() should return nil result on error, got %v", result)
	}

	if !strings.Contains(err.Error(), "container not started") {
		t.Errorf("Error should mention 'container not started', got: %v", err)
	}

	if mockSandbox.execCalled {
		t.Error("Exec should NOT have been called")
	}
}

// TestWorkerLifecycle_StartFailure tests Start() error handling
func TestWorkerLifecycle_StartFailure(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	mockSandbox := &MockSandboxForLifecycle{
		startErr: fmt.Errorf("Docker daemon not running"),
	}
	executor := &worker.Executor{
		Config:   cfg,
		Sandbox:  mockSandbox,
		RepoPath: "/test/repo",
	}

	ctx := context.Background()

	// Start should fail
	err := executor.Start(ctx)
	if err == nil {
		t.Fatal("Start() should fail when StartContainer fails")
	}

	if !strings.Contains(err.Error(), "failed to start container") {
		t.Errorf("Error should contain 'failed to start container', got: %v", err)
	}
}

// TestWorkerLifecycle_StopFailure tests Stop() error handling
func TestWorkerLifecycle_StopFailure(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	mockSandbox := &MockSandboxForLifecycle{
		stopErr: fmt.Errorf("container already removed"),
	}
	executor := &worker.Executor{
		Config:   cfg,
		Sandbox:  mockSandbox,
		RepoPath: "/test/repo",
	}

	ctx := context.Background()

	// Start container first
	err := executor.Start(ctx)
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Stop should fail but still clear containerID
	err = executor.Stop(ctx)
	if err == nil {
		t.Fatal("Stop() should fail when StopContainer fails")
	}

	if !strings.Contains(err.Error(), "failed to stop container") {
		t.Errorf("Error should contain 'failed to stop container', got: %v", err)
	}
}
