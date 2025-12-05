# IDE Package - Workspace メタデータ管理

このパッケージは multiverse IDE の Workspace 管理機能を提供します。

## 責務

- **Workspace ID 生成**: プロジェクトルートパスから決定的な ID を生成
- **Workspace 永続化**: workspace.json の読み書き
- **ディレクトリ管理**: Workspace ディレクトリ構造の管理

## ファイル構成

| ファイル | 役割 |
|---------|------|
| workspace_store.go | Workspace メタデータの永続化 |
| workspace_store_test.go | WorkspaceStore のユニットテスト |

## 主要データモデル

### Workspace 構造体

```go
type Workspace struct {
    Version      string    `json:"version"`      // スキーマバージョン（"1.0"）
    ProjectRoot  string    `json:"projectRoot"`  // プロジェクトルートの絶対パス
    DisplayName  string    `json:"displayName"`  // UI 表示用の名前
    CreatedAt    time.Time `json:"createdAt"`    // 作成日時
    LastOpenedAt time.Time `json:"lastOpenedAt"` // 最終アクセス日時
}
```

## WorkspaceStore API

```go
// 初期化（baseDir = $HOME/.multiverse/workspaces）
store := NewWorkspaceStore(baseDir)

// ID 生成
id := store.GetWorkspaceID(projectRoot)  // sha1(projectRoot)[:12]

// ディレクトリパス取得
dir := store.GetWorkspaceDir(id)         // <baseDir>/<id>

// Workspace 操作
ws, err := store.LoadWorkspace(id)       // workspace.json を読み込み
err := store.SaveWorkspace(ws)           // workspace.json を保存（ディレクトリ自動作成）
```

## Workspace ID 生成ロジック

```go
// sha1(projectRoot)[:12] で決定的な ID を生成
func (s *WorkspaceStore) GetWorkspaceID(projectRoot string) string {
    h := sha1.New()
    h.Write([]byte(projectRoot))
    sum := h.Sum(nil)
    return hex.EncodeToString(sum)[:12]
}
```

**特徴**:
- **決定的**: 同じプロジェクトルートからは常に同じ ID
- **衝突耐性**: 12 文字の hex（48 ビット）で実用上十分
- **可読性**: ファイルシステム上で確認しやすい長さ

## ディレクトリ構造

```
$HOME/.multiverse/workspaces/
└── <workspace-id>/           # sha1(projectRoot)[:12]
    ├── workspace.json        # Workspace メタデータ
    ├── config/               # WorkerPool 設定等
    ├── tasks/                # Task JSONL ファイル
    ├── attempts/             # Attempt JSON ファイル
    ├── ipc/                  # IPC Queue
    └── logs/                 # ログファイル
```

## 設計原則

### 単一責任

- WorkspaceStore は Workspace メタデータのみを管理
- Task/Attempt は orchestrator パッケージが管理
- 各 Store は独立して動作可能

### 自動ディレクトリ作成

- `SaveWorkspace()` は必要に応じてディレクトリを自動作成
- 呼び出し側は事前のディレクトリ作成が不要

### バージョニング

- `Version` フィールドでスキーマバージョンを管理
- 将来的なマイグレーションに対応可能

## テスト戦略

### ユニットテスト

- `workspace_store_test.go`: 読み書きの基本動作をテスト
- 一時ディレクトリを使用

### テストパターン

```go
func TestWorkspaceStore(t *testing.T) {
    tmpDir := t.TempDir()
    store := NewWorkspaceStore(tmpDir)

    // ID 生成テスト
    id := store.GetWorkspaceID("/path/to/project")

    // 保存・読み込みテスト
    ws := &Workspace{
        Version:     "1.0",
        ProjectRoot: "/path/to/project",
        DisplayName: "Test Project",
    }
    err := store.SaveWorkspace(ws)
    loaded, err := store.LoadWorkspace(id)
}
```

## 拡張予定

### 短期

- [ ] Workspace 一覧取得 API
- [ ] LastOpenedAt の自動更新
- [ ] Workspace 削除 API

### 長期

- [ ] Workspace 設定（WorkerPool 等）の統合管理
- [ ] マルチユーザー対応（ユーザー別 baseDir）

## 関連ドキュメント

- [../orchestrator/CLAUDE.md](../orchestrator/CLAUDE.md): Task/Attempt 永続化
- [../../cmd/multiverse-ide/CLAUDE.md](../../cmd/multiverse-ide/CLAUDE.md): IDE バックエンド
- [../../frontend/ide/CLAUDE.md](../../frontend/ide/CLAUDE.md): フロントエンド
