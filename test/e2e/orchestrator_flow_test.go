package e2e_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/biwakonbu/agent-runner/internal/ide"
	"github.com/biwakonbu/agent-runner/internal/orchestrator"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/ipc"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrchestratorFlow(t *testing.T) {
	// 1. Setup specific test workspace
	tempDir, err := os.MkdirTemp("", "multiverse-e2e-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Logf("Test Workspace: %s", tempDir)

	// 2. Build binaries (Orchestrator)
	// Assuming running from test/e2e
	wd, _ := os.Getwd()
	rootDir := filepath.Dir(filepath.Dir(wd)) // ../../

	orchBin := filepath.Join(tempDir, "multiverse-orchestrator")
	buildCmd := exec.Command("go", "build", "-o", orchBin, "./cmd/multiverse-orchestrator")
	buildCmd.Dir = rootDir
	out, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "Failed to build orchestrator: %s", string(out))

	// Mock Agent Runner
	mockRunner := filepath.Join(rootDir, "test/e2e/mock_runner.sh")

	// 3. Initialize Backend Components (Simulating App logic)
	// We use the internal packages directly to ensure the flow works from data perspective
	wsStore := ide.NewWorkspaceStore(filepath.Join(tempDir, "workspaces"))

	// Create a workspace
	projectRoot := filepath.Join(tempDir, "project")
	ws := &ide.Workspace{
		ProjectRoot: projectRoot,
		Version:     "1.0",
		DisplayName: "E2E Test Project",
	}
	err = wsStore.SaveWorkspace(ws)
	require.NoError(t, err)

	wsID := wsStore.GetWorkspaceID(projectRoot)
	wsDir := wsStore.GetWorkspaceDir(wsID)

	// Use Persistence Repo
	repo := persistence.NewWorkspaceRepository(wsDir)
	require.NoError(t, repo.Init())

	queue := ipc.NewFilesystemQueue(wsDir)
	scheduler := orchestrator.NewScheduler(repo, queue, nil)

	// 4. Create Task (Simulating IDE creates task)
	// In V2, we create TaskState + NodeDesign
	taskID := "e2e-task-1"
	nodeID := "node-1"

	tasksState := &persistence.TasksState{
		Tasks: []persistence.TaskState{
			{
				TaskID:    taskID,
				NodeID:    nodeID,
				Status:    string(orchestrator.TaskStatusPending),
				CreatedAt: time.Now(),
				Inputs: map[string]interface{}{
					"pool_id": "default",
				},
			},
		},
	}
	err = repo.State().SaveTasks(tasksState)
	require.NoError(t, err)

	err = repo.Design().SaveNode(&persistence.NodeDesign{
		NodeID: nodeID,
		Name:   "E2E Test Task",
	})
	require.NoError(t, err)

	// 5. Schedule Task (Simulating IDE schedules task)
	err = scheduler.ScheduleTask(taskID)
	require.NoError(t, err)

	// Verify task is READY (TaskState updated + Queued)
	// Note: ScheduleTask updates Status to READY and enqueues it.
	loadedState, err := repo.State().LoadTasks()
	require.NoError(t, err)

	var loadedTask *persistence.TaskState
	for _, t := range loadedState.Tasks {
		if t.TaskID == taskID {
			loadedTask = &t
			break
		}
	}
	require.NotNil(t, loadedTask)
	assert.Equal(t, string(orchestrator.TaskStatusReady), loadedTask.Status)

	// 6. Start Orchestrator Process
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, orchBin,
		"--workspace", wsDir,
		"--agent-runner", mockRunner,
		"--pool", "default")

	// Capture output for debugging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	require.NoError(t, err)

	// 7. Wait for completion (Polling)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(500 * time.Millisecond):
				state, err := repo.State().LoadTasks()
				if err == nil {
					// Find task
					for _, t := range state.Tasks {
						if t.TaskID == taskID {
							if t.Status == string(orchestrator.TaskStatusSucceeded) {
								done <- true
								return
							}
							if t.Status == string(orchestrator.TaskStatusFailed) {
								// Fail fast
								return
							}
						}
					}
				}
			}
		}
	}()

	select {
	case <-done:
		t.Log("Task succeeded!")
	case <-ctx.Done():
		t.Fatal("Timeout waiting for task success")
	}

	// Cleanup process
	_ = cmd.Process.Kill()
}
