package chat

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestChatSessionStore_CreateSession(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewChatSessionStore(tmpDir)

	session, err := store.CreateSession("test-session-1", "workspace-1")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if session.ID != "test-session-1" {
		t.Errorf("expected session ID 'test-session-1', got '%s'", session.ID)
	}
	if session.WorkspaceID != "workspace-1" {
		t.Errorf("expected workspace ID 'workspace-1', got '%s'", session.WorkspaceID)
	}

	// メタファイルが作成されていることを確認
	metaPath := filepath.Join(tmpDir, "chat", "test-session-1.meta.json")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Error("session meta file was not created")
	}
}

func TestChatSessionStore_LoadSession(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewChatSessionStore(tmpDir)

	// セッションを作成
	_, err := store.CreateSession("test-session-2", "workspace-2")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// セッションを読み込み
	loaded, err := store.LoadSession("test-session-2")
	if err != nil {
		t.Fatalf("LoadSession failed: %v", err)
	}

	if loaded.ID != "test-session-2" {
		t.Errorf("expected session ID 'test-session-2', got '%s'", loaded.ID)
	}
	if loaded.WorkspaceID != "workspace-2" {
		t.Errorf("expected workspace ID 'workspace-2', got '%s'", loaded.WorkspaceID)
	}
}

func TestChatSessionStore_LoadSession_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewChatSessionStore(tmpDir)

	_, err := store.LoadSession("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session, got nil")
	}
}

func TestChatSessionStore_AppendMessage(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewChatSessionStore(tmpDir)

	// セッションを作成
	_, err := store.CreateSession("test-session-3", "workspace-3")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// メッセージを追加
	msg := &ChatMessage{
		ID:        "msg-1",
		SessionID: "test-session-3",
		Role:      "user",
		Content:   "Hello, world!",
		Timestamp: time.Now(),
	}
	if err := store.AppendMessage(msg); err != nil {
		t.Fatalf("AppendMessage failed: %v", err)
	}

	// メッセージファイルが作成されていることを確認
	msgPath := filepath.Join(tmpDir, "chat", "test-session-3.jsonl")
	if _, err := os.Stat(msgPath); os.IsNotExist(err) {
		t.Error("message file was not created")
	}
}

func TestChatSessionStore_LoadMessages(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewChatSessionStore(tmpDir)

	// セッションを作成
	_, err := store.CreateSession("test-session-4", "workspace-4")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// 複数のメッセージを追加
	messages := []*ChatMessage{
		{
			ID:        "msg-1",
			SessionID: "test-session-4",
			Role:      "user",
			Content:   "First message",
			Timestamp: time.Now(),
		},
		{
			ID:        "msg-2",
			SessionID: "test-session-4",
			Role:      "assistant",
			Content:   "Second message",
			Timestamp: time.Now(),
		},
		{
			ID:        "msg-3",
			SessionID: "test-session-4",
			Role:      "user",
			Content:   "Third message",
			Timestamp: time.Now(),
		},
	}

	for _, msg := range messages {
		if err := store.AppendMessage(msg); err != nil {
			t.Fatalf("AppendMessage failed: %v", err)
		}
	}

	// メッセージを読み込み
	loaded, err := store.LoadMessages("test-session-4")
	if err != nil {
		t.Fatalf("LoadMessages failed: %v", err)
	}

	if len(loaded) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(loaded))
	}

	// 順序を確認
	if loaded[0].Content != "First message" {
		t.Errorf("expected first message 'First message', got '%s'", loaded[0].Content)
	}
	if loaded[1].Content != "Second message" {
		t.Errorf("expected second message 'Second message', got '%s'", loaded[1].Content)
	}
	if loaded[2].Content != "Third message" {
		t.Errorf("expected third message 'Third message', got '%s'", loaded[2].Content)
	}
}

func TestChatSessionStore_LoadMessages_EmptySession(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewChatSessionStore(tmpDir)

	// メッセージがないセッションを読み込み
	messages, err := store.LoadMessages("nonexistent-session")
	if err != nil {
		t.Fatalf("LoadMessages failed: %v", err)
	}

	if len(messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(messages))
	}
}

func TestChatSessionStore_GetRecentMessages(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewChatSessionStore(tmpDir)

	// セッションを作成
	_, err := store.CreateSession("test-session-5", "workspace-5")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// 5つのメッセージを追加
	for i := 1; i <= 5; i++ {
		msg := &ChatMessage{
			ID:        "msg-" + string(rune('0'+i)),
			SessionID: "test-session-5",
			Role:      "user",
			Content:   "Message " + string(rune('0'+i)),
			Timestamp: time.Now(),
		}
		if err := store.AppendMessage(msg); err != nil {
			t.Fatalf("AppendMessage failed: %v", err)
		}
	}

	// 最新3件を取得
	recent, err := store.GetRecentMessages("test-session-5", 3)
	if err != nil {
		t.Fatalf("GetRecentMessages failed: %v", err)
	}

	if len(recent) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(recent))
	}

	// 最新3件であることを確認（3, 4, 5）
	if recent[0].Content != "Message 3" {
		t.Errorf("expected 'Message 3', got '%s'", recent[0].Content)
	}
}

func TestChatSessionStore_GetRecentMessages_LimitGreaterThanTotal(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewChatSessionStore(tmpDir)

	// セッションを作成
	_, err := store.CreateSession("test-session-6", "workspace-6")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// 2つのメッセージを追加
	for i := 1; i <= 2; i++ {
		msg := &ChatMessage{
			ID:        "msg-" + string(rune('0'+i)),
			SessionID: "test-session-6",
			Role:      "user",
			Content:   "Message " + string(rune('0'+i)),
			Timestamp: time.Now(),
		}
		if err := store.AppendMessage(msg); err != nil {
			t.Fatalf("AppendMessage failed: %v", err)
		}
	}

	// limit > total の場合、全件を返す
	recent, err := store.GetRecentMessages("test-session-6", 10)
	if err != nil {
		t.Fatalf("GetRecentMessages failed: %v", err)
	}

	if len(recent) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(recent))
	}
}

func TestChatSessionStore_ListSessions(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewChatSessionStore(tmpDir)

	// 複数のセッションを作成
	for i := 1; i <= 3; i++ {
		_, err := store.CreateSession("session-"+string(rune('0'+i)), "workspace-1")
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}
	}

	// セッション一覧を取得
	sessions, err := store.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}

	if len(sessions) != 3 {
		t.Errorf("expected 3 sessions, got %d", len(sessions))
	}
}

func TestChatSessionStore_ListSessions_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewChatSessionStore(tmpDir)

	// セッションがない場合
	sessions, err := store.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}

	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestChatSessionStore_AppendMessage_WithGeneratedTasks(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewChatSessionStore(tmpDir)

	// セッションを作成
	_, err := store.CreateSession("test-session-7", "workspace-7")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// GeneratedTasks 付きのメッセージを追加
	msg := &ChatMessage{
		ID:             "msg-1",
		SessionID:      "test-session-7",
		Role:           "assistant",
		Content:        "Tasks created",
		Timestamp:      time.Now(),
		GeneratedTasks: []string{"task-1", "task-2", "task-3"},
	}
	if err := store.AppendMessage(msg); err != nil {
		t.Fatalf("AppendMessage failed: %v", err)
	}

	// メッセージを読み込み
	loaded, err := store.LoadMessages("test-session-7")
	if err != nil {
		t.Fatalf("LoadMessages failed: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("expected 1 message, got %d", len(loaded))
	}

	if len(loaded[0].GeneratedTasks) != 3 {
		t.Errorf("expected 3 generated tasks, got %d", len(loaded[0].GeneratedTasks))
	}
}

func TestChatSessionStore_PathTraversalIsRejected(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewChatSessionStore(tmpDir)

	badID := "../evil"

	if _, err := store.CreateSession(badID, "ws"); err == nil {
		t.Fatalf("expected error for invalid session id, got nil")
	}

	msg := &ChatMessage{
		ID:        "m1",
		SessionID: badID,
		Role:      "user",
		Content:   "hi",
		Timestamp: time.Now(),
	}
	if err := store.AppendMessage(msg); err == nil {
		t.Fatalf("expected error for append with invalid session id, got nil")
	}

	if _, err := store.LoadMessages(badID); err == nil {
		t.Fatalf("expected error for load with invalid session id, got nil")
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "chat", "evil.jsonl")); err == nil {
		t.Fatalf("expected no file to be created for invalid session id")
	}
}
