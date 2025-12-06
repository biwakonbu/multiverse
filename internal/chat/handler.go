package chat

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/biwakonbu/agent-runner/internal/logging"
	"github.com/biwakonbu/agent-runner/internal/meta"
	"github.com/biwakonbu/agent-runner/internal/orchestrator"
	"github.com/google/uuid"
)

// MetaClient は Meta-agent クライアントのインターフェース
type MetaClient interface {
	Decompose(ctx context.Context, req *meta.DecomposeRequest) (*meta.DecomposeResponse, error)
}

// ChatResponse はチャット応答を表す
type ChatResponse struct {
	Message        ChatMessage              `json:"message"`        // アシスタントの応答メッセージ
	GeneratedTasks []orchestrator.Task      `json:"generatedTasks"` // 生成されたタスク
	Understanding  string                   `json:"understanding"`  // ユーザー意図の理解
	Conflicts      []meta.PotentialConflict `json:"conflicts"`      // 潜在的なコンフリクト
}

// Handler はチャットメッセージを処理するハンドラ
type Handler struct {
	Meta         MetaClient
	TaskStore    *orchestrator.TaskStore
	SessionStore *ChatSessionStore
	WorkspaceID  string
	ProjectRoot  string
	logger       *slog.Logger
	events       orchestrator.EventEmitter
	metaTimeout  time.Duration
}

// NewHandler は新しい ChatHandler を作成する
func NewHandler(
	metaClient MetaClient,
	taskStore *orchestrator.TaskStore,
	sessionStore *ChatSessionStore,
	workspaceID string,
	projectRoot string,
	events orchestrator.EventEmitter,
) *Handler {
	return &Handler{
		Meta:         metaClient,
		TaskStore:    taskStore,
		SessionStore: sessionStore,
		WorkspaceID:  workspaceID,
		ProjectRoot:  projectRoot,
		logger:       logging.WithComponent(slog.Default(), "chat-handler"),
		events:       events,
		metaTimeout:  30 * time.Second,
	}
}

// SetLogger はカスタムロガーを設定する
func (h *Handler) SetLogger(logger *slog.Logger) {
	h.logger = logging.WithComponent(logger, "chat-handler")
}

// HandleMessage はユーザーメッセージを処理し、タスクを生成する
func (h *Handler) HandleMessage(ctx context.Context, sessionID, message string) (*ChatResponse, error) {
	logger := logging.WithTraceID(h.logger, ctx)
	start := time.Now()

	emitProgress := func(step, msg string) {
		if h.events != nil {
			h.events.Emit(orchestrator.EventChatProgress, orchestrator.ChatProgressEvent{
				SessionID: sessionID,
				Step:      step,
				Message:   msg,
				Timestamp: time.Now(),
			})
		}
	}

	emitFailed := func(msg string) {
		emitProgress("Failed", msg)
	}

	logger.Info("handling chat message",
		slog.String("session_id", sessionID),
		slog.Int("message_length", len(message)),
	)

	// 1. ユーザーメッセージを保存 (Event: processing)
	emitProgress("Processing", "メッセージを受信しました...")

	userMsg := &ChatMessage{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      "user",
		Content:   message,
		Timestamp: time.Now(),
	}
	if err := h.SessionStore.AppendMessage(userMsg); err != nil {
		emitFailed(fmt.Sprintf("ユーザーメッセージの保存に失敗しました: %v", err))
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	existingTasks, err := h.TaskStore.ListAllTasks()
	if err != nil {
		emitFailed(fmt.Sprintf("既存タスクの取得に失敗しました: %v", err))
		return nil, fmt.Errorf("failed to list existing tasks: %w", err)
	}
	existingTaskIDs := make(map[string]struct{}, len(existingTasks))
	for _, t := range existingTasks {
		existingTaskIDs[t.ID] = struct{}{}
	}

	// 2. コンテキスト情報を収集 (Event: analyzing)
	emitProgress("Analyzing", "コンテキスト情報を収集中...")
	decomposeReq := h.buildDecomposeRequest(sessionID, message, existingTasks)

	// 3. Meta-agent を呼び出してタスク分解 (Event: decomposing)
	emitProgress("Decomposing", "Meta-agent がタスクを分解中...")
	logger.Debug("calling meta-agent for decompose")
	metaCtx, cancel := context.WithTimeout(ctx, h.metaTimeout)
	defer cancel()

	decomposeResp, err := h.Meta.Decompose(metaCtx, decomposeReq)
	if err != nil {
		emitFailed(fmt.Sprintf("タスク分解に失敗しました: %v", err))
		// エラー時もアシスタントメッセージを返す
		errMsg := &ChatMessage{
			ID:        uuid.New().String(),
			SessionID: sessionID,
			Role:      "assistant",
			Content:   fmt.Sprintf("申し訳ありません。タスク分解中にエラーが発生しました: %v", err),
			Timestamp: time.Now(),
		}
		if appendErr := h.SessionStore.AppendMessage(errMsg); appendErr != nil {
			return nil, fmt.Errorf("meta-agent decompose failed: %v (assistant message save failed: %w)", err, appendErr)
		}
		return &ChatResponse{
			Message: *errMsg,
		}, fmt.Errorf("meta-agent decompose failed: %w", err)
	}

	// 4. タスクを永続化 (Event: persisting)
	emitProgress("Persisting", fmt.Sprintf("%d 個のタスクを保存中...", countTotalTasks(decomposeResp)))
	generatedTasks, err := h.persistTasks(ctx, sessionID, decomposeResp, existingTaskIDs)
	if err != nil {
		emitFailed(fmt.Sprintf("タスク保存に失敗しました: %v", err))
		return nil, fmt.Errorf("failed to persist tasks: %w", err)
	}

	// 5. アシスタント応答メッセージを作成 (Event: completed)
	emitProgress("Completed", "処理が完了しました。")
	responseContent := h.buildResponseContent(decomposeResp, generatedTasks)
	taskIDs := make([]string, len(generatedTasks))
	for i, t := range generatedTasks {
		taskIDs[i] = t.ID
	}

	assistantMsg := &ChatMessage{
		ID:             uuid.New().String(),
		SessionID:      sessionID,
		Role:           "assistant",
		Content:        responseContent,
		Timestamp:      time.Now(),
		GeneratedTasks: taskIDs,
	}
	if err := h.SessionStore.AppendMessage(assistantMsg); err != nil {
		return nil, fmt.Errorf("failed to save assistant message: %w", err)
	}

	logger.Info("chat message handled",
		slog.Int("generated_tasks", len(generatedTasks)),
		logging.LogDuration(start),
	)

	return &ChatResponse{
		Message:        *assistantMsg,
		GeneratedTasks: generatedTasks,
		Understanding:  decomposeResp.Understanding,
		Conflicts:      decomposeResp.PotentialConflicts,
	}, nil
}

// buildDecomposeRequest は Meta-agent への分解リクエストを構築する
func (h *Handler) buildDecomposeRequest(sessionID, message string, existingTasks []orchestrator.Task) *meta.DecomposeRequest {
	taskSummaries := make([]meta.ExistingTaskSummary, len(existingTasks))
	for i, t := range existingTasks {
		taskSummaries[i] = meta.ExistingTaskSummary{
			ID:           t.ID,
			Title:        t.Title,
			Status:       string(t.Status),
			Dependencies: t.Dependencies,
			PhaseName:    t.PhaseName,
		}
	}

	// 会話履歴を取得（最新10件）
	recentMessages, err := h.SessionStore.GetRecentMessages(sessionID, 10)
	if err != nil {
		recentMessages = []ChatMessage{}
	}

	conversationHistory := make([]meta.ConversationMessage, len(recentMessages))
	for i, m := range recentMessages {
		conversationHistory[i] = meta.ConversationMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	return &meta.DecomposeRequest{
		UserInput: message,
		Context: meta.DecomposeContext{
			WorkspacePath:       h.ProjectRoot,
			ExistingTasks:       taskSummaries,
			ConversationHistory: conversationHistory,
		},
	}
}

// persistTasks は分解されたタスクを永続化する
func (h *Handler) persistTasks(ctx context.Context, sessionID string, resp *meta.DecomposeResponse, existingTaskIDs map[string]struct{}) ([]orchestrator.Task, error) {
	logger := logging.WithTraceID(h.logger, ctx)

	// 一時ID → 正式ID のマッピング
	idMapping := make(map[string]string)
	var allTasks []orchestrator.Task

	now := time.Now()

	// 1st pass: すべての新規タスクに正式 ID を割り当てる
	for _, phase := range resp.Phases {
		for _, decomposedTask := range phase.Tasks {
			idMapping[decomposedTask.ID] = uuid.New().String()
		}
	}

	// 2nd pass: 依存解決しつつ Task を構築
	var tasksToSave []orchestrator.Task
	var unresolvedDeps []string

	for _, phase := range resp.Phases {
		for _, decomposedTask := range phase.Tasks {
			taskID := idMapping[decomposedTask.ID]

			dependencies := make([]string, 0, len(decomposedTask.Dependencies))
			for _, depID := range decomposedTask.Dependencies {
				if realID, ok := idMapping[depID]; ok {
					dependencies = append(dependencies, realID)
					continue
				}
				if _, ok := existingTaskIDs[depID]; ok {
					dependencies = append(dependencies, depID)
					continue
				}
				unresolvedDeps = append(unresolvedDeps, depID)
			}

			task := orchestrator.Task{
				ID:                 taskID,
				Title:              decomposedTask.Title,
				Description:        decomposedTask.Description,
				Status:             orchestrator.TaskStatusPending,
				PoolID:             "default",
				CreatedAt:          now,
				UpdatedAt:          now,
				Dependencies:       dependencies,
				WBSLevel:           decomposedTask.WBSLevel,
				PhaseName:          phase.Name,
				SourceChatID:       &sessionID,
				AcceptanceCriteria: decomposedTask.AcceptanceCriteria,
			}

			tasksToSave = append(tasksToSave, task)
		}
	}

	if len(unresolvedDeps) > 0 {
		seen := make(map[string]struct{})
		unique := make([]string, 0, len(unresolvedDeps))
		for _, dep := range unresolvedDeps {
			if _, ok := seen[dep]; ok {
				continue
			}
			seen[dep] = struct{}{}
			unique = append(unique, dep)
		}
		return nil, fmt.Errorf("unresolved dependencies: %s", strings.Join(unique, ", "))
	}

	for _, task := range tasksToSave {
		if err := h.TaskStore.SaveTask(&task); err != nil {
			logger.Error("failed to save task",
				slog.String("task_id", task.ID),
				slog.Any("error", err),
			)
			return nil, fmt.Errorf("failed to save task %s: %w", task.ID, err)
		}

		allTasks = append(allTasks, task)
		logger.Debug("task created",
			slog.String("task_id", task.ID),
			slog.String("title", task.Title),
			slog.String("phase", task.PhaseName),
		)
	}

	return allTasks, nil
}

// buildResponseContent はアシスタント応答メッセージを構築する
func (h *Handler) buildResponseContent(resp *meta.DecomposeResponse, tasks []orchestrator.Task) string {
	var content string

	content += resp.Understanding + "\n\n"

	if len(tasks) > 0 {
		content += fmt.Sprintf("以下の %d 個のタスクを作成しました：\n\n", len(tasks))

		currentPhase := ""
		for _, task := range tasks {
			if task.PhaseName != currentPhase {
				currentPhase = task.PhaseName
				content += fmt.Sprintf("### %s\n", currentPhase)
			}
			content += fmt.Sprintf("- **%s**: %s\n", task.Title, task.Description)
		}
	}

	if len(resp.PotentialConflicts) > 0 {
		content += "\n\n**注意**: 以下のファイルで潜在的なコンフリクトが検出されました：\n"
		for _, conflict := range resp.PotentialConflicts {
			content += fmt.Sprintf("- `%s`: %s\n", conflict.File, conflict.Warning)
		}
	}

	return content
}

// CreateSession は新しいチャットセッションを作成する
func (h *Handler) CreateSession(ctx context.Context) (*ChatSession, error) {
	sessionID := uuid.New().String()
	session, err := h.SessionStore.CreateSession(sessionID, h.WorkspaceID)
	if err != nil {
		return nil, err
	}

	// システムメッセージを追加
	sysMsg := &ChatMessage{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      "system",
		Content:   "チャットセッションが開始されました。開発したい機能や解決したい課題を教えてください。",
		Timestamp: time.Now(),
	}
	if err := h.SessionStore.AppendMessage(sysMsg); err != nil {
		return nil, err
	}

	return session, nil
}

// GetHistory はセッションのメッセージ履歴を取得する
func (h *Handler) GetHistory(ctx context.Context, sessionID string) ([]ChatMessage, error) {
	return h.SessionStore.LoadMessages(sessionID)
}

// countTotalTasks counts tasks across all phases
func countTotalTasks(resp *meta.DecomposeResponse) int {
	count := 0
	for _, p := range resp.Phases {
		count += len(p.Tasks)
	}
	return count
}
