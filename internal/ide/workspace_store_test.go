package ide

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWorkspaceStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "workspace_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewWorkspaceStore(tmpDir)
	projectRoot := "/tmp/my-project"
	id := store.GetWorkspaceID(projectRoot)

	ws := &Workspace{
		Version:      "1.0",
		ProjectRoot:  projectRoot,
		DisplayName:  "My Project",
		CreatedAt:    time.Now(),
		LastOpenedAt: time.Now(),
	}

	if err := store.SaveWorkspace(ws); err != nil {
		t.Fatalf("SaveWorkspace failed: %v", err)
	}

	loadedWs, err := store.LoadWorkspace(id)
	if err != nil {
		t.Fatalf("LoadWorkspace failed: %v", err)
	}

	if loadedWs.ProjectRoot != ws.ProjectRoot {
		t.Errorf("expected ProjectRoot %s, got %s", ws.ProjectRoot, loadedWs.ProjectRoot)
	}
	if loadedWs.DisplayName != ws.DisplayName {
		t.Errorf("expected DisplayName %s, got %s", ws.DisplayName, loadedWs.DisplayName)
	}

	// Verify directory structure
	expectedDir := filepath.Join(tmpDir, id)
	if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
		t.Errorf("workspace directory not created: %s", expectedDir)
	}
}

func TestListWorkspaces(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "workspace_list_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewWorkspaceStore(tmpDir)

	// 空の一覧
	summaries, err := store.ListWorkspaces()
	if err != nil {
		t.Fatalf("ListWorkspaces failed: %v", err)
	}
	if len(summaries) != 0 {
		t.Errorf("expected 0 workspaces, got %d", len(summaries))
	}

	// 複数のワークスペースを作成
	now := time.Now()
	ws1 := &Workspace{
		Version:      "1.0",
		ProjectRoot:  "/tmp/project-a",
		DisplayName:  "Project A",
		CreatedAt:    now,
		LastOpenedAt: now.Add(-time.Hour), // 1時間前
	}
	ws2 := &Workspace{
		Version:      "1.0",
		ProjectRoot:  "/tmp/project-b",
		DisplayName:  "Project B",
		CreatedAt:    now,
		LastOpenedAt: now, // 現在
	}

	if err := store.SaveWorkspace(ws1); err != nil {
		t.Fatalf("SaveWorkspace ws1 failed: %v", err)
	}
	if err := store.SaveWorkspace(ws2); err != nil {
		t.Fatalf("SaveWorkspace ws2 failed: %v", err)
	}

	// 一覧を取得（lastOpenedAt でソート、新しい順）
	summaries, err = store.ListWorkspaces()
	if err != nil {
		t.Fatalf("ListWorkspaces failed: %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("expected 2 workspaces, got %d", len(summaries))
	}

	// 最新のものが先頭
	if summaries[0].DisplayName != "Project B" {
		t.Errorf("expected first workspace to be Project B, got %s", summaries[0].DisplayName)
	}
	if summaries[1].DisplayName != "Project A" {
		t.Errorf("expected second workspace to be Project A, got %s", summaries[1].DisplayName)
	}
}

func TestRemoveWorkspace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "workspace_remove_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewWorkspaceStore(tmpDir)

	// ワークスペースを作成
	ws := &Workspace{
		Version:      "1.0",
		ProjectRoot:  "/tmp/project-remove",
		DisplayName:  "Remove Me",
		CreatedAt:    time.Now(),
		LastOpenedAt: time.Now(),
	}
	if err := store.SaveWorkspace(ws); err != nil {
		t.Fatalf("SaveWorkspace failed: %v", err)
	}

	id := store.GetWorkspaceID(ws.ProjectRoot)

	// 削除前に存在確認
	if _, err := store.LoadWorkspace(id); err != nil {
		t.Fatalf("workspace should exist before removal: %v", err)
	}

	// 削除
	if err := store.RemoveWorkspace(id); err != nil {
		t.Fatalf("RemoveWorkspace failed: %v", err)
	}

	// 削除後に存在しないことを確認
	if _, err := store.LoadWorkspace(id); err == nil {
		t.Error("workspace should not exist after removal")
	}

	// 存在しないワークスペースの削除はエラー
	if err := store.RemoveWorkspace("nonexistent"); err == nil {
		t.Error("RemoveWorkspace should fail for nonexistent workspace")
	}
}
