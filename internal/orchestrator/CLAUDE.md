# Orchestrator Package - タスク永続化・スケジューリング・IPC

このパッケージは multiverse IDE と AgentRunner Core の間の中間層で、タスクの永続化、スケジューリング、Worker との IPC を管理します。

## 責務

- **Task/Attempt の永続化**: JSONL/JSON 形式でタスク状態を永続化
- **スケジューリング**: Task を READY 状態に更新し、Queue にジョブを投入
- **IPC Queue**: ファイルベースの Orchestrator ↔ Worker 通信

## ファイル構成

| ファイル | 役割 |
|---------|------|
| task_store.go | Task/Attempt の JSONL/JSON 永続化 |
| task_store_test.go | TaskStore のユニットテスト |
| scheduler.go | タスクスケジューリング |
| executor.go | Orchestrator から AgentRunner Core への実行委譲（予定） |
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

## 拡張予定

### 短期

- [ ] ファイルロックによる並行アクセス制御
- [ ] Worker → Orchestrator の結果キュー実装
- [ ] Attempt のライフサイクル管理

### 長期

- [ ] Worker Pool 設定の永続化
- [ ] タスク依存関係管理
- [ ] 優先度ベーススケジューリング

## 関連ドキュメント

- [../ide/CLAUDE.md](../ide/CLAUDE.md): Workspace 管理
- [../core/CLAUDE.md](../core/CLAUDE.md): AgentRunner Core FSM
- [../../cmd/multiverse-ide/CLAUDE.md](../../cmd/multiverse-ide/CLAUDE.md): IDE バックエンド
