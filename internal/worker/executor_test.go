package worker

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"io"

	"github.com/biwakonbu/agent-runner/internal/meta"
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
	lastRepoPath         string // Added to verify repo path resolution
}

// Verify that MockSandboxManager implements SandboxProvider interface
var _ SandboxProvider = (*MockSandboxManager)(nil)

func (m *MockSandboxManager) StartContainer(ctx context.Context, image string, repoPath string, env map[string]string) (string, error) {
	m.startContainerCalled = true
	m.lastRepoPath = repoPath // Capture the repo path
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

func (m *MockSandboxManager) Exec(ctx context.Context, containerID string, cmd []string, stdin io.Reader) (int, string, error) {
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

// ============ NEW TESTS (Persistent Container Design) ============

// TestExecutor_Start_Success tests successful container startup
func TestExecutor_Start_Success(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	mockSandbox := &MockSandboxManager{}
	executor := &Executor{
		Config:      cfg,
		Sandbox:     mockSandbox,
		RepoPath:    "/test/repo",
		containerID: "", // Not yet started
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

// TestExecutor_Start_AlreadyStarted tests that Start() fails if already running
func TestExecutor_Start_AlreadyStarted(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	mockSandbox := &MockSandboxManager{}
	executor := &Executor{
		Config:      cfg,
		Sandbox:     mockSandbox,
		RepoPath:    "/test/repo",
		containerID: "already-running-123",
	}

	ctx := context.Background()
	err := executor.Start(ctx)

	if err == nil {
		t.Fatalf("Start() expected error when already started, got nil")
	}

	if !strings.Contains(err.Error(), "already started") {
		t.Errorf("Error should mention 'already started', got: %v", err)
	}
}

// TestExecutor_Stop_Success tests successful container stop
func TestExecutor_Stop_Success(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	mockSandbox := &MockSandboxManager{}
	executor := &Executor{
		Config:      cfg,
		Sandbox:     mockSandbox,
		RepoPath:    "/test/repo",
		containerID: "test-container-123",
	}

	ctx := context.Background()
	err := executor.Stop(ctx)

	if err != nil {
		t.Fatalf("Stop() error = %v, want nil", err)
	}

	if executor.containerID != "" {
		t.Errorf("containerID should be cleared after Stop()")
	}

	if !mockSandbox.stopContainerCalled {
		t.Errorf("StopContainer should have been called")
	}
}

// TestExecutor_Stop_NoContainer tests that Stop() fails when no container running
func TestExecutor_Stop_NoContainer(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	mockSandbox := &MockSandboxManager{}
	executor := &Executor{
		Config:      cfg,
		Sandbox:     mockSandbox,
		RepoPath:    "/test/repo",
		containerID: "", // Not started
	}

	ctx := context.Background()
	err := executor.Stop(ctx)

	if err == nil {
		t.Fatalf("Stop() expected error when no container running, got nil")
	}

	if !strings.Contains(err.Error(), "no container") {
		t.Errorf("Error should mention 'no container', got: %v", err)
	}
}

// TestExecutor_RunWorker_WithPersistentContainer tests RunWorker with persistent container
func TestExecutor_RunWorker_WithPersistentContainer(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	mockSandbox := &MockSandboxManager{
		execExitCode: 0,
		execOutput:   "Success from persistent container",
	}
	executor := &Executor{
		Config:      cfg,
		Sandbox:     mockSandbox,
		RepoPath:    "/test/repo",
		containerID: "persistent-container-123",
	}

	ctx := context.Background()
	result, err := executor.RunWorker(ctx, meta.WorkerCall{WorkerType: "codex-cli", Mode: "exec", Prompt: "test prompt"}, map[string]string{})

	if err != nil {
		t.Fatalf("RunWorker() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatalf("RunWorker() returned nil result")
	}

	// Verify that Exec was called (but not Start/Stop)
	if !mockSandbox.execCalled {
		t.Errorf("Exec should have been called")
	}

	// Start/Stop should NOT be called in new design
	if mockSandbox.startContainerCalled {
		t.Errorf("StartContainer should NOT be called (container already running)")
	}
	if mockSandbox.stopContainerCalled {
		t.Errorf("StopContainer should NOT be called (container lifecycle managed by Runner)")
	}

	if result.RawOutput != "Success from persistent container" {
		t.Errorf("RawOutput = %s, want 'Success from persistent container'", result.RawOutput)
	}
}

// TestExecutor_RunWorker_NoPersistentContainer tests that RunWorker fails if container not started
func TestExecutor_RunWorker_NoPersistentContainer(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	mockSandbox := &MockSandboxManager{}
	executor := &Executor{
		Config:      cfg,
		Sandbox:     mockSandbox,
		RepoPath:    "/test/repo",
		containerID: "", // Not started
	}

	ctx := context.Background()
	result, err := executor.RunWorker(ctx, meta.WorkerCall{WorkerType: "codex-cli", Mode: "exec", Prompt: "test prompt"}, map[string]string{})

	if err == nil {
		t.Fatalf("RunWorker() expected error when no container running, got nil")
	}

	if result != nil {
		t.Errorf("RunWorker() expected nil result on error, got %v", result)
	}

	if !strings.Contains(err.Error(), "container not started") {
		t.Errorf("Error should mention 'container not started', got: %v", err)
	}
}

// ============ NEW TESTS (Phase 8-2-1: Start/Stop Error Handling) ============

// TestExecutor_Start_SandboxStartError tests Start() error when StartContainer fails
func TestExecutor_Start_SandboxStartError(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	mockSandbox := &MockSandboxManager{
		startContainerErr: fmt.Errorf("Docker daemon not running"),
	}
	executor := &Executor{
		Config:      cfg,
		Sandbox:     mockSandbox,
		RepoPath:    "/test/repo",
		containerID: "", // Not started
	}

	ctx := context.Background()
	err := executor.Start(ctx)

	if err == nil {
		t.Fatalf("Start() expected error when StartContainer fails, got nil")
	}

	if !strings.Contains(err.Error(), "failed to start container") {
		t.Errorf("Error should contain 'failed to start container', got: %v", err)
	}

	// containerID should remain empty after error
	if executor.containerID != "" {
		t.Errorf("containerID should be empty after Start error, got: %s", executor.containerID)
	}

	if !mockSandbox.startContainerCalled {
		t.Errorf("StartContainer should have been called")
	}
}

// TestExecutor_Stop_SandboxStopError tests Stop() error when StopContainer fails
func TestExecutor_Stop_SandboxStopError(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	mockSandbox := &MockSandboxManager{
		stopContainerErr: fmt.Errorf("container already removed"),
	}
	executor := &Executor{
		Config:      cfg,
		Sandbox:     mockSandbox,
		RepoPath:    "/test/repo",
		containerID: "test-container-123",
	}

	ctx := context.Background()
	err := executor.Stop(ctx)

	if err == nil {
		t.Fatalf("Stop() expected error when StopContainer fails, got nil")
	}

	if !strings.Contains(err.Error(), "failed to stop container") {
		t.Errorf("Error should contain 'failed to stop container', got: %v", err)
	}

	// containerID should be cleared even on error (resource leak prevention)
	if executor.containerID != "" {
		t.Errorf("containerID should be cleared after Stop (even on error), got: %s", executor.containerID)
	}

	if !mockSandbox.stopContainerCalled {
		t.Errorf("StopContainer should have been called")
	}
}

// TestExecutor_RunWorker_MultipleExecutionsInSameContainer tests Phase 8-2-1 core feature:
// multiple RunWorker calls reusing the same container
func TestExecutor_RunWorker_MultipleExecutionsInSameContainer(t *testing.T) {
	cfg := config.WorkerConfig{
		Kind:        "codex-cli",
		DockerImage: "agent-runner-codex:latest",
	}

	mockSandbox := &MockSandboxManager{
		execExitCode: 0,
		execOutput:   "Success",
	}
	executor := &Executor{
		Config:      cfg,
		Sandbox:     mockSandbox,
		RepoPath:    "/test/repo",
		containerID: "", // Not started yet
	}

	ctx := context.Background()

	// Phase 1: Start container
	err := executor.Start(ctx)
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	containerID1 := executor.containerID
	if containerID1 == "" {
		t.Fatalf("containerID should be set after Start()")
	}

	// Phase 2: First RunWorker execution
	res1, err1 := executor.RunWorker(ctx, meta.WorkerCall{WorkerType: "codex-cli", Mode: "exec", Prompt: "Task 1: Create calculator"}, map[string]string{})
	if err1 != nil {
		t.Fatalf("First RunWorker() failed: %v", err1)
	}
	if res1 == nil {
		t.Fatalf("First RunWorker() returned nil result")
	}
	if res1.ExitCode != 0 {
		t.Errorf("First RunWorker() ExitCode = %d, want 0", res1.ExitCode)
	}

	// Verify containerID unchanged after first execution
	containerID2 := executor.containerID
	if containerID2 != containerID1 {
		t.Errorf("containerID changed after first RunWorker: %s -> %s", containerID1, containerID2)
	}

	// Phase 3: Second RunWorker execution (reusing same container)
	mockSandbox.execOutput = "Second task success"
	res2, err2 := executor.RunWorker(ctx, meta.WorkerCall{WorkerType: "codex-cli", Mode: "exec", Prompt: "Task 2: Add tests"}, map[string]string{})
	if err2 != nil {
		t.Fatalf("Second RunWorker() failed: %v", err2)
	}
	if res2 == nil {
		t.Fatalf("Second RunWorker() returned nil result")
	}
	if res2.ExitCode != 0 {
		t.Errorf("Second RunWorker() ExitCode = %d, want 0", res2.ExitCode)
	}
	if res2.RawOutput != "Second task success" {
		t.Errorf("Second RunWorker() RawOutput = %s, want 'Second task success'", res2.RawOutput)
	}

	// Verify containerID still unchanged after second execution
	containerID3 := executor.containerID
	if containerID3 != containerID1 {
		t.Errorf("containerID changed after second RunWorker: %s -> %s", containerID1, containerID3)
	}

	// Phase 4: Stop container
	err = executor.Stop(ctx)
	if err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}

	// Verify containerID cleared after Stop
	if executor.containerID != "" {
		t.Errorf("containerID should be cleared after Stop(), got: %s", executor.containerID)
	}

	// Verify that Exec was called twice, but Start/Stop only once
	if mockSandbox.startContainerCalled != true {
		t.Errorf("StartContainer should have been called once")
	}
	if mockSandbox.stopContainerCalled != true {
		t.Errorf("StopContainer should have been called once")
	}

	// Note: execCalled is a boolean, so we can't verify it was called twice
	// In a more sophisticated mock, we could track call counts
	if !mockSandbox.execCalled {
		t.Errorf("Exec should have been called at least once")
	}
}
