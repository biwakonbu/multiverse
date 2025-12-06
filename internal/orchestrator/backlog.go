package orchestrator

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/biwakonbu/agent-runner/internal/logging"
	"github.com/google/uuid"
)

// BacklogType はバックログアイテムの種類を表す
type BacklogType string

const (
	BacklogTypeFailure  BacklogType = "FAILURE"  // タスク失敗
	BacklogTypeQuestion BacklogType = "QUESTION" // Meta-agent からの質問
	BacklogTypeBlocker  BacklogType = "BLOCKER"  // 外部ブロッカー
)

// BacklogItem はバックログアイテムを表す
type BacklogItem struct {
	ID          string         `json:"id"`
	TaskID      string         `json:"taskId"`
	Type        BacklogType    `json:"type"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Priority    int            `json:"priority"` // 1-5（5が最高）
	CreatedAt   time.Time      `json:"createdAt"`
	ResolvedAt  *time.Time     `json:"resolvedAt,omitempty"`
	Resolution  string         `json:"resolution,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"` // エラー詳細等
}

// BacklogStore はバックログアイテムを永続化する
type BacklogStore struct {
	workspaceDir string
	logger       *slog.Logger
}

// NewBacklogStore は BacklogStore を作成する
func NewBacklogStore(workspaceDir string) *BacklogStore {
	return &BacklogStore{
		workspaceDir: workspaceDir,
		logger:       logging.WithComponent(slog.Default(), "backlog-store"),
	}
}

// backlogDir はバックログディレクトリのパスを返す
func (s *BacklogStore) backlogDir() string {
	return filepath.Join(s.workspaceDir, "backlog")
}

// itemPath は特定アイテムのファイルパスを返す
func (s *BacklogStore) itemPath(id string) string {
	return filepath.Join(s.backlogDir(), id+".json")
}

// ensureDir はディレクトリを作成する
func (s *BacklogStore) ensureDir() error {
	return os.MkdirAll(s.backlogDir(), 0755)
}

// Add はバックログアイテムを追加する
func (s *BacklogStore) Add(item *BacklogItem) error {
	if err := s.ensureDir(); err != nil {
		return fmt.Errorf("failed to create backlog dir: %w", err)
	}

	// ID が未設定なら生成
	if item.ID == "" {
		item.ID = uuid.New().String()
	}

	// CreatedAt が未設定なら現在時刻
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}

	data, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal backlog item: %w", err)
	}

	path := s.itemPath(item.ID)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write backlog item: %w", err)
	}

	s.logger.Info("backlog item added",
		slog.String("id", item.ID),
		slog.String("task_id", item.TaskID),
		slog.String("type", string(item.Type)),
	)

	return nil
}

// Get はバックログアイテムを取得する
func (s *BacklogStore) Get(id string) (*BacklogItem, error) {
	path := s.itemPath(id)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("backlog item not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read backlog item: %w", err)
	}

	var item BacklogItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal backlog item: %w", err)
	}

	return &item, nil
}

// List は全バックログアイテムを取得する
func (s *BacklogStore) List() ([]BacklogItem, error) {
	dir := s.backlogDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BacklogItem{}, nil
		}
		return nil, fmt.Errorf("failed to read backlog dir: %w", err)
	}

	var items []BacklogItem
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		id := entry.Name()[:len(entry.Name())-5] // .json を除去
		item, err := s.Get(id)
		if err != nil {
			s.logger.Warn("failed to load backlog item", slog.String("id", id), slog.Any("error", err))
			continue
		}
		items = append(items, *item)
	}

	// 優先度（降順）→ 作成日時（昇順）でソート
	sort.Slice(items, func(i, j int) bool {
		if items[i].Priority != items[j].Priority {
			return items[i].Priority > items[j].Priority
		}
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})

	return items, nil
}

// ListUnresolved は未解決のバックログアイテムを取得する
func (s *BacklogStore) ListUnresolved() ([]BacklogItem, error) {
	all, err := s.List()
	if err != nil {
		return nil, err
	}

	var unresolved []BacklogItem
	for _, item := range all {
		if item.ResolvedAt == nil {
			unresolved = append(unresolved, item)
		}
	}

	return unresolved, nil
}

// Resolve はバックログアイテムを解決済みにする
func (s *BacklogStore) Resolve(id string, resolution string) error {
	item, err := s.Get(id)
	if err != nil {
		return err
	}

	now := time.Now()
	item.ResolvedAt = &now
	item.Resolution = resolution

	data, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal backlog item: %w", err)
	}

	path := s.itemPath(id)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write backlog item: %w", err)
	}

	s.logger.Info("backlog item resolved",
		slog.String("id", id),
		slog.String("resolution", resolution),
	)

	return nil
}

// Delete はバックログアイテムを削除する
func (s *BacklogStore) Delete(id string) error {
	path := s.itemPath(id)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil // 既に存在しない場合は成功とみなす
		}
		return fmt.Errorf("failed to delete backlog item: %w", err)
	}

	s.logger.Info("backlog item deleted", slog.String("id", id))
	return nil
}

// CreateFailureItem はタスク失敗からバックログアイテムを作成する
func CreateFailureItem(taskID string, taskTitle string, err error, attemptCount int) *BacklogItem {
	return &BacklogItem{
		TaskID:      taskID,
		Type:        BacklogTypeFailure,
		Title:       fmt.Sprintf("タスク失敗: %s", taskTitle),
		Description: fmt.Sprintf("タスク '%s' が %d 回の試行後に失敗しました。", taskTitle, attemptCount),
		Priority:    4, // 高優先度
		Metadata: map[string]any{
			"error":        err.Error(),
			"attemptCount": attemptCount,
		},
	}
}
