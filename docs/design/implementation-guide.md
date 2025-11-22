# AgentRunner 実装ガイド

最終更新: 2025-11-22

## 概要

本ドキュメントは AgentRunner の Go 実装に関するガイドです。パッケージ構成、依存性注入パターン、実装パターン、テスト戦略を説明します。

## パッケージ構成

### ディレクトリ構造

```
agent-runner/
├── cmd/
│   └── agent-runner/
│       └── main.go              # CLI エントリポイント
├── internal/
│   ├── core/                    # タスク実行エンジン
│   │   ├── runner.go            # FSM オーケストレーション
│   │   ├── context.go           # TaskContext・TaskState定義
│   │   └── runner_test.go       # プロパティベーステスト
│   ├── meta/                    # Meta-agent通信層
│   │   ├── client.go            # OpenAI API通信
│   │   └── protocol.go          # YAMLプロトコル定義
│   ├── worker/                  # Worker実行・Dockerサンドボックス
│   │   ├── executor.go          # Worker CLI実行の抽象化
│   │   └── sandbox.go           # Docker API管理
│   ├── note/                    # Task Note生成
│   │   └── writer.go            # Markdown テンプレート出力
│   └── mock/                    # テスト用モック実装
│       ├── meta.go
│       ├── worker.go
│       └── note.go
├── pkg/
│   └── config/                  # 公開パッケージ（YAML設定）
│       └── config.go            # TaskConfig構造体定義
└── test/
    ├── integration/             # Mock統合テスト
    ├── sandbox/                 # Docker Sandboxテスト
    └── codex/                   # Codex統合テスト
```

### パッケージの役割

| パッケージ         | 責務                            | 公開/内部 |
| ------------------ | ------------------------------- | --------- |
| `cmd/agent-runner` | CLI エントリポイント            | 実行可能  |
| `internal/core`    | FSM・TaskContext・状態遷移      | 内部      |
| `internal/meta`    | LLM 通信・YAML プロトコル       | 内部      |
| `internal/worker`  | CLI 実行・Docker サンドボックス | 内部      |
| `internal/note`    | Task Note 生成・テンプレート    | 内部      |
| `internal/mock`    | テストダブル・FuncField 注入    | 内部      |
| `pkg/config`       | YAML 設定スキーマ               | 公開      |

## 依存性注入パターン

### Runner 構造体

`Runner` はインターフェースを受け入れることで、テスト時にモックを注入できます。

```go
type Runner struct {
    Config *config.TaskConfig
    Meta   MetaClient        // インターフェース
    Worker WorkerExecutor    // インターフェース
    Note   NoteWriter        // インターフェース
}
```

### インターフェース定義

#### MetaClient

```go
type MetaClient interface {
    PlanTask(ctx context.Context, tc *TaskContext) (*PlanTaskResponse, error)
    NextAction(ctx context.Context, tc *TaskContext) (*NextActionResponse, error)
    CompletionAssessment(ctx context.Context, tc *TaskContext) (*CompletionAssessmentResponse, error)
}
```

#### WorkerExecutor

```go
type WorkerExecutor interface {
    Start(ctx context.Context) error
    RunWorker(ctx context.Context, prompt string) (*WorkerRunResult, error)
    Stop(ctx context.Context) error
}
```

#### NoteWriter

```go
type NoteWriter interface {
    WriteTaskNote(ctx context.Context, tc *TaskContext) error
}
```

### モック実装

`internal/mock` パッケージで Function Field Injection パターンを使用します。

```go
type MockMetaClient struct {
    PlanTaskFunc             func(ctx context.Context, tc *TaskContext) (*PlanTaskResponse, error)
    NextActionFunc           func(ctx context.Context, tc *TaskContext) (*NextActionResponse, error)
    CompletionAssessmentFunc func(ctx context.Context, tc *TaskContext) (*CompletionAssessmentResponse, error)
}

func (m *MockMetaClient) PlanTask(ctx context.Context, tc *TaskContext) (*PlanTaskResponse, error) {
    if m.PlanTaskFunc != nil {
        return m.PlanTaskFunc(ctx, tc)
    }
    return nil, errors.New("PlanTaskFunc not set")
}
```

## 実装パターン

### 1. TaskContext の伝播

実行状態はすべて `TaskContext` に保持し、FSM を通じて伝播します。

```go
type TaskContext struct {
    ID        string
    Title     string
    RepoPath  string
    State     TaskState

    PRDText            string
    AcceptanceCriteria []AcceptanceCriterion
    MetaCalls          []MetaCallLog
    WorkerRuns         []WorkerRunResult

    StartedAt  time.Time
    FinishedAt time.Time
}
```

### 2. FSM 状態遷移

```go
func (r *Runner) Run(ctx context.Context) error {
    tc := r.buildTaskContext()

    for tc.State != StateComplete && tc.State != StateFailed {
        switch tc.State {
        case StatePending:
            tc.State = StatePlanning
        case StatePlanning:
            if err := r.planTask(ctx, tc); err != nil {
                tc.State = StateFailed
                return err
            }
            tc.State = StateRunning
        case StateRunning:
            if err := r.runWorker(ctx, tc); err != nil {
                tc.State = StateFailed
                return err
            }
            tc.State = StateValidating
        case StateValidating:
            if err := r.validate(ctx, tc); err != nil {
                tc.State = StateFailed
                return err
            }
            // Meta の判断で RUNNING or COMPLETE に遷移
        }
    }

    return r.Note.WriteTaskNote(ctx, tc)
}
```

### 3. エラーハンドリング

#### Exponential Backoff

```go
func (c *Client) callWithRetry(ctx context.Context, fn func() error) error {
    backoff := 1 * time.Second
    maxRetries := 3

    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }

        if !isRetryable(err) {
            return err
        }

        time.Sleep(backoff)
        backoff *= 2
    }

    return errors.New("max retries exceeded")
}
```

#### コンテキストキャンセル

```go
func (w *WorkerExecutor) RunWorker(ctx context.Context, prompt string) (*WorkerRunResult, error) {
    timeout := time.Duration(w.config.MaxRunTimeSec) * time.Second
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()

    // Worker 実行
    result, err := w.sandbox.Exec(ctx, w.containerID, cmd)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return nil, errors.New("worker execution timeout")
        }
        return nil, err
    }

    return result, nil
}
```

### 4. YAML パース

```go
func parseYAML[T any](data []byte) (*T, error) {
    var result T
    if err := yaml.Unmarshal(data, &result); err != nil {
        return nil, fmt.Errorf("failed to parse YAML: %w", err)
    }
    return &result, nil
}
```

### 5. テンプレート生成

```go
func (w *Writer) WriteTaskNote(ctx context.Context, tc *TaskContext) error {
    tmpl, err := template.New("task_note").Parse(taskNoteTemplate)
    if err != nil {
        return err
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, tc); err != nil {
        return err
    }

    path := filepath.Join(tc.RepoPath, ".agent-runner", fmt.Sprintf("task-%s.md", tc.ID))
    return os.WriteFile(path, buf.Bytes(), 0644)
}
```

## テスト戦略

### テストの種類

| テスト種別                 | ツール    | タグ           | 実行時間 |
| -------------------------- | --------- | -------------- | -------- |
| **ユニットテスト**         | `go test` | なし           | < 1s     |
| **プロパティベーステスト** | `gopter`  | なし           | 1-5s     |
| **Mock 統合テスト**        | `go test` | なし           | 1-10s    |
| **Docker Sandbox テスト**  | `go test` | `-tags=docker` | 10-30s   |
| **Codex 統合テスト**       | `go test` | `-tags=codex`  | 60-300s  |

### テスト実行コマンド

```bash
# ユニットテスト（依存なし、高速）
go test ./...

# 全テスト実行（推奨、Docker + Codex CLI 必須）
go test -tags=docker,codex -timeout=15m ./...

# カバレッジレポート生成
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

### プロパティベーステスト

```go
func TestRunnerProperties(t *testing.T) {
    properties := gopter.NewProperties(nil)

    properties.Property("TaskContext は常に有効な状態遷移を行う", prop.ForAll(
        func(initialState TaskState) bool {
            tc := &TaskContext{State: initialState}
            nextState := transition(tc)
            return isValidTransition(initialState, nextState)
        },
        gen.OneConstOf(StatePending, StatePlanning, StateRunning, StateValidating),
    ))

    properties.TestingRun(t)
}
```

### モックを使った統合テスト

```go
func TestRunnerIntegration(t *testing.T) {
    mockMeta := &mock.MockMetaClient{
        PlanTaskFunc: func(ctx context.Context, tc *TaskContext) (*PlanTaskResponse, error) {
            return &PlanTaskResponse{
                AcceptanceCriteria: []AcceptanceCriterion{
                    {ID: "AC-1", Description: "Test criterion"},
                },
            }, nil
        },
        NextActionFunc: func(ctx context.Context, tc *TaskContext) (*NextActionResponse, error) {
            return &NextActionResponse{
                Decision: Decision{Action: "mark_complete"},
            }, nil
        },
    }

    runner := &Runner{
        Meta: mockMeta,
        // ...
    }

    err := runner.Run(context.Background())
    assert.NoError(t, err)
}
```

## ベストプラクティス

### 1. ロギング

構造化ログ（`log/slog`）を使用します。

```go
slog.Info("starting worker execution",
    "task_id", tc.ID,
    "worker_type", "codex-cli",
)
```

### 2. エラーメッセージ

コンテキスト情報を含めます。

```go
return fmt.Errorf("failed to start container for task %s: %w", tc.ID, err)
```

### 3. リソース管理

`defer` でリソースを確実に解放します。

```go
func (w *WorkerExecutor) Start(ctx context.Context) error {
    containerID, err := w.sandbox.StartContainer(ctx, ...)
    if err != nil {
        return err
    }
    w.containerID = containerID

    // 確実に停止するように defer で登録
    runtime.SetFinalizer(w, func(w *WorkerExecutor) {
        w.Stop(context.Background())
    })

    return nil
}
```

### 4. コンテキスト伝播

すべての I/O 操作でコンテキストを受け取ります。

```go
func (c *Client) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
    // コンテキストを HTTP リクエストに渡す
    httpReq, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, body)
    // ...
}
```

## 実装状況

### 実装済み機能

- ✅ CLI エントリポイント
- ✅ FSM オーケストレーション
- ✅ Meta-agent 通信（OpenAI API）
- ✅ Worker 実行（Codex CLI）
- ✅ Docker サンドボックス管理
- ✅ Task Note 生成
- ✅ 依存性注入とモック
- ✅ プロパティベーステスト
- ✅ Exponential Backoff
- ✅ コンテナライフサイクル最適化

### 今後の拡張

- [ ] 複数 Worker サポート
- [ ] TaskContext の永続化（DB）
- [ ] Web UI
- [ ] メトリクス収集

## 参考ドキュメント

- [アーキテクチャ](architecture.md)
- [データフロー](data-flow.md)
- [テストガイド](../guides/testing.md)
- [コア仕様](../specifications/core-specification.md)
