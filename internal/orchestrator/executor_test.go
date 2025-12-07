package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestExecutor_ExecuteTask_Cancellation verifies that canceling the context kills the process.
func TestExecutor_ExecuteTask_Cancellation(t *testing.T) {
	// 1. Create a dummy agent-runner script
	tmpDir := t.TempDir()
	mockRunnerPath := filepath.Join(tmpDir, "mock_runner.sh")

	// Script using exec to replace shell process so kill works on it directly
	scriptContent := `#!/bin/sh
exec sleep 10
`
	err := os.WriteFile(mockRunnerPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("failed to write mock runner: %v", err)
	}

	// 2. Setup Executor
	taskStore := NewTaskStore(tmpDir) // Reuse temp dir
	executor := NewExecutor(mockRunnerPath, tmpDir, taskStore)
	// Suppress logs during test or keep them for debug
	// executor.SetLogger(...)

	// 3. Create a Task
	task := &Task{
		ID:     "task-cancel-test",
		Title:  "Sleep Task",
		Status: TaskStatusPending,
		PoolID: "default",
	}
	// TaskStore requires file existence for SaveTask? It writes to .multiverse/tasks/ID.jsonl
	// NewTaskStore uses workspaceDir. checks tasks dir.
	// We need to ensure taskStore is usable.
	// task_store.go: SaveTask writes to file.
	_ = taskStore.SaveTask(task)

	// 4. Execute with cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cleanup

	started := make(chan struct{})
	done := make(chan error)

	go func() {
		close(started)
		_, err := executor.ExecuteTask(ctx, task)
		done <- err
	}()

	<-started
	// Wait a bit to let the process start
	time.Sleep(500 * time.Millisecond)

	// 5. Cancel context
	cancel()

	// 6. Wait for result
	select {
	case err := <-done:
		// ExecuteTask returns error on cancellation
		assert.Error(t, err)
		// Check error message or type if possible, usually "signal: killed" or "context canceled"
		// exec.CommandContext returns the error from Wait.
		t.Logf("ExecuteTask returned: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("ExecuteTask did not return after cancellation")
	}

	// 7. Verify Task Attempt marks as Failed (or handled by Executor)
	// Executor logic: if err != nil -> AttemptStatusFailed
	attempts, _ := taskStore.ListAttemptsByTaskID(task.ID)
	if assert.NotEmpty(t, attempts) {
		assert.Equal(t, AttemptStatusFailed, attempts[0].Status)
	}
}
