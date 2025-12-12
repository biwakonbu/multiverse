package orchestrator

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/biwakonbu/agent-runner/internal/orchestrator/persistence"
	"github.com/stretchr/testify/assert"
)

func TestExecutorV2_Execute_Integration(t *testing.T) {
	// 1. Setup Repo
	tmpDir, err := os.MkdirTemp("", "executor_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	repo := persistence.NewWorkspaceRepository(tmpDir)
	err = repo.Init()
	assert.NoError(t, err)

	// Pre-seed task
	task := persistence.TaskState{
		TaskID: "task-exec-1",
		NodeID: "node-1",
		Kind:   "implementation",
		Status: "running",
		Inputs: map[string]interface{}{
			"goal": "Test goal",
		},
	}
	tasks := persistence.TasksState{
		Tasks: []persistence.TaskState{task},
	}
	err = repo.State().SaveTasks(&tasks)
	assert.NoError(t, err)

	// Verify pre-seed
	preTasks, err := repo.State().LoadTasks()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(preTasks.Tasks))

	// 2. Setup Executor with "echo" as agent
	// We use "sh" to emulate agent-runner
	logger := slog.Default()
	_ = NewExecutorV2("sh", tmpDir, repo, logger)

	// 3. Execute
	// We need to pass args to sh to make it behave like agent-runner (consume stdin, produce stdout)
	// But ExecutorV2 hardcodes args? No, it hardcodes `exec.CommandContext(ctx, e.AgentRunnerPath)`
	// Wait, if AgentRunnerPath is "sh", it waits for stdin? sh without args is interactive.
	// We need ExecutorV2 to allow args or use a wrapper script.
	// Simplest hack: create a wrapper script "agent-runner-mock" in tmpDir that ignores stdin and exits 0.

	mockRunnerPath := tmpDir + "/agent-runner-mock.sh"
	scriptContent := `#!/bin/sh
cat > /dev/null # Consume stdin
echo "Mock Output"
exit 0
`
	err = os.WriteFile(mockRunnerPath, []byte(scriptContent), 0755)
	assert.NoError(t, err)

	// Re-init with mock script
	executor := NewExecutorV2(mockRunnerPath, tmpDir, repo, logger)

	err = executor.Execute(context.TODO(), task)
	assert.NoError(t, err)

	// 4. Verify State Update
	updatedTasks, err := repo.State().LoadTasks()
	assert.NoError(t, err)
	if assert.NotEmpty(t, updatedTasks.Tasks) {
		assert.Equal(t, "succeeded", updatedTasks.Tasks[0].Status)
	}

	// 5. Verify History
	actions, _ := repo.History().ListActions(time.Time{}, time.Now())
	assert.Equal(t, 2, len(actions)) // started + succeeded
	assert.Equal(t, "task.attempt_started", actions[0].Kind)
	assert.Equal(t, "task.succeeded", actions[1].Kind)
	assert.Equal(t, "Mock Output", strings.TrimSpace(actions[1].Payload["output"].(string)))
}
