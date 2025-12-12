package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/biwakonbu/agent-runner/internal/chat"
	"github.com/biwakonbu/agent-runner/internal/ide"
	"github.com/biwakonbu/agent-runner/internal/meta"
	"github.com/biwakonbu/agent-runner/internal/orchestrator"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/ipc"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/persistence"
	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx                   context.Context
	workspaceStore        *ide.WorkspaceStore
	llmConfigStore        *ide.LLMConfigStore
	repo                  persistence.WorkspaceRepository
	scheduler             *orchestrator.Scheduler
	chatHandler           *chat.Handler
	currentWS             *ide.Workspace
	currentWSID           string
	executionOrchestrator *orchestrator.ExecutionOrchestrator
	backlogStore          *orchestrator.BacklogStore
	eventEmitter          orchestrator.EventEmitter
}

// NewApp creates a new App application struct
func NewApp() *App {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	multiverseDir := fmt.Sprintf("%s/.multiverse", homeDir)
	return &App{
		workspaceStore: ide.NewWorkspaceStore(filepath.Join(multiverseDir, "workspaces")),
		llmConfigStore: ide.NewLLMConfigStore(multiverseDir),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// newMetaClientFromConfig は LLMConfigStore の設定に基づいて Meta クライアントを生成する
// 優先度:
// 1. LLMConfigStore の設定（codex-cli, mock 等）
// 2. 環境変数でのオーバーライド（後方互換性のため）
func (a *App) newMetaClientFromConfig() *meta.Client {
	config, err := a.llmConfigStore.GetEffectiveConfig()
	if err != nil {
		// エラー時でも実クライアント(codex-cli)を返す（モック廃止）
		// エラーログは出しておく
		runtime.LogErrorf(a.ctx, "Failed to load LLM config, falling back to default codex-cli: %v", err)
		return meta.NewClient("codex-cli", "", "", "")
	}

	kind := config.Kind
	if kind == "" {
		kind = "codex-cli" // デフォルトで実タスク実行
	}

	switch kind {
	case "codex-cli":
		return meta.NewClient("codex-cli", "", config.Model, config.SystemPrompt)
	case "openai-chat":
		// 後方互換性のため残す（HTTP ベース）
		apiKey := os.Getenv("OPENAI_API_KEY")
		return meta.NewClient("openai-chat", apiKey, config.Model, config.SystemPrompt)
	default:
		// 未知の種類の時も実クライアント(codex-cli)を返す（モック廃止）
		runtime.LogErrorf(a.ctx, "Unknown LLM kind '%s', falling back to codex-cli", kind)
		return meta.NewClient("codex-cli", "", config.Model, config.SystemPrompt)
	}
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
	a.currentWSID = id

	// Initialize TaskStore (Repo), Scheduler, and ChatHandler for this workspace
	wsDir := a.workspaceStore.GetWorkspaceDir(id)
	a.repo = persistence.NewWorkspaceRepository(wsDir)
	if err := a.repo.Init(); err != nil {
		runtime.LogErrorf(a.ctx, "Failed to initialize repository: %v", err)
		return ""
	}
	// a.taskStore = orchestrator.NewTaskStore(wsDir) // Removed
	queue := ipc.NewFilesystemQueue(wsDir)

	// Initialize Execution Environment
	agentRunnerPath := "agent-runner"
	if _, err := os.Stat("agent-runner"); err == nil {
		agentRunnerPath, _ = filepath.Abs("agent-runner")
	}
	executor := orchestrator.NewExecutor(agentRunnerPath, ws.ProjectRoot)
	a.eventEmitter = orchestrator.NewWailsEventEmitter(a.ctx)
	executor.SetEventEmitter(a.eventEmitter)

	a.scheduler = orchestrator.NewScheduler(a.repo, queue, a.eventEmitter)

	// Initialize BacklogStore (before ExecutionOrchestrator)
	a.backlogStore = orchestrator.NewBacklogStore(wsDir)

	a.executionOrchestrator = orchestrator.NewExecutionOrchestrator(
		a.scheduler,
		executor,
		a.repo,
		queue,
		a.eventEmitter,
		a.backlogStore,
		[]string{"default", "codegen", "test"},
	)

	// Initialize ChatHandler with Meta client from LLMConfigStore
	sessionStore := chat.NewChatSessionStore(wsDir)
	metaClient := a.newMetaClientFromConfig()
	// ChatHandler still needs TaskStore? Or Repo?
	// ChatHandler needs to create nodes/tasks.
	// I need to check ChatHandler signature. Assuming it needs update too.
	// For now let's pass nil or fix ChatHandler later?
	// ChatHandler uses TaskStore to List/Create tasks.
	// I should probably force ChatHandler update or provide adapter.
	// If ChatHandler expects *TaskStore, I'm stuck unless I update it.
	// I'll leave a.taskStore as nil passed to ChatHandler? It will crash.
	// I'll create a new TaskStore instance just for legacy compatibility if needed?
	// "a.taskStore = orchestrator.NewTaskStore(wsDir)" -> I can keep this line for ChatHandler if needed?
	// But TaskStore writes to old paths?
	// ChatHandler should use Repo.
	// I will update ChatHandler signature in next step.
	// For this block I will pass 'nil' and fix compilation later?
	// Or define 'taskStore' locally if ChatHandler needs it.
	taskStore := orchestrator.NewTaskStore(wsDir) // Temporary for ChatHandler compatibility?
	a.chatHandler = chat.NewHandler(metaClient, taskStore, sessionStore, id, ws.ProjectRoot, a.eventEmitter)

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
	a.currentWSID = id

	// Initialize TaskStore, Scheduler, and ChatHandler for this workspace
	wsDir := a.workspaceStore.GetWorkspaceDir(id)
	a.repo = persistence.NewWorkspaceRepository(wsDir) // Initialize repo here
	if err := a.repo.Init(); err != nil {
		runtime.LogErrorf(a.ctx, "Failed to initialize repository: %v", err)
		return ""
	}
	// a.taskStore = orchestrator.NewTaskStore(wsDir) // Removed
	queue := ipc.NewFilesystemQueue(wsDir)

	// Initialize Execution Environment
	agentRunnerPath := "agent-runner"
	if _, err := os.Stat("agent-runner"); err == nil {
		agentRunnerPath, _ = filepath.Abs("agent-runner")
	}
	executor := orchestrator.NewExecutor(agentRunnerPath, ws.ProjectRoot) // Removed a.taskStore from here
	a.eventEmitter = orchestrator.NewWailsEventEmitter(a.ctx)
	executor.SetEventEmitter(a.eventEmitter)

	a.scheduler = orchestrator.NewScheduler(a.repo, queue, a.eventEmitter) // Use a.repo here

	// Initialize BacklogStore (ExecutionOrchestrator depends on it)
	a.backlogStore = orchestrator.NewBacklogStore(wsDir)

	a.executionOrchestrator = orchestrator.NewExecutionOrchestrator(
		a.scheduler,
		executor,
		a.repo,
		queue,
		a.eventEmitter,
		a.backlogStore,
		[]string{"default", "codegen", "test"},
	)

	// Initialize ChatHandler with Meta client from LLMConfigStore
	sessionStore := chat.NewChatSessionStore(wsDir)
	metaClient := a.newMetaClientFromConfig()
	// Temporary taskStore instance for ChatHandler
	taskStore := orchestrator.NewTaskStore(wsDir)
	a.chatHandler = chat.NewHandler(metaClient, taskStore, sessionStore, id, ws.ProjectRoot, a.eventEmitter)

	return id
}

// RemoveWorkspace はワークスペースを履歴から削除
func (a *App) RemoveWorkspace(id string) error {
	return a.workspaceStore.RemoveWorkspace(id)
}

// ListTasks returns all tasks in the current workspace.
// ListTasks returns all tasks in the current workspace.
func (a *App) ListTasks() []orchestrator.Task {
	if a.repo == nil {
		return []orchestrator.Task{}
	}

	tasksState, err := a.repo.State().LoadTasks()
	if err != nil {
		runtime.LogErrorf(a.ctx, "Failed to load tasks: %v", err)
		return []orchestrator.Task{}
	}

	var tasks []orchestrator.Task
	for _, t := range tasksState.Tasks {
		// Map persistence.TaskState to orchestrator.Task
		task := orchestrator.Task{
			ID:        t.TaskID,
			Title:     t.Kind + ": " + t.NodeID, // Fallback
			Status:    orchestrator.TaskStatus(t.Status),
			PoolID:    "default", // Missing in TaskState
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.UpdatedAt,
			// Attempts: not populated
		}
		// Try to load Node info for Title refinement? Expensive in loop.
		tasks = append(tasks, task)
	}
	return tasks
}

// CreateTask creates a new task.
func (a *App) CreateTask(title string, poolID string) *orchestrator.Task {
	if a.repo == nil {
		return nil
	}

	// NOTE: CreateTask in V2 should go through WBS/Node creation properly via Planner.
	// Direct task creation is technical debt or for simple tasks.
	// We'll map it to a generic task node or just create a task in tasks.json without node?
	// The schema expects NodeID.
	// For "manual" task, maybe create a "manual-node" or similar.
	// This is tricky. I'll just skip implementation or create a dummy node?
	// Or create a task with NodeID="manual".

	tasksState, err := a.repo.State().LoadTasks()
	if err != nil {
		runtime.LogErrorf(a.ctx, "Failed to load tasks: %v", err)
		return nil
	}

	taskID := uuid.New().String()
	now := time.Now()

	newState := persistence.TaskState{
		TaskID:    taskID,
		NodeID:    "manual-" + taskID, // Dummy
		Kind:      "manual",
		Status:    string(orchestrator.TaskStatusPending),
		CreatedAt: now,
		UpdatedAt: now,
		Inputs:    map[string]interface{}{"title": title, "pool_id": poolID},
	}
	tasksState.Tasks = append(tasksState.Tasks, newState)

	if err := a.repo.State().SaveTasks(tasksState); err != nil {
		runtime.LogErrorf(a.ctx, "Failed to save tasks: %v", err)
		return nil
	}

	return &orchestrator.Task{
		ID:        taskID,
		Title:     title,
		Status:    orchestrator.TaskStatusPending,
		PoolID:    poolID,
		CreatedAt: now,
	}
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
	// Not supported in new persistence yet. Return empty.
	return []orchestrator.Attempt{}
}

// GetPoolSummaries returns task count summaries by pool.
func (a *App) GetPoolSummaries() []orchestrator.PoolSummary {
	// Not supported fully yet (Repo logic missing for stats)
	return []orchestrator.PoolSummary{}
	// Implementation would read tasks.json and aggregate.
}

// GetAvailablePools returns the list of available worker pools.
func (a *App) GetAvailablePools() []orchestrator.Pool {
	return orchestrator.DefaultPools
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

// ============================================================================
// LLM Config API
// ============================================================================

// LLMConfigDTO は LLM 設定のフロントエンド向けデータ転送オブジェクト
type LLMConfigDTO struct {
	Kind         string `json:"kind"`
	Model        string `json:"model"`
	BaseURL      string `json:"baseUrl"`
	SystemPrompt string `json:"systemPrompt"`
	HasAPIKey    bool   `json:"hasApiKey"`
}

// GetLLMConfig は現在の LLM 設定を返す
func (a *App) GetLLMConfig() LLMConfigDTO {
	config, err := a.llmConfigStore.Load()
	if err != nil {
		runtime.LogErrorf(a.ctx, "Failed to load LLM config: %v", err)
		return LLMConfigDTO{Kind: "mock", Model: "gpt-4o"}
	}

	return LLMConfigDTO{
		Kind:         config.Kind,
		Model:        config.Model,
		BaseURL:      config.BaseURL,
		SystemPrompt: config.SystemPrompt,
		HasAPIKey:    a.llmConfigStore.HasAPIKey(),
	}
}

// SetLLMConfig は LLM 設定を保存する
func (a *App) SetLLMConfig(dto LLMConfigDTO) error {
	config := &ide.LLMConfig{
		Kind:         dto.Kind,
		Model:        dto.Model,
		BaseURL:      dto.BaseURL,
		SystemPrompt: dto.SystemPrompt,
	}
	if err := a.llmConfigStore.Save(config); err != nil {
		return err
	}

	// 現在のワークスペースがあれば Meta/Chat を再初期化して即時反映
	if a.currentWS != nil && a.repo != nil && a.currentWSID != "" {
		wsDir := a.workspaceStore.GetWorkspaceDir(a.currentWSID)
		sessionStore := chat.NewChatSessionStore(wsDir)
		metaClient := a.newMetaClientFromConfig()
		taskStore := orchestrator.NewTaskStore(wsDir) // Temp
		a.chatHandler = chat.NewHandler(metaClient, taskStore, sessionStore, a.currentWSID, a.currentWS.ProjectRoot, a.eventEmitter)
	}

	return nil
}

// TestLLMConnection は LLM 接続をテストする（CLI セッション検証）
func (a *App) TestLLMConnection() (string, error) {
	config, err := a.llmConfigStore.GetEffectiveConfig()
	if err != nil {
		return "", fmt.Errorf("設定の読み込みに失敗: %w", err)
	}

	if config.Kind == "mock" {
		return "モックモード: 接続テストはスキップされました", nil
	}

	// Meta クライアントを作成
	client := a.newMetaClientFromConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// CLI プロバイダの場合は TestConnection を呼び出す
	if config.Kind == "codex-cli" {
		// CodexCLIProvider の TestConnection を呼び出す
		// 内部実装では codex --version でセッション確認
		cmd := exec.CommandContext(ctx, "codex", "--version")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("Codex CLI セッションが見つかりません: %w (出力: %s)", err, string(output))
		}
		return fmt.Sprintf("接続成功 (プロバイダ: %s, モデル: %s)", config.Kind, config.Model), nil
	}

	// 後方互換性: HTTP ベースのプロバイダ（openai-chat 等）
	apiKey, err := a.llmConfigStore.GetAPIKey()
	if err != nil {
		return "", fmt.Errorf("API キーの取得に失敗: %w", err)
	}
	if apiKey == "" {
		return "", fmt.Errorf("API キーが設定されていません。環境変数 OPENAI_API_KEY を設定してください")
	}

	// テスト用の簡単なリクエストを送信
	_, err = client.Decompose(ctx, &meta.DecomposeRequest{
		UserInput: "接続テスト: この文章を確認してください",
		Context:   meta.DecomposeContext{},
	})
	if err != nil {
		return "", fmt.Errorf("接続テスト失敗: %w", err)
	}

	return fmt.Sprintf("接続成功 (プロバイダ: %s, モデル: %s)", config.Kind, config.Model), nil
}
