# multiverse-ide - Wails デスクトップアプリケーション

このディレクトリは multiverse IDE の Wails デスクトップアプリケーションを提供します。

## 責務

- **デスクトップアプリ起動**: Wails フレームワークによるウィンドウ管理
- **バックエンド API**: フロントエンドへの Go API 公開
- **コンポーネント統合**: WorkspaceStore, TaskStore, Scheduler の統合

## ファイル構成

| ファイル | 役割 |
|---------|------|
| main.go | Wails アプリケーション初期化、アセット埋め込み |
| app.go | バックエンドロジック、フロントエンド API |

## アーキテクチャ

```
┌─────────────────────────────────────┐
│  Svelte Frontend (frontend/ide/)   │
│  - WorkspaceSelector               │
│  - TaskList / TaskDetail           │
│  - TaskCreate                      │
└──────────────┬──────────────────────┘
               │ Wails IPC
┌──────────────▼──────────────────────┐
│  App struct (app.go)               │
│  - SelectWorkspace()               │
│  - GetWorkspace()                  │
│  - ListTasks() / CreateTask()      │
│  - RunTask()                       │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│  internal/ide + orchestrator       │
│  - WorkspaceStore                  │
│  - TaskStore + Scheduler           │
└─────────────────────────────────────┘
```

## App 構造体

```go
type App struct {
    ctx            context.Context
    workspaceStore *ide.WorkspaceStore
    taskStore      *orchestrator.TaskStore
    scheduler      *orchestrator.Scheduler
    currentWS      *ide.Workspace
}
```

## 公開 API（フロントエンドから呼び出し可能）

### SelectWorkspace

```go
func (a *App) SelectWorkspace() string
```

ディレクトリ選択ダイアログを表示し、Workspace を初期化します。

**戻り値**: Workspace ID（空文字列はキャンセル）

**処理フロー**:
1. `runtime.OpenDirectoryDialog()` でディレクトリ選択
2. `WorkspaceStore.GetWorkspaceID()` で ID 生成
3. 既存 Workspace をロード、なければ新規作成
4. TaskStore, Scheduler を初期化

### GetWorkspace

```go
func (a *App) GetWorkspace(id string) *ide.Workspace
```

Workspace 情報を取得します。

### ListTasks

```go
func (a *App) ListTasks() []orchestrator.Task
```

現在の Workspace 内の全タスクを取得します。

**処理フロー**:
1. `<workspace-dir>/tasks/` 内の JSONL ファイルを列挙
2. 各ファイルから最新状態をロード

### CreateTask

```go
func (a *App) CreateTask(title string, poolID string) *orchestrator.Task
```

新規タスクを作成します。

**パラメータ**:
- `title`: タスクタイトル
- `poolID`: Worker Pool ID

### RunTask

```go
func (a *App) RunTask(taskID string) error
```

タスクをスケジュールして実行を開始します。

## Wails 設定（main.go）

```go
wails.Run(&options.App{
    Title:  "multiverse-ide",
    Width:  1024,
    Height: 768,
    AssetServer: &assetserver.Options{
        Assets: assets,  // embed.FS で埋め込み
    },
    BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
    OnStartup:        app.startup,
    Bind: []interface{}{
        app,  // App 構造体の公開メソッドをバインド
    },
})
```

## ライフサイクル

### startup

```go
func (a *App) startup(ctx context.Context) {
    a.ctx = ctx  // Wails ランタイムコンテキストを保存
}
```

- Wails ランタイム関数（`runtime.OpenDirectoryDialog` 等）の使用に必要

## データディレクトリ

```
$HOME/.multiverse/workspaces/<workspace-id>/
├── workspace.json        # Workspace メタデータ
├── tasks/                # Task JSONL
├── attempts/             # Attempt JSON
└── ipc/                  # IPC Queue
```

## 開発・ビルド

### 開発モード

```bash
wails dev
```

- ホットリロード有効
- フロントエンド変更が即座に反映

### 本番ビルド

```bash
wails build
```

- 単一バイナリ生成
- フロントエンドアセットを埋め込み

## テスト戦略

### ユニットテスト

- App のメソッドは Wails ランタイムに依存するため、直接的なユニットテストは困難
- 内部ロジックは `internal/` パッケージでテスト

### 統合テスト

- E2E テストは手動または自動化ツール（Playwright 等）で実施

## 拡張予定

### 短期

- [ ] タスク削除 API
- [ ] タスク更新 API（タイトル変更等）
- [ ] Workspace 一覧 API

### 長期

- [ ] リアルタイムタスク状態更新（WebSocket 相当）
- [ ] Worker Pool 設定 UI
- [ ] ログビューア

## 関連ドキュメント

- [../../internal/ide/CLAUDE.md](../../internal/ide/CLAUDE.md): WorkspaceStore
- [../../internal/orchestrator/CLAUDE.md](../../internal/orchestrator/CLAUDE.md): TaskStore, Scheduler
- [../../frontend/ide/CLAUDE.md](../../frontend/ide/CLAUDE.md): Svelte フロントエンド
- [../../wails.json](../../wails.json): Wails プロジェクト設定
