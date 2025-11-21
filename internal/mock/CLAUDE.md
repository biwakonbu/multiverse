# Mock Package - テスト用モック実装

このパッケージは依存性注入パターンを実現し、外部システム（Meta-agent、Worker）に依存しないユニット・統合テストを可能にします。

## 概要

- **meta.go**: MetaClient のモック実装（OpenAI API呼び出しなし）
- **worker.go**: WorkerExecutor のモック実装（Docker実行なし）
- **note.go**: NoteWriter のモック実装（ファイルI/O なし）

## 設計パターン：Function Field Injection

このパッケージは「Function Field Injection」と呼ばれるパターンを採用しており、テスト時に関数を差し替えることで動的な検証が可能になります。

### パターンの全体像

```go
// Meta-agent モック実装
type MetaClient struct {
    PlanTaskFunc   func(...) (*PlanTaskResponse, error)
    NextActionFunc func(...) (*NextActionResponse, error)
}

func (m *MetaClient) PlanTask(...) (*PlanTaskResponse, error) {
    if m.PlanTaskFunc != nil {
        return m.PlanTaskFunc(...)
    }
    return nil, nil  // デフォルト（無実装）
}
```

**特徴**:
1. 構造体フィールドに **関数オブジェクト** を保持
2. インターフェースメソッドから関数呼び出し
3. テスト時に関数ロジックをカスタマイズ

### 利点

| 利点 | 説明 |
|------|------|
| **セットアップ簡潔** | `mock := &mock.MetaClient{ PlanTaskFunc: func(...) {...} }` |
| **シナリオ分岐** | 同一テスト内で成功・失敗を検証可能 |
| **トレース可能** | モック関数内で呼び出し回数・引数をログ記録 |
| **Nil-safe** | 関数が nil の場合、デフォルト動作（nil 返却）を実行 |

## 各モック実装の詳細

### 1. MetaClient モック（meta.go）

```go
type MetaClient struct {
    PlanTaskFunc   func(ctx context.Context, prdText string) (*meta.PlanTaskResponse, error)
    NextActionFunc func(ctx context.Context, taskSummary *meta.TaskSummary) (*meta.NextActionResponse, error)
}

func (m *MetaClient) PlanTask(ctx context.Context, prdText string) (*meta.PlanTaskResponse, error) {
    if m.PlanTaskFunc != nil {
        return m.PlanTaskFunc(ctx, prdText)
    }
    return nil, nil
}

func (m *MetaClient) NextAction(ctx context.Context, taskSummary *meta.TaskSummary) (*meta.NextActionResponse, error) {
    if m.NextActionFunc != nil {
        return m.NextActionFunc(ctx, taskSummary)
    }
    return nil, nil
}
```

**用途**: Meta-agent（LLM）の呼び出しをシミュレート

**テストケース例**:

```go
// ケース1: PlanTask 成功、NextAction で完了を指示
mockMeta := &mock.MetaClient{
    PlanTaskFunc: func(ctx context.Context, prdText string) (*meta.PlanTaskResponse, error) {
        return &meta.PlanTaskResponse{
            TaskID: "TASK-TEST",
            AcceptanceCriteria: []meta.AcceptanceCriterion{
                {ID: "AC-1", Description: "Feature X", Critical: true},
            },
        }, nil
    },
    NextActionFunc: func(ctx context.Context, taskSummary *meta.TaskSummary) (*meta.NextActionResponse, error) {
        return &meta.NextActionResponse{
            Decision: meta.Decision{Action: "mark_complete"},
        }, nil
    },
}

// Runner テスト
runner := core.NewRunner(cfg, mockMeta, mockWorker, mockNote)
taskCtx, err := runner.Run(context.Background())
// 検証: taskCtx.State == core.StateComplete
```

```go
// ケース2: PlanTask 失敗（API エラー）
mockMeta := &mock.MetaClient{
    PlanTaskFunc: func(ctx context.Context, prdText string) (*meta.PlanTaskResponse, error) {
        return nil, fmt.Errorf("API timeout")
    },
}

// 検証: Runner が StateFailed に遷移
```

### 2. WorkerExecutor モック（worker.go）

```go
type WorkerExecutor struct {
    RunWorkerFunc func(ctx context.Context, prompt string, env map[string]string) (*core.WorkerRunResult, error)
}

func (w *WorkerExecutor) RunWorker(ctx context.Context, prompt string, env map[string]string) (*core.WorkerRunResult, error) {
    if w.RunWorkerFunc != nil {
        return w.RunWorkerFunc(ctx, prompt, env)
    }
    return nil, nil
}
```

**用途**: Worker CLI（Docker実行）をシミュレート

**テストケース例**:

```go
// ケース1: Worker 実行成功
mockWorker := &mock.WorkerExecutor{
    RunWorkerFunc: func(ctx context.Context, prompt string, env map[string]string) (*core.WorkerRunResult, error) {
        return &core.WorkerRunResult{
            ID:        "run-001",
            ExitCode:  0,
            RawOutput: "✓ Feature X implemented successfully",
            Summary:   "Implementation complete",
        }, nil
    },
}
```

```go
// ケース2: Worker 実行失敗（Docker エラー）
mockWorker := &mock.WorkerExecutor{
    RunWorkerFunc: func(ctx context.Context, prompt string, env map[string]string) (*core.WorkerRunResult, error) {
        return &core.WorkerRunResult{
            ID:      "run-001",
            ExitCode: 1,
            Error:   fmt.Errorf("Docker container crashed"),
        }, nil  // エラーを結果に記録、Runner は continue
    },
}
```

```go
// ケース3: 複数回 Worker 実行（ループシミュレーション）
runCount := 0
mockWorker := &mock.WorkerExecutor{
    RunWorkerFunc: func(ctx context.Context, prompt string, env map[string]string) (*core.WorkerRunResult, error) {
        runCount++
        return &core.WorkerRunResult{
            ID:       fmt.Sprintf("run-%d", runCount),
            ExitCode: 0,
        }, nil
    },
}

// 検証: runCount == 期待値
```

### 3. NoteWriter モック（note.go）

```go
type NoteWriter struct {
    WriteFunc func(taskCtx *core.TaskContext) error
}

func (n *NoteWriter) Write(taskCtx *core.TaskContext) error {
    if n.WriteFunc != nil {
        return n.WriteFunc(taskCtx)
    }
    return nil
}
```

**用途**: Task Note 出力（ファイルI/O）をシミュレート

**テストケース例**:

```go
// ケース1: Note 出力成功
mockNote := &mock.NoteWriter{
    WriteFunc: func(taskCtx *core.TaskContext) error {
        // TaskContext を検証（値が正しく伝播しているか）
        if taskCtx.ID != "TASK-001" {
            return fmt.Errorf("invalid task ID")
        }
        return nil  // 成功
    },
}
```

```go
// ケース2: Note 出力失敗（ディスク容量不足）
mockNote := &mock.NoteWriter{
    WriteFunc: func(taskCtx *core.TaskContext) error {
        return fmt.Errorf("disk full")
    },
}

// 検証: Runner はログのみ出力し、COMPLETE 状態に遷移
```

## テスト設計パターン

### パターン1: 単体テスト（Unit Test）

**対象**: Runner の FSM 遷移ロジック

**モック戦略**: 全モック使用

```go
func TestRunner_PLANNING_to_RUNNING_Transition(t *testing.T) {
    cfg := &config.TaskConfig{
        Task: config.TaskDetails{
            ID: "TASK-TEST",
            PRD: config.PRDDetails{
                Text: "Implement feature X",
            },
        },
    }

    // モック：PlanTask は AC を返す
    mockMeta := &mock.MetaClient{
        PlanTaskFunc: func(ctx context.Context, prdText string) (*meta.PlanTaskResponse, error) {
            return &meta.PlanTaskResponse{
                AcceptanceCriteria: []meta.AcceptanceCriterion{
                    {ID: "AC-1", Description: "Feature X", Critical: true},
                },
            }, nil
        },
    }

    mockWorker := &mock.WorkerExecutor{}
    mockNote := &mock.NoteWriter{}

    runner := core.NewRunner(cfg, mockMeta, mockWorker, mockNote)
    taskCtx, _ := runner.Run(context.Background())

    // 検証
    if len(taskCtx.AcceptanceCriteria) != 1 {
        t.Errorf("expected 1 AC, got %d", len(taskCtx.AcceptanceCriteria))
    }
}
```

### パターン2: 統合テスト（Integration Test）

**対象**: end-to-end フロー（PENDING → PLANNING → RUNNING → COMPLETE）

**モック戦略**: 全モック、リアルな動作をシミュレート

```go
func TestRunFlow_Success(t *testing.T) {
    // 設定
    cfg := &config.TaskConfig{
        Task: config.TaskDetails{
            ID:   "TASK-001",
            Repo: ".",
            PRD:  config.PRDDetails{Text: "Build API endpoint"},
        },
    }

    // Meta モック：AC を生成 → Worker実行指示 → 完了指示
    callCount := 0
    mockMeta := &mock.MetaClient{
        PlanTaskFunc: func(ctx context.Context, prdText string) (*meta.PlanTaskResponse, error) {
            return &meta.PlanTaskResponse{
                AcceptanceCriteria: []meta.AcceptanceCriterion{
                    {ID: "AC-1", Description: "Implement POST /api/users", Critical: true},
                },
            }, nil
        },
        NextActionFunc: func(ctx context.Context, taskSummary *meta.TaskSummary) (*meta.NextActionResponse, error) {
            callCount++
            if callCount == 1 {
                // 初回：Worker 実行指示
                return &meta.NextActionResponse{
                    Decision: meta.Decision{Action: "run_worker"},
                    WorkerCall: meta.WorkerCall{
                        Prompt: "Implement POST /api/users endpoint",
                    },
                }, nil
            }
            // 2回目：完了指示
            return &meta.NextActionResponse{
                Decision: meta.Decision{Action: "mark_complete"},
            }, nil
        },
    }

    // Worker モック：実行結果を返す
    mockWorker := &mock.WorkerExecutor{
        RunWorkerFunc: func(ctx context.Context, prompt string, env map[string]string) (*core.WorkerRunResult, error) {
            return &core.WorkerRunResult{
                ID:        "run-1",
                ExitCode:  0,
                RawOutput: "✓ API endpoint implemented",
            }, nil
        },
    }

    mockNote := &mock.NoteWriter{}

    // 実行
    runner := core.NewRunner(cfg, mockMeta, mockWorker, mockNote)
    taskCtx, err := runner.Run(context.Background())

    // 検証
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if taskCtx.State != core.StateComplete {
        t.Errorf("expected StateComplete, got %s", taskCtx.State)
    }
    if len(taskCtx.WorkerRuns) != 1 {
        t.Errorf("expected 1 Worker run, got %d", len(taskCtx.WorkerRuns))
    }
}
```

### パターン3: エラーハンドリングテスト

**対象**: 各レイヤーでのエラー処理

```go
func TestRunner_MetaError_BecomesStateFailed(t *testing.T) {
    cfg := &config.TaskConfig{
        Task: config.TaskDetails{
            ID:  "TASK-001",
            PRD: config.PRDDetails{Text: "..."},
        },
    }

    // Meta が エラーを返す
    mockMeta := &mock.MetaClient{
        PlanTaskFunc: func(ctx context.Context, prdText string) (*meta.PlanTaskResponse, error) {
            return nil, fmt.Errorf("OpenAI API timeout")
        },
    }

    mockWorker := &mock.WorkerExecutor{}
    mockNote := &mock.NoteWriter{}

    runner := core.NewRunner(cfg, mockMeta, mockWorker, mockNote)
    taskCtx, err := runner.Run(context.Background())

    // 検証
    if taskCtx.State != core.StateFailed {
        t.Errorf("expected StateFailed, got %s", taskCtx.State)
    }
    if err == nil {
        t.Errorf("expected error, got nil")
    }
}
```

## モック拡張ガイド

### 新しいインターフェース作成時の手順

**例: ValidationClient インターフェースを追加する場合**

1. **internal/core/runner.go** に インターフェース定義
   ```go
   type ValidationClient interface {
       Validate(ctx context.Context, taskCtx *TaskContext) error
   }
   ```

2. **Runner** に フィールド追加
   ```go
   type Runner struct {
       ...
       Validation ValidationClient  // 新規
   }
   ```

3. **internal/mock/validation.go** に モック実装
   ```go
   package mock

   type ValidationClient struct {
       ValidateFunc func(ctx context.Context, taskCtx *core.TaskContext) error
   }

   func (v *ValidationClient) Validate(ctx context.Context, taskCtx *core.TaskContext) error {
       if v.ValidateFunc != nil {
           return v.ValidateFunc(ctx, taskCtx)
       }
       return nil
   }
   ```

4. **テスト** で 使用
   ```go
   mockValidation := &mock.ValidationClient{
       ValidateFunc: func(ctx context.Context, taskCtx *core.TaskContext) error {
           if taskCtx.State != core.StateComplete {
               return fmt.Errorf("validation failed: task not complete")
           }
           return nil
       },
   }
   ```

## テスト実行方法

```bash
# 全テスト実行
go test ./...

# モック関連テストのみ実行
go test ./internal/mock

# 統合テスト実行
go test ./test/integration/...

# カバレッジ確認
go test -cover ./...

# 詳細ログ付き実行
go test -v ./test/integration/...
```

## ベストプラクティス

### 1. モック関数の詳細なカスタマイズ

```go
// Good: 複雑なロジックをモック内に実装
mockMeta.PlanTaskFunc = func(ctx context.Context, prdText string) (*meta.PlanTaskResponse, error) {
    // PRD テキストの内容に応じて異なるレスポンスを返す
    if strings.Contains(prdText, "urgent") {
        return &meta.PlanTaskResponse{
            AcceptanceCriteria: []meta.AcceptanceCriterion{
                {ID: "AC-1", Critical: true},
            },
        }, nil
    }
    return nil, fmt.Errorf("PRD incomplete")
}
```

### 2. 呼び出し履歴の記録

```go
// モック内で呼び出し回数・引数を記録
type MockMeta struct {
    PlanTaskFunc func(...) (...)
    PlanTaskCalls int
    PlanTaskArgs []string
}

func (m *MockMeta) PlanTask(ctx context.Context, prdText string) (*meta.PlanTaskResponse, error) {
    m.PlanTaskCalls++
    m.PlanTaskArgs = append(m.PlanTaskArgs, prdText)
    if m.PlanTaskFunc != nil {
        return m.PlanTaskFunc(ctx, prdText)
    }
    return nil, nil
}

// テスト検証
if mockMeta.PlanTaskCalls != 1 {
    t.Errorf("expected 1 call, got %d", mockMeta.PlanTaskCalls)
}
```

### 3. Nil-safety の活用

```go
// 関数が nil の場合のデフォルト動作を利用
mockMeta := &mock.MetaClient{}
// mockMeta.PlanTaskFunc = nil（デフォルト）
// 呼び出し時は nil, nil を返す
```

## 既知の制限事項

### 1. Context タイムアウトのシミュレーション

**現在**: モック関数内で `context.WithTimeout` を明示的に処理

**改善案**: Helper 関数で Context タイムアウト シミュレーション提供

```go
func WithTimeout(fn func(context.Context) error, timeout time.Duration) error {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    return fn(ctx)
}
```

### 2. Concurrency テスト

**現在**: シングルスレード テストのみ

**将来対応**: Goroutine セーフなモック（sync.Mutex 使用）

## 関連ドキュメント

- [core/CLAUDE.md](../core/CLAUDE.md) - Runner FSM と インターフェース定義
- [meta/CLAUDE.md](../meta/CLAUDE.md) - Meta クライアント インターフェース
- [worker/CLAUDE.md](../worker/CLAUDE.md) - Worker インターフェース
- [note/CLAUDE.md](../note/CLAUDE.md) - Note Writer インターフェース
- `/TESTING.md` - テスト全体のベストプラクティス
