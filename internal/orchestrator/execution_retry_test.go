package orchestrator

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/biwakonbu/agent-runner/internal/orchestrator/ipc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryPersistence_Integration(t *testing.T) {
	// Setup temporary workspace
	tmpDir, err := os.MkdirTemp("", "multiverse-retry-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Setup components
	taskStore := NewTaskStore(tmpDir)
	backlogStore := NewBacklogStore(tmpDir)
	queue := ipc.NewFilesystemQueue(tmpDir)
	scheduler := NewScheduler(taskStore, queue, nil)

	// Create Orchestrator with nil Executor/EventEmitter (not needed for this test)
	orch := NewExecutionOrchestrator(scheduler, nil, taskStore, queue, nil, backlogStore, []string{"default"})

	t.Run("HandleFailure persists retry state", func(t *testing.T) {
		// Create a task
		task := &Task{
			ID:        "task-retry-1",
			Title:     "Retry Test Task",
			Status:    TaskStatusRunning,
			CreatedAt: time.Now(),
			PoolID:    "default",
		}
		err := taskStore.SaveTask(task)
		require.NoError(t, err)

		// Simulate failure (Attempt 1)
		execErr := fmt.Errorf("simulated failure")
		err = orch.HandleFailure(task, execErr, 1)
		require.NoError(t, err)

		// Verify task state in store
		loadedTask, err := taskStore.LoadTask(task.ID)
		require.NoError(t, err)
		assert.Equal(t, TaskStatusRetryWait, loadedTask.Status)
		assert.Equal(t, 1, loadedTask.AttemptCount)
		assert.NotNil(t, loadedTask.NextRetryAt)
		// Default backoff is 5s, so NextRetryAt should be in the future
		assert.True(t, loadedTask.NextRetryAt.After(time.Now()))
	})

	t.Run("ResetRetryTasks resets tasks when time comes", func(t *testing.T) {
		// Create a task that is ready to retry (NextRetryAt in past)
		pastTime := time.Now().Add(-1 * time.Minute)
		task := &Task{
			ID:           "task-retry-2",
			Title:        "Reset Test Task",
			Status:       TaskStatusRetryWait,
			CreatedAt:    time.Now(),
			PoolID:       "default",
			AttemptCount: 1,
			NextRetryAt:  &pastTime,
		}
		err := taskStore.SaveTask(task)
		require.NoError(t, err)

		// Create another task that is NOT ready (NextRetryAt in future)
		futureTime := time.Now().Add(1 * time.Hour)
		futureTask := &Task{
			ID:           "task-retry-future",
			Title:        "Future Test Task",
			Status:       TaskStatusRetryWait,
			CreatedAt:    time.Now(),
			PoolID:       "default",
			AttemptCount: 1,
			NextRetryAt:  &futureTime,
		}
		err = taskStore.SaveTask(futureTask)
		require.NoError(t, err)

		// Execute ResetRetryTasks
		resetIDs, err := scheduler.ResetRetryTasks()
		require.NoError(t, err)

		// Verify result
		assert.Contains(t, resetIDs, "task-retry-2")
		assert.NotContains(t, resetIDs, "task-retry-future")

		// Verify DB state for ready task
		loadedTask, err := taskStore.LoadTask("task-retry-2")
		require.NoError(t, err)
		assert.Equal(t, TaskStatusPending, loadedTask.Status)
		assert.Nil(t, loadedTask.NextRetryAt)
		assert.Equal(t, 1, loadedTask.AttemptCount) // Count remains

		// Verify DB state for future task
		loadedFutureTask, err := taskStore.LoadTask("task-retry-future")
		require.NoError(t, err)
		assert.Equal(t, TaskStatusRetryWait, loadedFutureTask.Status)
		assert.NotNil(t, loadedFutureTask.NextRetryAt)
	})
}
