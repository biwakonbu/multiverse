# Core Package - タスク実行エンジン

このパッケージはAgentRunnerの心臓部で、タスク状態機械（FSM）のオーケストレーションと、Meta-agentとWorkerの調整を管理します。

## 概要

- **runner.go**: メインのFSMオーケストレーター、Runner構造体、タスク実行フロー
- **context.go**: TaskContext、TaskState定義、実行状態の保持
- **runner_test.go**: プロパティベーステスト（gopterを使用）

## タスク状態機械（FSM）

### 状態定義（TaskState）

```
PENDING → PLANNING → RUNNING → VALIDATING → COMPLETE
                                              ↓
                                            FAILED
```

- **PENDING**: 初期状態、タスク初期化直前
- **PLANNING**: PRD読み込み完了、Meta-agentからAcceptanceCriteria（AC）を生成中
- **RUNNING**: Worker実行ループ中（Meta.NextAction → Worker.RunWorker のループ）
- **VALIDATING**: 予約済み状態（現在未使用、将来的な完了評価フェーズ用）
- **COMPLETE**: タスク正常完了
- **FAILED**: エラーで失敗

### 主要な遷移ルール

#### 1. PENDING → PLANNING
- **トリガー**: Runner.Run() 呼び出し
- **処理内容**:
  - TaskContext初期化（ID、Title、RepoPath）
  - リポジトリパスを絶対パスに解決
  - PRD読み込み（テキストまたはファイルから）
  - TaskContextにPRDTextを記録

#### 2. PLANNING → RUNNING
- **トリガー**: Meta.PlanTask() 成功
- **処理内容**:
  - LLMがPRDからAcceptanceCriteriaを生成
  - 各AC（Acceptance Criterion）をTaskContext.AcceptanceCriteriaに記録
  - 初期状態: ACはすべて`Passed=false`
  - **エラー時**: PLANNING → FAILED（Meta APIエラー時）

#### 3. RUNNING ループ
- **実行フロー**:
  ```
  for i := 0; i < maxLoops; i++ {
    1. TaskSummary生成（現在の状態をMeta用に要約）
    2. Meta.NextAction() 呼び出し（次のアクション決定）
    3. Decision確認:
       - "run_worker": Worker.RunWorker() 実行、結果をWorkerRuns[]に記録
       - "mark_complete": ループ脱出、COMPLETE遷移
       - その他: FAILED遷移
  }
  ```

- **TaskSummary内容**:
  - Title: タスクタイトル
  - State: 現在のタスク状態
  - AcceptanceCriteria: AC一覧（各ACの`Passed`フラグ含む）
  - WorkerRunsCount: これまでのWorker実行回数

- **maxLoops = 10**: 無限ループ防止の安全弁

#### 4. RUNNING → FAILED
- **エラーケース**:
  - Meta.NextAction() API呼び出し失敗 → 即座にFAILED
  - Worker実行エラー → 結果に記録、Meta.NextAction()で再試行判定を任せる
  - Unknown action（"run_worker"と"mark_complete"以外） → 即座にFAILED

#### 5. RUNNING → COMPLETE
- **トリガー**: Meta.NextAction()から`decision.action == "mark_complete"`
- **処理内容**:
  - タスク完了判定
  - FinishedAtタイムスタンプ記録
  - Task Note出力

## TaskContext構造体

タスク全体の状態を単一の構造体で保持し、FSM遷移時に伝播させます。

```go
type TaskContext struct {
    ID                  string                      // タスク一意識別子
    Title               string                      // タスクタイトル
    RepoPath            string                      // リポジトリの絶対パス
    State               TaskState                   // 現在の状態
    PRDText             string                      // PRDテキスト全体
    AcceptanceCriteria  []AcceptanceCriterion       // AC一覧
    MetaCalls           []MetaCallLog               // Meta API呼び出し履歴
    WorkerRuns          []WorkerRunResult           // Worker実行履歴
    TestConfig          *config.TestDetails        // テスト設定
    TestResult          *TestResult                 // テスト実行結果
    StartedAt           time.Time                   // タスク開始時刻
    FinishedAt          time.Time                   // タスク終了時刻
}
```

### 主要フィールドの役割

- **AcceptanceCriteria**: Meta-agentが生成したAC。各ACは`Passed`フラグを持つ。
  - 初期状態: `Passed=false`
  - Meta-agentが更新（runner.go内で現在は明示的に更新していないが、MetaCallLogで記録）

- **WorkerRuns**: Worker.RunWorker() の実行結果を時系列で記録
  - ExitCode（0=成功、非0=失敗）
  - RawOutput（標準出力・標準エラー）
  - Error（実行時エラーがあれば記録）

- **MetaCalls**: Meta API（PlanTask, NextAction）の呼び出し履歴
  - Request YAML、Response YAMLを記録
  - 監査証跡とデバッグ用

## 依存性注入（Dependency Injection）パターン

Runner構造体は3つのインターフェースに依存し、モック化が容易です。

```go
type Runner struct {
    Config *config.TaskConfig
    Meta   MetaClient      // インターフェース
    Worker WorkerExecutor  // インターフェース
    Note   NoteWriter      // インターフェース
}
```

### インターフェース定義

```go
// MetaClient: Meta-agent通信
type MetaClient interface {
    PlanTask(ctx context.Context, prdText string) (*meta.PlanTaskResponse, error)
    NextAction(ctx context.Context, taskSummary *meta.TaskSummary) (*meta.NextActionResponse, error)
}

// WorkerExecutor: Worker CLI実行
type WorkerExecutor interface {
    RunWorker(ctx context.Context, prompt string, env map[string]string) (*WorkerRunResult, error)
}

// NoteWriter: Task Note出力
type NoteWriter interface {
    Write(taskCtx *TaskContext) error
}
```

### 設計上の利点

- **テスト容易性**: モック実装を差し込めば、外部依存なしでテスト可能
  - `internal/mock` パッケージにモック実装が存在
  - 統合テストで各パッケージを個別にモック化可能

- **独立性**: Runner は具体的な実装（MetaClient の OpenAI実装など）に依存しない
  - 将来、別のLLMプロバイダーに切り替え可能
  - Worker も Codex CLI以外のツールに変更可能

- **テスタビリティ**: プロパティベーステスト対応
  - `gopter` を使用した状態不変条件のテスト
  - FSM の遷移ルールが正しく実装されていることを検証

## エラーハンドリング戦略

### エラーの分類

#### 1. システムエラー（致命的）→ 即座にFAILED
- Meta API呼び出し失敗（PlanTask, NextAction）
- 不正なMeta レスポンス（YAMLパース失敗など）
- PRDファイル読み込み失敗

**対応**: `taskCtx.State = StateFailed` → 即座にReturn

#### 2. Worker実行エラー → 記録のみ、Meta判定に委譲
- Worker CLI実行失敗（ネットワークエラー、コマンド失敗など）

**対応**:
```go
if err != nil {
    res = &WorkerRunResult{
        Error:   err,
        Summary: "Worker execution failed: " + err.Error(),
    }
}
taskCtx.WorkerRuns = append(taskCtx.WorkerRuns, *res)
// ループ継続 → Meta.NextAction() で再試行を判定
```

**理由**: 一時的なWorkerエラーは自動再試行の余地があるため、Meta-agentに判定を委譲する方がロバスト

#### 3. Task Note出力エラー → 警告のみ、タスク失敗にしない
```go
if err := r.Note.Write(taskCtx); err != nil {
    fmt.Printf("Warning: failed to write task note: %v\n", err)
}
```

**理由**: Note出力は監査用途であり、コア機能の失敗ではない

## パフォーマンスと安全性

### maxLoops = 10
- Meta-agentが無限ループを引き起こす場合の防止弁
- 実際のタスクでは通常3〜5ループで完了
- 複雑なタスク場合は増加を検討（設定化推奨）

### TaskContext の完全性
- 全ての状態遷移、Meta呼び出し、Worker実行を TaskContext に記録
- Task Noteに全履歴が出力される
- 再実行やデバッグ時に完全な監査証跡を確保

## テスト戦略

### 単体テスト（runner_test.go）
- `gopter` を使用したプロパティベーステスト
- FSM の状態不変条件検証
  - 例: "PENDING状態から直接COMPLETE に遷移しない"
  - 例: "FAILED状態からの遷移は許可されない"

### 統合テスト
- `internal/mock` のモック実装を使用
- end-to-endフロー（成功・失敗シナリオ）
- `test/integration/run_flow_test.go` 参照

## 既知の制限事項

### 1. AC（AcceptanceCriteria）の Passed フラグ更新
- 現在、Meta.NextAction() レスポンスでAC状態が更新される仕様が未実装
- runner.go では AC の Passed フラグを明示的に更新していない
- 将来的に、Meta レスポンスに AC 更新情報を含める設計が必要

### 2. VALIDATING 状態の未使用
- FSM定義には VALIDATING 状態があるが、現在使用されていない
- 将来的に、完了評価フェーズを分離する場合に活用予定

## 関連ドキュメント

- [meta/CLAUDE.md](../meta/CLAUDE.md): Meta-agent通信の詳細
- [worker/CLAUDE.md](../worker/CLAUDE.md): Worker実行の詳細
- [../note/CLAUDE.md](../note/CLAUDE.md): Task Note生成ロジック（予定）
- `/docs/AgentRunner-architecture.md`: アーキテクチャ全体
- `/docs/agentrunner-spec-v1.md`: MVP仕様
