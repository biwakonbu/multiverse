# ロギング仕様書

最終更新: 2025-12-06

## 概要

Multiverse サービス全体で統一的なロギングを実現するための仕様書です。IDE、バックエンド、AI 処理フローを通じてデバッグとトレースを可能にします。

## 設計原則

### 1. Trace ID (相関 ID)

- 各タスク実行に一意の **Trace ID** (UUID) を付与
- IDE → Orchestrator → AgentRunner → Worker の全フローで同一 ID を伝播
- ログ検索・フィルタリングで処理フロー全体を追跡可能

### 2. 構造化ログ

- **Go バックエンド**: `log/slog` を使用
- **フロントエンド**: カスタム `Logger` クラスを使用
- JSON フォーマット（本番環境）/ Text フォーマット（開発環境）対応

### 3. ログレベル

| レベル  | 用途                                                      |
| ------- | --------------------------------------------------------- |
| `DEBUG` | 詳細なデバッグ情報（変数値、リクエスト/レスポンス全文等） |
| `INFO`  | 重要なイベント（タスク開始/終了、状態遷移等）             |
| `WARN`  | 警告（リトライ、軽微なエラー等）                          |
| `ERROR` | エラー（処理失敗、例外等）                                |

## Go バックエンド

### ロギングパッケージ

`internal/logging/logging.go`

```go
package logging

// Trace ID をコンテキストに設定
func ContextWithTraceID(ctx context.Context, traceID string) context.Context

// コンテキストから Trace ID を取得
func TraceIDFromContext(ctx context.Context) string

// 新しいロガーを作成
func NewLogger(cfg Config) *slog.Logger

// Trace ID 付きロガーを返す
func WithTraceID(logger *slog.Logger, ctx context.Context) *slog.Logger

// コンポーネント名付きロガーを返す
func WithComponent(logger *slog.Logger, component string) *slog.Logger
```

### 設定

```go
type Config struct {
    Level      slog.Level  // 最小ログレベル
    JSONFormat bool        // JSON 形式で出力
    AddSource  bool        // ソースファイル情報を追加
}

// プリセット設定
DefaultConfig()    // 開発用（INFO、Text）
ProductionConfig() // 本番用（INFO、JSON、ソース付き）
DebugConfig()      // デバッグ用（DEBUG、Text、ソース付き）
```

### 使用例

```go
import "github.com/biwakonbu/agent-runner/internal/logging"

// タスク実行開始時に Trace ID を生成
traceID := uuid.New().String()
ctx := logging.ContextWithTraceID(ctx, traceID)

// ロガーに Trace ID とコンポーネント名を付与
logger := logging.WithTraceID(slog.Default(), ctx)
logger = logging.WithComponent(logger, "runner")

// ログ出力
logger.Info("starting task execution",
    slog.String("task_id", taskID),
    slog.String("state", "PENDING"),
)
```

### 出力例

```
2025/12/06 00:48:28 INFO starting task execution component=runner trace_id=abc123 task_id=test-task state=PENDING
2025/12/06 00:48:28 INFO state transition component=runner trace_id=abc123 from=PENDING to=PLANNING
2025/12/06 00:48:28 INFO calling Meta.PlanTask component=runner trace_id=abc123
2025/12/06 00:48:28 INFO PlanTask completed component=runner trace_id=abc123 criteria_count=2 duration_ms=1234
```

## フロントエンド

### ロガークラス

`frontend/ide/src/services/logger.ts`

```typescript
type LogLevel = "debug" | "info" | "warn" | "error";

class Logger {
  static setLevel(level: LogLevel): void;
  static setTraceId(id: string | null): void;
  static withComponent(component: string): ComponentLogger;

  static debug(message: string, context?: Record<string, unknown>): void;
  static info(message: string, context?: Record<string, unknown>): void;
  static warn(message: string, context?: Record<string, unknown>): void;
  static error(message: string, context?: Record<string, unknown>): void;
}
```

### 使用例

```typescript
import { Logger } from './services/logger';

// コンポーネント別ロガーを作成
const log = Logger.withComponent('TaskCreate');

// ログ出力
log.info('creating task', { title: 'タスク名', poolId: 'default' });
log.debug('task details', { data: {...} });
log.error('task creation failed', { error: e });
```

### 出力例

```
[00:48:28.123] INFO  [TaskCreate] creating task { title: 'タスク名', poolId: 'default' }
[00:48:28.456] ERROR [TaskCreate] task creation failed { error: Error(...) }
```

## ログポイント

### Core Runner (`internal/core/runner.go`)

| ログポイント           | レベル | 内容                                                          |
| ---------------------- | ------ | ------------------------------------------------------------- |
| タスク開始             | INFO   | task_id, title, state                                         |
| 状態遷移               | INFO   | from, to                                                      |
| Meta.PlanTask 呼び出し | INFO   | -                                                             |
| PlanTask 完了          | INFO   | criteria_count, duration_ms                                   |
| Worker 実行開始        | INFO   | prompt_length                                                 |
| Worker 実行完了        | INFO   | exit_code, output_length, duration_ms                         |
| Worker 出力            | DEBUG  | output (全文)                                                 |
| タスク完了             | INFO   | final_state, worker_runs_count, meta_calls_count, duration_ms |

### Meta Client (`internal/meta/client.go`)

| ログポイント     | レベル | 内容                                       |
| ---------------- | ------ | ------------------------------------------ |
| LLM 呼び出し開始 | INFO   | model, request_size                        |
| リクエスト内容   | DEBUG  | system_prompt, user_prompt                 |
| リトライ         | WARN   | attempt, max_retries, delay_seconds, error |
| LLM 呼び出し完了 | INFO   | response_size, duration_ms                 |
| レスポンス内容   | DEBUG  | content (全文)                             |

### Worker Executor (`internal/worker/executor.go`)

| ログポイント        | レベル | 内容                                     |
| ------------------- | ------ | ---------------------------------------- |
| コンテナ起動開始    | INFO   | image, repo_path                         |
| コンテナ起動完了    | INFO   | container_id, duration_ms                |
| Worker コマンド実行 | INFO   | container_id, prompt_length, timeout_sec |
| Worker 実行完了     | INFO   | exit_code, output_length, duration_ms    |
| コンテナ停止        | INFO   | container_id, duration_ms                |

### IDE App (`app.go`)

| ログポイント           | レベル     | 内容                    |
| ---------------------- | ---------- | ----------------------- |
| アプリ起動             | INFO       | -                       |
| ワークスペース選択     | INFO       | path                    |
| ワークスペース読み込み | INFO       | id, workspace_dir       |
| タスク作成             | INFO       | title, pool_id, task_id |
| タスク実行開始         | INFO       | task_id, trace_id       |
| タスク実行完了/失敗    | INFO/ERROR | task_id                 |

## デバッグ手順

### 1. Trace ID でログを検索

```bash
# 特定の Trace ID のログを抽出
grep "trace_id=abc123" app.log
```

### 2. DEBUG レベルで詳細ログを出力

```go
// Go バックエンド
logger := logging.NewLogger(logging.DebugConfig())
slog.SetDefault(logger)
```

```typescript
// フロントエンド
Logger.setLevel("debug");
```

### 3. 問題の特定

1. エラーログから問題発生箇所を特定
2. Trace ID を取得
3. 同一 Trace ID のログを時系列で追跡
4. DEBUG レベルで詳細情報を確認
