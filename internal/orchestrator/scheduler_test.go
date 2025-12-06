package orchestrator

import (
	"testing"
	"time"

	"github.com/biwakonbu/agent-runner/internal/orchestrator/ipc"
)

func TestScheduler_ScheduleTask_WithDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)
	queue := ipc.NewFilesystemQueue(tmpDir)
	scheduler := NewScheduler(store, queue)

	now := time.Now()

	// 依存元タスク（未完了）
	depTask := &Task{
		ID:        "dep-task",
		Title:     "Dependency Task",
		Status:    TaskStatusPending,
		PoolID:    "default",
		CreatedAt: now,
	}

	// 依存先タスク
	mainTask := &Task{
		ID:           "main-task",
		Title:        "Main Task",
		Status:       TaskStatusPending,
		PoolID:       "default",
		CreatedAt:    now,
		Dependencies: []string{"dep-task"},
	}

	if err := store.SaveTask(depTask); err != nil {
		t.Fatalf("failed to save dep task: %v", err)
	}
	if err := store.SaveTask(mainTask); err != nil {
		t.Fatalf("failed to save main task: %v", err)
	}

	// 依存が満たされていないのでスケジュール失敗＆BLOCKED状態に
	err := scheduler.ScheduleTask("main-task")
	if err == nil {
		t.Error("expected error for unsatisfied dependencies")
	}

	// タスクがBLOCKED状態になっているか確認
	task, err := store.LoadTask("main-task")
	if err != nil {
		t.Fatalf("failed to load task: %v", err)
	}
	if task.Status != TaskStatusBlocked {
		t.Errorf("expected status BLOCKED, got %s", task.Status)
	}
}

func TestScheduler_ScheduleTask_SatisfiedDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)
	queue := ipc.NewFilesystemQueue(tmpDir)
	scheduler := NewScheduler(store, queue)

	now := time.Now()

	// 依存元タスク（完了済み）
	depTask := &Task{
		ID:        "dep-task",
		Title:     "Dependency Task",
		Status:    TaskStatusSucceeded,
		PoolID:    "default",
		CreatedAt: now,
	}

	// 依存先タスク
	mainTask := &Task{
		ID:           "main-task",
		Title:        "Main Task",
		Status:       TaskStatusPending,
		PoolID:       "default",
		CreatedAt:    now,
		Dependencies: []string{"dep-task"},
	}

	if err := store.SaveTask(depTask); err != nil {
		t.Fatalf("failed to save dep task: %v", err)
	}
	if err := store.SaveTask(mainTask); err != nil {
		t.Fatalf("failed to save main task: %v", err)
	}

	// 依存が満たされているのでスケジュール成功
	err := scheduler.ScheduleTask("main-task")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// タスクがREADY状態になっているか確認
	task, err := store.LoadTask("main-task")
	if err != nil {
		t.Fatalf("failed to load task: %v", err)
	}
	if task.Status != TaskStatusReady {
		t.Errorf("expected status READY, got %s", task.Status)
	}
}

func TestScheduler_ScheduleTask_NoDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)
	queue := ipc.NewFilesystemQueue(tmpDir)
	scheduler := NewScheduler(store, queue)

	now := time.Now()

	task := &Task{
		ID:        "task-1",
		Title:     "Task 1",
		Status:    TaskStatusPending,
		PoolID:    "default",
		CreatedAt: now,
	}

	if err := store.SaveTask(task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	// 依存なしなのでスケジュール成功
	err := scheduler.ScheduleTask("task-1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// タスクがREADY状態になっているか確認
	loaded, err := store.LoadTask("task-1")
	if err != nil {
		t.Fatalf("failed to load task: %v", err)
	}
	if loaded.Status != TaskStatusReady {
		t.Errorf("expected status READY, got %s", loaded.Status)
	}
}

func TestScheduler_ScheduleReadyTasks(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)
	queue := ipc.NewFilesystemQueue(tmpDir)
	scheduler := NewScheduler(store, queue)

	now := time.Now()

	tasks := []*Task{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{}},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"task-1"}},
		{ID: "task-3", Title: "Task 3", Status: TaskStatusSucceeded, PoolID: "default", CreatedAt: now, Dependencies: []string{}},
		{ID: "task-4", Title: "Task 4", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"task-3"}},
	}

	for _, task := range tasks {
		if err := store.SaveTask(task); err != nil {
			t.Fatalf("failed to save task: %v", err)
		}
	}

	// Ready なタスクをスケジュール
	scheduled, err := scheduler.ScheduleReadyTasks()
	if err != nil {
		t.Fatalf("ScheduleReadyTasks failed: %v", err)
	}

	// task-1 と task-4 がスケジュールされるはず
	if len(scheduled) != 2 {
		t.Errorf("expected 2 scheduled tasks, got %d: %v", len(scheduled), scheduled)
	}

	// task-1 がREADYになっているか確認
	task1, _ := store.LoadTask("task-1")
	if task1.Status != TaskStatusReady {
		t.Errorf("expected task-1 status READY, got %s", task1.Status)
	}

	// task-4 がREADYになっているか確認
	task4, _ := store.LoadTask("task-4")
	if task4.Status != TaskStatusReady {
		t.Errorf("expected task-4 status READY, got %s", task4.Status)
	}
}

func TestScheduler_UpdateBlockedTasks(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)
	queue := ipc.NewFilesystemQueue(tmpDir)
	scheduler := NewScheduler(store, queue)

	now := time.Now()

	// 依存元タスク（最初は PENDING）
	depTask := &Task{
		ID:        "dep-task",
		Title:     "Dependency Task",
		Status:    TaskStatusPending,
		PoolID:    "default",
		CreatedAt: now,
	}

	// BLOCKED 状態のタスク
	blockedTask := &Task{
		ID:           "blocked-task",
		Title:        "Blocked Task",
		Status:       TaskStatusBlocked,
		PoolID:       "default",
		CreatedAt:    now,
		Dependencies: []string{"dep-task"},
	}

	if err := store.SaveTask(depTask); err != nil {
		t.Fatalf("failed to save dep task: %v", err)
	}
	if err := store.SaveTask(blockedTask); err != nil {
		t.Fatalf("failed to save blocked task: %v", err)
	}

	// 依存が満たされていないので unblock されない
	unblocked, err := scheduler.UpdateBlockedTasks()
	if err != nil {
		t.Fatalf("UpdateBlockedTasks failed: %v", err)
	}
	if len(unblocked) != 0 {
		t.Errorf("expected 0 unblocked tasks, got %d", len(unblocked))
	}

	// 依存元を完了させる
	depTask.Status = TaskStatusSucceeded
	if err := store.SaveTask(depTask); err != nil {
		t.Fatalf("failed to update dep task: %v", err)
	}

	// 今度は unblock される
	unblocked, err = scheduler.UpdateBlockedTasks()
	if err != nil {
		t.Fatalf("UpdateBlockedTasks failed: %v", err)
	}
	if len(unblocked) != 1 {
		t.Errorf("expected 1 unblocked task, got %d", len(unblocked))
	}

	// タスクが PENDING に戻っているか確認
	task, _ := store.LoadTask("blocked-task")
	if task.Status != TaskStatusPending {
		t.Errorf("expected status PENDING, got %s", task.Status)
	}
}

func TestScheduler_SetBlockedStatusForPendingWithUnsatisfiedDeps(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)
	queue := ipc.NewFilesystemQueue(tmpDir)
	scheduler := NewScheduler(store, queue)

	now := time.Now()

	tasks := []*Task{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{}},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"task-1"}},
		{ID: "task-3", Title: "Task 3", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"task-1", "task-2"}},
	}

	for _, task := range tasks {
		if err := store.SaveTask(task); err != nil {
			t.Fatalf("failed to save task: %v", err)
		}
	}

	// 依存が満たされていない PENDING タスクを BLOCKED に設定
	blocked, err := scheduler.SetBlockedStatusForPendingWithUnsatisfiedDeps()
	if err != nil {
		t.Fatalf("SetBlockedStatusForPendingWithUnsatisfiedDeps failed: %v", err)
	}

	// task-2 と task-3 が BLOCKED になるはず
	if len(blocked) != 2 {
		t.Errorf("expected 2 blocked tasks, got %d: %v", len(blocked), blocked)
	}

	// task-1 は依存なしなので PENDING のまま
	task1, _ := store.LoadTask("task-1")
	if task1.Status != TaskStatusPending {
		t.Errorf("expected task-1 status PENDING, got %s", task1.Status)
	}

	// task-2 は BLOCKED
	task2, _ := store.LoadTask("task-2")
	if task2.Status != TaskStatusBlocked {
		t.Errorf("expected task-2 status BLOCKED, got %s", task2.Status)
	}

	// task-3 は BLOCKED
	task3, _ := store.LoadTask("task-3")
	if task3.Status != TaskStatusBlocked {
		t.Errorf("expected task-3 status BLOCKED, got %s", task3.Status)
	}
}

func TestScheduler_AllDependenciesSatisfied(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)
	queue := ipc.NewFilesystemQueue(tmpDir)
	scheduler := NewScheduler(store, queue)

	now := time.Now()

	t.Run("no dependencies", func(t *testing.T) {
		task := &Task{
			ID:           "task-no-deps",
			Title:        "No Dependencies",
			Status:       TaskStatusPending,
			PoolID:       "default",
			CreatedAt:    now,
			Dependencies: []string{},
		}
		if err := store.SaveTask(task); err != nil {
			t.Fatalf("failed to save task: %v", err)
		}

		if !scheduler.allDependenciesSatisfied(task) {
			t.Error("expected allDependenciesSatisfied=true for task with no dependencies")
		}
	})

	t.Run("all dependencies succeeded", func(t *testing.T) {
		dep1 := &Task{ID: "dep1", Title: "Dep 1", Status: TaskStatusSucceeded, PoolID: "default", CreatedAt: now}
		dep2 := &Task{ID: "dep2", Title: "Dep 2", Status: TaskStatusCanceled, PoolID: "default", CreatedAt: now}
		task := &Task{
			ID:           "task-with-deps",
			Title:        "With Dependencies",
			Status:       TaskStatusPending,
			PoolID:       "default",
			CreatedAt:    now,
			Dependencies: []string{"dep1", "dep2"},
		}

		for _, t := range []*Task{dep1, dep2, task} {
			if err := store.SaveTask(t); err != nil {
				panic(err)
			}
		}

		if !scheduler.allDependenciesSatisfied(task) {
			t.Error("expected allDependenciesSatisfied=true when all dependencies are completed")
		}
	})

	t.Run("some dependencies not completed", func(t *testing.T) {
		dep3 := &Task{ID: "dep3", Title: "Dep 3", Status: TaskStatusSucceeded, PoolID: "default", CreatedAt: now}
		dep4 := &Task{ID: "dep4", Title: "Dep 4", Status: TaskStatusRunning, PoolID: "default", CreatedAt: now}
		task := &Task{
			ID:           "task-partial-deps",
			Title:        "Partial Dependencies",
			Status:       TaskStatusPending,
			PoolID:       "default",
			CreatedAt:    now,
			Dependencies: []string{"dep3", "dep4"},
		}

		for _, t := range []*Task{dep3, dep4, task} {
			if err := store.SaveTask(t); err != nil {
				panic(err)
			}
		}

		if scheduler.allDependenciesSatisfied(task) {
			t.Error("expected allDependenciesSatisfied=false when some dependencies are not completed")
		}
	})

	t.Run("missing dependency", func(t *testing.T) {
		task := &Task{
			ID:           "task-missing-dep",
			Title:        "Missing Dependency",
			Status:       TaskStatusPending,
			PoolID:       "default",
			CreatedAt:    now,
			Dependencies: []string{"non-existent-task"},
		}
		if err := store.SaveTask(task); err != nil {
			t.Fatalf("failed to save task: %v", err)
		}

		if scheduler.allDependenciesSatisfied(task) {
			t.Error("expected allDependenciesSatisfied=false when dependency is missing")
		}
	})
}
