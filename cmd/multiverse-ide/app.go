package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/biwakonbu/agent-runner/internal/chat"
	"github.com/biwakonbu/agent-runner/internal/ide"
	"github.com/biwakonbu/agent-runner/internal/meta"
	"github.com/biwakonbu/agent-runner/internal/orchestrator"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/ipc"
	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx                   context.Context
	workspaceStore        *ide.WorkspaceStore
	taskStore             *orchestrator.TaskStore
	scheduler             *orchestrator.Scheduler
	chatHandler           *chat.Handler
	currentWS             *ide.Workspace
	executionOrchestrator *orchestrator.ExecutionOrchestrator
	backlogStore          *orchestrator.BacklogStore
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

	// Initialize TaskStore, Scheduler, and ChatHandler for this workspace
	wsDir := a.workspaceStore.GetWorkspaceDir(id)
	a.taskStore = orchestrator.NewTaskStore(wsDir)
	queue := ipc.NewFilesystemQueue(wsDir)
	a.scheduler = orchestrator.NewScheduler(a.taskStore, queue)

	// Initialize Execution Environment
	agentRunnerPath := "agent-runner"
	if _, err := os.Stat("agent-runner"); err == nil {
		agentRunnerPath, _ = filepath.Abs("agent-runner")
	}
	executor := orchestrator.NewExecutor(agentRunnerPath, a.taskStore)
	eventEmitter := orchestrator.NewWailsEventEmitter(a.ctx)
	a.executionOrchestrator = orchestrator.NewExecutionOrchestrator(
		a.scheduler,
		executor,
		a.taskStore,
		queue,
		eventEmitter,
	)

	// Initialize ChatHandler with mock Meta client (TODO: configurable)
	sessionStore := chat.NewChatSessionStore(wsDir)
	metaClient := meta.NewMockClient()
	a.chatHandler = chat.NewHandler(metaClient, a.taskStore, sessionStore, id, ws.ProjectRoot, eventEmitter)

	// Initialize BacklogStore
	a.backlogStore = orchestrator.NewBacklogStore(wsDir)

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

	// Initialize TaskStore, Scheduler, and ChatHandler for this workspace
	wsDir := a.workspaceStore.GetWorkspaceDir(id)
	a.taskStore = orchestrator.NewTaskStore(wsDir)
	queue := ipc.NewFilesystemQueue(wsDir)
	a.scheduler = orchestrator.NewScheduler(a.taskStore, queue)

	// Initialize Execution Environment
	agentRunnerPath := "agent-runner"
	if _, err := os.Stat("agent-runner"); err == nil {
		agentRunnerPath, _ = filepath.Abs("agent-runner")
	}
	executor := orchestrator.NewExecutor(agentRunnerPath, a.taskStore)
	eventEmitter := orchestrator.NewWailsEventEmitter(a.ctx)
	a.executionOrchestrator = orchestrator.NewExecutionOrchestrator(
		a.scheduler,
		executor,
		a.taskStore,
		queue,
		eventEmitter,
	)

	// Initialize ChatHandler with mock Meta client (TODO: configurable)
	sessionStore := chat.NewChatSessionStore(wsDir)
	metaClient := meta.NewMockClient()
	a.chatHandler = chat.NewHandler(metaClient, a.taskStore, sessionStore, id, ws.ProjectRoot, eventEmitter)

	// Initialize BacklogStore
	a.backlogStore = orchestrator.NewBacklogStore(wsDir)

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

// GetAvailablePools returns the list of available worker pools.
func (a *App) GetAvailablePools() []orchestrator.Pool {
	if a.taskStore == nil {
		// taskStore がない場合でもデフォルト Pool を返す
		return orchestrator.DefaultPools
	}
	return a.taskStore.GetAvailablePools()
}

// ============================================================================
// Chat API (v2.0): チャット駆動タスク生成
// ============================================================================

// ChatResponseDTO はフロントエンドに返すチャット応答
type ChatResponseDTO struct {
	Message        chat.ChatMessage         `json:"message"`
	GeneratedTasks []orchestrator.Task      `json:"generatedTasks"`
	Understanding  string                   `json:"understanding"`
	Conflicts      []meta.PotentialConflict `json:"conflicts,omitempty"`
	Error          string                   `json:"error,omitempty"`
}

// CreateChatSession は新しいチャットセッションを作成する
func (a *App) CreateChatSession() *chat.ChatSession {
	if a.chatHandler == nil {
		runtime.LogErrorf(a.ctx, "ChatHandler not initialized")
		return nil
	}

	session, err := a.chatHandler.CreateSession(a.ctx)
	if err != nil {
		runtime.LogErrorf(a.ctx, "Failed to create chat session: %v", err)
		return nil
	}
	return session
}

// SendChatMessage はチャットメッセージを送信し、タスクを生成する
func (a *App) SendChatMessage(sessionID string, message string) *ChatResponseDTO {
	if a.chatHandler == nil {
		return &ChatResponseDTO{
			Error: "ChatHandler not initialized",
		}
	}

	resp, err := a.chatHandler.HandleMessage(a.ctx, sessionID, message)
	if err != nil {
		return &ChatResponseDTO{
			Error: err.Error(),
		}
	}

	return &ChatResponseDTO{
		Message:        resp.Message,
		GeneratedTasks: resp.GeneratedTasks,
		Understanding:  resp.Understanding,
		Conflicts:      resp.Conflicts,
	}
}

// GetChatHistory はチャット履歴を取得する
func (a *App) GetChatHistory(sessionID string) []chat.ChatMessage {
	if a.chatHandler == nil {
		return []chat.ChatMessage{}
	}

	messages, err := a.chatHandler.GetHistory(a.ctx, sessionID)
	if err != nil {
		runtime.LogErrorf(a.ctx, "Failed to get chat history: %v", err)
		return []chat.ChatMessage{}
	}
	return messages
}

// ============================================================================
// Execution Control API
// ============================================================================

// StartExecution starts the autonomous execution loop.
func (a *App) StartExecution() error {
	if a.executionOrchestrator == nil {
		return fmt.Errorf("execution orchestrator not initialized")
	}
	return a.executionOrchestrator.Start(a.ctx)
}

// PauseExecution pauses the autonomous execution loop.
func (a *App) PauseExecution() error {
	if a.executionOrchestrator == nil {
		return fmt.Errorf("execution orchestrator not initialized")
	}
	return a.executionOrchestrator.Pause()
}

// ResumeExecution resumes the autonomous execution loop.
func (a *App) ResumeExecution() error {
	if a.executionOrchestrator == nil {
		return fmt.Errorf("execution orchestrator not initialized")
	}
	return a.executionOrchestrator.Resume()
}

// StopExecution stops the autonomous execution loop.
func (a *App) StopExecution() error {
	if a.executionOrchestrator == nil {
		return fmt.Errorf("execution orchestrator not initialized")
	}
	return a.executionOrchestrator.Stop()
}

// GetExecutionState returns the current execution state.
func (a *App) GetExecutionState() string {
	if a.executionOrchestrator == nil {
		return "IDLE"
	}
	return string(a.executionOrchestrator.State())
}

// ============================================================================
// Backlog API
// ============================================================================

// GetBacklogItems returns all backlog items (unresolved).
func (a *App) GetBacklogItems() []orchestrator.BacklogItem {
	if a.backlogStore == nil {
		return []orchestrator.BacklogItem{}
	}

	items, err := a.backlogStore.ListUnresolved()
	if err != nil {
		runtime.LogErrorf(a.ctx, "Failed to list backlog items: %v", err)
		return []orchestrator.BacklogItem{}
	}
	return items
}

// GetAllBacklogItems returns all backlog items (including resolved).
func (a *App) GetAllBacklogItems() []orchestrator.BacklogItem {
	if a.backlogStore == nil {
		return []orchestrator.BacklogItem{}
	}

	items, err := a.backlogStore.List()
	if err != nil {
		runtime.LogErrorf(a.ctx, "Failed to list backlog items: %v", err)
		return []orchestrator.BacklogItem{}
	}
	return items
}

// ResolveBacklogItem marks a backlog item as resolved.
func (a *App) ResolveBacklogItem(id string, resolution string) error {
	if a.backlogStore == nil {
		return fmt.Errorf("backlog store not initialized")
	}
	return a.backlogStore.Resolve(id, resolution)
}

// DeleteBacklogItem deletes a backlog item.
func (a *App) DeleteBacklogItem(id string) error {
	if a.backlogStore == nil {
		return fmt.Errorf("backlog store not initialized")
	}
	return a.backlogStore.Delete(id)
}
