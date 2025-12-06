# TODO: multiverse v2.0 Implementation

Based on PRD v2.0

---

## 進捗サマリ

| Phase | Status | 備考 |
|-------|--------|------|
| Phase 1: チャット→タスク生成 | 🟢 ほぼ完了 | E2Eテストのみ残 |
| Phase 2: 依存グラフ・WBS表示 | 🟢 完了 | Week 3-4 + Scheduler拡張 完了 |
| Phase 3: 自律実行ループ | 🟢 ほぼ完了 | Week 5-6 完了、失敗処理統合のみ残 |

---

## Phase 1: チャット → タスク生成（MVP）

### Week 1: バックエンド実装

#### 1.1 Task 構造体拡張

- [x] `internal/orchestrator/task_store.go`
  - [x] `Description string` フィールド追加
  - [x] `Dependencies []string` フィールド追加
  - [x] `ParentID *string` フィールド追加
  - [x] `WBSLevel int` フィールド追加
  - [x] `PhaseName string` フィールド追加
  - [x] `SourceChatID *string` フィールド追加
  - [x] `AcceptanceCriteria []string` フィールド追加

#### 1.2 Meta-agent decompose プロトコル

- [x] `internal/meta/protocol.go`
  - [x] `DecomposeRequest` 構造体追加
  - [x] `DecomposeResponse` 構造体追加
  - [x] `DecomposedTask` 構造体追加
  - [x] `DecomposedPhase` 構造体追加
- [x] `internal/meta/client.go`
  - [x] `Decompose(ctx, request)` メソッド追加
  - [x] decompose 用システムプロンプト定義

#### 1.3 ChatHandler 実装

- [x] `internal/chat/handler.go` (新規)
  - [x] `ChatHandler` 構造体
  - [x] `HandleMessage()` メソッド
  - [x] Meta-agent 呼び出しロジック
  - [x] タスク生成・保存ロジック
- [x] `internal/chat/session_store.go` (新規)
  - [x] `ChatSession` 構造体
  - [x] `ChatMessage` 構造体
  - [x] JSONL 永続化
- [x] `internal/chat/CLAUDE.md` (新規)

#### 1.4 IDE バックエンド API

- [x] `cmd/multiverse-ide/app.go`
  - [x] `SendChatMessage(sessionID, message string) (*ChatResponse, error)`
  - [x] `GetChatHistory(sessionID string) ([]ChatMessage, error)`
  - [x] `CreateChatSession() (string, error)`
  - [x] ChatHandler 初期化

### Week 2: フロントエンド連携

#### 2.1 チャットUI連携

- [x] `frontend/ide/src/lib/components/chat/FloatingChatWindow.svelte`
  - [x] Wails API 呼び出し（SendChatMessage）
  - [x] 応答メッセージの表示
  - [x] タスク生成結果のインライン表示
- [x] `frontend/ide/src/stores/chat.ts`
  - [x] セッション管理
  - [x] メッセージ履歴管理
  - [x] Wails API 連携

#### 2.2 タスク表示更新

- [x] `frontend/ide/src/stores/taskStore.ts`
  - [x] 新規タスク追加時の状態更新
  - [x] 依存関係情報の保持（taskEdges, blockedTasks, readyTasks）
- [x] `frontend/ide/src/lib/grid/GridNode.svelte`
  - [x] フェーズ別色分け（概念設計/実装設計/実装/検証）

#### 2.3 テスト

- [x] ChatHandler ユニットテスト（handler_test.go）
- [x] Meta-agent decompose モックテスト（MockMetaClient）
- [ ] E2E テスト（チャット→タスク生成フロー）

---

## Phase 2: 依存関係グラフ・WBS表示

### Week 3: グラフ管理

#### 3.1 TaskGraphManager

- [x] `internal/orchestrator/task_graph.go` (新規)
  - [x] `TaskGraphManager` 構造体
  - [x] `TaskGraph` 構造体
  - [x] `GraphNode` 構造体
  - [x] `TaskEdge` 構造体
  - [x] `BuildGraph()` メソッド
  - [x] `GetExecutionOrder()` メソッド（トポロジカルソート）
  - [x] `GetBlockedTasks()` メソッド
  - [x] `GetReadyTasks()` メソッド
  - [x] `DetectCycle()` メソッド（サイクル検出）
  - [x] `GetTaskDependencyInfo()` メソッド
- [x] `internal/orchestrator/task_graph_test.go` (新規)
  - [x] BuildGraph テスト
  - [x] GetExecutionOrder テスト
  - [x] サイクル検出テスト
  - [x] ブロックタスク検出テスト

#### 3.2 Scheduler 拡張

- [x] `internal/orchestrator/scheduler.go`
  - [x] `ScheduleReadyTasks()` メソッド
  - [x] `allDependenciesSatisfied()` メソッド
  - [x] BLOCKED 状態の自動設定
  - [x] `UpdateBlockedTasks()` メソッド
  - [x] `SetBlockedStatusForPendingWithUnsatisfiedDeps()` メソッド
- [x] `internal/orchestrator/scheduler_test.go` (新規)

#### 3.3 ConnectionLine コンポーネント

- [x] `frontend/ide/src/lib/grid/ConnectionLine.svelte` (新規)
  - [x] SVG ベジェ曲線パス計算
  - [x] 依存状態による色分け（satisfied: 緑, unsatisfied: オレンジ破線）
  - [x] 矢印マーカー
  - [x] ダッシュアニメーション（未満の依存）
- [x] `frontend/ide/src/lib/grid/GridCanvas.svelte`
  - [x] ConnectionLine のレンダリング
  - [x] 矢印マーカー定義（SVG defs）

### Week 4: WBS・視覚化

#### 4.1 WBS ビュー

- [x] `frontend/ide/src/lib/wbs/WBSView.svelte` (新規)
  - [x] ツリー構造表示
  - [x] 折りたたみ/展開
  - [x] マイルストーン表示
- [x] `frontend/ide/src/lib/wbs/WBSNode.svelte` (新規)
- [x] `frontend/ide/src/stores/wbsStore.ts` (新規)

#### 4.2 進捗率表示

- [x] `frontend/ide/src/lib/toolbar/Toolbar.svelte`
  - [x] 進捗率バー
  - [x] Graph/WBS 切り替えボタン

---

## Phase 3: 自律実行ループ

### Week 5: 実行オーケストレーション

#### 5.1 ExecutionOrchestrator（バックエンド）

- [x] `internal/orchestrator/execution_orchestrator.go` (新規)
  - [x] `ExecutionState` 型定義（IDLE/RUNNING/PAUSED）
  - [x] `ExecutionOrchestrator` 構造体
  - [x] `NewExecutionOrchestrator()` コンストラクタ
  - [x] `Start(ctx)` メソッド（非ブロッキング実行開始）
  - [x] `Pause()` メソッド（新規タスク開始停止）
  - [x] `Resume()` メソッド（一時停止解除）
  - [x] `Stop()` メソッド（ループ終了）
  - [x] `State()` メソッド（現在状態取得）
  - [x] `runLoop(ctx)` 内部メソッド（自律実行ループ）
- [x] `internal/orchestrator/execution_orchestrator_test.go` (新規)
  - [x] Start/Pause/Resume/Stop の状態遷移テスト
  - [ ] 依存順実行テスト（モック使用）
  - [ ] 並行実行制御テスト

#### 5.2 EventEmitter インターフェース

- [x] `internal/orchestrator/events.go` (新規)
  - [x] `EventEmitter` インターフェース定義
  - [x] `WailsEventEmitter` 実装
  - [x] イベント名定数（EventTaskStateChange, EventExecutionStateChange）
  - [x] `TaskStateChangeEvent` 構造体
  - [x] `ExecutionStateChangeEvent` 構造体
- [x] `internal/orchestrator/execution_orchestrator_test.go` 内
  - [x] `MockEventEmitter` テスト用実装（testify/mock 使用）

#### 5.3 IDE バックエンド API 拡張

- [x] `cmd/multiverse-ide/app.go`
  - [x] `executionOrchestrator` フィールド追加
  - [x] `StartExecution()` API
  - [x] `PauseExecution()` API
  - [x] `ResumeExecution()` API
  - [x] `StopExecution()` API
  - [x] `GetExecutionState()` API
  - [x] SelectWorkspace/OpenWorkspaceByID で ExecutionOrchestrator 初期化

#### 5.4 フロントエンド実行状態管理

- [x] `frontend/ide/src/stores/executionStore.ts` (新規)
  - [x] `executionState` ストア
  - [x] `initExecutionEvents()` 関数（スタブ実装）
  - [x] `startExecution()` アクション（スタブ実装）
  - [x] `pauseExecution()` アクション（スタブ実装）
  - [x] `resumeExecution()` アクション（スタブ実装）
  - [x] `stopExecution()` アクション（スタブ実装）
  - [ ] Wails バインディング接続（スタブ→実API）
- [x] `frontend/ide/src/lib/toolbar/ExecutionControls.svelte` (新規)
  - [x] 開始ボタン（IDLE 時）
  - [x] 一時停止ボタン（RUNNING 時）
  - [x] 再開ボタン（PAUSED 時）
  - [x] 停止ボタン
  - [x] 状態ラベル表示
- [x] `frontend/ide/src/lib/toolbar/Toolbar.svelte`
  - [x] ExecutionControls 統合

#### 5.5 リアルタイム通知（Wails Events）

- [x] `frontend/ide/src/stores/taskStore.ts`
  - [x] `initTaskEvents()` 関数追加
  - [x] `task:stateChange` リスナー
- [x] `frontend/ide/src/App.svelte`
  - [x] `initTaskEvents()` 呼び出し
  - [x] `initExecutionEvents()` 呼び出し
  - [x] ポーリング間隔を 10 秒に延長

### Week 6: エラーハンドリング

#### 6.1 RetryPolicy

- [x] `internal/orchestrator/retry.go` (新規)
  - [x] `RetryPolicy` 構造体
  - [x] `DefaultRetryPolicy()` 関数
  - [x] `CalculateBackoff()` メソッド（指数バックオフ）
  - [x] `ShouldRetry()` メソッド
  - [x] `DetermineNextAction()` メソッド
- [x] `internal/orchestrator/retry_test.go` (新規)
  - [x] バックオフ計算テスト
  - [x] リトライ判定テスト
  - [x] 次アクション決定テスト

#### 6.2 ExecutionOrchestrator 失敗処理

- [ ] `internal/orchestrator/execution_orchestrator.go`
  - [ ] `HandleFailure()` メソッド
  - [ ] `retryQueue` チャネル追加
  - [ ] `addToBacklog()` 内部メソッド
  - [ ] リトライ回数トラッキング（attemptCount map）

#### 6.3 BacklogStore

- [x] `internal/orchestrator/backlog.go` (新規)
  - [x] `BacklogType` 型定義（FAILURE/QUESTION/BLOCKER）
  - [x] `BacklogItem` 構造体
  - [x] `BacklogStore` 構造体
  - [x] `NewBacklogStore()` コンストラクタ
  - [x] `Add()` メソッド
  - [x] `Get()` メソッド
  - [x] `List()` メソッド
  - [x] `ListUnresolved()` メソッド
  - [x] `Resolve()` メソッド
  - [x] `Delete()` メソッド
  - [x] `CreateFailureItem()` ヘルパー関数
- [x] `internal/orchestrator/backlog_test.go` (新規)
  - [x] CRUD テスト
  - [x] 未解決フィルタテスト

#### 6.4 バックログ API

- [x] `cmd/multiverse-ide/app.go`
  - [x] `backlogStore` フィールド追加
  - [x] `GetBacklogItems()` API
  - [x] `GetAllBacklogItems()` API
  - [x] `ResolveBacklogItem()` API
  - [x] `DeleteBacklogItem()` API

#### 6.5 バックログ UI

- [x] `frontend/ide/src/stores/backlogStore.ts` (新規)
  - [x] `backlogItems` ストア
  - [x] `initBacklogEvents()` 関数
  - [x] `loadBacklogItems()` 関数
  - [x] `resolveItem()` アクション
  - [x] `deleteItem()` アクション
- [x] `frontend/ide/src/lib/backlog/BacklogPanel.svelte` (新規)
  - [x] アイテム一覧表示
  - [x] タイプ別バッジ（FAILURE/QUESTION/BLOCKER）
  - [x] 解決・削除ボタン
  - [x] 空状態表示
  - [x] 解決ダイアログ
- [x] `frontend/ide/src/App.svelte`
  - [x] BacklogPanel 配置（サイドバー）
  - [x] バックログ表示FABボタン
  - [x] `initBacklogEvents()` 呼び出し

---

## 実装済みファイル一覧

### Phase 1 で作成予定

| ファイル | 種別 | 説明 |
|---------|------|------|
| `internal/chat/handler.go` | 新規 | ChatHandler |
| `internal/chat/session_store.go` | 新規 | ChatSession 永続化 |
| `internal/chat/CLAUDE.md` | 新規 | パッケージドキュメント |

### Phase 2 で作成予定

| ファイル | 種別 | 説明 |
|---------|------|------|
| `internal/orchestrator/task_graph.go` | 新規 | TaskGraphManager |
| `frontend/ide/src/lib/grid/ConnectionLine.svelte` | 新規 | 依存矢印 |
| `frontend/ide/src/lib/wbs/WBSView.svelte` | 新規 | WBS ビュー |
| `frontend/ide/src/lib/wbs/WBSNode.svelte` | 新規 | WBS ノード |
| `frontend/ide/src/stores/wbsStore.ts` | 新規 | WBS 状態管理 |

### Phase 3 で作成予定

| ファイル | 種別 | 説明 |
|---------|------|------|
| `internal/orchestrator/execution_orchestrator.go` | 新規 | ExecutionOrchestrator（自律実行ループ） |
| `internal/orchestrator/execution_orchestrator_test.go` | 新規 | ExecutionOrchestrator テスト |
| `internal/orchestrator/events.go` | 新規 | EventEmitter インターフェース |
| `internal/orchestrator/retry.go` | 新規 | RetryPolicy（リトライポリシー） |
| `internal/orchestrator/retry_test.go` | 新規 | RetryPolicy テスト |
| `internal/orchestrator/backlog.go` | 新規 | BacklogStore（バックログ永続化） |
| `internal/orchestrator/backlog_test.go` | 新規 | BacklogStore テスト |
| `internal/mock/event_emitter.go` | 新規 | MockEventEmitter（テスト用） |
| `frontend/ide/src/stores/executionStore.ts` | 新規 | 実行状態管理 |
| `frontend/ide/src/stores/backlogStore.ts` | 新規 | バックログ状態管理 |
| `frontend/ide/src/lib/toolbar/ExecutionControls.svelte` | 新規 | 実行制御ボタン |
| `frontend/ide/src/lib/backlog/BacklogPanel.svelte` | 新規 | バックログ UI |

---

## 次のアクション

1. **Phase 3 残作業**: ExecutionOrchestrator 失敗処理統合
   - `HandleFailure()` メソッド実装
   - RetryPolicy と BacklogStore の統合
   - リトライキューとバックオフ処理
2. **Phase 1 E2E テスト**: チャット→タスク生成フローのテスト
3. **テスト拡充**: ExecutionOrchestrator の依存順実行・並行実行テスト
