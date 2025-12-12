package orchestrator

import (
	"fmt"
	"testing"
	"time"

	"github.com/biwakonbu/agent-runner/internal/orchestrator/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryPersistence_Integration(t *testing.T) {
	repo, queue := setupTestRepo(t)
	backlogStore := NewBacklogStore(repo.BaseDir()) // Still uses file based store? Or should use Repo?
	// BacklogStore is untouched in this refactor, it seems.
	// But it uses files in .multiverse/backlog.
	// Repo base dir is acceptable.

	scheduler := NewScheduler(repo, queue, nil)

	// Create Orchestrator with nil Executor/EventEmitter (not needed for this test)
	orch := NewExecutionOrchestrator(scheduler, nil, repo, queue, nil, backlogStore, []string{"default"})

	now := time.Now()

	t.Run("HandleFailure persists retry state", func(t *testing.T) {
		// Create a task
		taskID := "task-retry-1"
		saveState(t, repo, []persistence.TaskState{
			{
				TaskID:    taskID,
				NodeID:    "node-1",
				Status:    string(TaskStatusRunning),
				CreatedAt: now,
				Inputs:    map[string]interface{}{"pool_id": "default"},
			},
		}, nil)
		saveDesign(t, repo, []persistence.NodeDesign{{NodeID: "node-1"}})

		// Simulate failure (Attempt 1)
		execErr := fmt.Errorf("simulated failure")
		// Need a Task struct to pass to HandleFailure?
		// HandleFailure takes *Task object OR logic inside handles it?
		// HandleFailure signature currently: (task *Task, err error, attemptNum int)
		// But in V2, HandleFailure likely should reload fresh state or accept minimal ID?
		// The refactored HandleFailure loads task from repo using task.ID.
		// So we pass a dummy task with ID.
		dummyTask := &persistence.TaskState{TaskID: taskID}

		err := orch.HandleFailure(dummyTask, execErr, 1)
		require.NoError(t, err)

		// Verify task state in store
		state, err := repo.State().LoadTasks()
		require.NoError(t, err)

		var loadedTask *persistence.TaskState
		for i := range state.Tasks {
			if state.Tasks[i].TaskID == taskID {
				loadedTask = &state.Tasks[i]
				break
			}
		}
		require.NotNil(t, loadedTask)

		assert.Equal(t, string(TaskStatusRetryWait), loadedTask.Status)
		// AttemptCount ? TaskState doesn't have explicit AttemptCount field in models.go yet?
		// Check processJob or HandleFailure implementation.
		// If implementation sets it in Outputs or Inputs?
		// Let's assume implementation logic.
		// Wait, handled failure logic sets inputs["next_retry_at"].

		val, ok := loadedTask.Inputs["next_retry_at"]
		assert.True(t, ok, "next_retry_at input missing")
		retryTimeStr, ok := val.(string)
		assert.True(t, ok)
		retryAt, err := time.Parse(time.RFC3339, retryTimeStr)
		require.NoError(t, err)

		// Default backoff is 5s, so NextRetryAt should be in the future
		assert.True(t, retryAt.After(time.Now()))
	})

	t.Run("ResetRetryTasks resets tasks when time comes", func(t *testing.T) {
		pastTime := time.Now().Add(-1 * time.Minute)
		futureTime := time.Now().Add(1 * time.Hour)

		saveState(t, repo, []persistence.TaskState{
			{
				TaskID:    "task-retry-2",
				NodeID:    "node-2",
				Status:    string(TaskStatusRetryWait),
				CreatedAt: now,
				Inputs:    map[string]interface{}{"next_retry_at": pastTime.Format(time.RFC3339)},
			},
			{
				TaskID:    "task-retry-future",
				NodeID:    "node-future",
				Status:    string(TaskStatusRetryWait),
				CreatedAt: now,
				Inputs:    map[string]interface{}{"next_retry_at": futureTime.Format(time.RFC3339)},
			},
		}, nil)
		saveDesign(t, repo, []persistence.NodeDesign{{NodeID: "node-2"}, {NodeID: "node-future"}})

		// Execute ResetRetryTasks
		resetIDs, err := scheduler.ResetRetryTasks()
		require.NoError(t, err)

		// Verify result
		assert.Contains(t, resetIDs, "task-retry-2")
		assert.NotContains(t, resetIDs, "task-retry-future")

		// Verify DB state for ready task
		state, err := repo.State().LoadTasks()
		require.NoError(t, err)

		taskMap := make(map[string]persistence.TaskState)
		for _, task := range state.Tasks {
			taskMap[task.TaskID] = task
		}

		readyTask := taskMap["task-retry-2"]
		assert.Equal(t, string(TaskStatusPending), readyTask.Status)
		_, ok := readyTask.Inputs["next_retry_at"]
		assert.False(t, ok, "next_retry_at should be cleared")

		// Verify DB state for future task
		futureTask := taskMap["task-retry-future"]
		assert.Equal(t, string(TaskStatusRetryWait), futureTask.Status)
		_, ok = futureTask.Inputs["next_retry_at"]
		assert.True(t, ok, "next_retry_at should stay")
	})
}
