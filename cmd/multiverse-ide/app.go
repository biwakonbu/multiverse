package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/biwakonbu/agent-runner/internal/ide"
	"github.com/biwakonbu/agent-runner/internal/orchestrator"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/ipc"
	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx            context.Context
	workspaceStore *ide.WorkspaceStore
	taskStore      *orchestrator.TaskStore
	scheduler      *orchestrator.Scheduler
	currentWS      *ide.Workspace
}

// NewApp creates a new App application struct
func NewApp() *App {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	baseDir := fmt.Sprintf("%s/.multiverse/workspaces", homeDir)
	return &App{
		workspaceStore: ide.NewWorkspaceStore(baseDir),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SelectWorkspace opens a directory selection dialog and loads the workspace.
func (a *App) SelectWorkspace() string {
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Project Root",
	})
	if err != nil {
		return ""
	}
	if selection == "" {
		return ""
	}

	id := a.workspaceStore.GetWorkspaceID(selection)
	ws, err := a.workspaceStore.LoadWorkspace(id)
	if err != nil {
		// Create new workspace if not exists
		now := time.Now()
		ws = &ide.Workspace{
			Version:      "1.0",
			ProjectRoot:  selection,
			DisplayName:  filepath.Base(selection),
			CreatedAt:    now,
			LastOpenedAt: now,
		}
	} else {
		// Update lastOpenedAt for existing workspace
		ws.LastOpenedAt = time.Now()
	}

	if err := a.workspaceStore.SaveWorkspace(ws); err != nil {
		runtime.LogErrorf(a.ctx, "Failed to save workspace: %v", err)
		return ""
	}

	a.currentWS = ws

	// Initialize TaskStore and Scheduler for this workspace
	wsDir := a.workspaceStore.GetWorkspaceDir(id)
	a.taskStore = orchestrator.NewTaskStore(wsDir)
	queue := ipc.NewFilesystemQueue(wsDir)
	a.scheduler = orchestrator.NewScheduler(a.taskStore, queue)

	return id
}

// GetWorkspace returns the workspace details.
func (a *App) GetWorkspace(id string) *ide.Workspace {
	ws, err := a.workspaceStore.LoadWorkspace(id)
	if err != nil {
		return nil
	}
	return ws
}

// ListRecentWorkspaces は最近使用したワークスペース一覧を返す
func (a *App) ListRecentWorkspaces() []ide.WorkspaceSummary {
	summaries, err := a.workspaceStore.ListWorkspaces()
	if err != nil {
		runtime.LogErrorf(a.ctx, "Failed to list workspaces: %v", err)
		return []ide.WorkspaceSummary{}
	}
	return summaries
}

// OpenWorkspaceByID は既存ワークスペースを ID で開く
func (a *App) OpenWorkspaceByID(id string) string {
	ws, err := a.workspaceStore.LoadWorkspace(id)
	if err != nil {
		runtime.LogErrorf(a.ctx, "Failed to load workspace: %v", err)
		return ""
	}

	// Update lastOpenedAt
	ws.LastOpenedAt = time.Now()
	if err := a.workspaceStore.SaveWorkspace(ws); err != nil {
		runtime.LogErrorf(a.ctx, "Failed to save workspace: %v", err)
		return ""
	}

	a.currentWS = ws

	// Initialize TaskStore and Scheduler for this workspace
	wsDir := a.workspaceStore.GetWorkspaceDir(id)
	a.taskStore = orchestrator.NewTaskStore(wsDir)
	queue := ipc.NewFilesystemQueue(wsDir)
	a.scheduler = orchestrator.NewScheduler(a.taskStore, queue)

	return id
}

// RemoveWorkspace はワークスペースを履歴から削除
func (a *App) RemoveWorkspace(id string) error {
	return a.workspaceStore.RemoveWorkspace(id)
}

// ListTasks returns all tasks in the current workspace.
// Note: In a real app, we might want pagination or filtering.
// For now, we'll just list all jsonl files in the tasks dir.
func (a *App) ListTasks() []orchestrator.Task {
	if a.taskStore == nil {
		return []orchestrator.Task{}
	}

	dir := a.taskStore.GetTaskDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []orchestrator.Task{}
	}

	var tasks []orchestrator.Task
	for _, entry := range entries {
		if !entry.IsDir() && len(entry.Name()) > 6 && entry.Name()[len(entry.Name())-6:] == ".jsonl" {
			id := entry.Name()[:len(entry.Name())-6]
			task, err := a.taskStore.LoadTask(id)
			if err == nil {
				tasks = append(tasks, *task)
			}
		}
	}
	return tasks
}

// CreateTask creates a new task.
func (a *App) CreateTask(title string, poolID string) *orchestrator.Task {
	if a.taskStore == nil {
		return nil
	}

	task := &orchestrator.Task{
		ID:        uuid.New().String(),
		Title:     title,
		Status:    orchestrator.TaskStatusPending,
		PoolID:    poolID,
		CreatedAt: time.Now(),
	}

	if err := a.taskStore.SaveTask(task); err != nil {
		runtime.LogErrorf(a.ctx, "Failed to save task: %v", err)
		return nil
	}
	return task
}

// RunTask schedules a task for execution.
func (a *App) RunTask(taskID string) error {
	if a.scheduler == nil {
		return fmt.Errorf("scheduler not initialized")
	}
	return a.scheduler.ScheduleTask(taskID)
}

// ListAttempts returns all attempts for a given task.
func (a *App) ListAttempts(taskID string) []orchestrator.Attempt {
	if a.taskStore == nil {
		return []orchestrator.Attempt{}
	}

	attempts, err := a.taskStore.ListAttemptsByTaskID(taskID)
	if err != nil {
		runtime.LogErrorf(a.ctx, "Failed to list attempts: %v", err)
		return []orchestrator.Attempt{}
	}
	return attempts
}

// GetPoolSummaries returns task count summaries by pool.
func (a *App) GetPoolSummaries() []orchestrator.PoolSummary {
	if a.taskStore == nil {
		return []orchestrator.PoolSummary{}
	}

	summaries, err := a.taskStore.GetPoolSummaries()
	if err != nil {
		runtime.LogErrorf(a.ctx, "Failed to get pool summaries: %v", err)
		return []orchestrator.PoolSummary{}
	}
	return summaries
}
