# PRD v2.0: multiverse - チャット駆動AI開発支援プラットフォーム

## 1. プロダクトビジョン

### 1.1 ビジョンステートメント

**multiverse** は、チャットインターフェースを通じて開発者の意図を理解し、
Meta-agentが自律的にタスクを分解・実行・評価する AI 開発支援プラットフォームです。

**コアコンセプト:**
- チャットウィンドウが全ての入力経路（AIとの対話）
- Meta-agentによる徹底的なタスク分解
  - 概念設計 → 実装設計 → 実装計画 → タスクマネジメント → アサイン
- 2D俯瞰UIでタスクグラフを視覚化（有向グラフ）
- WBSはリリースマイルストーンとして別枠管理
- 自律実行（計画→実行まで全自動、一時停止機能あり）

### 1.2 解決する課題

| 現状の課題 | multiverse v2.0 での解決 |
|-----------|-------------------------|
| タスク作成が手動・煩雑 | チャットから自然言語でタスク生成 |
| タスク間依存関係の管理が困難 | 有向グラフで依存関係を可視化 |
| 達成判定が曖昧 | 細分化されたタスクで個別・シンプルな達成判定 |
| 人間の介入が頻繁に必要 | 自律実行ループで人間待ち不要 |
| 問題・検討材料の散逸 | バックログで一元管理 |

### 1.3 ターゲットユーザー

- ソフトウェア開発者（個人・チーム）
- AIアシスタントと協調して開発を進めたいエンジニア
- 複数の並行タスクを俯瞰的に管理したい開発リーダー

---

## 2. 機能要件（フェーズ別）

### Phase 1: チャット → タスク生成（MVP）【優先度: 最高】

#### FR-P1-001: チャット入力UI

- 既存 FloatingChatWindow を拡張
- テキスト入力・送信
- メッセージ履歴表示（user/assistant/system）
- Wails IPC 経由でバックエンドと通信
- タスク生成結果のインライン表示

#### FR-P1-002: ChatHandler（バックエンド新規）

```go
// internal/chat/handler.go
type ChatHandler struct {
    Meta          MetaClient
    TaskStore     *orchestrator.TaskStore
    SessionStore  *ChatSessionStore
}

func (h *ChatHandler) HandleMessage(ctx context.Context, sessionID, message string) (*ChatResponse, error)
```

処理フロー:
1. ユーザーメッセージを ChatSession に保存
2. Meta-agent の `decompose` を呼び出し
3. 生成されたタスクを TaskStore に永続化
4. フロントエンドに結果を返却

#### FR-P1-003: Meta-agent decompose プロトコル

リクエスト:
```yaml
type: decompose
version: 1
payload:
  user_input: "認証機能を実装してほしい"
  context:
    workspace_path: "/path/to/project"
    existing_tasks: [...]
    conversation_history: [...]
```

レスポンス:
```yaml
type: decompose
version: 1
payload:
  understanding: "認証機能の実装を要求..."
  phases:
    - name: "概念設計"
      milestone: "M1-Auth-Design"
      tasks:
        - id: "task-001"
          title: "認証フロー設計"
          description: "..."
          acceptance_criteria: [...]
          dependencies: []
          wbs_level: 1
    - name: "実装設計"
      tasks: [...]
    - name: "実装"
      tasks: [...]
  potential_conflicts:
    - file: "src/auth/login.ts"
      tasks: ["task-004"]
      warning: "既存ファイルを変更"
```

#### FR-P1-004: Task 構造体拡張

```go
type Task struct {
    // 既存
    ID        string     `json:"id"`
    Title     string     `json:"title"`
    Status    TaskStatus `json:"status"`
    PoolID    string     `json:"poolId"`
    CreatedAt time.Time  `json:"createdAt"`
    UpdatedAt time.Time  `json:"updatedAt"`
    StartedAt *time.Time `json:"startedAt,omitempty"`
    DoneAt    *time.Time `json:"doneAt,omitempty"`

    // 新規
    Description        string   `json:"description,omitempty"`
    Dependencies       []string `json:"dependencies,omitempty"`
    ParentID           *string  `json:"parentId,omitempty"`
    WBSLevel           int      `json:"wbsLevel,omitempty"`
    PhaseName          string   `json:"phaseName,omitempty"`
    SourceChatID       *string  `json:"sourceChatId,omitempty"`
    AcceptanceCriteria []string `json:"acceptanceCriteria,omitempty"`
}
```

#### FR-P1-005: ノード表示（GridCanvas拡張）

- 新規タスク生成時のアニメーション
- 依存関係インジケーター（Phase 2 準備）
- フェーズ（概念設計/実装設計/実装）の色分け

---

### Phase 2: 依存関係グラフ・WBS表示【優先度: 高】

#### FR-P2-001: TaskGraphManager

```go
// internal/orchestrator/task_graph.go
type TaskGraphManager struct {
    TaskStore *TaskStore
}

type TaskGraph struct {
    Nodes map[string]*GraphNode
    Edges []TaskEdge
}

func (m *TaskGraphManager) BuildGraph() (*TaskGraph, error)
func (m *TaskGraphManager) GetExecutionOrder() ([]string, error)
func (m *TaskGraphManager) GetBlockedTasks() ([]string, error)
```

#### FR-P2-002: ConnectionLine（依存矢印）

```svelte
<!-- frontend/ide/src/lib/grid/ConnectionLine.svelte -->
<svg class="connection-line">
  <path d={calculatePath(fromNode, toNode)} class={status} />
  <marker id="arrowhead" ... />
</svg>
```

視覚表現:
- 完了した依存: 緑色の実線
- 未完了の依存: オレンジの破線
- ブロック状態: 赤色の太線

#### FR-P2-003: WBS表示モード

- ツールバーに WBS/Graph 切り替えボタン
- WBS ビュー: マイルストーン別のツリー表示
- 折りたたみ/展開機能
- 進捗率表示（完了タスク / 全タスク）

#### FR-P2-004: 依存に基づくスケジューリング

```go
func (s *Scheduler) ScheduleReadyTasks() error {
    for _, task := range s.GetPendingTasks() {
        if s.allDependenciesSatisfied(task) {
            s.ScheduleTask(task.ID)
        }
    }
}
```

---

### Phase 3: 自律実行ループ【優先度: 中】

#### FR-P3-001: ExecutionOrchestrator

**設計方針:**
- 既存の `Executor` を拡張せず、新規 `ExecutionOrchestrator` を作成
- `Executor` は単一タスク実行、`ExecutionOrchestrator` は複数タスクの自律実行ループを担当
- goroutine による非同期実行とチャネルによる状態制御

```go
// internal/orchestrator/execution_orchestrator.go

// ExecutionState は自律実行の状態を表す
type ExecutionState string

const (
    ExecutionStateIdle    ExecutionState = "IDLE"    // 未開始・停止済み
    ExecutionStateRunning ExecutionState = "RUNNING" // 実行中
    ExecutionStatePaused  ExecutionState = "PAUSED"  // 一時停止中
)

// ExecutionOrchestrator は自律実行ループを管理する
type ExecutionOrchestrator struct {
    Scheduler    *Scheduler
    Executor     *Executor
    GraphManager *TaskGraphManager
    TaskStore    *TaskStore
    EventEmitter EventEmitter          // Wails Events 抽象化

    state        ExecutionState
    stateMu      sync.RWMutex          // 状態の排他制御

    pauseCh      chan struct{}         // 一時停止シグナル
    resumeCh     chan struct{}         // 再開シグナル
    stopCh       chan struct{}         // 停止シグナル

    maxConcurrent int                  // 同時実行タスク数（デフォルト: 1）
    runningTasks  map[string]context.CancelFunc  // 実行中タスクのキャンセル関数
    runningMu     sync.Mutex

    logger       *slog.Logger
}

// NewExecutionOrchestrator は ExecutionOrchestrator を作成する
func NewExecutionOrchestrator(
    scheduler *Scheduler,
    executor *Executor,
    taskStore *TaskStore,
    eventEmitter EventEmitter,
) *ExecutionOrchestrator

// Start は自律実行ループを開始する（非ブロッキング）
func (e *ExecutionOrchestrator) Start(ctx context.Context) error

// Pause は新規タスク開始を一時停止する（実行中タスクは継続）
func (e *ExecutionOrchestrator) Pause() error

// Resume は一時停止状態から再開する
func (e *ExecutionOrchestrator) Resume() error

// Stop は自律実行ループを停止する
func (e *ExecutionOrchestrator) Stop() error

// State は現在の実行状態を返す
func (e *ExecutionOrchestrator) State() ExecutionState

// runLoop は自律実行のメインループ（内部goroutine）
func (e *ExecutionOrchestrator) runLoop(ctx context.Context)
```

**実行ループの処理フロー:**

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

#### FR-P3-002: EventEmitter インターフェース

**設計方針:**
- Wails runtime への依存を抽象化してテスト可能にする
- 本番は Wails Events、テストはモック実装

```go
// internal/orchestrator/events.go

// EventEmitter はイベント発火を抽象化するインターフェース
type EventEmitter interface {
    Emit(eventName string, data any)
}

// WailsEventEmitter は Wails runtime を使ったイベント発火
type WailsEventEmitter struct {
    ctx context.Context  // Wails startup で渡される context
}

func (w *WailsEventEmitter) Emit(eventName string, data any) {
    runtime.EventsEmit(w.ctx, eventName, data)
}

// イベント名定義
const (
    EventTaskStateChange      = "task:stateChange"
    EventExecutionStateChange = "execution:stateChange"
    EventTaskProgress         = "task:progress"
    EventBacklogAdded         = "backlog:added"
)

// TaskStateChangeEvent はタスク状態変更イベントのペイロード
type TaskStateChangeEvent struct {
    TaskID    string     `json:"taskId"`
    OldStatus TaskStatus `json:"oldStatus"`
    NewStatus TaskStatus `json:"newStatus"`
    Timestamp time.Time  `json:"timestamp"`
}

// ExecutionStateChangeEvent は実行状態変更イベントのペイロード
type ExecutionStateChangeEvent struct {
    OldState  ExecutionState `json:"oldState"`
    NewState  ExecutionState `json:"newState"`
    Timestamp time.Time      `json:"timestamp"`
}
```

#### FR-P3-003: 一時停止・再開機能

**バックエンド API:**

```go
// cmd/multiverse-ide/app.go に追加

// StartExecution は自律実行を開始する
func (a *App) StartExecution() error {
    return a.executionOrchestrator.Start(a.ctx)
}

// PauseExecution は自律実行を一時停止する
func (a *App) PauseExecution() error {
    return a.executionOrchestrator.Pause()
}

// ResumeExecution は自律実行を再開する
func (a *App) ResumeExecution() error {
    return a.executionOrchestrator.Resume()
}

// StopExecution は自律実行を停止する
func (a *App) StopExecution() error {
    return a.executionOrchestrator.Stop()
}

// GetExecutionState は現在の実行状態を返す
func (a *App) GetExecutionState() string {
    return string(a.executionOrchestrator.State())
}
```

**フロントエンド:**

```typescript
// frontend/ide/src/stores/executionStore.ts

import { writable, derived } from 'svelte/store';
import { StartExecution, PauseExecution, ResumeExecution, StopExecution, GetExecutionState } from '$lib/wailsjs/go/main/App';
import { EventsOn } from '$lib/wailsjs/runtime/runtime';

export type ExecutionState = 'IDLE' | 'RUNNING' | 'PAUSED';

export const executionState = writable<ExecutionState>('IDLE');

// Wails Events リスナー設定
export function initExecutionEvents() {
    EventsOn('execution:stateChange', (event: { newState: ExecutionState }) => {
        executionState.set(event.newState);
    });
}

// アクション
export async function startExecution(): Promise<void> {
    await StartExecution();
}

export async function pauseExecution(): Promise<void> {
    await PauseExecution();
}

export async function resumeExecution(): Promise<void> {
    await ResumeExecution();
}

export async function stopExecution(): Promise<void> {
    await StopExecution();
}
```

**Toolbar UI:**

```svelte
<!-- frontend/ide/src/lib/toolbar/ExecutionControls.svelte -->
<script lang="ts">
    import { executionState, startExecution, pauseExecution, resumeExecution, stopExecution } from '../../stores/executionStore';
</script>

<div class="execution-controls">
    {#if $executionState === 'IDLE'}
        <button on:click={startExecution} title="実行開始">
            ▶️ 開始
        </button>
    {:else if $executionState === 'RUNNING'}
        <button on:click={pauseExecution} title="一時停止">
            ⏸️ 一時停止
        </button>
        <button on:click={stopExecution} title="停止">
            ⏹️ 停止
        </button>
    {:else if $executionState === 'PAUSED'}
        <button on:click={resumeExecution} title="再開">
            ▶️ 再開
        </button>
        <button on:click={stopExecution} title="停止">
            ⏹️ 停止
        </button>
    {/if}
    <span class="state-label">{$executionState}</span>
</div>
```

#### FR-P3-004: 自動リトライ/人間判断

**RetryPolicy 設計:**

```go
// internal/orchestrator/retry.go

// RetryPolicy はタスク失敗時のリトライポリシーを定義する
type RetryPolicy struct {
    MaxAttempts     int           // 最大試行回数（デフォルト: 3）
    BackoffBase     time.Duration // バックオフ基準時間（デフォルト: 5秒）
    BackoffMax      time.Duration // バックオフ最大時間（デフォルト: 5分）
    BackoffFactor   float64       // バックオフ乗数（デフォルト: 2.0）
    RequireHuman    bool          // 最大試行後に人間判断を要求するか
}

// DefaultRetryPolicy はデフォルトのリトライポリシー
func DefaultRetryPolicy() *RetryPolicy {
    return &RetryPolicy{
        MaxAttempts:   3,
        BackoffBase:   5 * time.Second,
        BackoffMax:    5 * time.Minute,
        BackoffFactor: 2.0,
        RequireHuman:  true,
    }
}

// CalculateBackoff は次のリトライまでの待機時間を計算する
func (p *RetryPolicy) CalculateBackoff(attemptNumber int) time.Duration {
    backoff := float64(p.BackoffBase) * math.Pow(p.BackoffFactor, float64(attemptNumber-1))
    if backoff > float64(p.BackoffMax) {
        backoff = float64(p.BackoffMax)
    }
    return time.Duration(backoff)
}

// ShouldRetry はリトライすべきかを判定する
func (p *RetryPolicy) ShouldRetry(attemptNumber int) bool {
    return attemptNumber < p.MaxAttempts
}
```

**HandleFailure 実装:**

```go
// ExecutionOrchestrator に追加

func (e *ExecutionOrchestrator) HandleFailure(task *Task, err error, attemptNumber int) error {
    logger := e.logger.With(
        slog.String("task_id", task.ID),
        slog.Int("attempt", attemptNumber),
        slog.Any("error", err),
    )

    if e.retryPolicy.ShouldRetry(attemptNumber) {
        // リトライ
        backoff := e.retryPolicy.CalculateBackoff(attemptNumber)
        logger.Info("scheduling retry", slog.Duration("backoff", backoff))

        // バックオフ後にリトライキューに追加
        time.AfterFunc(backoff, func() {
            e.retryQueue <- task.ID
        })
        return nil
    }

    if e.retryPolicy.RequireHuman {
        // バックログに追加
        logger.Info("adding to backlog for human review")
        return e.addToBacklog(task, err, BacklogTypeFailure)
    }

    // FAILED としてマーク（既に Executor で実施済み）
    logger.Warn("task permanently failed")
    return nil
}
```

#### FR-P3-005: バックログ管理

**BacklogStore 設計:**

```go
// internal/orchestrator/backlog.go

// BacklogType はバックログアイテムの種類を表す
type BacklogType string

const (
    BacklogTypeFailure  BacklogType = "FAILURE"  // タスク失敗
    BacklogTypeQuestion BacklogType = "QUESTION" // Meta-agent からの質問
    BacklogTypeBlocker  BacklogType = "BLOCKER"  // 外部ブロッカー
)

// BacklogItem はバックログアイテムを表す
type BacklogItem struct {
    ID          string      `json:"id"`
    TaskID      string      `json:"taskId"`
    Type        BacklogType `json:"type"`
    Title       string      `json:"title"`
    Description string      `json:"description"`
    Priority    int         `json:"priority"`    // 1-5（5が最高）
    CreatedAt   time.Time   `json:"createdAt"`
    ResolvedAt  *time.Time  `json:"resolvedAt,omitempty"`
    Resolution  string      `json:"resolution,omitempty"`
    Metadata    map[string]any `json:"metadata,omitempty"` // エラー詳細等
}

// BacklogStore はバックログアイテムを永続化する
type BacklogStore struct {
    WorkspaceDir string
    logger       *slog.Logger
}

// NewBacklogStore は BacklogStore を作成する
func NewBacklogStore(workspaceDir string) *BacklogStore

// Add はバックログアイテムを追加する
func (s *BacklogStore) Add(item *BacklogItem) error

// Get はバックログアイテムを取得する
func (s *BacklogStore) Get(id string) (*BacklogItem, error)

// List は全バックログアイテムを取得する
func (s *BacklogStore) List() ([]BacklogItem, error)

// ListUnresolved は未解決のバックログアイテムを取得する
func (s *BacklogStore) ListUnresolved() ([]BacklogItem, error)

// Resolve はバックログアイテムを解決済みにする
func (s *BacklogStore) Resolve(id string, resolution string) error

// Delete はバックログアイテムを削除する
func (s *BacklogStore) Delete(id string) error
```

**永続化形式:**

```
<workspace-dir>/backlog/
├── <item-id>.json  # 各アイテム
└── index.json      # 一覧（optional, 高速化用）
```

**バックログ API:**

```go
// cmd/multiverse-ide/app.go に追加

func (a *App) GetBacklogItems() ([]BacklogItem, error)
func (a *App) ResolveBacklogItem(id string, resolution string) error
func (a *App) DeleteBacklogItem(id string) error
```

**バックログ UI:**

```svelte
<!-- frontend/ide/src/lib/backlog/BacklogPanel.svelte -->
<script lang="ts">
    import { backlogItems, resolveItem, deleteItem } from '../../stores/backlogStore';
</script>

<aside class="backlog-panel">
    <h3>バックログ ({$backlogItems.length})</h3>
    {#each $backlogItems as item}
        <div class="backlog-item" class:failure={item.type === 'FAILURE'}>
            <span class="type-badge">{item.type}</span>
            <h4>{item.title}</h4>
            <p>{item.description}</p>
            <div class="actions">
                <button on:click={() => resolveItem(item.id)}>解決</button>
                <button on:click={() => deleteItem(item.id)}>削除</button>
            </div>
        </div>
    {/each}
</aside>
```

#### FR-P3-006: リアルタイム進捗表示（タスク状態の即時反映）

**設計方針:**
- 現在のポーリング（2秒間隔）を Wails Events で置き換え
- タスク状態変更時に即座にフロントエンドへ通知
- ポーリングはフォールバックとして維持（間隔を10秒に延長）

**フロントエンド Events 統合:**

```typescript
// frontend/ide/src/stores/taskStore.ts に追加

import { EventsOn } from '$lib/wailsjs/runtime/runtime';

// Wails Events リスナー設定
export function initTaskEvents() {
    EventsOn('task:stateChange', (event: TaskStateChangeEvent) => {
        tasks.updateTask(event.taskId, { status: event.newStatus });
    });
}

// App.svelte の onMount で呼び出し
```

```svelte
<!-- App.svelte 変更 -->
<script>
    import { initTaskEvents } from './stores/taskStore';
    import { initExecutionEvents } from './stores/executionStore';

    onMount(() => {
        initTaskEvents();
        initExecutionEvents();
        // ポーリング間隔を10秒に延長（フォールバック）
        interval = setInterval(loadData, 10000);
    });
</script>
```

---

## 3. データモデル

### Task（拡張）

| フィールド | 型 | 説明 |
|-----------|-----|------|
| dependencies | []string | 依存タスクIDリスト |
| parentId | *string | 親タスクID（WBS階層用） |
| wbsLevel | int | WBS階層レベル（1=概念設計, 2=実装設計, 3=実装） |
| phaseName | string | フェーズ名 |
| sourceChatId | *string | 生成元チャットセッションID |
| acceptanceCriteria | []string | 達成条件リスト |

### ChatSession

| フィールド | 型 | 説明 |
|-----------|-----|------|
| id | string | セッションID |
| workspaceId | string | ワークスペースID |
| messages | []ChatMessage | メッセージ一覧 |
| createdAt | time.Time | 作成日時 |
| updatedAt | time.Time | 更新日時 |

### ChatMessage

| フィールド | 型 | 説明 |
|-----------|-----|------|
| id | string | メッセージID |
| role | string | user / assistant / system |
| content | string | メッセージ本文 |
| timestamp | time.Time | タイムスタンプ |
| generatedTasks | []string | このメッセージで生成されたタスクID |

### BacklogItem

| フィールド | 型 | 説明 |
|-----------|-----|------|
| id | string | バックログID |
| taskId | string | 関連タスクID |
| type | BacklogType | FAILURE / QUESTION / BLOCKER |
| title | string | タイトル |
| description | string | 説明 |
| priority | int | 優先度 |
| createdAt | time.Time | 作成日時 |
| resolvedAt | *time.Time | 解決日時 |
| resolution | string | 解決方法 |

---

## 4. アーキテクチャ

### 4層構造（維持 + 拡張）

```
┌─────────────────────────────────────────────────────┐
│  multiverse-ide (Desktop UI)                        │
│  - ChatWindow → タスク生成                           │
│  - GridCanvas → 依存グラフ表示                       │
│  - WBSView → マイルストーン表示                      │
│  - BacklogPanel → バックログ管理                     │
└──────────────┬──────────────────────────────────────┘
               │ Wails IPC + Events
┌──────────────▼──────────────────────────────────────┐
│  Orchestrator Layer                                 │
│  - ChatHandler (NEW)                                │
│  - TaskGraphManager (NEW)                           │
│  - ExecutionOrchestrator (NEW)                      │
│  - BacklogStore (NEW)                               │
│  - TaskStore / Scheduler                            │
└──────────────┬──────────────────────────────────────┘
               │
┌──────────────▼──────────────────────────────────────┐
│  AgentRunner Core + Meta-agent                      │
│  - FSM（既存維持）                                   │
│  - decompose プロトコル (NEW)                        │
└──────────────┬──────────────────────────────────────┘
               │
┌──────────────▼──────────────────────────────────────┘
│  Worker (Docker Sandbox)                            │
└─────────────────────────────────────────────────────┘
```

### 新規コンポーネント

| コンポーネント | 場所 | 責務 |
|--------------|------|------|
| ChatHandler | internal/chat/handler.go | チャット入力のMeta-agent転送、タスク生成 |
| TaskGraphManager | internal/orchestrator/task_graph.go | 依存関係グラフの構築・管理 |
| ExecutionOrchestrator | internal/orchestrator/executor.go | 自律実行ループ、一時停止/再開 |
| BacklogStore | internal/orchestrator/backlog.go | 問題・検討材料の永続化 |
| ChatSessionStore | internal/chat/session_store.go | チャット履歴の永続化 |

---

## 5. マイルストーン

### M1: チャット→タスク生成（2週間）

**Week 1:**
- Task 構造体拡張
- Meta-agent decompose プロトコル
- ChatHandler 実装
- ChatSession 永続化

**Week 2:**
- FloatingChatWindow バックエンド連携
- タスク生成結果のUI表示
- E2Eテスト

### M2: 依存グラフ・WBS表示（2週間）

**Week 3:**
- TaskGraphManager
- Scheduler 依存チェック拡張
- ConnectionLine コンポーネント

**Week 4:**
- WBS ツリービュー
- マイルストーン表示
- 進捗率計算

### M3: 自律実行ループ（2週間）

**Week 5:**
- ExecutionOrchestrator
- 一時停止/再開
- Wails Events リアルタイム通知

**Week 6:**
- 自動リトライ
- BacklogStore
- バックログUI

---

## 6. 受け入れ条件

### Phase 1 完了条件

| ID | 条件 |
|----|------|
| AC-P1-01 | チャットからテキストを送信できる |
| AC-P1-02 | Meta-agent がタスク分解を行い、複数タスクが生成される |
| AC-P1-03 | 生成タスクが tasks/*.jsonl に永続化される |
| AC-P1-04 | タスクに依存関係情報が含まれる |
| AC-P1-05 | GridCanvas にノードとして表示される |

### Phase 2 完了条件

| ID | 条件 |
|----|------|
| AC-P2-01 | タスク間依存が矢印で表示される |
| AC-P2-02 | 依存タスク未完了時に BLOCKED 状態になる |
| AC-P2-03 | WBS ビューでツリー表示できる |
| AC-P2-04 | マイルストーン別の進捗率が表示される |

### Phase 3 完了条件

| ID | 条件 |
|----|------|
| AC-P3-01 | 自動実行で依存順にタスクが実行される |
| AC-P3-02 | 一時停止で新規タスク開始が停止する |
| AC-P3-03 | 再開で実行が継続する |
| AC-P3-04 | 失敗時に自動リトライまたはバックログ追加 |

---

## 7. 技術的リスクと対策

| リスク | 影響度 | 対策 |
|--------|--------|------|
| Meta-agent のタスク分解精度が低い | 高 | プロンプトエンジニアリング、人間レビュー機能 |
| 依存関係の循環参照 | 中 | グラフ構築時にサイクル検出 |
| 大量タスク時のUI性能劣化 | 中 | 仮想化描画（可視領域のみレンダリング） |
| ファイルコンフリクト検出漏れ | 高 | Meta-agent に明示的なコンフリクト分析を依頼 |
| 自律実行中のエラー連鎖 | 高 | 失敗回数閾値でループ停止、人間判断モード |

---

## 8. 既存設計との差分

### 削除

- タスク作成ダイアログ（FR-IDE-012）→ チャットに置換

### 維持

- 4層アーキテクチャ
- $HOME/.multiverse/workspaces/ 構造
- JSONL/JSON 永続化形式
- FSM 状態遷移
- Task/Attempt のステータス定義

### 変更

- Task 構造体: 依存関係、WBS、生成元情報追加
- Meta-agent プロトコル: decompose 追加
- Scheduler: 依存チェック追加

---

## 9. 技術スタック

### バックエンド（維持）

| カテゴリ | 技術 | バージョン |
|---------|------|-----------|
| 言語 | Go | 1.23+ |
| デスクトップ | Wails | v2 |
| コンテナ | Docker | - |
| LLM | OpenAI API | - |

### フロントエンド（維持）

| カテゴリ | 技術 | バージョン |
|---------|------|-----------|
| フレームワーク | Svelte | 4 |
| 型安全 | TypeScript | 5 |
| ビルド | Vite | 5 |
| パッケージ管理 | pnpm | - |

### 新規追加

| カテゴリ | 技術 | 用途 |
|---------|------|------|
| グラフ描画 | SVG | 依存関係の矢印描画 |
| リアルタイム通信 | Wails Events | 状態変更通知 |
