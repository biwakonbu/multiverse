# Orchestrator Package - タスク永続化・スケジューリング・自律実行

このパッケージは multiverse IDE と AgentRunner Core の間の中間層で、タスクの永続化、スケジューリング、依存グラフ管理、自律実行ループを管理します。

## 責務

- **Task/Attempt の永続化**: JSONL/JSON 形式でタスク状態を永続化
- **スケジューリング**: Task を READY 状態に更新し、Queue にジョブを投入
- **依存グラフ管理**: タスク間依存関係のグラフ構築・トポロジカルソート
- **自律実行ループ**: 依存順にタスクを自動実行、一時停止/再開機能
- **バックログ管理**: 失敗タスク・質問・ブロッカーの永続化
- **IPC Queue**: ファイルベースの Orchestrator ↔ Worker 通信

## ファイル構成

| ファイル | 役割 |
|---------|------|
| task_store.go | Task/Attempt の JSONL/JSON 永続化 |
| task_store_test.go | TaskStore のユニットテスト |
| scheduler.go | タスクスケジューリング・依存チェック |
| scheduler_test.go | Scheduler のユニットテスト |
| task_graph.go | TaskGraphManager（依存グラフ管理） |
| task_graph_test.go | TaskGraphManager のユニットテスト |
| executor.go | Executor（単一タスク実行） |
| execution_orchestrator.go | ExecutionOrchestrator（自律実行ループ）★Phase 3 |
| execution_orchestrator_test.go | ExecutionOrchestrator テスト★Phase 3 |
| events.go | EventEmitter インターフェース★Phase 3 |
| retry.go | RetryPolicy（リトライポリシー）★Phase 3 |
| retry_test.go | RetryPolicy テスト★Phase 3 |
| backlog.go | BacklogStore（バックログ永続化）★Phase 3 |
| backlog_test.go | BacklogStore テスト★Phase 3 |
| ipc/filesystem_queue.go | ファイルベース IPC キュー |

## 主要データモデル

### TaskStatus

```go
const (
    TaskStatusPending   TaskStatus = "PENDING"   // 初期状態
    TaskStatusReady     TaskStatus = "READY"     // スケジュール済み、実行待ち
    TaskStatusRunning   TaskStatus = "RUNNING"   // 実行中
    TaskStatusSucceeded TaskStatus = "SUCCEEDED" // 正常完了
    TaskStatusFailed    TaskStatus = "FAILED"    // 失敗
    TaskStatusCanceled  TaskStatus = "CANCELED"  // キャンセル
    TaskStatusBlocked   TaskStatus = "BLOCKED"   // ブロック（依存タスク待ち等）
)
```

### Task 構造体

```go
type Task struct {
    ID        string     `json:"id"`
    Title     string     `json:"title"`
    Status    TaskStatus `json:"status"`
    PoolID    string     `json:"poolId"`      // Worker Pool ID
    CreatedAt time.Time  `json:"createdAt"`
    UpdatedAt time.Time  `json:"updatedAt"`
    StartedAt *time.Time `json:"startedAt,omitempty"`
    DoneAt    *time.Time `json:"doneAt,omitempty"`
}
```

### AttemptStatus

```go
const (
    AttemptStatusStarting  AttemptStatus = "STARTING"  // 開始処理中
    AttemptStatusRunning   AttemptStatus = "RUNNING"   // 実行中
    AttemptStatusSucceeded AttemptStatus = "SUCCEEDED" // 成功
    AttemptStatusFailed    AttemptStatus = "FAILED"    // 失敗
    AttemptStatusTimeout   AttemptStatus = "TIMEOUT"   // タイムアウト
    AttemptStatusCanceled  AttemptStatus = "CANCELED"  // キャンセル
)
```

### Attempt 構造体

```go
type Attempt struct {
    ID           string        `json:"id"`
    TaskID       string        `json:"taskId"`
    Status       AttemptStatus `json:"status"`
    StartedAt    time.Time     `json:"startedAt"`
    FinishedAt   *time.Time    `json:"finishedAt,omitempty"`
    ErrorSummary string        `json:"errorSummary,omitempty"`
}
```

## 永続化パターン

### Task: JSONL 形式（イベントソーシング風）

- ファイル: `<workspace-dir>/tasks/<task-id>.jsonl`
- 1 行 = 1 JSON オブジェクト（状態のスナップショット）
- 最後の行が最新状態
- 状態履歴を保持し、監査証跡として機能

```jsonl
{"id":"task-1","title":"Feature A","status":"PENDING",...}
{"id":"task-1","title":"Feature A","status":"READY",...}
{"id":"task-1","title":"Feature A","status":"RUNNING",...}
```

### Attempt: JSON 形式（単一ファイル）

- ファイル: `<workspace-dir>/attempts/<attempt-id>.json`
- 1 Attempt 1 ファイル
- 上書き保存

## TaskStore API

```go
// 初期化
store := NewTaskStore(workspaceDir)

// Task 操作
task, err := store.LoadTask(id)      // JSONL の最終行を読み込み
err := store.SaveTask(task)          // JSONL にアペンド

// Attempt 操作
attempt, err := store.LoadAttempt(id) // JSON を読み込み
err := store.SaveAttempt(attempt)     // JSON を上書き保存

// ディレクトリ取得
taskDir := store.GetTaskDir()         // <workspace-dir>/tasks
attemptDir := store.GetAttemptDir()   // <workspace-dir>/attempts
```

## Scheduler API

```go
// 初期化
scheduler := NewScheduler(taskStore, queue)

// タスクスケジューリング
err := scheduler.ScheduleTask(taskID)
```

### ScheduleTask の処理フロー

1. TaskStore から Task をロード
2. Status が PENDING または FAILED であることを確認
3. Status を READY に更新、TaskStore に保存
4. IPC Queue にジョブを投入

## IPC Queue（ipc/filesystem_queue.go）

### Job 構造体

```go
type Job struct {
    ID      string `json:"id"`
    TaskID  string `json:"taskId"`
    PoolID  string `json:"poolId"`
    Payload any    `json:"payload"`
}
```

### FilesystemQueue API

```go
// 初期化
queue := NewFilesystemQueue(workspaceDir)

// ジョブ操作
err := queue.Enqueue(job)           // ipc/queue/<pool-id>/<job-id>.json に保存
jobIDs, err := queue.ListJobs(poolID) // キュー内ジョブ一覧
```

### ディレクトリ構造

```
<workspace-dir>/
├── ipc/
│   ├── queue/
│   │   └── <pool-id>/
│   │       └── <job-id>.json
│   └── results/              # Worker → Orchestrator（実装予定）
```

## 設計原則

### ファイルベース IPC の理由

- **シンプルさ**: データベース不要、ファイルシステムのみで動作
- **可視性**: JSON ファイルを直接確認可能
- **ポータビリティ**: 特別なインフラ不要
- **障害耐性**: ファイルロックで並行アクセスを制御（将来実装）

### Task vs Attempt の分離

- **Task**: 論理的なタスク単位、複数回リトライ可能
- **Attempt**: 物理的な実行単位、1 回の実行試行
- 1 Task に対して N Attempt（リトライ）

## テスト戦略

### ユニットテスト

- `task_store_test.go`: JSONL/JSON の読み書きをテスト
- 一時ディレクトリを使用し、実際のファイルシステム操作をテスト

### 統合テスト

- Scheduler + TaskStore + FilesystemQueue の連携テスト
- `test/integration/` で実施（予定）

## TaskGraphManager

タスク間の依存関係をグラフとして管理する。

### 主要機能

- **BuildGraph()**: 全タスクから依存グラフを構築
- **GetExecutionOrder()**: トポロジカルソートによる実行順序を返す
- **GetBlockedTasks()**: 依存が満たされていないタスクを返す
- **GetReadyTasks()**: 実行可能タスク（PENDING で全依存満）を返す
- **DetectCycle()**: サイクル検出、関与ノードを返す

### データ構造

```go
type TaskGraph struct {
    Nodes map[string]*GraphNode
    Edges []TaskEdge
}

type GraphNode struct {
    Task       *Task
    InDegree   int      // 入次数
    OutDegree  int      // 出次数
    Dependents []string // このタスクに依存しているタスクID
}

type TaskEdge struct {
    From      string // 依存元タスクID
    To        string // 依存先タスクID
    Satisfied bool   // 依存が満たされているか
}
```

## ExecutionOrchestrator（Phase 3）

自律実行ループを管理するコンポーネント。依存関係を考慮してタスクを順次実行し、一時停止/再開機能を提供する。

### 設計方針

- 既存の `Executor` は単一タスク実行のまま維持
- `ExecutionOrchestrator` は複数タスクの自律実行ループを担当
- goroutine による非同期実行とチャネルによる状態制御
- EventEmitter インターフェースで Wails Events への依存を抽象化

### 状態遷移

```
IDLE ──Start()──> RUNNING ──Pause()──> PAUSED
  ↑                  │                    │
  └──────Stop()──────┴──────Resume()──────┘
```

### 実行ループの処理フロー

```
runLoop:
  1. UpdateBlockedTasks() で BLOCKED → PENDING 解除
  2. ScheduleReadyTasks() で実行可能タスクを READY に
  3. READY タスクを取得
  4. maxConcurrent まで並列実行
  5. タスク完了時に EventEmit
  6. 全タスク完了まで繰り返し
  7. 一時停止シグナル受信時は待機
  8. 停止シグナル受信時はループ終了
```

### EventEmitter インターフェース

```go
type EventEmitter interface {
    Emit(eventName string, data any)
}
```

- **WailsEventEmitter**: 本番用、Wails runtime.EventsEmit を呼び出す
- **MockEventEmitter**: テスト用、発火イベントを記録

### RetryPolicy

失敗タスクのリトライポリシーを定義。指数バックオフでリトライ間隔を調整。

```go
type RetryPolicy struct {
    MaxAttempts   int           // 最大試行回数（デフォルト: 3）
    BackoffBase   time.Duration // バックオフ基準時間（5秒）
    BackoffMax    time.Duration // バックオフ最大時間（5分）
    BackoffFactor float64       // バックオフ乗数（2.0）
    RequireHuman  bool          // 最大試行後に人間判断を要求
}
```

### BacklogStore

失敗タスク・質問・ブロッカーを永続化する。

```go
type BacklogItem struct {
    ID          string      // バックログID
    TaskID      string      // 関連タスクID
    Type        BacklogType // FAILURE / QUESTION / BLOCKER
    Title       string
    Description string
    Priority    int         // 1-5（5が最高）
    CreatedAt   time.Time
    ResolvedAt  *time.Time
    Resolution  string
    Metadata    map[string]any
}
```

永続化形式: `<workspace-dir>/backlog/<item-id>.json`

## 拡張予定

### Phase 3（現在）

- [ ] ExecutionOrchestrator 実装
- [ ] EventEmitter インターフェース
- [ ] RetryPolicy 実装
- [ ] BacklogStore 実装

### 将来

- [ ] ファイルロックによる並行アクセス制御
- [ ] Worker Pool 設定の永続化
- [ ] 優先度ベーススケジューリング

## 関連ドキュメント

- [../ide/CLAUDE.md](../ide/CLAUDE.md): Workspace 管理
- [../core/CLAUDE.md](../core/CLAUDE.md): AgentRunner Core FSM
- [../../cmd/multiverse-ide/CLAUDE.md](../../cmd/multiverse-ide/CLAUDE.md): IDE バックエンド
- [../../PRD.md](../../PRD.md): Phase 3 詳細設計
