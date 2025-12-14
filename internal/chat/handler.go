package chat

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/biwakonbu/agent-runner/internal/logging"
	"github.com/biwakonbu/agent-runner/internal/meta"
	"github.com/biwakonbu/agent-runner/internal/orchestrator"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/persistence"
	"github.com/google/uuid"
)

// DefaultChatMetaTimeout は Meta-agent 呼び出しのデフォルトタイムアウト（15分）
// LLM によるタスク分解は時間がかかるため、十分な時間を確保する
const DefaultChatMetaTimeout = 15 * time.Minute

// MetaClient は Meta-agent クライアントのインターフェース
type MetaClient interface {
	Decompose(ctx context.Context, req *meta.DecomposeRequest) (*meta.DecomposeResponse, error)
	PlanPatch(ctx context.Context, req *meta.PlanPatchRequest) (*meta.PlanPatchResponse, error)
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
	Repo         persistence.WorkspaceRepository
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
	repo persistence.WorkspaceRepository,
	events orchestrator.EventEmitter,
) *Handler {
	return &Handler{
		Meta:         metaClient,
		Repo:         repo,
		TaskStore:    taskStore,
		SessionStore: sessionStore,
		WorkspaceID:  workspaceID,
		ProjectRoot:  projectRoot,
		logger:       logging.WithComponent(slog.Default(), "chat-handler"),
		events:       events,
		metaTimeout:  DefaultChatMetaTimeout,
	}
}

// SetLogger はカスタムロガーを設定する
func (h *Handler) SetLogger(logger *slog.Logger) {
	h.logger = logging.WithComponent(logger, "chat-handler")
}

// SetMetaTimeout は Meta-agent 呼び出しのタイムアウトを設定する
func (h *Handler) SetMetaTimeout(timeout time.Duration) {
	h.metaTimeout = timeout
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
	if h.Repo != nil {
		// design/state を真実源として扱い、state/tasks.json に存在するタスクのみを
		// 既存タスクとして Meta へ渡す（削除済みタスクの再登場を防ぐ）。
		if err := h.Repo.Init(); err == nil {
			if tasksState, err := h.Repo.State().LoadTasks(); err == nil && tasksState != nil {
				activeIDs := make(map[string]struct{}, len(tasksState.Tasks))
				for _, ts := range tasksState.Tasks {
					if ts.TaskID != "" {
						activeIDs[ts.TaskID] = struct{}{}
					}
				}
				filtered := make([]orchestrator.Task, 0, len(existingTasks))
				for _, t := range existingTasks {
					if _, ok := activeIDs[t.ID]; ok {
						filtered = append(filtered, t)
					}
				}
				existingTasks = filtered
			}
		}
	}
	existingTaskIDs := make(map[string]struct{}, len(existingTasks))
	existingTasksByID := make(map[string]orchestrator.Task, len(existingTasks))
	for _, t := range existingTasks {
		existingTaskIDs[t.ID] = struct{}{}
		existingTasksByID[t.ID] = t
	}

	// 2. コンテキスト情報を収集 (Event: analyzing)
	emitProgress("Analyzing", "コンテキスト情報を収集中...")
	planPatchReq := h.buildPlanPatchRequest(sessionID, message, existingTasks)

	// 3. Meta-agent を呼び出して計画更新 (Event: planning)
	emitProgress("Planning", "Meta-agent が計画を更新中...")
	logger.Debug("calling meta-agent for plan_patch")
	metaCtx, cancel := context.WithTimeout(ctx, h.metaTimeout)
	defer cancel()

	patchResp, err := h.Meta.PlanPatch(metaCtx, planPatchReq)
	if err != nil {
		emitFailed(fmt.Sprintf("計画更新に失敗しました: %v", err))
		// エラー時もアシスタントメッセージを返す
		errMsg := &ChatMessage{
			ID:        uuid.New().String(),
			SessionID: sessionID,
			Role:      "assistant",
			Content:   fmt.Sprintf("申し訳ありません。計画更新中にエラーが発生しました: %v", err),
			Timestamp: time.Now(),
		}
		if appendErr := h.SessionStore.AppendMessage(errMsg); appendErr != nil {
			return nil, fmt.Errorf("meta-agent plan_patch failed: %v (assistant message save failed: %w)", err, appendErr)
		}
		return &ChatResponse{
			Message: *errMsg,
		}, fmt.Errorf("meta-agent plan_patch failed: %w", err)
	}

	// 4. 計画変更を永続化 (Event: persisting)
	emitProgress("Persisting", fmt.Sprintf("%d 個の変更を保存中...", len(patchResp.Operations)))
	applyRes, err := h.applyPlanPatch(ctx, sessionID, patchResp, existingTaskIDs, existingTasksByID)
	if err != nil {
		emitFailed(fmt.Sprintf("計画変更の保存に失敗しました: %v", err))
		return nil, fmt.Errorf("failed to apply plan patch: %w", err)
	}

	// Filter potential conflicts against actual workspace files.
	filteredConflicts := h.filterPotentialConflicts(patchResp.PotentialConflicts)
	if len(filteredConflicts) != len(patchResp.PotentialConflicts) {
		patchResp.PotentialConflicts = filteredConflicts
	}

	// 5. アシスタント応答メッセージを作成 (Event: completed)
	emitProgress("Completed", "処理が完了しました。")
	responseContent := h.buildPlanPatchResponseContent(patchResp, applyRes)
	taskIDs := make([]string, len(applyRes.CreatedTasks))
	for i, t := range applyRes.CreatedTasks {
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
		slog.Int("created_tasks", len(applyRes.CreatedTasks)),
		logging.LogDuration(start),
	)

	return &ChatResponse{
		Message:        *assistantMsg,
		GeneratedTasks: applyRes.CreatedTasks,
		Understanding:  patchResp.Understanding,
		Conflicts:      filteredConflicts,
	}, nil
}

// BuildDecomposeRequest は Meta-agent への分解リクエストを構築する
// NOTE: 現在は plan_patch が主要フロー。decompose は将来の batched generation 用途で維持。
// PRD 13.3 #5: 削除は将来タスクとし、現時点では維持する。
func (h *Handler) BuildDecomposeRequest(sessionID, message string, existingTasks []orchestrator.Task) *meta.DecomposeRequest {
	taskSummaries := make([]meta.ExistingTaskSummary, len(existingTasks))
	for i, t := range existingTasks {
		taskSummaries[i] = meta.ExistingTaskSummary{
			ID:           t.ID,
			Title:        t.Title,
			Status:       string(t.Status),
			Dependencies: t.Dependencies,
			PhaseName:    t.PhaseName,
			Milestone:    t.Milestone,
			WBSLevel:     t.WBSLevel,
			ParentID:     t.ParentID,
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

// PersistTasks は分解されたタスクを永続化する
func (h *Handler) PersistTasks(
	ctx context.Context,
	sessionID string,
	resp *meta.DecomposeResponse,
	existingTaskIDs map[string]struct{},
	existingTasksByID map[string]orchestrator.Task,
) ([]orchestrator.Task, error) {
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
				Milestone:          phase.Milestone,
				SourceChatID:       &sessionID,
				AcceptanceCriteria: decomposedTask.AcceptanceCriteria,
				Runner: &orchestrator.RunnerSpec{
					MaxLoops:   orchestrator.DefaultRunnerMaxLoops,
					WorkerKind: orchestrator.DefaultWorkerKind,
				},
			}

			if decomposedTask.SuggestedImpl != nil {
				// Validate file paths
				validatedPaths := h.validateFilePaths(decomposedTask.SuggestedImpl.FilePaths)

				task.SuggestedImpl = &orchestrator.SuggestedImpl{
					Language:    decomposedTask.SuggestedImpl.Language,
					FilePaths:   validatedPaths,
					Constraints: decomposedTask.SuggestedImpl.Constraints,
				}
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

	if err := h.persistDesignAndState(ctx, sessionID, tasksToSave, existingTasksByID); err != nil {
		logger.Error("failed to persist design/state", slog.Any("error", err))
		return nil, fmt.Errorf("failed to persist design/state: %w", err)
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

		// Emit real-time event
		if h.events != nil {
			h.events.Emit(orchestrator.EventTaskCreated, orchestrator.TaskCreatedEvent{
				Task: task,
			})
		}
	}

	return allTasks, nil
}

// persistDesignAndState は decompose されたタスクを design/state に反映する（Repo が nil の場合はスキップ）。
func (h *Handler) persistDesignAndState(
	ctx context.Context,
	sessionID string,
	tasks []orchestrator.Task,
	existingTasksByID map[string]orchestrator.Task,
) error {
	if h.Repo == nil {
		return nil
	}
	logger := logging.WithTraceID(h.logger, ctx)

	if err := h.Repo.Init(); err != nil {
		return fmt.Errorf("failed to init workspace repo: %w", err)
	}

	now := time.Now()

	// Load or create WBS root
	wbs, err := h.Repo.Design().LoadWBS()
	if err != nil {
		if os.IsNotExist(err) {
			wbs = &persistence.WBS{
				WBSID:       uuid.New().String(),
				ProjectRoot: h.ProjectRoot,
				CreatedAt:   now,
				UpdatedAt:   now,
				RootNodeID:  "node-root",
				NodeIndex:   []persistence.NodeIndex{},
			}
		} else {
			return fmt.Errorf("failed to load wbs: %w", err)
		}
	}

	if wbs.WBSID == "" {
		wbs.WBSID = uuid.New().String()
		wbs.CreatedAt = now
	}
	if wbs.ProjectRoot == "" {
		wbs.ProjectRoot = h.ProjectRoot
	}
	if wbs.RootNodeID == "" {
		wbs.RootNodeID = "node-root"
	}
	wbs.UpdatedAt = now

	// Index lookup (store positions to avoid slice pointer invalidation).
	indexPosByID := make(map[string]int)
	for i := range wbs.NodeIndex {
		indexPosByID[wbs.NodeIndex[i].NodeID] = i
	}
	rootPos, ok := indexPosByID[wbs.RootNodeID]
	if !ok {
		rootID := wbs.RootNodeID
		wbs.NodeIndex = append(wbs.NodeIndex, persistence.NodeIndex{
			NodeID:   rootID,
			ParentID: nil,
			Children: []string{},
		})
		rootPos = len(wbs.NodeIndex) - 1
		indexPosByID[rootID] = rootPos
	}

	contains := func(xs []string, x string) bool {
		for _, v := range xs {
			if v == x {
				return true
			}
		}
		return false
	}

	// Track new nodes for dependency handling
	newNodeIDs := make(map[string]struct{}, len(tasks))

	// Save NodeDesign for each task and update WBS index
	for _, t := range tasks {
		newNodeIDs[t.ID] = struct{}{}

		suggested := persistence.SuggestedImpl{}
		if t.SuggestedImpl != nil {
			paths := make([]string, 0, len(t.SuggestedImpl.FilePaths))
			for _, p := range t.SuggestedImpl.FilePaths {
				paths = append(paths, strings.TrimSuffix(p, " (New File)"))
			}
			suggested.Language = t.SuggestedImpl.Language
			suggested.FilePaths = paths
			suggested.Constraints = t.SuggestedImpl.Constraints
		}

		node := &persistence.NodeDesign{
			NodeID:             t.ID,
			WBSID:              wbs.WBSID,
			Name:               t.Title,
			Summary:            t.Description,
			PhaseName:          t.PhaseName,
			Milestone:          t.Milestone,
			WBSLevel:           t.WBSLevel,
			Kind:               "feature",
			Priority:           "medium",
			Estimate:           persistence.Estimate{},
			Dependencies:       t.Dependencies,
			AcceptanceCriteria: t.AcceptanceCriteria,
			DesignNotes:        []string{},
			SuggestedImpl:      suggested,
			CreatedAt:          now,
			UpdatedAt:          now,
			CreatedBy:          "agent:planner",
		}

		if err := h.Repo.Design().SaveNode(node); err != nil {
			return fmt.Errorf("failed to save node %s: %w", node.NodeID, err)
		}

		if _, exists := indexPosByID[t.ID]; !exists {
			parentID := wbs.RootNodeID
			wbs.NodeIndex = append(wbs.NodeIndex, persistence.NodeIndex{
				NodeID:   t.ID,
				ParentID: &parentID,
				Children: []string{},
			})
			pos := len(wbs.NodeIndex) - 1
			indexPosByID[t.ID] = pos
			if !contains(wbs.NodeIndex[rootPos].Children, t.ID) {
				wbs.NodeIndex[rootPos].Children = append(wbs.NodeIndex[rootPos].Children, t.ID)
			}
		}
	}

	if err := h.Repo.Design().SaveWBS(wbs); err != nil {
		return fmt.Errorf("failed to save wbs: %w", err)
	}

	// Load state
	nodesRuntime, err := h.Repo.State().LoadNodesRuntime()
	if err != nil {
		return fmt.Errorf("failed to load nodes runtime: %w", err)
	}
	tasksState, err := h.Repo.State().LoadTasks()
	if err != nil {
		return fmt.Errorf("failed to load tasks state: %w", err)
	}

	runtimeByID := make(map[string]*persistence.NodeRuntime)
	for i := range nodesRuntime.Nodes {
		runtimeByID[nodesRuntime.Nodes[i].NodeID] = &nodesRuntime.Nodes[i]
	}

	taskByID := make(map[string]struct{})
	for _, ts := range tasksState.Tasks {
		taskByID[ts.TaskID] = struct{}{}
	}

	// Ensure runtime entries for existing dependency nodes to avoid permanent blocking.
	for _, t := range tasks {
		for _, depID := range t.Dependencies {
			if _, isNew := newNodeIDs[depID]; isNew {
				continue
			}
			if _, exists := runtimeByID[depID]; exists {
				continue
			}
			existing, ok := existingTasksByID[depID]
			if !ok {
				continue
			}
			status := "planned"
			if existing.Status == orchestrator.TaskStatusSucceeded || existing.Status == orchestrator.TaskStatusCompleted {
				status = "implemented"
			}
			nodesRuntime.Nodes = append(nodesRuntime.Nodes, persistence.NodeRuntime{
				NodeID: depID,
				Status: status,
				Implementation: persistence.NodeImplementation{
					Files:          []string{},
					LastModifiedAt: now,
					LastModifiedBy: "chat-handler",
				},
				Verification: persistence.NodeVerification{
					Status: "not_tested",
				},
				Notes: []persistence.NodeNote{
					{At: now, By: "chat-handler", Text: "imported from existing task"},
				},
			})
			runtimeByID[depID] = &nodesRuntime.Nodes[len(nodesRuntime.Nodes)-1]
		}
	}

	// Upsert runtime/tasks for new nodes.
	for _, t := range tasks {
		if _, exists := runtimeByID[t.ID]; !exists {
			nodesRuntime.Nodes = append(nodesRuntime.Nodes, persistence.NodeRuntime{
				NodeID: t.ID,
				Status: "planned",
				Implementation: persistence.NodeImplementation{
					Files:          []string{},
					LastModifiedAt: now,
					LastModifiedBy: "chat-handler",
				},
				Verification: persistence.NodeVerification{
					Status: "not_tested",
				},
				Notes: []persistence.NodeNote{
					{At: now, By: "chat-handler", Text: fmt.Sprintf("created from chat session %s", sessionID)},
				},
			})
			runtimeByID[t.ID] = &nodesRuntime.Nodes[len(nodesRuntime.Nodes)-1]
		}

		if _, exists := taskByID[t.ID]; !exists {
			tasksState.Tasks = append(tasksState.Tasks, persistence.TaskState{
				TaskID:        t.ID,
				NodeID:        t.ID,
				Kind:          "implementation",
				Status:        string(orchestrator.TaskStatusPending),
				CreatedAt:     now,
				UpdatedAt:     now,
				ScheduledBy:   "chat-handler",
				AssignedAgent: "",
				Priority:      0,
				Inputs: map[string]interface{}{
					orchestrator.InputKeyAttemptCount:     0,
					orchestrator.InputKeyRunnerMaxLoops:   orchestrator.DefaultRunnerMaxLoops,
					orchestrator.InputKeyRunnerWorkerKind: orchestrator.DefaultWorkerKind,
				},
				Outputs: persistence.TaskOutputs{},
			})
			taskByID[t.ID] = struct{}{}
		}
	}

	if err := h.Repo.State().SaveNodesRuntime(nodesRuntime); err != nil {
		return fmt.Errorf("failed to save nodes runtime: %w", err)
	}
	if err := h.Repo.State().SaveTasks(tasksState); err != nil {
		return fmt.Errorf("failed to save tasks state: %w", err)
	}

	logger.Debug("persisted design/state from chat",
		slog.Int("nodes", len(tasks)),
	)

	return nil
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

// CountTotalTasks counts tasks across all phases
func CountTotalTasks(resp *meta.DecomposeResponse) int {
	count := 0
	for _, p := range resp.Phases {
		count += len(p.Tasks)
	}
	return count
}

// validateFilePaths checks if files exist and annotates them if they are new.
func (h *Handler) validateFilePaths(paths []string) []string {
	validated := make([]string, len(paths))
	for i, path := range paths {
		// Resolve path to absolute for checking
		checkPath := path
		if !filepath.IsAbs(path) {
			checkPath = filepath.Join(h.ProjectRoot, path)
		}

		if _, err := os.Stat(checkPath); os.IsNotExist(err) {
			validated[i] = fmt.Sprintf("%s (New File)", path)
		} else {
			validated[i] = path
		}
	}
	return validated
}

// filterPotentialConflicts removes conflicts for files that do not exist in the workspace.
// potential_conflicts はヒューリスティック出力なので false positive を含む。
func (h *Handler) filterPotentialConflicts(conflicts []meta.PotentialConflict) []meta.PotentialConflict {
	if h.ProjectRoot == "" || len(conflicts) == 0 {
		return conflicts
	}

	filtered := make([]meta.PotentialConflict, 0, len(conflicts))
	for _, c := range conflicts {
		file := strings.TrimSpace(c.File)
		if file == "" {
			continue
		}
		checkPath := file
		if !filepath.IsAbs(file) {
			checkPath = filepath.Join(h.ProjectRoot, file)
		}

		if _, err := os.Stat(checkPath); err == nil {
			filtered = append(filtered, c)
			continue
		} else if os.IsNotExist(err) {
			if h.logger != nil {
				h.logger.Debug("dropping potential conflict for non-existent file",
					slog.String("file", file),
				)
			}
			continue
		}

		// Non-ENOENT errors (permission, etc.): keep to avoid hiding real conflicts.
		filtered = append(filtered, c)
	}

	return filtered
}
