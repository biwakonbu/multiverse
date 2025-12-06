package orchestrator

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBacklogStore_Add(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewBacklogStore(tmpDir)

	item := &BacklogItem{
		TaskID:      "task-1",
		Type:        BacklogTypeFailure,
		Title:       "Test Failure",
		Description: "Something went wrong",
		Priority:    3,
	}

	err := store.Add(item)
	require.NoError(t, err)

	// ID と CreatedAt が自動設定されること
	assert.NotEmpty(t, item.ID)
	assert.False(t, item.CreatedAt.IsZero())

	// ファイルが作成されていること
	path := filepath.Join(tmpDir, "backlog", item.ID+".json")
	_, err = os.Stat(path)
	assert.NoError(t, err)
}

func TestBacklogStore_Get(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewBacklogStore(tmpDir)

	// アイテムを追加
	item := &BacklogItem{
		TaskID:      "task-1",
		Type:        BacklogTypeQuestion,
		Title:       "Need clarification",
		Description: "What should we do?",
		Priority:    5,
	}
	err := store.Add(item)
	require.NoError(t, err)

	// 取得
	retrieved, err := store.Get(item.ID)
	require.NoError(t, err)

	assert.Equal(t, item.ID, retrieved.ID)
	assert.Equal(t, item.TaskID, retrieved.TaskID)
	assert.Equal(t, item.Type, retrieved.Type)
	assert.Equal(t, item.Title, retrieved.Title)
	assert.Equal(t, item.Priority, retrieved.Priority)
}

func TestBacklogStore_Get_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewBacklogStore(tmpDir)

	_, err := store.Get("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBacklogStore_List(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewBacklogStore(tmpDir)

	// 複数アイテムを追加
	items := []*BacklogItem{
		{TaskID: "task-1", Type: BacklogTypeFailure, Title: "Failure 1", Priority: 2},
		{TaskID: "task-2", Type: BacklogTypeQuestion, Title: "Question 1", Priority: 5},
		{TaskID: "task-3", Type: BacklogTypeBlocker, Title: "Blocker 1", Priority: 3},
	}

	for _, item := range items {
		err := store.Add(item)
		require.NoError(t, err)
	}

	// リスト取得
	list, err := store.List()
	require.NoError(t, err)
	assert.Len(t, list, 3)

	// 優先度でソートされていること
	assert.Equal(t, 5, list[0].Priority)
	assert.Equal(t, 3, list[1].Priority)
	assert.Equal(t, 2, list[2].Priority)
}

func TestBacklogStore_List_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewBacklogStore(tmpDir)

	list, err := store.List()
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestBacklogStore_ListUnresolved(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewBacklogStore(tmpDir)

	// 未解決アイテムを追加
	item1 := &BacklogItem{TaskID: "task-1", Type: BacklogTypeFailure, Title: "Failure 1", Priority: 3}
	err := store.Add(item1)
	require.NoError(t, err)

	// 解決済みアイテムを追加
	item2 := &BacklogItem{TaskID: "task-2", Type: BacklogTypeQuestion, Title: "Question 1", Priority: 3}
	now := time.Now()
	item2.ResolvedAt = &now
	item2.Resolution = "Resolved"
	err = store.Add(item2)
	require.NoError(t, err)

	// 未解決のみ取得
	unresolved, err := store.ListUnresolved()
	require.NoError(t, err)
	assert.Len(t, unresolved, 1)
	assert.Equal(t, item1.ID, unresolved[0].ID)
}

func TestBacklogStore_Resolve(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewBacklogStore(tmpDir)

	// アイテムを追加
	item := &BacklogItem{
		TaskID:   "task-1",
		Type:     BacklogTypeFailure,
		Title:    "Failure 1",
		Priority: 3,
	}
	err := store.Add(item)
	require.NoError(t, err)

	// 解決
	err = store.Resolve(item.ID, "Fixed the issue")
	require.NoError(t, err)

	// 確認
	resolved, err := store.Get(item.ID)
	require.NoError(t, err)
	assert.NotNil(t, resolved.ResolvedAt)
	assert.Equal(t, "Fixed the issue", resolved.Resolution)
}

func TestBacklogStore_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewBacklogStore(tmpDir)

	// アイテムを追加
	item := &BacklogItem{
		TaskID:   "task-1",
		Type:     BacklogTypeFailure,
		Title:    "Failure 1",
		Priority: 3,
	}
	err := store.Add(item)
	require.NoError(t, err)

	// 削除
	err = store.Delete(item.ID)
	require.NoError(t, err)

	// 取得できないこと
	_, err = store.Get(item.ID)
	assert.Error(t, err)
}

func TestBacklogStore_Delete_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewBacklogStore(tmpDir)

	// 存在しないアイテムの削除は成功
	err := store.Delete("non-existent")
	assert.NoError(t, err)
}

func TestCreateFailureItem(t *testing.T) {
	err := errors.New("connection timeout")
	item := CreateFailureItem("task-123", "Deploy to Production", err, 3)

	assert.Equal(t, "task-123", item.TaskID)
	assert.Equal(t, BacklogTypeFailure, item.Type)
	assert.Contains(t, item.Title, "Deploy to Production")
	assert.Contains(t, item.Description, "3 回の試行")
	assert.Equal(t, 4, item.Priority)
	assert.Equal(t, "connection timeout", item.Metadata["error"])
	assert.Equal(t, 3, item.Metadata["attemptCount"])
}
