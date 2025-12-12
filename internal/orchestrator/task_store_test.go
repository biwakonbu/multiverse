package orchestrator

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTaskStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "task_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewTaskStore(tmpDir)

	// Verify directories are created
	if _, err := os.Stat(store.GetTaskDir()); os.IsNotExist(err) {
		t.Errorf("expected task dir to be created")
	}
	if _, err := os.Stat(store.GetAttemptDir()); os.IsNotExist(err) {
		t.Errorf("expected attempt dir to be created")
	}

	task := &Task{
		ID:        "task-1",
		Title:     "Test Task",
		Status:    TaskStatusPending,
		PoolID:    "pool-1",
		CreatedAt: time.Now(),
	}

	if err := store.SaveTask(task); err != nil {
		t.Fatalf("SaveTask failed: %v", err)
	}

	loadedTask, err := store.LoadTask("task-1")
	if err != nil {
		t.Fatalf("LoadTask failed: %v", err)
	}

	if loadedTask.Title != task.Title {
		t.Errorf("expected Title %s, got %s", task.Title, loadedTask.Title)
	}
	if loadedTask.Status != TaskStatusPending {
		t.Errorf("expected Status %s, got %s", TaskStatusPending, loadedTask.Status)
	}

	// Update task
	task.Status = TaskStatusRunning
	if err := store.SaveTask(task); err != nil {
		t.Fatalf("SaveTask update failed: %v", err)
	}

	loadedTask2, err := store.LoadTask("task-1")
	if err != nil {
		t.Fatalf("LoadTask failed: %v", err)
	}

	if loadedTask2.Status != TaskStatusRunning {
		t.Errorf("expected Status %s, got %s", TaskStatusRunning, loadedTask2.Status)
	}
}

func TestTaskStore_InvalidTaskID(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)

	invalidIDs := []string{
		"../escape",
		"task/1",
		"task\\1",
		"task with space",
		"..",
	}

	for _, id := range invalidIDs {
		t.Run(id, func(t *testing.T) {
			task := &Task{
				ID:        id,
				Title:     "Invalid",
				Status:    TaskStatusPending,
				PoolID:    "p",
				CreatedAt: time.Now(),
			}
			if err := store.SaveTask(task); err == nil {
				t.Fatalf("expected validation error for id %q", id)
			}
		})
	}
}

func TestAttemptStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "attempt_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewTaskStore(tmpDir)

	attempt := &Attempt{
		ID:        "attempt-1",
		TaskID:    "task-1",
		Status:    AttemptStatusRunning,
		StartedAt: time.Now(),
	}

	if err := store.SaveAttempt(attempt); err != nil {
		t.Fatalf("SaveAttempt failed: %v", err)
	}

	loadedAttempt, err := store.LoadAttempt("attempt-1")
	if err != nil {
		t.Fatalf("LoadAttempt failed: %v", err)
	}

	if loadedAttempt.Status != attempt.Status {
		t.Errorf("expected Status %s, got %s", attempt.Status, loadedAttempt.Status)
	}
}

func TestAttemptStore_InvalidIDs(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)

	cases := []struct {
		name    string
		Attempt Attempt
	}{
		{name: "invalid attempt id", Attempt: Attempt{ID: "../evil", TaskID: "task-1", Status: AttemptStatusRunning, StartedAt: time.Now()}},
		{name: "invalid task id", Attempt: Attempt{ID: "attempt-1", TaskID: "task/1", Status: AttemptStatusRunning, StartedAt: time.Now()}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := store.SaveAttempt(&tc.Attempt); err == nil {
				t.Fatalf("expected validation error for %s", tc.name)
			}
		})
	}
}

func TestListAttemptsByTaskID(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "list_attempts_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewTaskStore(tmpDir)

	// 空のディレクトリの場合
	attempts, err := store.ListAttemptsByTaskID("task-1")
	if err != nil {
		t.Fatalf("ListAttemptsByTaskID failed on empty dir: %v", err)
	}
	if len(attempts) != 0 {
		t.Errorf("expected 0 attempts, got %d", len(attempts))
	}

	// 複数のAttemptを保存
	attempt1 := &Attempt{
		ID:        "attempt-1",
		TaskID:    "task-1",
		Status:    AttemptStatusSucceeded,
		StartedAt: time.Now(),
	}
	attempt2 := &Attempt{
		ID:        "attempt-2",
		TaskID:    "task-1",
		Status:    AttemptStatusFailed,
		StartedAt: time.Now(),
	}
	attempt3 := &Attempt{
		ID:        "attempt-3",
		TaskID:    "task-2", // 別のタスク
		Status:    AttemptStatusRunning,
		StartedAt: time.Now(),
	}

	if err := store.SaveAttempt(attempt1); err != nil {
		t.Fatalf("SaveAttempt failed: %v", err)
	}
	if err := store.SaveAttempt(attempt2); err != nil {
		t.Fatalf("SaveAttempt failed: %v", err)
	}
	if err := store.SaveAttempt(attempt3); err != nil {
		t.Fatalf("SaveAttempt failed: %v", err)
	}

	// task-1のAttemptsを取得
	attempts, err = store.ListAttemptsByTaskID("task-1")
	if err != nil {
		t.Fatalf("ListAttemptsByTaskID failed: %v", err)
	}
	if len(attempts) != 2 {
		t.Errorf("expected 2 attempts for task-1, got %d", len(attempts))
	}

	// task-2のAttemptsを取得
	attempts, err = store.ListAttemptsByTaskID("task-2")
	if err != nil {
		t.Fatalf("ListAttemptsByTaskID failed: %v", err)
	}
	if len(attempts) != 1 {
		t.Errorf("expected 1 attempt for task-2, got %d", len(attempts))
	}
}

func TestGetPoolSummaries(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pool_summary_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewTaskStore(tmpDir)

	// 空のディレクトリの場合
	summaries, err := store.GetPoolSummaries()
	if err != nil {
		t.Fatalf("GetPoolSummaries failed on empty dir: %v", err)
	}
	if len(summaries) != 0 {
		t.Errorf("expected 0 summaries, got %d", len(summaries))
	}

	// 複数のタスクを保存
	tasks := []*Task{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusRunning, PoolID: "codegen", CreatedAt: time.Now()},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusRunning, PoolID: "codegen", CreatedAt: time.Now()},
		{ID: "task-3", Title: "Task 3", Status: TaskStatusPending, PoolID: "codegen", CreatedAt: time.Now()},
		{ID: "task-4", Title: "Task 4", Status: TaskStatusFailed, PoolID: "codegen", CreatedAt: time.Now()},
		{ID: "task-5", Title: "Task 5", Status: TaskStatusRunning, PoolID: "test", CreatedAt: time.Now()},
		{ID: "task-6", Title: "Task 6", Status: TaskStatusSucceeded, PoolID: "test", CreatedAt: time.Now()},
	}

	for _, task := range tasks {
		if err := store.SaveTask(task); err != nil {
			t.Fatalf("SaveTask failed: %v", err)
		}
	}

	summaries, err = store.GetPoolSummaries()
	if err != nil {
		t.Fatalf("GetPoolSummaries failed: %v", err)
	}
	if len(summaries) != 2 {
		t.Errorf("expected 2 pools, got %d", len(summaries))
	}

	// サマリを検証
	poolMap := make(map[string]PoolSummary)
	for _, s := range summaries {
		poolMap[s.PoolID] = s
	}

	codegen, ok := poolMap["codegen"]
	if !ok {
		t.Fatal("codegen pool not found")
	}
	if codegen.Running != 2 {
		t.Errorf("expected codegen running=2, got %d", codegen.Running)
	}
	if codegen.Queued != 1 {
		t.Errorf("expected codegen queued=1, got %d", codegen.Queued)
	}
	if codegen.Failed != 1 {
		t.Errorf("expected codegen failed=1, got %d", codegen.Failed)
	}
	if codegen.Total != 4 {
		t.Errorf("expected codegen total=4, got %d", codegen.Total)
	}

	testPool, ok := poolMap["test"]
	if !ok {
		t.Fatal("test pool not found")
	}
	if testPool.Running != 1 {
		t.Errorf("expected test running=1, got %d", testPool.Running)
	}
	if testPool.Total != 2 {
		t.Errorf("expected test total=2, got %d", testPool.Total)
	}
}

func TestGetAvailablePools(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "available_pools_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewTaskStore(tmpDir)

	pools := store.GetAvailablePools()

	// DefaultPools が返されることを確認
	if len(pools) != len(DefaultPools) {
		t.Errorf("expected %d pools, got %d", len(DefaultPools), len(pools))
	}

	// 各 Pool の ID と Name を検証
	expectedPools := map[string]string{
		"default": "Default",
		"codegen": "Codegen",
		"test":    "Test",
	}

	for _, pool := range pools {
		expectedName, ok := expectedPools[pool.ID]
		if !ok {
			t.Errorf("unexpected pool ID: %s", pool.ID)
			continue
		}
		if pool.Name != expectedName {
			t.Errorf("pool %s: expected name %s, got %s", pool.ID, expectedName, pool.Name)
		}
		// Description が空でないことを確認
		if pool.Description == "" {
			t.Errorf("pool %s: expected non-empty description", pool.ID)
		}
	}
}

func TestPoolStructJSON(t *testing.T) {
	// Pool 構造体の JSON シリアライゼーションをテスト
	pool := Pool{
		ID:          "test-pool",
		Name:        "Test Pool",
		Description: "A test pool",
	}

	// encoding/json は標準的な Go のシリアライゼーションなので、
	// タグが正しく設定されていることを確認
	if pool.ID != "test-pool" {
		t.Errorf("expected ID test-pool, got %s", pool.ID)
	}
	if pool.Name != "Test Pool" {
		t.Errorf("expected Name Test Pool, got %s", pool.Name)
	}
	if pool.Description != "A test pool" {
		t.Errorf("expected Description 'A test pool', got %s", pool.Description)
	}
}

func TestDefaultPoolsContent(t *testing.T) {
	// DefaultPools の内容を検証
	if len(DefaultPools) != 3 {
		t.Fatalf("expected 3 default pools, got %d", len(DefaultPools))
	}

	// ID の一意性を確認
	ids := make(map[string]bool)
	for _, pool := range DefaultPools {
		if ids[pool.ID] {
			t.Errorf("duplicate pool ID: %s", pool.ID)
		}
		ids[pool.ID] = true
	}

	// 必須 Pool が存在することを確認
	requiredIDs := []string{"default", "codegen", "test"}
	for _, id := range requiredIDs {
		if !ids[id] {
			t.Errorf("required pool %s not found in DefaultPools", id)
		}
	}
}

func TestListTasksBySourceChat(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)

	now := time.Now()
	chatID1 := "chat-session-1"
	chatID2 := "chat-session-2"

	tasks := []*Task{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, SourceChatID: &chatID1},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, SourceChatID: &chatID1},
		{ID: "task-3", Title: "Task 3", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, SourceChatID: &chatID2},
		{ID: "task-4", Title: "Task 4", Status: TaskStatusPending, PoolID: "default", CreatedAt: now}, // SourceChatID なし
	}

	for _, task := range tasks {
		if err := store.SaveTask(task); err != nil {
			t.Fatalf("failed to save task: %v", err)
		}
	}

	// chat-session-1 のタスクを取得
	result, err := store.ListTasksBySourceChat("chat-session-1")
	if err != nil {
		t.Fatalf("ListTasksBySourceChat failed: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 tasks for chat-session-1, got %d", len(result))
	}

	// chat-session-2 のタスクを取得
	result, err = store.ListTasksBySourceChat("chat-session-2")
	if err != nil {
		t.Fatalf("ListTasksBySourceChat failed: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 task for chat-session-2, got %d", len(result))
	}

	// 存在しない chat ID
	result, err = store.ListTasksBySourceChat("non-existent")
	if err != nil {
		t.Fatalf("ListTasksBySourceChat failed: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 tasks for non-existent chat, got %d", len(result))
	}
}

func TestListAllTasks(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)

	// 空のディレクトリ
	tasks, err := store.ListAllTasks()
	if err != nil {
		t.Fatalf("ListAllTasks failed on empty dir: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}

	now := time.Now()
	taskList := []*Task{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusPending, PoolID: "default", CreatedAt: now},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusRunning, PoolID: "default", CreatedAt: now},
		{ID: "task-3", Title: "Task 3", Status: TaskStatusSucceeded, PoolID: "codegen", CreatedAt: now},
	}

	for _, task := range taskList {
		if err := store.SaveTask(task); err != nil {
			t.Fatalf("failed to save task: %v", err)
		}
	}

	tasks, err = store.ListAllTasks()
	if err != nil {
		t.Fatalf("ListAllTasks failed: %v", err)
	}
	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(tasks))
	}
}

func TestListTasksByStatus(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)

	now := time.Now()
	taskList := []*Task{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusPending, PoolID: "default", CreatedAt: now},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusPending, PoolID: "default", CreatedAt: now},
		{ID: "task-3", Title: "Task 3", Status: TaskStatusRunning, PoolID: "default", CreatedAt: now},
		{ID: "task-4", Title: "Task 4", Status: TaskStatusSucceeded, PoolID: "default", CreatedAt: now},
	}

	for _, task := range taskList {
		if err := store.SaveTask(task); err != nil {
			t.Fatalf("failed to save task: %v", err)
		}
	}

	// PENDING タスク
	pending, err := store.ListTasksByStatus(TaskStatusPending)
	if err != nil {
		t.Fatalf("ListTasksByStatus failed: %v", err)
	}
	if len(pending) != 2 {
		t.Errorf("expected 2 pending tasks, got %d", len(pending))
	}

	// RUNNING タスク
	running, err := store.ListTasksByStatus(TaskStatusRunning)
	if err != nil {
		t.Fatalf("ListTasksByStatus failed: %v", err)
	}
	if len(running) != 1 {
		t.Errorf("expected 1 running task, got %d", len(running))
	}

	// BLOCKED タスク（なし）
	blocked, err := store.ListTasksByStatus(TaskStatusBlocked)
	if err != nil {
		t.Fatalf("ListTasksByStatus failed: %v", err)
	}
	if len(blocked) != 0 {
		t.Errorf("expected 0 blocked tasks, got %d", len(blocked))
	}
}

func TestLoadTask_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)

	_, err := store.LoadTask("non-existent-task")
	if err == nil {
		t.Error("expected error for non-existent task, got nil")
	}
}

func TestLoadAttempt_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)

	_, err := store.LoadAttempt("non-existent-attempt")
	if err == nil {
		t.Error("expected error for non-existent attempt, got nil")
	}
}

func TestTaskStore_PathTraversalIsRejected(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)

	badTask := &Task{
		ID:        "../escape",
		Title:     "Bad",
		Status:    TaskStatusPending,
		PoolID:    "default",
		CreatedAt: time.Now(),
	}
	if err := store.SaveTask(badTask); err == nil {
		t.Fatalf("expected error for invalid task id, got nil")
	}

	if _, err := store.LoadTask("../escape"); err == nil {
		t.Fatalf("expected error for loading invalid task id, got nil")
	}

	// Ensure no file got created
	if _, err := os.Stat(filepath.Join(tmpDir, "tasks", "escape.jsonl")); err == nil {
		t.Fatalf("expected no task file to be created for invalid id")
	}

	badAttempt := &Attempt{
		ID:        "../escape-attempt",
		TaskID:    "task-1",
		Status:    AttemptStatusRunning,
		StartedAt: time.Now(),
	}
	if err := store.SaveAttempt(badAttempt); err == nil {
		t.Fatalf("expected error for invalid attempt id, got nil")
	}

	if _, err := store.LoadAttempt("../escape-attempt"); err == nil {
		t.Fatalf("expected error for loading invalid attempt id, got nil")
	}
}

func TestTaskStore_SuggestedImpl(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)

	task := &Task{
		ID:        "task-with-impl",
		Title:     "Impl Task",
		Status:    TaskStatusPending,
		CreatedAt: time.Now(),
		SuggestedImpl: &SuggestedImpl{
			Language:    "go",
			FilePaths:   []string{"main.go"},
			Constraints: []string{"no panic"},
		},
		Artifacts: &Artifacts{
			Files: []string{"output.txt"},
			Logs:  []string{"run.log"},
		},
	}

	if err := store.SaveTask(task); err != nil {
		t.Fatalf("SaveTask failed: %v", err)
	}

	loaded, err := store.LoadTask("task-with-impl")
	if err != nil {
		t.Fatalf("LoadTask failed: %v", err)
	}

	if loaded.SuggestedImpl == nil {
		t.Fatal("expected SuggestedImpl to be not nil")
	}
	if loaded.SuggestedImpl.Language != "go" {
		t.Errorf("expected Language 'go', got %s", loaded.SuggestedImpl.Language)
	}
	if len(loaded.SuggestedImpl.FilePaths) != 1 || loaded.SuggestedImpl.FilePaths[0] != "main.go" {
		t.Errorf("unexpected file paths: %v", loaded.SuggestedImpl.FilePaths)
	}
	if len(loaded.SuggestedImpl.Constraints) != 1 || loaded.SuggestedImpl.Constraints[0] != "no panic" {
		t.Errorf("unexpected constraints: %v", loaded.SuggestedImpl.Constraints)
	}

	if loaded.Artifacts == nil {
		t.Fatal("expected Artifacts to be not nil")
	}
	if len(loaded.Artifacts.Files) != 1 || loaded.Artifacts.Files[0] != "output.txt" {
		t.Errorf("unexpected artifact files: %v", loaded.Artifacts.Files)
	}
	if len(loaded.Artifacts.Logs) != 1 || loaded.Artifacts.Logs[0] != "run.log" {
		t.Errorf("unexpected artifact logs: %v", loaded.Artifacts.Logs)
	}
}
