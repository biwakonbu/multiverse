package chat

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/biwakonbu/agent-runner/internal/meta"
	"github.com/biwakonbu/agent-runner/internal/orchestrator"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/persistence"
)

// MockMetaClient は MetaClient のモック実装
type MockMetaClient struct {
	DecomposeFunc func(ctx context.Context, req *meta.DecomposeRequest) (*meta.DecomposeResponse, error)
}

func (m *MockMetaClient) Decompose(ctx context.Context, req *meta.DecomposeRequest) (*meta.DecomposeResponse, error) {
	if m.DecomposeFunc != nil {
		return m.DecomposeFunc(ctx, req)
	}
	return nil, errors.New("DecomposeFunc not set")
}

func TestHandler_NewHandler(t *testing.T) {
	tmpDir := t.TempDir()
	taskStore := orchestrator.NewTaskStore(tmpDir)
	sessionStore := NewChatSessionStore(tmpDir)
	mockMeta := &MockMetaClient{}

	handler := NewHandler(mockMeta, taskStore, sessionStore, "workspace-1", "/project", nil, nil)

	if handler.Meta != mockMeta {
		t.Error("Meta client not set correctly")
	}
	if handler.TaskStore != taskStore {
		t.Error("TaskStore not set correctly")
	}
	if handler.SessionStore != sessionStore {
		t.Error("SessionStore not set correctly")
	}
	if handler.WorkspaceID != "workspace-1" {
		t.Errorf("WorkspaceID expected 'workspace-1', got '%s'", handler.WorkspaceID)
	}
	if handler.ProjectRoot != "/project" {
		t.Errorf("ProjectRoot expected '/project', got '%s'", handler.ProjectRoot)
	}
}

func TestHandler_CreateSession(t *testing.T) {
	tmpDir := t.TempDir()
	taskStore := orchestrator.NewTaskStore(tmpDir)
	sessionStore := NewChatSessionStore(tmpDir)
	mockMeta := &MockMetaClient{}

	handler := NewHandler(mockMeta, taskStore, sessionStore, "workspace-1", "/project", nil, nil)

	ctx := context.Background()
	session, err := handler.CreateSession(ctx)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if session.ID == "" {
		t.Error("session ID should not be empty")
	}
	if session.WorkspaceID != "workspace-1" {
		t.Errorf("expected workspace ID 'workspace-1', got '%s'", session.WorkspaceID)
	}

	// システムメッセージが作成されていることを確認
	messages, err := handler.GetHistory(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("expected 1 system message, got %d", len(messages))
	}

	if messages[0].Role != "system" {
		t.Errorf("expected role 'system', got '%s'", messages[0].Role)
	}
}

func TestHandler_HandleMessage_Success(t *testing.T) {
	tmpDir := t.TempDir()
	taskStore := orchestrator.NewTaskStore(tmpDir)
	sessionStore := NewChatSessionStore(tmpDir)

	mockMeta := &MockMetaClient{
		DecomposeFunc: func(ctx context.Context, req *meta.DecomposeRequest) (*meta.DecomposeResponse, error) {
			return &meta.DecomposeResponse{
				Understanding: "ユーザーは認証機能の実装を要求しています。",
				Phases: []meta.DecomposedPhase{
					{
						Name:      "概念設計",
						Milestone: "M1-Auth",
						Tasks: []meta.DecomposedTask{
							{
								ID:                 "temp-task-1",
								Title:              "認証フロー設計",
								Description:        "ログイン/ログアウトフローを設計",
								AcceptanceCriteria: []string{"フロー図が作成される"},
								Dependencies:       []string{},
								WBSLevel:           1,
							},
						},
					},
					{
						Name:      "実装",
						Milestone: "M2-Auth",
						Tasks: []meta.DecomposedTask{
							{
								ID:                 "temp-task-2",
								Title:              "ログイン画面実装",
								Description:        "ログイン画面のUI実装",
								AcceptanceCriteria: []string{"ログイン画面が表示される"},
								Dependencies:       []string{"temp-task-1"},
								WBSLevel:           2,
							},
						},
					},
				},
				PotentialConflicts: []meta.PotentialConflict{},
			}, nil
		},
	}

	handler := NewHandler(mockMeta, taskStore, sessionStore, "workspace-1", "/project", nil, nil)

	ctx := context.Background()

	// セッションを作成
	session, err := handler.CreateSession(ctx)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// メッセージを送信
	resp, err := handler.HandleMessage(ctx, session.ID, "認証機能を実装してほしい")
	if err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	// レスポンスを検証
	if resp.Understanding != "ユーザーは認証機能の実装を要求しています。" {
		t.Errorf("unexpected understanding: %s", resp.Understanding)
	}

	if len(resp.GeneratedTasks) != 2 {
		t.Fatalf("expected 2 generated tasks, got %d", len(resp.GeneratedTasks))
	}

	// タスクの依存関係が正しく変換されていることを確認
	task2 := resp.GeneratedTasks[1]
	if len(task2.Dependencies) != 1 {
		t.Errorf("expected 1 dependency, got %d", len(task2.Dependencies))
	}
	// 依存先が正式IDに変換されていることを確認
	if task2.Dependencies[0] == "temp-task-1" {
		t.Error("dependency should be converted to real ID")
	}
	if task2.Dependencies[0] != resp.GeneratedTasks[0].ID {
		t.Errorf("dependency should point to first task's ID")
	}

	// メッセージ履歴を確認
	messages, err := handler.GetHistory(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	// システム + ユーザー + アシスタント = 3件
	if len(messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(messages))
	}

	// ユーザーメッセージ
	if messages[1].Role != "user" {
		t.Errorf("expected role 'user', got '%s'", messages[1].Role)
	}
	if messages[1].Content != "認証機能を実装してほしい" {
		t.Errorf("unexpected user message content")
	}

	// アシスタントメッセージ
	if messages[2].Role != "assistant" {
		t.Errorf("expected role 'assistant', got '%s'", messages[2].Role)
	}
	if len(messages[2].GeneratedTasks) != 2 {
		t.Errorf("expected 2 generated task IDs, got %d", len(messages[2].GeneratedTasks))
	}
}

func TestHandler_HandleMessage_PersistsDesignAndState(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatalf("failed to create project root: %v", err)
	}

	taskStore := orchestrator.NewTaskStore(tmpDir)
	sessionStore := NewChatSessionStore(tmpDir)
	repo := persistence.NewWorkspaceRepository(tmpDir)
	if err := repo.Init(); err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	mockMeta := &MockMetaClient{
		DecomposeFunc: func(ctx context.Context, req *meta.DecomposeRequest) (*meta.DecomposeResponse, error) {
			return &meta.DecomposeResponse{
				Understanding: "理解しました",
				Phases: []meta.DecomposedPhase{
					{
						Name:      "概念設計",
						Milestone: "M1",
						Tasks: []meta.DecomposedTask{
							{
								ID:           "temp-task-1",
								Title:        "設計タスク",
								Description:  "設計の説明",
								Dependencies: []string{},
								WBSLevel:     1,
							},
						},
					},
					{
						Name:      "実装",
						Milestone: "M2",
						Tasks: []meta.DecomposedTask{
							{
								ID:           "temp-task-2",
								Title:        "実装タスク",
								Description:  "実装の説明",
								Dependencies: []string{"temp-task-1"},
								WBSLevel:     2,
							},
						},
					},
				},
			}, nil
		},
	}

	handler := NewHandler(mockMeta, taskStore, sessionStore, "workspace-1", projectRoot, repo, nil)
	ctx := context.Background()
	session, err := handler.CreateSession(ctx)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	resp, err := handler.HandleMessage(ctx, session.ID, "テスト")
	if err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}
	if len(resp.GeneratedTasks) != 2 {
		t.Fatalf("expected 2 generated tasks, got %d", len(resp.GeneratedTasks))
	}

	// WBS が作成され、Root にタスクが紐づく
	wbs, err := repo.Design().LoadWBS()
	if err != nil {
		t.Fatalf("failed to load wbs: %v", err)
	}
	if wbs.RootNodeID == "" {
		t.Fatalf("root node id should not be empty")
	}
	rootChildren := map[string]struct{}{}
	for _, idx := range wbs.NodeIndex {
		if idx.NodeID == wbs.RootNodeID {
			for _, c := range idx.Children {
				rootChildren[c] = struct{}{}
			}
		}
	}

	// NodesRuntime / TasksState を一度だけロードして検証する
	nodesRuntime, err := repo.State().LoadNodesRuntime()
	if err != nil {
		t.Fatalf("failed to load nodes runtime: %v", err)
	}
	runtimeIDs := map[string]struct{}{}
	for _, rt := range nodesRuntime.Nodes {
		runtimeIDs[rt.NodeID] = struct{}{}
	}

	tasksState, err := repo.State().LoadTasks()
	if err != nil {
		t.Fatalf("failed to load tasks state: %v", err)
	}
	taskStatesByID := map[string]persistence.TaskState{}
	for _, ts := range tasksState.Tasks {
		taskStatesByID[ts.TaskID] = ts
	}

	for _, task := range resp.GeneratedTasks {
		if _, ok := rootChildren[task.ID]; !ok {
			t.Errorf("expected task %s to be a child of root in wbs", task.ID)
		}

		// NodeDesign が保存されている
		node, err := repo.Design().GetNode(task.ID)
		if err != nil {
			t.Fatalf("failed to load node design %s: %v", task.ID, err)
		}
		if node.Name != task.Title {
			t.Errorf("expected node name %q, got %q", task.Title, node.Name)
		}

		// NodesRuntime が作られている
		if _, ok := runtimeIDs[task.ID]; !ok {
			t.Errorf("expected node runtime for %s", task.ID)
		}

		// TasksState が作られている
		ts, ok := taskStatesByID[task.ID]
		if !ok {
			t.Errorf("expected task state for %s", task.ID)
			continue
		}
		if ts.NodeID != task.ID {
			t.Errorf("expected task state node_id %s, got %s", task.ID, ts.NodeID)
		}
	}
}

func TestHandler_HandleMessage_MetaError(t *testing.T) {
	tmpDir := t.TempDir()
	taskStore := orchestrator.NewTaskStore(tmpDir)
	sessionStore := NewChatSessionStore(tmpDir)

	mockMeta := &MockMetaClient{
		DecomposeFunc: func(ctx context.Context, req *meta.DecomposeRequest) (*meta.DecomposeResponse, error) {
			return nil, errors.New("API error")
		},
	}

	handler := NewHandler(mockMeta, taskStore, sessionStore, "workspace-1", "/project", nil, nil)

	ctx := context.Background()

	// セッションを作成
	session, err := handler.CreateSession(ctx)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// メッセージを送信（エラーが返る）
	resp, err := handler.HandleMessage(ctx, session.ID, "テストメッセージ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// エラー時もレスポンスが返ることを確認
	if resp == nil {
		t.Fatal("response should not be nil even on error")
	}

	// エラーメッセージがアシスタントメッセージとして保存されている
	if resp.Message.Role != "assistant" {
		t.Errorf("expected role 'assistant', got '%s'", resp.Message.Role)
	}
}

func TestHandler_HandleMessage_WithExistingTasks(t *testing.T) {
	tmpDir := t.TempDir()
	taskStore := orchestrator.NewTaskStore(tmpDir)
	sessionStore := NewChatSessionStore(tmpDir)

	// 既存タスクを作成
	existingTask := &orchestrator.Task{
		ID:        "existing-task-1",
		Title:     "既存タスク",
		Status:    orchestrator.TaskStatusPending,
		PoolID:    "default",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := taskStore.SaveTask(existingTask); err != nil {
		t.Fatalf("SaveTask failed: %v", err)
	}

	var capturedReq *meta.DecomposeRequest
	mockMeta := &MockMetaClient{
		DecomposeFunc: func(ctx context.Context, req *meta.DecomposeRequest) (*meta.DecomposeResponse, error) {
			capturedReq = req
			return &meta.DecomposeResponse{
				Understanding: "理解しました",
				Phases:        []meta.DecomposedPhase{},
			}, nil
		},
	}

	handler := NewHandler(mockMeta, taskStore, sessionStore, "workspace-1", "/project", nil, nil)

	ctx := context.Background()

	// セッションを作成
	session, err := handler.CreateSession(ctx)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// メッセージを送信
	_, err = handler.HandleMessage(ctx, session.ID, "テスト")
	if err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	// リクエストに既存タスクが含まれていることを確認
	if capturedReq == nil {
		t.Fatal("request was not captured")
	}
	if len(capturedReq.Context.ExistingTasks) != 1 {
		t.Errorf("expected 1 existing task, got %d", len(capturedReq.Context.ExistingTasks))
	}
	if capturedReq.Context.ExistingTasks[0].ID != "existing-task-1" {
		t.Errorf("expected existing task ID 'existing-task-1', got '%s'", capturedReq.Context.ExistingTasks[0].ID)
	}
}

func TestHandler_HandleMessage_WithConversationHistory(t *testing.T) {
	tmpDir := t.TempDir()
	taskStore := orchestrator.NewTaskStore(tmpDir)
	sessionStore := NewChatSessionStore(tmpDir)

	var capturedReq *meta.DecomposeRequest
	callCount := 0
	mockMeta := &MockMetaClient{
		DecomposeFunc: func(ctx context.Context, req *meta.DecomposeRequest) (*meta.DecomposeResponse, error) {
			capturedReq = req
			callCount++
			return &meta.DecomposeResponse{
				Understanding: "理解しました",
				Phases:        []meta.DecomposedPhase{},
			}, nil
		},
	}

	handler := NewHandler(mockMeta, taskStore, sessionStore, "workspace-1", "/project", nil, nil)

	ctx := context.Background()

	// セッションを作成
	session, err := handler.CreateSession(ctx)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// 最初のメッセージ
	_, err = handler.HandleMessage(ctx, session.ID, "最初のメッセージ")
	if err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	// 2番目のメッセージ
	_, err = handler.HandleMessage(ctx, session.ID, "2番目のメッセージ")
	if err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	// リクエストに会話履歴が含まれていることを確認
	if capturedReq == nil {
		t.Fatal("request was not captured")
	}

	// システム + ユーザー1 + アシスタント1 + ユーザー2 = 4件の履歴
	// ただし2番目のメッセージ送信時点では ユーザー2 はまだ保存されていないので3件
	// 最新10件を取得しているので全て含まれる
	if len(capturedReq.Context.ConversationHistory) < 2 {
		t.Errorf("expected at least 2 conversation history messages, got %d", len(capturedReq.Context.ConversationHistory))
	}
}

func TestHandler_HandleMessage_PotentialConflicts(t *testing.T) {
	tmpDir := t.TempDir()
	taskStore := orchestrator.NewTaskStore(tmpDir)
	sessionStore := NewChatSessionStore(tmpDir)
	projectRoot := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(filepath.Join(projectRoot, "src", "auth"), 0o755); err != nil {
		t.Fatalf("failed to create project root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, "src", "auth", "login.ts"), []byte(""), 0o644); err != nil {
		t.Fatalf("failed to create conflict file: %v", err)
	}

	mockMeta := &MockMetaClient{
		DecomposeFunc: func(ctx context.Context, req *meta.DecomposeRequest) (*meta.DecomposeResponse, error) {
			return &meta.DecomposeResponse{
				Understanding: "理解しました",
				Phases: []meta.DecomposedPhase{
					{
						Name: "実装",
						Tasks: []meta.DecomposedTask{
							{
								ID:          "temp-task-1",
								Title:       "タスク1",
								Description: "説明",
							},
						},
					},
				},
				PotentialConflicts: []meta.PotentialConflict{
					{
						File:    "src/auth/login.ts",
						Warning: "既存ファイルを変更",
						Tasks:   []string{"temp-task-1"},
					},
				},
			}, nil
		},
	}

	handler := NewHandler(mockMeta, taskStore, sessionStore, "workspace-1", projectRoot, nil, nil)

	ctx := context.Background()

	// セッションを作成
	session, err := handler.CreateSession(ctx)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// メッセージを送信
	resp, err := handler.HandleMessage(ctx, session.ID, "テスト")
	if err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	// コンフリクト情報が返されていることを確認
	if len(resp.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(resp.Conflicts))
	}
	if resp.Conflicts[0].File != "src/auth/login.ts" {
		t.Errorf("unexpected conflict file: %s", resp.Conflicts[0].File)
	}
}

func TestHandler_GetHistory(t *testing.T) {
	tmpDir := t.TempDir()
	taskStore := orchestrator.NewTaskStore(tmpDir)
	sessionStore := NewChatSessionStore(tmpDir)
	mockMeta := &MockMetaClient{
		DecomposeFunc: func(ctx context.Context, req *meta.DecomposeRequest) (*meta.DecomposeResponse, error) {
			return &meta.DecomposeResponse{
				Understanding: "理解しました",
				Phases:        []meta.DecomposedPhase{},
			}, nil
		},
	}

	handler := NewHandler(mockMeta, taskStore, sessionStore, "workspace-1", "/project", nil, nil)

	ctx := context.Background()

	// セッションを作成
	session, err := handler.CreateSession(ctx)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// メッセージを送信
	_, err = handler.HandleMessage(ctx, session.ID, "テスト")
	if err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	// 履歴を取得
	history, err := handler.GetHistory(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	// システム + ユーザー + アシスタント = 3件
	if len(history) != 3 {
		t.Errorf("expected 3 messages, got %d", len(history))
	}
}

func TestHandler_buildResponseContent(t *testing.T) {
	handler := &Handler{}

	resp := &meta.DecomposeResponse{
		Understanding: "テストの理解",
		Phases: []meta.DecomposedPhase{
			{Name: "概念設計"},
			{Name: "実装"},
		},
		PotentialConflicts: []meta.PotentialConflict{
			{File: "test.ts", Warning: "警告"},
		},
	}

	tasks := []orchestrator.Task{
		{Title: "タスク1", Description: "説明1", PhaseName: "概念設計"},
		{Title: "タスク2", Description: "説明2", PhaseName: "実装"},
	}

	content := handler.buildResponseContent(resp, tasks)

	// 内容の検証
	if content == "" {
		t.Error("content should not be empty")
	}

	// 理解が含まれている
	if !containsString(content, "テストの理解") {
		t.Error("content should contain understanding")
	}

	// タスク数が含まれている
	if !containsString(content, "2 個のタスク") {
		t.Error("content should contain task count")
	}

	// コンフリクト警告が含まれている
	if !containsString(content, "test.ts") {
		t.Error("content should contain conflict file")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestHandler_HandleMessage_WithSuggestedImpl(t *testing.T) {
	tmpDir := t.TempDir()
	taskStore := orchestrator.NewTaskStore(tmpDir)
	sessionStore := NewChatSessionStore(tmpDir)

	mockMeta := &MockMetaClient{
		DecomposeFunc: func(ctx context.Context, req *meta.DecomposeRequest) (*meta.DecomposeResponse, error) {
			return &meta.DecomposeResponse{
				Understanding: "AI understands impl request",
				Phases: []meta.DecomposedPhase{
					{
						Name: "実装",
						Tasks: []meta.DecomposedTask{
							{
								ID:          "temp-1",
								Title:       "Impl Task",
								Description: "Do code",
								SuggestedImpl: &meta.SuggestedImpl{
									Language:    "go",
									FilePaths:   []string{"app/main.go"},
									Constraints: []string{"use context"},
								},
							},
						},
					},
				},
			}, nil
		},
	}

	handler := NewHandler(mockMeta, taskStore, sessionStore, "workspace-1", "/project", nil, nil)
	ctx := context.Background()
	session, _ := handler.CreateSession(ctx)

	resp, err := handler.HandleMessage(ctx, session.ID, "implement this")
	if err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	if len(resp.GeneratedTasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(resp.GeneratedTasks))
	}

	task := resp.GeneratedTasks[0]
	if task.SuggestedImpl == nil {
		t.Fatal("expected SuggestedImpl to be mapped")
	}
	if task.SuggestedImpl.Language != "go" {
		t.Errorf("expected go, got %s", task.SuggestedImpl.Language)
	}
	if len(task.SuggestedImpl.FilePaths) != 1 || task.SuggestedImpl.FilePaths[0] != "app/main.go (New File)" {
		t.Errorf("unexpected file paths: %v", task.SuggestedImpl.FilePaths)
	}
}
