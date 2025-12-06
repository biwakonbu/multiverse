package chat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ChatSession はチャットセッションを表す
type ChatSession struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspaceId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ChatMessage はチャットメッセージを表す
type ChatMessage struct {
	ID             string    `json:"id"`
	SessionID      string    `json:"sessionId"`
	Role           string    `json:"role"` // user | assistant | system
	Content        string    `json:"content"`
	Timestamp      time.Time `json:"timestamp"`
	GeneratedTasks []string  `json:"generatedTasks,omitempty"` // このメッセージで生成されたタスクID
}

// ChatSessionStore はチャットセッションの永続化を管理する
type ChatSessionStore struct {
	WorkspaceDir string
}

// NewChatSessionStore は新しい ChatSessionStore を作成する
func NewChatSessionStore(workspaceDir string) *ChatSessionStore {
	return &ChatSessionStore{WorkspaceDir: workspaceDir}
}

// GetChatDir はチャットディレクトリのパスを返す
func (s *ChatSessionStore) GetChatDir() string {
	return filepath.Join(s.WorkspaceDir, "chat")
}

// ensureSafeSessionID validates the sessionID to avoid path traversal.
func ensureSafeSessionID(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("sessionID is empty")
	}
	if filepath.IsAbs(sessionID) {
		return fmt.Errorf("absolute sessionID is not allowed")
	}
	if strings.Contains(sessionID, "..") || strings.ContainsAny(sessionID, `/\`) {
		return fmt.Errorf("sessionID contains invalid path characters")
	}
	return nil
}

// getSessionFilePath はセッションファイルのパスを返す
func (s *ChatSessionStore) getSessionFilePath(sessionID string) string {
	return filepath.Join(s.GetChatDir(), sessionID+".jsonl")
}

// getSessionMetaFilePath はセッションメタデータファイルのパスを返す
func (s *ChatSessionStore) getSessionMetaFilePath(sessionID string) string {
	return filepath.Join(s.GetChatDir(), sessionID+".meta.json")
}

// CreateSession は新しいチャットセッションを作成する
func (s *ChatSessionStore) CreateSession(id, workspaceID string) (*ChatSession, error) {
	if err := ensureSafeSessionID(id); err != nil {
		return nil, err
	}

	dir := s.GetChatDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create chat directory: %w", err)
	}

	now := time.Now()
	session := &ChatSession{
		ID:          id,
		WorkspaceID: workspaceID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.saveSessionMeta(session); err != nil {
		return nil, err
	}

	return session, nil
}

// saveSessionMeta はセッションメタデータを保存する
func (s *ChatSessionStore) saveSessionMeta(session *ChatSession) error {
	path := s.getSessionMetaFilePath(session.ID)
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write session meta: %w", err)
	}

	return nil
}

// LoadSession はセッションを読み込む
func (s *ChatSessionStore) LoadSession(sessionID string) (*ChatSession, error) {
	path := s.getSessionMetaFilePath(sessionID)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var session ChatSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// AppendMessage はメッセージをセッションに追加する
func (s *ChatSessionStore) AppendMessage(msg *ChatMessage) error {
	if err := ensureSafeSessionID(msg.SessionID); err != nil {
		return err
	}

	dir := s.GetChatDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create chat directory: %w", err)
	}

	path := s.getSessionFilePath(msg.SessionID)
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open session file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	// セッションメタデータの UpdatedAt を更新
	session, err := s.LoadSession(msg.SessionID)
	if err == nil {
		session.UpdatedAt = time.Now()
		_ = s.saveSessionMeta(session)
	}

	return nil
}

// LoadMessages はセッションの全メッセージを読み込む
func (s *ChatSessionStore) LoadMessages(sessionID string) ([]ChatMessage, error) {
	if err := ensureSafeSessionID(sessionID); err != nil {
		return nil, err
	}

	path := s.getSessionFilePath(sessionID)
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return []ChatMessage{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var messages []ChatMessage
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var msg ChatMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue // 破損した行はスキップ
		}
		messages = append(messages, msg)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

// ListSessions はワークスペース内の全セッションを一覧する
func (s *ChatSessionStore) ListSessions() ([]ChatSession, error) {
	dir := s.GetChatDir()
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return []ChatSession{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read chat directory: %w", err)
	}

	var sessions []ChatSession
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		// .meta.json ファイルのみ処理
		if len(entry.Name()) < 10 || entry.Name()[len(entry.Name())-10:] != ".meta.json" {
			continue
		}

		sessionID := entry.Name()[:len(entry.Name())-10]
		session, err := s.LoadSession(sessionID)
		if err != nil {
			continue
		}
		sessions = append(sessions, *session)
	}

	return sessions, nil
}

// GetRecentMessages は最新の N 件のメッセージを返す
func (s *ChatSessionStore) GetRecentMessages(sessionID string, limit int) ([]ChatMessage, error) {
	messages, err := s.LoadMessages(sessionID)
	if err != nil {
		return nil, err
	}

	if len(messages) <= limit {
		return messages, nil
	}

	return messages[len(messages)-limit:], nil
}
