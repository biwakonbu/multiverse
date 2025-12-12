package orchestrator

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/biwakonbu/agent-runner/internal/orchestrator/ipc"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/persistence"
)

func setupTestRepo(t *testing.T) (persistence.WorkspaceRepository, *ipc.FilesystemQueue) {
	tmpDir := t.TempDir()
	repo := persistence.NewWorkspaceRepository(tmpDir)
	if err := repo.Init(); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}
	queue := ipc.NewFilesystemQueue(tmpDir)
	return repo, queue
}

func saveState(t *testing.T, repo persistence.WorkspaceRepository, tasks []persistence.TaskState, nodes []persistence.NodeRuntime) {
	if tasks != nil {
		if err := repo.State().SaveTasks(&persistence.TasksState{Tasks: tasks}); err != nil {
			t.Fatalf("failed to save tasks: %v", err)
		}
	} else {
		// Ensure empty state exists
		repo.State().SaveTasks(&persistence.TasksState{Tasks: []persistence.TaskState{}})
	}

	if nodes != nil {
		if err := repo.State().SaveNodesRuntime(&persistence.NodesRuntime{Nodes: nodes}); err != nil {
			t.Fatalf("failed to save nodes runtime: %v", err)
		}
	} else {
		repo.State().SaveNodesRuntime(&persistence.NodesRuntime{Nodes: []persistence.NodeRuntime{}})
	}
}

func saveDesign(t *testing.T, repo persistence.WorkspaceRepository, nodes []persistence.NodeDesign) {
	// We need to save nodes individually or via WBS?
	// Repo.Design().SaveNode() is available.
	for _, n := range nodes {
		if err := repo.Design().SaveNode(&n); err != nil {
			t.Fatalf("failed to save node design: %v", err)
		}
	}
}

func TestScheduler_ScheduleTask_WithDependencies(t *testing.T) {
	repo, queue := setupTestRepo(t)
	scheduler := NewScheduler(repo, queue, nil)

	now := time.Now()

	// Design: MainNode depends on DepNode
	saveDesign(t, repo, []persistence.NodeDesign{
		{NodeID: "dep-node", Dependencies: []string{}},
		{NodeID: "main-node", Dependencies: []string{"dep-node"}},
	})

	// Runtime: DepNode is NOT implemented yet
	saveState(t, repo, []persistence.TaskState{
		{TaskID: "main-task", NodeID: "main-node", Status: string(TaskStatusPending), CreatedAt: now},
	}, []persistence.NodeRuntime{
		{NodeID: "dep-node", Status: "in_progress"},
	})

	// Schedule should fail and block
	err := scheduler.ScheduleTask("main-task")
	if err == nil {
		t.Error("expected error for unsatisfied dependencies")
	}

	// Verify BLOCKED
	state, _ := repo.State().LoadTasks()
	if state.Tasks[0].Status != string(TaskStatusBlocked) {
		t.Errorf("expected status BLOCKED, got %s", state.Tasks[0].Status)
	}
}

func TestScheduler_ScheduleTask_SatisfiedDependencies(t *testing.T) {
	repo, queue := setupTestRepo(t)
	scheduler := NewScheduler(repo, queue, nil)

	now := time.Now()

	// Design: MainNode depends on DepNode
	saveDesign(t, repo, []persistence.NodeDesign{
		{NodeID: "dep-node", Dependencies: []string{}},
		{NodeID: "main-node", Dependencies: []string{"dep-node"}},
	})

	// Runtime: DepNode IS implemented
	saveState(t, repo, []persistence.TaskState{
		{TaskID: "main-task", NodeID: "main-node", Status: string(TaskStatusPending), CreatedAt: now},
	}, []persistence.NodeRuntime{
		{NodeID: "dep-node", Status: "implemented"},
	})

	// Schedule should succeed
	err := scheduler.ScheduleTask("main-task")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify READY
	state, _ := repo.State().LoadTasks()
	if state.Tasks[0].Status != string(TaskStatusReady) {
		t.Errorf("expected status READY, got %s", state.Tasks[0].Status)
	}
}

func TestScheduler_ScheduleTask_NoDependencies(t *testing.T) {
	repo, queue := setupTestRepo(t)
	// Suppress logs for cleaner test output
	scheduler := NewScheduler(repo, queue, nil)
	scheduler.logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	now := time.Now()

	saveDesign(t, repo, []persistence.NodeDesign{
		{NodeID: "node-1", Dependencies: []string{}},
	})
	saveState(t, repo, []persistence.TaskState{
		{TaskID: "task-1", NodeID: "node-1", Status: string(TaskStatusPending), CreatedAt: now},
	}, nil)

	err := scheduler.ScheduleTask("task-1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	state, _ := repo.State().LoadTasks()
	if state.Tasks[0].Status != string(TaskStatusReady) {
		t.Errorf("expected status READY, got %s", state.Tasks[0].Status)
	}
}

func TestScheduler_ScheduleReadyTasks(t *testing.T) {
	repo, queue := setupTestRepo(t)
	scheduler := NewScheduler(repo, queue, nil)

	now := time.Now()

	// task-1: node-1 (no deps)
	// task-2: node-2 (deps node-1)
	// task-3: node-3 (deps node-2) -- node-2 not ready
	saveDesign(t, repo, []persistence.NodeDesign{
		{NodeID: "node-1", Dependencies: []string{}},
		{NodeID: "node-2", Dependencies: []string{"node-1"}},
		{NodeID: "node-3", Dependencies: []string{"node-2"}},
	})

	// Runtime: node-1 implemented. node-2 pending.
	saveState(t, repo, []persistence.TaskState{
		{TaskID: "task-1", NodeID: "node-1", Status: string(TaskStatusPending) /* logic check: if node-1 implemented, task-1 should ideally be done, but for test logic assume task-1 is pending for re-run? OR maybe this test checks task-2 readiness? */, CreatedAt: now},
		{TaskID: "task-2", NodeID: "node-2", Status: string(TaskStatusPending), CreatedAt: now},
		{TaskID: "task-3", NodeID: "node-3", Status: string(TaskStatusPending), CreatedAt: now},
	}, []persistence.NodeRuntime{
		{NodeID: "node-1", Status: "implemented"},
		{NodeID: "node-2", Status: "planned"},
	})

	// ScheduleReadyTasks
	// task-1: deps=[] -> Ready (Logic: pending task with satisfied deps)
	// task-2: deps=[node-1] -> node-1 implemented -> Ready
	// task-3: deps=[node-2] -> node-2 planned (not implemented) -> Not Ready
	scheduled, err := scheduler.ScheduleReadyTasks()
	if err != nil {
		t.Fatalf("ScheduleReadyTasks failed: %v", err)
	}

	// Expect task-1 and task-2
	// Wait, task-1 has no deps, so it is satisfied.
	// task-2 has node-1 dep, node-1 is implemented, so satisfied.
	if len(scheduled) != 2 {
		t.Errorf("expected 2 scheduled tasks, got %d: %v", len(scheduled), scheduled)
	}

	state, _ := repo.State().LoadTasks()
	statusMap := make(map[string]string)
	for _, t := range state.Tasks {
		statusMap[t.TaskID] = t.Status
	}

	if statusMap["task-1"] != string(TaskStatusReady) {
		t.Errorf("expected task-1 READY")
	}
	if statusMap["task-2"] != string(TaskStatusReady) {
		t.Errorf("expected task-2 READY")
	}
	if statusMap["task-3"] == string(TaskStatusReady) {
		t.Errorf("expected task-3 NOT READY")
	}
}

func TestScheduler_UpdateBlockedTasks(t *testing.T) {
	repo, queue := setupTestRepo(t)
	scheduler := NewScheduler(repo, queue, nil)

	now := time.Now()

	saveDesign(t, repo, []persistence.NodeDesign{
		{NodeID: "node-dep", Dependencies: []string{}},
		{NodeID: "node-blocked", Dependencies: []string{"node-dep"}},
	})

	// Case 1: Dependency NOT met
	saveState(t, repo, []persistence.TaskState{
		{TaskID: "blocked-task", NodeID: "node-blocked", Status: string(TaskStatusBlocked), CreatedAt: now},
	}, []persistence.NodeRuntime{
		{NodeID: "node-dep", Status: "in_progress"},
	})

	unblocked, _ := scheduler.UpdateBlockedTasks()
	if len(unblocked) != 0 {
		t.Errorf("expected 0 unblocked, got %d", len(unblocked))
	}

	// Case 2: Dependency MET
	saveState(t, repo, []persistence.TaskState{
		{TaskID: "blocked-task", NodeID: "node-blocked", Status: string(TaskStatusBlocked), CreatedAt: now},
	}, []persistence.NodeRuntime{
		{NodeID: "node-dep", Status: "implemented"},
	})

	unblocked, _ = scheduler.UpdateBlockedTasks()
	if len(unblocked) != 1 {
		t.Errorf("expected 1 unblocked, got %d", len(unblocked))
	}

	state, _ := repo.State().LoadTasks()
	if state.Tasks[0].Status != string(TaskStatusPending) {
		t.Errorf("expected status PENDING, got %s", state.Tasks[0].Status)
	}
}

func TestScheduler_SetBlockedStatusForPendingWithUnsatisfiedDeps(t *testing.T) {
	repo, queue := setupTestRepo(t)
	scheduler := NewScheduler(repo, queue, nil)

	now := time.Now()

	saveDesign(t, repo, []persistence.NodeDesign{
		{NodeID: "node-1", Dependencies: []string{}},
		{NodeID: "node-2", Dependencies: []string{"node-1"}},
	})

	// task-1 (node-1): satisfied (no deps)
	// task-2 (node-2): unsatisfied (node-1 not implemented)
	saveState(t, repo, []persistence.TaskState{
		{TaskID: "task-1", NodeID: "node-1", Status: string(TaskStatusPending), CreatedAt: now},
		{TaskID: "task-2", NodeID: "node-2", Status: string(TaskStatusPending), CreatedAt: now},
	}, []persistence.NodeRuntime{
		{NodeID: "node-1", Status: "in_progress"},
	})

	blocked, err := scheduler.SetBlockedStatusForPendingWithUnsatisfiedDeps()
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	if len(blocked) != 1 || blocked[0] != "task-2" {
		t.Errorf("expected task-2 blocked, got %v", blocked)
	}

	state, _ := repo.State().LoadTasks()
	for _, task := range state.Tasks {
		if task.TaskID == "task-1" && task.Status != string(TaskStatusPending) {
			t.Errorf("task-1 should coincide PENDING")
		}
		if task.TaskID == "task-2" && task.Status != string(TaskStatusBlocked) {
			t.Errorf("task-2 should be BLOCKED")
		}
	}
}
