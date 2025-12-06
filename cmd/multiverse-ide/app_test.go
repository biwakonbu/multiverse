package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/biwakonbu/agent-runner/internal/chat"
	"github.com/biwakonbu/agent-runner/internal/ide"
	"github.com/biwakonbu/agent-runner/internal/meta"
	"github.com/biwakonbu/agent-runner/internal/orchestrator"
)

func TestNewApp(t *testing.T) {
	app := NewApp()

	if app == nil {
		t.Fatal("NewApp returned nil")
	}

	if app.workspaceStore == nil {
		t.Error("workspaceStore should not be nil")
	}
}

func TestGetAvailablePools_WithoutTaskStore(t *testing.T) {
	app := NewApp()

	// taskStore が nil の場合でも DefaultPools を返す
	pools := app.GetAvailablePools()

	if len(pools) != len(orchestrator.DefaultPools) {
		t.Errorf("expected %d pools, got %d", len(orchestrator.DefaultPools), len(pools))
	}

	// DefaultPools と一致することを確認
	for i, pool := range pools {
		if pool.ID != orchestrator.DefaultPools[i].ID {
			t.Errorf("pool %d: expected ID %s, got %s", i, orchestrator.DefaultPools[i].ID, pool.ID)
		}
	}
}

func TestGetAvailablePools_WithTaskStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "app_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	app := NewApp()
	app.taskStore = orchestrator.NewTaskStore(tmpDir)

	pools := app.GetAvailablePools()

	if len(pools) != len(orchestrator.DefaultPools) {
		t.Errorf("expected %d pools, got %d", len(orchestrator.DefaultPools), len(pools))
	}
}

func TestGetPoolSummaries_WithoutTaskStore(t *testing.T) {
	app := NewApp()

	// taskStore が nil の場合は空を返す
	summaries := app.GetPoolSummaries()

	if summaries == nil {
		t.Error("GetPoolSummaries should return empty slice, not nil")
	}
	if len(summaries) != 0 {
		t.Errorf("expected 0 summaries without taskStore, got %d", len(summaries))
	}
}

func TestGetPoolSummaries_WithTaskStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "app_summaries_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	app := NewApp()
	app.taskStore = orchestrator.NewTaskStore(tmpDir)

	// タスクを追加
	task := &orchestrator.Task{
		ID:        "task-1",
		Title:     "Test Task",
		Status:    orchestrator.TaskStatusRunning,
		PoolID:    "codegen",
		CreatedAt: time.Now(),
	}
	if err := app.taskStore.SaveTask(task); err != nil {
		t.Fatalf("SaveTask failed: %v", err)
	}

	summaries := app.GetPoolSummaries()

	if len(summaries) != 1 {
		t.Errorf("expected 1 summary, got %d", len(summaries))
	}
	if len(summaries) > 0 && summaries[0].PoolID != "codegen" {
		t.Errorf("expected poolID codegen, got %s", summaries[0].PoolID)
	}
}

func TestListTasks_WithoutTaskStore(t *testing.T) {
	app := NewApp()

	tasks := app.ListTasks()

	if tasks == nil {
		t.Error("ListTasks should return empty slice, not nil")
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks without taskStore, got %d", len(tasks))
	}
}

func TestListTasks_WithTaskStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "app_list_tasks_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	app := NewApp()
	app.taskStore = orchestrator.NewTaskStore(tmpDir)

	// タスクを追加
	for i := 1; i <= 3; i++ {
		task := &orchestrator.Task{
			ID:        "task-" + string(rune('0'+i)),
			Title:     "Test Task",
			Status:    orchestrator.TaskStatusPending,
			PoolID:    "default",
			CreatedAt: time.Now(),
		}
		if err := app.taskStore.SaveTask(task); err != nil {
			t.Fatalf("SaveTask failed: %v", err)
		}
	}

	tasks := app.ListTasks()

	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(tasks))
	}
}

func TestCreateTask_WithoutTaskStore(t *testing.T) {
	app := NewApp()

	task := app.CreateTask("Test Title", "default")

	if task != nil {
		t.Error("CreateTask should return nil when taskStore is nil")
	}
}

func TestCreateTask_WithTaskStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "app_create_task_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	app := NewApp()
	app.taskStore = orchestrator.NewTaskStore(tmpDir)

	task := app.CreateTask("Test Title", "codegen")

	if task == nil {
		t.Fatal("CreateTask should return non-nil task")
	}
	if task.Title != "Test Title" {
		t.Errorf("expected title 'Test Title', got '%s'", task.Title)
	}
	if task.PoolID != "codegen" {
		t.Errorf("expected poolID 'codegen', got '%s'", task.PoolID)
	}
	if task.Status != orchestrator.TaskStatusPending {
		t.Errorf("expected status PENDING, got %s", task.Status)
	}
	if task.ID == "" {
		t.Error("task ID should not be empty")
	}

	// ファイルが作成されたことを確認
	taskPath := filepath.Join(tmpDir, "tasks", task.ID+".jsonl")
	if _, err := os.Stat(taskPath); os.IsNotExist(err) {
		t.Errorf("expected task file at %s", taskPath)
	}
}

func TestListAttempts_WithoutTaskStore(t *testing.T) {
	app := NewApp()

	attempts := app.ListAttempts("task-1")

	if attempts == nil {
		t.Error("ListAttempts should return empty slice, not nil")
	}
	if len(attempts) != 0 {
		t.Errorf("expected 0 attempts without taskStore, got %d", len(attempts))
	}
}

func TestListAttempts_WithTaskStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "app_list_attempts_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	app := NewApp()
	app.taskStore = orchestrator.NewTaskStore(tmpDir)

	// Attempt を追加
	attempt := &orchestrator.Attempt{
		ID:        "attempt-1",
		TaskID:    "task-1",
		Status:    orchestrator.AttemptStatusRunning,
		StartedAt: time.Now(),
	}
	if err := app.taskStore.SaveAttempt(attempt); err != nil {
		t.Fatalf("SaveAttempt failed: %v", err)
	}

	attempts := app.ListAttempts("task-1")

	if len(attempts) != 1 {
		t.Errorf("expected 1 attempt, got %d", len(attempts))
	}
}

func TestGetWorkspace_NotFound(t *testing.T) {
	app := NewApp()

	ws := app.GetWorkspace("nonexistent-id")

	if ws != nil {
		t.Error("GetWorkspace should return nil for non-existent workspace")
	}
}

func TestGetWorkspace_Found(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "app_get_workspace_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// WorkspaceStore を直接使って Workspace を保存
	store := ide.NewWorkspaceStore(tmpDir)
	projectRoot := "/test/project"
	wsID := store.GetWorkspaceID(projectRoot)

	ws := &ide.Workspace{
		Version:      "1.0",
		ProjectRoot:  projectRoot,
		DisplayName:  "Test Project",
		CreatedAt:    time.Now(),
		LastOpenedAt: time.Now(),
	}
	if err := store.SaveWorkspace(ws); err != nil {
		t.Fatalf("SaveWorkspace failed: %v", err)
	}

	// App を作成して workspaceStore を設定
	app := &App{
		workspaceStore: store,
	}

	loadedWS := app.GetWorkspace(wsID)

	if loadedWS == nil {
		t.Fatal("GetWorkspace should return non-nil workspace")
	}
	if loadedWS.DisplayName != "Test Project" {
		t.Errorf("expected DisplayName 'Test Project', got '%s'", loadedWS.DisplayName)
	}
}

func TestListRecentWorkspaces_Empty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "app_list_recent_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	app := &App{
		workspaceStore: ide.NewWorkspaceStore(tmpDir),
	}

	workspaces := app.ListRecentWorkspaces()

	// nil または空スライスのどちらかを許容
	if len(workspaces) != 0 {
		t.Errorf("expected 0 workspaces, got %d", len(workspaces))
	}
}

func TestListRecentWorkspaces_WithWorkspaces(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "app_list_recent_with_ws_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := ide.NewWorkspaceStore(tmpDir)

	// 複数の Workspace を保存
	for i := 1; i <= 3; i++ {
		ws := &ide.Workspace{
			Version:      "1.0",
			ProjectRoot:  "/test/project" + string(rune('0'+i)),
			DisplayName:  "Test Project " + string(rune('0'+i)),
			CreatedAt:    time.Now(),
			LastOpenedAt: time.Now(),
		}
		if err := store.SaveWorkspace(ws); err != nil {
			t.Fatalf("SaveWorkspace failed: %v", err)
		}
	}

	app := &App{
		workspaceStore: store,
	}

	workspaces := app.ListRecentWorkspaces()

	if len(workspaces) != 3 {
		t.Errorf("expected 3 workspaces, got %d", len(workspaces))
	}
}

func TestRunTask_WithoutScheduler(t *testing.T) {
	app := NewApp()

	err := app.RunTask("task-1")

	if err == nil {
		t.Error("RunTask should return error when scheduler is nil")
	}
}

func TestRemoveWorkspace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "app_remove_ws_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := ide.NewWorkspaceStore(tmpDir)

	// Workspace を保存
	ws := &ide.Workspace{
		Version:      "1.0",
		ProjectRoot:  "/test/project",
		DisplayName:  "Test Project",
		CreatedAt:    time.Now(),
		LastOpenedAt: time.Now(),
	}
	if err := store.SaveWorkspace(ws); err != nil {
		t.Fatalf("SaveWorkspace failed: %v", err)
	}

	wsID := store.GetWorkspaceID("/test/project")

	app := &App{
		workspaceStore: store,
	}

	// 削除
	if err := app.RemoveWorkspace(wsID); err != nil {
		t.Fatalf("RemoveWorkspace failed: %v", err)
	}

	// 削除されていることを確認
	loadedWS := app.GetWorkspace(wsID)
	if loadedWS != nil {
		t.Error("workspace should be removed")
	}
}

func TestSendChatMessage(t *testing.T) {
	// Setup temporary directory for workspace
	tmpDir, err := os.MkdirTemp("", "app_chat_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize dependencies
	taskStore := orchestrator.NewTaskStore(tmpDir)
	sessionStore := chat.NewChatSessionStore(tmpDir)
	metaClient := meta.NewMockClient()

	// Initialize Handler manually
	handler := chat.NewHandler(metaClient, taskStore, sessionStore, "test-ws-id", tmpDir, nil)

	// Initialize App with the handler
	app := NewApp() // This sets workspaceStore
	app.chatHandler = handler
	app.taskStore = taskStore      // needed for list tasks context
	app.ctx = context.Background() // Fix panic due to nil context

	// 1. Create Session
	session := app.CreateChatSession()
	if session == nil {
		t.Fatal("CreateChatSession returned nil")
	}
	if session.ID == "" {
		t.Error("Session ID should not be empty")
	}

	// 2. Send Message
	resp := app.SendChatMessage(session.ID, "Make a delicious ramen website")

	if resp.Error != "" {
		t.Errorf("SendChatMessage returned error: %s", resp.Error)
	}

	if len(resp.GeneratedTasks) == 0 {
		t.Error("Expected generated tasks, got 0")
	}

	if resp.Understanding == "" {
		t.Error("Expected understanding, got empty string")
	}
}
