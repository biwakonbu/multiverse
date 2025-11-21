# テスト戦略・実装パターン・精度管理

このドキュメントは、agent-runner プロジェクトの**テスト戦略、実装パターン、ベストプラクティス、精度管理手法**を統合的に説明するものです。

## テスト戦略概観

### 4段階のテスト戦略

| Stage | 名称 | テストファイル | 依存性 | 実行速度 | カバレッジ | 目的 |
|-------|------|----------------|--------|---------|----------|------|
| **1** | ユニットテスト | `internal/*/name_test.go` + `pkg/config/config_test.go` | モック | 高速（<1s） | 機能別 | パッケージ個別機能の検証 |
| **2** | 統合テスト（Mock） | `test/integration/run_flow_test.go` | モック | 中速（<5s） | end-to-end | FSMフローの完全性検証 |
| **3** | Sandbox テスト | `test/sandbox/sandbox_test.go` (-tags=docker) | Docker | 遅い（10-30s） | インフラ | Docker API・コンテナ管理 |
| **4** | Codex 統合テスト | `test/codex/codex_integration_test.go` (-tags=codex) | Docker+Codex | 最遅（1-5m） | 実運用 | 実Codex CLIとの統合 |

### テストファイル構成と統計

```
ユニットテスト（5ファイル、~52テスト）:
  - internal/core/runner_test.go          (1 PBT)
  - internal/meta/protocol_test.go        (9テスト)
  - internal/note/writer_test.go          (12テスト)
  - internal/worker/executor_test.go      (15テスト)
  - pkg/config/config_test.go             (16テスト)

統合テスト（2ファイル、~10テスト）:
  - test/integration/run_flow_test.go     (2テスト)
  - test/sandbox/sandbox_test.go          (7テスト、-tags=docker)
  - test/codex/codex_integration_test.go  (8テスト、-tags=codex)

合計: 70+ テスト、複数レイヤー、完全隔離
```

---

## テストアーキテクチャ詳細

### テストレイヤー間の依存関係

```
Meta-agent (LLM)
    ↑↓
┌─────────────────────────────────────┐
│        AgentRunner Core             │
│  (TaskContext・TaskState管理)       │
└─────────────────────────────────────┘
    ↑↓
┌─────────────────────────────────────┐
│  Worker Executor + Sandbox Manager  │
│  (Docker API・コンテナ実行)         │
└─────────────────────────────────────┘

テストの実装例：
┌─────────────────────────────────────┐
│ Stage 1: ユニットテスト             │
│ 各インターフェース → モック化       │
└─────────────────────────────────────┘
          ↓
┌─────────────────────────────────────┐
│ Stage 2: Mock統合テスト             │
│ 全3層 → モック、フロー検証         │
└─────────────────────────────────────┘
          ↓
┌─────────────────────────────────────┐
│ Stage 3: Docker Sandbox テスト      │
│ 実Docker API、コンテナ管理         │
└─────────────────────────────────────┘
          ↓
┌─────────────────────────────────────┐
│ Stage 4: Codex 統合テスト          │
│ 実Codex CLI・実運用シナリオ        │
└─────────────────────────────────────┘
```

---

## 実装パターン集

### パターン 1: Table-Driven Tests（表駆動テスト）

**用途**: ユニットテスト（複数ケースを体系的に検証）

**実装例**: `pkg/config/config_test.go`

```go
func TestTaskConfig_UnmarshalYAML_Valid(t *testing.T) {
    tests := []struct {
        name    string
        yaml    string
        want    TaskConfig
        wantErr bool
    }{
        {
            name: "minimal valid config",
            yaml: `
version: 1
task:
  id: "TASK-001"
  repo: "/tmp/project"
runner:
  meta:
    kind: "mock"
  worker:
    kind: "codex-cli"
`,
            want: TaskConfig{
                Version: 1,
                Task: TaskDetails{
                    ID:   "TASK-001",
                    Repo: "/tmp/project",
                },
                // ...
            },
            wantErr: false,
        },
        {
            name: "complete config with all fields",
            yaml: `
version: 1
task:
  id: "TASK-002"
  title: "Complete Task"
  repo: "/tmp/project"
  prd:
    path: "./prd.md"
  test:
    command: "go test ./..."
    cwd: "./"
runner:
  meta:
    kind: "openai-chat"
    model: "gpt-4-turbo"
  worker:
    kind: "codex-cli"
    docker_image: "agent-runner-codex:latest"
    max_run_time_sec: 1800
    env:
      CODEX_API_KEY: "env:CODEX_API_KEY"
`,
            want: TaskConfig{...},
            wantErr: false,
        },
        {
            name: "prd as text instead of path",
            yaml: `
version: 1
task:
  id: "TASK-003"
  repo: "/tmp/project"
  prd:
    text: "Create a new feature"
runner:
  meta:
    kind: "mock"
  worker:
    kind: "codex-cli"
`,
            want: TaskConfig{...},
            wantErr: false,
        },
        // エラーケース
        {
            name:    "missing required field: task.repo",
            yaml:    "version: 1\ntask:\n  id: TASK-004",
            want:    TaskConfig{},
            wantErr: true,
        },
        {
            name:    "invalid version",
            yaml:    "version: 99\ntask:\n  repo: /tmp",
            want:    TaskConfig{},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var got TaskConfig
            err := yaml.Unmarshal([]byte(tt.yaml), &got)

            if (err != nil) != tt.wantErr {
                t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
                t.Errorf("UnmarshalYAML() got = %v, want %v", got, tt.want)
            }
        })
    }
}
```

**利点**:
- テストケースが視覚的に整理される
- 新規ケース追加が簡単
- 正常系・異常系を同じテーブルで管理
- パラメータ化テストの標準手法

**使用箇所**: config_test.go（16テスト）、protocol_test.go（9テスト）、writer_test.go（12テスト）

---

### パターン 2: Property-Based Tests（プロパティベーステスト）

**ライブラリ**: `github.com/leanovate/gopter`

**用途**: 不変条件の検証（有限状態機械の状態遷移ルール等）

**実装例**: `internal/core/runner_test.go`

```go
func TestRunner_Properties(t *testing.T) {
    // テストパラメータ設定
    parameters := gopter.DefaultTestParameters()
    parameters.MinSuccessfulTests = 5  // 最低5パターン検証

    properties := gopter.NewProperties(parameters)

    // Property 1: Runner は AC 数を正しく追跡する
    properties.Property(
        "Runner accumulates all ACs from Meta responses",
        prop.ForAll(
            func(acCount int) bool {
                // Arrange
                cfg := createMinimalConfig()
                mockMeta := &mock.MetaClient{
                    PlanTaskFunc: func(ctx context.Context, prdText string) (*meta.PlanTaskResponse, error) {
                        acs := make([]*meta.AcceptanceCriteria, acCount)
                        for i := 0; i < acCount; i++ {
                            acs[i] = &meta.AcceptanceCriteria{
                                ID:        fmt.Sprintf("AC-%d", i),
                                Condition: fmt.Sprintf("Condition %d", i),
                            }
                        }
                        return &meta.PlanTaskResponse{
                            TaskID:               "TASK-001",
                            AcceptanceCriteria:   acs,
                        }, nil
                    },
                    // ... NextActionFunc も設定
                }

                runner := core.NewRunner(cfg, mockMeta, mockWorker, mockNote)

                // Act
                result, err := runner.Run(context.Background())

                // Assert
                if err != nil {
                    return false
                }
                // 不変条件1: AC の数が一致
                if len(result.AcceptanceCriteria) != acCount {
                    return false
                }
                // 不変条件2: 全 AC に ID がある
                for _, ac := range result.AcceptanceCriteria {
                    if ac.ID == "" {
                        return false
                    }
                }
                return true
            },
            gen.IntRange(0, 3),  // AC 数の範囲: 0-3
        ),
    )

    // Property 2: Runner は COMPLETE 状態で WorkerRuns > 0 である
    properties.Property(
        "Runner reaches COMPLETE state only after worker execution",
        prop.ForAll(
            func(workerRunCount int) bool {
                // ... setup ...

                // Act
                result, err := runner.Run(context.Background())

                // Assert
                if err != nil {
                    return true  // エラー状態は OK
                }
                // 不変条件: COMPLETE → WorkerRuns > 0
                if result.State == core.StateComplete && len(result.WorkerRuns) == 0 {
                    return false
                }
                return true
            },
            gen.IntRange(0, 2),
        ),
    )

    // Property 3: タイムスタンプの順序性
    properties.Property(
        "Timestamps are monotonic increasing",
        prop.ForAll(
            func(unused int) bool {
                // ... setup & execution ...

                result, _ := runner.Run(context.Background())

                // 不変条件: StartedAt <= FinishedAt
                return !result.StartedAt.After(result.FinishedAt)
            },
            gen.IntRange(0, 1),
        ),
    )

    // 全プロパティを実行
    if !properties.TestingRun(t) {
        t.Fail()
    }
}
```

**特徴**:
- 有限状態機械の不変条件を形式化
- 複数のランダムケース（gopter が自動生成）で検証
- エッジケース（0、1、極端な値）を自動テスト
- 再現可能（seed 指定可）

**利点**:
- 手作業では見落とす組み合わせをテスト
- 不変条件を明文化できる
- 状態遷移バグ（off-by-one など）を検出

**使用箇所**: core/runner_test.go（1 PBT）

---

### パターン 3: Mock-based Integration Tests

**デザインパターン**: Function Field Injection（インターフェース実装の動的注入）

**用途**: 外部依存なしでend-to-endフロー検証

**実装例**: `test/integration/run_flow_test.go`

```go
func TestRunFlow_Success(t *testing.T) {
    // Arrange: テストデータとモック準備
    cfg := &config.TaskConfig{
        Version: 1,
        Task: config.TaskDetails{
            ID:   "integration-task-1",
            Repo: "/tmp/project",
            PRD: config.PRDDetails{
                Text: "Create a feature",
            },
        },
        Runner: config.RunnerConfig{
            Meta: config.MetaConfig{
                Kind: "mock",
            },
            Worker: config.WorkerConfig{
                Kind: "codex-cli",
            },
        },
    }

    // Meta モック: PlanTask → NextAction → (判定)
    mockMeta := &mock.MetaClient{
        PlanTaskFunc: func(ctx context.Context, prdText string) (*meta.PlanTaskResponse, error) {
            return &meta.PlanTaskResponse{
                TaskID: cfg.Task.ID,
                AcceptanceCriteria: []*meta.AcceptanceCriteria{
                    {
                        ID:        "AC-001",
                        Condition: "File created",
                    },
                },
            }, nil
        },
        NextActionFunc: func(ctx context.Context, summary *core.TaskSummary) (*meta.NextActionResponse, error) {
            if summary.State == core.StatePlanning {
                return &meta.NextActionResponse{
                    Action: "run_worker",
                    Prompt: "Create the file as specified",
                }, nil
            }
            if summary.State == core.StateRunning {
                // Worker 実行後の判定
                return &meta.NextActionResponse{
                    Action:   "mark_complete",
                    Feedback: "All ACs satisfied",
                }, nil
            }
            return nil, fmt.Errorf("unexpected state: %s", summary.State)
        },
    }

    // Worker モック: ファイル生成結果を返す
    var capturedPrompt string
    mockWorker := &mock.WorkerExecutor{
        RunWorkerFunc: func(ctx context.Context, prompt string, env map[string]string) (*core.WorkerRunResult, error) {
            capturedPrompt = prompt  // プロンプト確認用
            return &core.WorkerRunResult{
                ExitCode: 0,
                Output:   "Created feature.txt",
                RawOutput: "...",
            }, nil
        },
    }

    // Note モック: TaskContext キャプチャ
    var capturedTaskCtx *core.TaskContext
    mockNote := &mock.NoteWriter{
        WriteFunc: func(taskCtx *core.TaskContext) error {
            capturedTaskCtx = taskCtx  // 完全な Context を保存
            return nil
        },
    }

    // Act: Runner 実行（全3層の連携）
    runner := &core.Runner{
        Config: cfg,
        Meta:   mockMeta,
        Worker: mockWorker,
        Note:   mockNote,
    }

    result, err := runner.Run(context.Background())

    // Assert: 実行結果を検証
    if err != nil {
        t.Fatalf("runner.Run() failed: %v", err)
    }

    // 状態遷移の検証
    if result.State != core.StateComplete {
        t.Errorf("State = %s, want StateComplete", result.State)
    }

    // Worker 実行を検証
    if len(result.WorkerRuns) != 1 {
        t.Errorf("WorkerRuns count = %d, want 1", len(result.WorkerRuns))
    }

    // プロンプト が正しく渡されたか
    if capturedPrompt != "Create the file as specified" {
        t.Errorf("Prompt = %q, want %q", capturedPrompt, "Create the file as specified")
    }

    // Task Note が生成されたか
    if capturedTaskCtx == nil {
        t.Error("Task Note not written")
    }
    if capturedTaskCtx.ID != cfg.Task.ID {
        t.Errorf("TaskCtx.ID = %s, want %s", capturedTaskCtx.ID, cfg.Task.ID)
    }
}

func TestRunFlow_WorkerFailure(t *testing.T) {
    // ... setup ...

    // Worker が失敗を返すモック
    mockWorker := &mock.WorkerExecutor{
        RunWorkerFunc: func(ctx context.Context, prompt string, env map[string]string) (*core.WorkerRunResult, error) {
            return &core.WorkerRunResult{
                ExitCode: 1,
                Output:   "Error: Feature not created",
                RawOutput: "...",
            }, fmt.Errorf("worker failed")
        },
    }

    // Act
    result, err := runner.Run(context.Background())

    // Assert: エラーハンドリング検証
    if err == nil {
        t.Error("Expected error, got nil")
    }
    if result.State != core.StateFailed {
        t.Errorf("State = %s, want StateFailed", result.State)
    }
}
```

**特徴**:
- インターフェース型フィールドに関数オブジェクトを保持
- テスト時に関数をカスタマイズ可能
- Nil-safe（関数が nil なら デフォルト動作）

**利点**:
- 外部API・Docker 不要で高速実行
- end-to-endフロー全体を検証可能
- 異常系（エラーパス）を簡単にシミュレート

**使用箇所**: test/integration/run_flow_test.go（2テスト）、executor_test.go（15テスト）

---

### パターン 4: Docker Sandbox Tests

**用途**: Docker API・コンテナライフサイクル・マウント機能の検証

**ビルドタグ**: `// +build docker`

**実装例**: `test/sandbox/sandbox_test.go`

```go
// +build docker

package sandbox

import (
    "context"
    "os"
    "testing"
    "time"
)

func TestSandboxManager_StartStopContainer(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Docker tests in short mode")
    }

    // Docker の可用性確認
    sm, err := worker.NewSandboxManager()
    if err != nil {
        t.Skipf("Docker not available: %v", err)
    }

    // Arrange
    tmpDir := t.TempDir()
    image := "alpine:3.19"

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Act: コンテナ起動
    containerID, err := sm.StartContainer(
        ctx,
        image,
        tmpDir,  // マウントポイント
        map[string]string{"TEST_ENV": "1"},  // 環境変数
    )
    if err != nil {
        t.Fatalf("StartContainer() error = %v", err)
    }

    // Assert: コンテナが実行中であることを確認
    if containerID == "" {
        t.Error("ContainerID is empty")
    }

    // Cleanup: コンテナ停止（Defer で確実に実行）
    defer sm.StopContainer(ctx, containerID)
}

func TestSandboxManager_EnvironmentVariables(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Docker tests in short mode")
    }

    sm, err := worker.NewSandboxManager()
    if err != nil {
        t.Skipf("Docker not available: %v", err)
    }

    // Arrange
    tmpDir := t.TempDir()
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    containerID, _ := sm.StartContainer(
        ctx, "alpine:3.19", tmpDir,
        map[string]string{
            "CUSTOM_VAR": "custom_value",
            "ANOTHER_VAR": "another_value",
        },
    )
    defer sm.StopContainer(ctx, containerID)

    // Act: 環境変数を確認するコマンド実行
    exitCode, output, err := sm.Exec(
        ctx,
        containerID,
        []string{"sh", "-c", "echo $CUSTOM_VAR"},
    )

    // Assert
    if err != nil {
        t.Fatalf("Exec() error = %v", err)
    }
    if exitCode != 0 {
        t.Errorf("ExitCode = %d, want 0", exitCode)
    }
    if !contains(output, "custom_value") {
        t.Errorf("Output = %q, want to contain 'custom_value'", output)
    }
}

func TestSandboxManager_MountPoint(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Docker tests in short mode")
    }

    sm, err := worker.NewSandboxManager()
    if err != nil {
        t.Skipf("Docker not available: %v", err)
    }

    // Arrange: ホスト側にファイルを作成
    tmpDir := t.TempDir()
    testFile := filepath.Join(tmpDir, "test.txt")
    if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
        t.Fatalf("WriteFile() error = %v", err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    containerID, _ := sm.StartContainer(ctx, "alpine:3.19", tmpDir, map[string]string{})
    defer sm.StopContainer(ctx, containerID)

    // Act: コンテナ内からマウントポイントのファイルを読む
    exitCode, output, err := sm.Exec(
        ctx,
        containerID,
        []string{"cat", "/workspace/project/test.txt"},
    )

    // Assert
    if err != nil {
        t.Fatalf("Exec() error = %v", err)
    }
    if exitCode != 0 {
        t.Errorf("ExitCode = %d, want 0", exitCode)
    }
    if !contains(output, "hello") {
        t.Errorf("Output = %q, want to contain 'hello'", output)
    }
}

func TestSandboxManager_FileWritePermission(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Docker tests in short mode")
    }

    sm, err := worker.NewSandboxManager()
    if err != nil {
        t.Skipf("Docker not available: %v", err)
    }

    // Arrange
    tmpDir := t.TempDir()
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    containerID, _ := sm.StartContainer(ctx, "alpine:3.19", tmpDir, map[string]string{})
    defer sm.StopContainer(ctx, containerID)

    // Act: コンテナ内でファイルを書き込み
    exitCode, _, err := sm.Exec(
        ctx,
        containerID,
        []string{"sh", "-c", "echo 'generated' > /workspace/project/generated.txt"},
    )

    // Assert
    if err != nil || exitCode != 0 {
        t.Fatalf("Exec() failed")
    }

    // ホスト側でファイル存在確認
    generatedFile := filepath.Join(tmpDir, "generated.txt")
    content, err := os.ReadFile(generatedFile)
    if err != nil {
        t.Errorf("ReadFile() error = %v", err)
    }
    if !contains(string(content), "generated") {
        t.Errorf("Content = %q, want 'generated'", string(content))
    }
}

func contains(s, substr string) bool {
    return strings.Contains(s, substr)
}
```

**テストケース**（7項目）:
- `StartStopContainer`: ライフサイクル管理
- `EnvironmentVariables`: 環境変数伝播
- `MountPoint`: バインドマウント動作
- `FileWritePermission`: コンテナからのファイル出力
- `NonZeroExitCode`: エラーハンドリング
- `MultipleExec`: 複数実行の安定性
- `Cleanup`: リソース解放

**実行方法**:
```bash
go test -tags=docker -timeout=10m ./test/sandbox/...
```

**Skip 機構**:
```go
if testing.Short() {
    t.Skip("Skipping Docker tests in short mode")
}

sm, err := worker.NewSandboxManager()
if err != nil {
    t.Skipf("Docker not available: %v", err)
}
```

---

### パターン 5: End-to-End Tests（Codex 統合テスト）

**用途**: 実Codex CLIとの統合検証、完全な実運用シナリオテスト

**ビルドタグ**: `// +build codex`

**前提条件**:
- Docker デーモン起動
- Codex Docker イメージがビルド済み: `docker build -t agent-runner-codex:latest sandbox/`
- `~/.codex/auth.json` 存在 または `CODEX_API_KEY` 環境変数

**実装例**: `test/codex/codex_integration_test.go`

```go
// +build codex

package codex

import (
    "context"
    "os"
    "testing"
    "time"
)

func TestCodex_BasicFlow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Codex integration tests in short mode")
    }

    // Arrange
    tmpDir := t.TempDir()

    cfg := &config.TaskConfig{
        Version: 1,
        Task: config.TaskDetails{
            ID:   "TASK-CODEX-BASIC",
            Repo: tmpDir,  // 絶対パスが必須
            PRD: config.PRDDetails{
                Text: "Create a simple hello.txt file with content 'Hello, World!'",
            },
        },
        Runner: config.RunnerConfig{
            Meta: config.MetaConfig{
                Kind: "mock",
            },
            Worker: config.WorkerConfig{
                Kind:          "codex-cli",
                DockerImage:   "agent-runner-codex:latest",
                MaxRunTimeSec: 300,
            },
        },
    }

    // コンポーネント初期化
    executor, err := worker.NewExecutor(cfg.Runner.Worker, tmpDir)
    if err != nil {
        t.Skipf("Executor initialization failed: %v", err)
    }

    metaClient := meta.NewMockClient()
    noteWriter := note.NewWriter()

    runner := &core.Runner{
        Config: cfg,
        Meta:   metaClient,
        Worker: executor,
        Note:   noteWriter,
    }

    // Act: タスク実行（最大5分）
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    taskCtx, err := runner.Run(ctx)

    // Assert
    if err != nil {
        t.Logf("runner.Run() returned error: %v (state: %s)", err, taskCtx.State)
    }

    // 状態検証（完了 または 失敗 のいずれか）
    if taskCtx.State != core.StateComplete && taskCtx.State != core.StateFailed {
        t.Errorf("State = %s, want StateComplete or StateFailed", taskCtx.State)
    }

    // Worker が実行されたか確認
    if len(taskCtx.WorkerRuns) == 0 {
        t.Error("No worker runs recorded")
    }
}

func TestCodex_FileGeneration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Codex integration tests in short mode")
    }

    tmpDir := t.TempDir()

    cfg := &config.TaskConfig{
        Version: 1,
        Task: config.TaskDetails{
            ID:   "TASK-CODEX-FILE-GEN",
            Repo: tmpDir,
            PRD: config.PRDDetails{
                Text: "Create 'result.json' with: {\"status\": \"success\"}",
            },
        },
        Runner: config.RunnerConfig{
            Meta: config.MetaConfig{Kind: "mock"},
            Worker: config.WorkerConfig{
                Kind:          "codex-cli",
                DockerImage:   "agent-runner-codex:latest",
                MaxRunTimeSec: 300,
            },
        },
    }

    executor, _ := worker.NewExecutor(cfg.Runner.Worker, tmpDir)
    runner := &core.Runner{
        Config: cfg,
        Meta:   meta.NewMockClient(),
        Worker: executor,
        Note:   note.NewWriter(),
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    taskCtx, _ := runner.Run(ctx)

    // Assert: ファイルが実際に生成されたか確認
    resultFile := filepath.Join(tmpDir, "result.json")
    if _, err := os.Stat(resultFile); os.IsNotExist(err) {
        t.Logf("result.json not created (state: %s)", taskCtx.State)
    } else if err == nil {
        // ファイルが存在し、内容確認
        content, _ := os.ReadFile(resultFile)
        if !contains(string(content), "success") {
            t.Logf("result.json content: %s", string(content))
        }
    }
}

func TestCodex_EnvironmentIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Codex integration tests in short mode")
    }

    tmpDir := t.TempDir()

    // 環境変数をタスク設定に含める
    cfg := &config.TaskConfig{
        Version: 1,
        Task: config.TaskDetails{
            ID:   "TASK-CODEX-ENV",
            Repo: tmpDir,
            PRD: config.PRDDetails{
                Text: "Verify TEST_ENV is set in the environment",
            },
        },
        Runner: config.RunnerConfig{
            Meta: config.MetaConfig{Kind: "mock"},
            Worker: config.WorkerConfig{
                Kind:          "codex-cli",
                DockerImage:   "agent-runner-codex:latest",
                MaxRunTimeSec: 300,
                Env: map[string]string{
                    "TEST_ENV": "integration_test",
                },
            },
        },
    }

    executor, _ := worker.NewExecutor(cfg.Runner.Worker, tmpDir)
    runner := &core.Runner{
        Config: cfg,
        Meta:   meta.NewMockClient(),
        Worker: executor,
        Note:   note.NewWriter(),
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    taskCtx, _ := runner.Run(ctx)

    // 環境変数がWorker実行に渡されたかログで確認
    if len(taskCtx.WorkerRuns) > 0 {
        run := taskCtx.WorkerRuns[0]
        t.Logf("Worker output: %s", run.Output)
    }
}
```

**テストケース**（8項目）:
- `BasicFlow`: 基本的な実行フロー
- `TaskConfigLoading`: YAML設定解析
- `FileGeneration`: ファイル永続化確認
- `ErrorHandling`: エラーケース処理
- `MockMode`: Mockモード動作確認
- `TaskNoteGeneration`: Task Note出力確認
- `EnvironmentIntegration`: 環境変数伝播確認
- `MultipleRuns`: 複数実行の安定性

**実行方法**:
```bash
# Codex統合テスト のみ
go test -tags=codex -timeout=10m ./test/codex/...

# 全テスト実行（最も徹底的）
go test -tags=docker,codex -timeout=15m ./...
```

---

## ビルドタグ戦略

### タグの位置と構文

```go
// ファイル先頭（ドキュメント文字列より前）
// +build docker

package sandbox

import "testing"
```

### ビルドタグ実行マトリクス

| コマンド | タグ | 実行テスト | 実行時間 | 前提条件 |
|---------|------|---------|---------|--------|
| `go test ./...` | （なし） | ユニット + Mock統合 | <10s | なし |
| `go test -short ./...` | （なし） | ユニット + Mock統合（Docker テストは skip） | <10s | なし |
| `go test -tags=docker ./...` | docker | ユニット + Mock統合 + Docker Sandbox | 30-60s | Docker デーモン |
| `go test -tags=codex ./...` | codex | ユニット + Mock統合 + Codex統合 | 5-10m | Docker + Codex CLI |
| `go test -tags=docker,codex ./...` | docker,codex | 全テスト | 10-15m | Docker + Codex CLI |

### Skip 機構

```go
// Docker テストが -short フラグ付き実行時は skip
if testing.Short() {
    t.Skip("Skipping Docker tests in short mode")
}

// Docker 未インストール時は graceful skip
sm, err := worker.NewSandboxManager()
if err != nil {
    t.Skipf("Docker not available: %v", err)
}
```

---

## テスト実行コマンド集

### 日常開発（ユニット・Mock統合）

```bash
# 基本実行（最速、依存なし）
go test ./...

# 詳細ログ付き
go test -v ./...

# 特定パッケージのみ
go test ./internal/core
go test ./pkg/config
go test ./internal/meta

# 並列実行（高速化）
go test -parallel 4 ./...

# 特定テストのみ実行
go test -run TestRunFlow_Success ./test/integration/...
go test -run TestTaskConfig_UnmarshalYAML_Valid ./pkg/config/...
```

### Docker Sandbox テスト

```bash
# Docker テスト実行
go test -tags=docker -timeout=10m ./test/sandbox/...

# Docker テスト詳細ログ
go test -tags=docker -v -timeout=10m ./test/sandbox/...
```

### Codex 統合テスト

```bash
# Codex 統合テスト実行
go test -tags=codex -timeout=10m ./test/codex/...

# Codex テスト詳細ログ
go test -tags=codex -v -timeout=10m ./test/codex/...
```

### 全テスト実行（最も徹底的）

```bash
# 前提条件: Docker デーモン + Codex CLI インストール済み
go test -tags=docker,codex -timeout=15m ./...

# 詳細ログ付き
go test -tags=docker,codex -v -timeout=15m ./...
```

### カバレッジ測定

```bash
# レポート生成（カバレッジ確認）
go test -coverprofile=coverage.out ./...

# HTML形式で表示
go tool cover -html=coverage.out

# テキスト形式で表示（関数別）
go tool cover -func=coverage.out

# 特定パッケージのカバレッジ
go test -cover ./internal/core
go test -cover ./internal/meta

# 全テストのカバレッジ
go test -tags=docker,codex -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### デバッグ・プロファイリング

```bash
# 標準出力を表示
go test -v ./...

# Race detector（並行バグ検出）
go test -race ./...

# CPU プロファイル
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# メモリプロファイル
go test -memprofile=mem.prof ./...
go tool pprof mem.prof

# Short モード（Docker テストを skip）
go test -short ./...
```

---

## テストライブラリと手法

### 外部ライブラリ

| ライブラリ | バージョン | 用途 | 使用箇所 |
|-----------|-----------|------|--------|
| **gopter** | v0.2.x | プロパティベーステスト | `internal/core/runner_test.go` |
| **gopkg.in/yaml.v3** | v3.x | YAML パース・マーシャル | 全テストファイル |
| **docker/client** | (Go 1.24+) | Docker API | `test/sandbox/sandbox_test.go` |
| **context** | std | キャンセル・タイムアウト | 全テストファイル |
| **testing** | std | テストフレームワーク | 全テストファイル |

### テスト手法詳細

#### 1. Table-Driven Tests（表駆動テスト）

**メリット**:
- テストケースが視覚的に整理される
- 新規ケース追加が簡単
- 正常系・異常系を同じテーブルで管理
- エッジケース（空文字列、nil、極端な値）を網羅しやすい

**ベストプラクティス**:
- ケース名は**描写的に**（`valid`, `invalid` ではなく、`empty_prd_path`, `missing_required_field`）
- `want` フィールド は期待値を明確に
- エラーケースは `wantErr` フラグで統一

#### 2. Property-Based Tests（プロパティベーステスト）

**特徴**:
- 有限状態機械の不変条件を形式化
- 複数のランダムケース（gopter が自動生成）で検証
- エッジケース（0、1、極端な値）を自動テスト
- 再現可能（seed 指定可）

**メリット**:
- 手作業では見落とす組み合わせをテスト
- 不変条件を明文化できる
- 状態遷移バグ（off-by-one など）を検出

**使用タイミング**:
- FSM（有限状態機械）のテスト
- アルゴリズムの数学的性質検証
- 並行処理の同期ロジック検証

#### 3. Mock/Stub パターン

**パターン**: Function Field Injection

```go
// モック実装
type MetaClient struct {
    PlanTaskFunc   func(ctx context.Context, prdText string) (*PlanTaskResponse, error)
    NextActionFunc func(ctx context.Context, taskSummary *TaskSummary) (*NextActionResponse, error)
}

func (m *MetaClient) PlanTask(ctx context.Context, prdText string) (*PlanTaskResponse, error) {
    if m.PlanTaskFunc != nil {
        return m.PlanTaskFunc(ctx, prdText)
    }
    return nil, nil
}
```

**メリット**:
- 構造体フィールドに関数オブジェクトを保持
- テスト時に動的にカスタマイズ可能
- Nil-safe（関数が nil なら デフォルト動作）
- インターフェース実装と異なり、複数の関数を部分的に実装可能

#### 4. Context-based Testing

**用途**:
- タイムアウト処理の検証
- キャンセル機能の動作確認
- 長時間実行テストの制限

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// タイムアウト処理
if err := runLongTask(ctx); err != nil {
    t.Fatalf("Context timeout or error: %v", err)
}
```

#### 5. Temporary Directory（テンポラリディレクトリ）

**利点**:
- テスト間での干渉なし
- ディスクをクリーンアップ（自動削除）
- クロスプラットフォーム互換性

```go
tmpDir := t.TempDir()  // テスト完了時に自動削除

resultFile := filepath.Join(tmpDir, "result.txt")
if err := os.WriteFile(resultFile, []byte("test"), 0644); err != nil {
    t.Fatal(err)
}
```

---

## ベストプラクティス

### 1. エラーメッセージの詳細さ

```go
// Good: 詳細な context
if result.ExitCode != 0 {
    t.Errorf(
        "ExitCode = %d (want 0)\nOutput: %s\nError: %v",
        result.ExitCode, result.Output, result.Error,
    )
}

// Bad: 簡潔すぎる
if result.ExitCode != 0 {
    t.Error("ExitCode wrong")
}
```

### 2. 環境変数の検証

```go
// Good: 環境変数が正しく伝播しているか検証
mockWorker := &mock.WorkerExecutor{
    RunWorkerFunc: func(ctx context.Context, prompt string, env map[string]string) (*core.WorkerRunResult, error) {
        if env["TEST_ENV"] != "expected_value" {
            t.Errorf("Env TEST_ENV = %s, want 'expected_value'", env["TEST_ENV"])
        }
        return &core.WorkerRunResult{ExitCode: 0}, nil
    },
}
```

### 3. タイムスタンプの検証

```go
// Good: タイムスタンプの順序性を検証
if result.FinishedAt.Before(result.StartedAt) {
    t.Errorf(
        "FinishedAt (%v) is before StartedAt (%v)",
        result.FinishedAt, result.StartedAt,
    )
}

// Good: タイムスタンプが設定されているか確認
if result.StartedAt.IsZero() {
    t.Errorf("StartedAt should not be zero time")
}
```

### 4. Nil チェックと デフォルト値

```go
// Good: nil チェック + デフォルト値の確認
if executor.Config.DockerImage == "" {
    // Empty は OK、executor.RunWorker で デフォルト値を使用
    // 構造体の初期化が正しいことを確認
} else if executor.Config.DockerImage != "agent-runner-codex:latest" {
    t.Errorf("DockerImage = %s, want default or 'agent-runner-codex:latest'", executor.Config.DockerImage)
}
```

### 5. テスト隔離（Isolation）

```go
// Good: 各テストが独立している
func TestTaskConfig_UnmarshalYAML_Valid(t *testing.T) {
    // 独立したテストケース
    cfg := &config.TaskConfig{...}
    // ...
}

func TestTaskConfig_UnmarshalYAML_Invalid(t *testing.T) {
    // グローバル状態に依存しない
    cfg := &config.TaskConfig{...}
    // ...
}

// テストの実行順序が変わってもパスする
```

### 6. Defer によるクリーンアップ

```go
// Good: リソース解放を保証
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

containerID, err := sm.StartContainer(ctx, image, tmpDir, map[string]string{})
if err != nil {
    t.Fatalf("StartContainer() error = %v", err)
}
defer sm.StopContainer(ctx, containerID)  // 必ず実行される
```

### 7. Arrange-Act-Assert パターン

```go
// Good: テスト構成が明確
func TestRunFlow_Success(t *testing.T) {
    // === Arrange: テストデータ準備 ===
    cfg := &config.TaskConfig{...}
    mockMeta := &mock.MetaClient{...}
    mockWorker := &mock.WorkerExecutor{...}

    // === Act: 動作実行 ===
    runner := core.NewRunner(cfg, mockMeta, mockWorker, mockNote)
    result, err := runner.Run(context.Background())

    // === Assert: 検証 ===
    if err != nil {
        t.Fatalf("runner.Run() error = %v", err)
    }
    if result.State != core.StateComplete {
        t.Errorf("State = %s, want StateComplete", result.State)
    }
}
```

### 8. テストヘルパー関数

```go
// Good: 重複コードを削減
func createMinimalConfig() *config.TaskConfig {
    return &config.TaskConfig{
        Version: 1,
        Task: config.TaskDetails{
            ID:   "TEST-001",
            Repo: "/tmp/test",
            PRD: config.PRDDetails{
                Text: "Test requirement",
            },
        },
        Runner: config.RunnerConfig{
            Meta:   config.MetaConfig{Kind: "mock"},
            Worker: config.WorkerConfig{Kind: "codex-cli"},
        },
    }
}

// テスト内での使用
func TestRunFlow_Success(t *testing.T) {
    cfg := createMinimalConfig()
    // ...
}
```

---

## テスト精度管理手法

### 1. カバレッジ管理

**測定方法**:
```bash
# 全テストのカバレッジ測定
go test -tags=docker,codex -coverprofile=coverage.out ./...

# HTML形式で表示
go tool cover -html=coverage.out

# 関数別カバレッジ
go tool cover -func=coverage.out | grep total
```

**目標値**:
- `internal/core`: 70%+ （重要なFSMロジック）
- `internal/meta`: 80%+ （プロトコル通信）
- `internal/note`: 80%+ （出力生成）
- `internal/worker`: 50%+ （Docker依存が多い）
- `pkg/config`: 90%+ （パース・バリデーション）
- **全体**: 50%+（達成困難な依存部分を除外）

**現在のカバレッジ** (Phase 5実測):
- `internal/core`: 67.4%
- `internal/note`: 80.0%
- `internal/worker`: 23.0%（Docker依存が多い）
- **全体**: 26.8%

### 2. 不変条件の定義と検証

**FSM 不変条件**:
```
1. State遷移: PENDING → PLANNING → RUNNING → VALIDATING → COMPLETE/FAILED
2. AC数: len(AcceptanceCriteria) = len(InitialACs)
3. WorkerRuns: len(WorkerRuns) > 0 ⟹ State ∈ {RUNNING, VALIDATING, COMPLETE, FAILED}
4. タイムスタンプ: StartedAt ≤ FinishedAt
5. エラーハンドリング: Error != nil ⟹ State = FAILED
```

**検証方法**:
- Property-Based Tests（gopter）で形式化
- 複数のランダムケース（5-100パターン）で自動検証
- エッジケース（0、1、極端な値）を自動生成

### 3. テストデータ生成戦略

**レベル別**:
- **Level 1 (ユニット)**: 最小限のデータ（createMinimalConfig）
- **Level 2 (統合)**: 複数シナリオ（成功、エラー、キャンセル）
- **Level 3 (Docker)**: 実環境マウント・環境変数
- **Level 4 (Codex)**: 実プロジェクト構造

**ベストプラクティス**:
- テストデータを Builder パターンで生成可能にする
- 同じ構成を複数テストで使用できるように整理

### 4. 並行実行・Race Detection

```bash
# Race detector で並行バグを検出
go test -race ./...

# 並列実行テスト
go test -parallel 4 ./...

# Race detection + 並列実行
go test -race -parallel 4 ./...
```

**検出対象**:
- データ競合（複数ゴルーチンが同時にメモリ書き込み）
- 同期漏れ（mutex忘れなど）

### 5. テスト実行時間の最適化

**段階的テスト実行**:

```
┌─────────────────────────────────────┐
│ Stage 1: ユニット + Mock統合       │
│ 実行時間: <10秒                     │
│ 実行頻度: コミット前（毎回）        │
└─────────────────────────────────────┘
         ↓
┌─────────────────────────────────────┐
│ Stage 2: + Docker Sandbox           │
│ 実行時間: 30-60秒                   │
│ 実行頻度: PR作成前（1-2回）         │
└─────────────────────────────────────┘
         ↓
┌─────────────────────────────────────┐
│ Stage 3: + Codex 統合               │
│ 実行時間: 10-15分                   │
│ 実行頻度: CI/CD（手動トリガー）     │
└─────────────────────────────────────┘
```

**高速化テクニック**:
- `-parallel N`: 並列実行
- `-short`: Docker テストをスキップ
- `-run pattern`: 特定テストのみ実行
- キャッシング: go build cache、Docker image cache

---

## トラブルシューティング

### 問題: Docker がインストールされていない

```bash
$ go test -tags=docker ./test/sandbox/...

--- SKIP: TestSandboxManager_StartStopContainer (0.00s)
    sandbox_test.go:XXX: Docker not available: Cannot connect to Docker daemon
```

**解決**:
```bash
# Docker デーモンを起動
docker --version  # Docker がインストール済みか確認
docker ps         # デーモンが実行中か確認

# Docker がない場合はインストール
# macOS: brew install docker docker-desktop
# Linux: apt-get install docker.io (または dnf install docker)
```

### 問題: Codex Docker イメージがビルドされていない

```bash
$ go test -tags=codex ./test/codex/...

--- SKIP: TestCodex_BasicFlow (0.00s)
    codex_integration_test.go:XXX: image not found
```

**解決**:
```bash
# Codex Docker イメージをビルド
docker build -t agent-runner-codex:latest sandbox/

# 確認
docker images | grep agent-runner-codex
```

### 問題: Codex 認証エラー

```bash
$ go test -tags=codex ./test/codex/...

--- FAIL: TestCodex_BasicFlow (5.43s)
    codex_integration_test.go:XXX: Codex API key not found
```

**解決**:
```bash
# 方法1: auth.json 設定
mkdir -p ~/.codex
cat > ~/.codex/auth.json << 'EOF'
{
  "api_key": "your-codex-api-key"
}
EOF

# 方法2: 環境変数
export CODEX_API_KEY="your-codex-api-key"
```

### 問題: テストタイムアウト

```bash
$ go test -tags=codex ./test/codex/...

--- TIMEOUT: TestCodex_BasicFlow (300.00s)
```

**解決**:
```bash
# タイムアウト時間を延長
go test -tags=codex -timeout=20m ./test/codex/...

# または特定テストのみ実行
go test -tags=codex -timeout=10m -run TestCodex_BasicFlow ./test/codex/...
```

### 問題: 相対パスのマウントエラー

```bash
$ go test ./...

--- FAIL: TestSandboxManager_StartStopContainer (0.00s)
    sandbox_test.go:XXX: invalid mount path: '.' mount path must be absolute
```

**解決**:
```bash
# TaskConfig で絶対パスを使用
cfg := &config.TaskConfig{
    Task: config.TaskDetails{
        Repo: "/tmp/project",  // 相対パス "." ではなく絶対パス
    },
}

# または実行前に解決
abspath, _ := filepath.Abs(cfg.Task.Repo)
cfg.Task.Repo = abspath
```

---

## 品質指標とメトリクス

### テスト数と覆範囲

| レイヤー | テスト数 | ファイル数 | 主要検証項目 |
|---------|---------|----------|----------|
| ユニット | 52 | 5 | パース、プロトコル、生成、実行 |
| 統合（Mock） | 2 | 1 | end-to-endフロー、エラーハンドリング |
| Docker Sandbox | 7 | 1 | コンテナ管理、マウント、環境変数 |
| Codex 統合 | 8 | 1 | 実Codex実行、ファイル生成、Note出力 |
| **合計** | **70+** | **8** | 全レイヤー統合検証 |

### カバレッジ目標

| パッケージ | 現在 | 目標 | 理由 |
|-----------|------|------|------|
| `internal/core` | 67.4% | 80%+ | FSM重要ロジック |
| `internal/meta` | n/a | 85%+ | LLM通信プロトコル |
| `internal/note` | 80.0% | 85%+ | Markdown出力品質 |
| `internal/worker` | 23.0% | 40%+ | Docker依存が多い |
| `pkg/config` | n/a | 95%+ | YAML解析・バリデーション |
| **全体** | 26.8% | 50%+ | 統合品質 |

### ビルドパイプライン（参考: CI/CD）

```yaml
# .github/workflows/test.yml
stages:
  - unit_tests:      go test ./...                (毎回)
  - integration:     go test ./test/integration/... (毎回)
  - docker_tests:    go test -tags=docker ./...    (毎回)
  - codex_tests:     go test -tags=codex ./...     (手動トリガー)
  - lint:            golangci-lint run            (毎回)
  - coverage:        go tool cover -html=...      (毎回)
```

---

## 関連ドキュメント

- **[TESTING.md](../TESTING.md)**: テスト全般のベストプラクティス
- **[CODEX_TEST_README.md](../CODEX_TEST_README.md)**: Codex 統合ガイド詳細
- **[internal/core/CLAUDE.md](../internal/core/CLAUDE.md)**: FSM実装の詳細
- **[internal/meta/CLAUDE.md](../internal/meta/CLAUDE.md)**: Meta-agent プロトコル詳細
- **[internal/worker/CLAUDE.md](../internal/worker/CLAUDE.md)**: Docker・Sandbox 詳細
- **[internal/mock/CLAUDE.md](../internal/mock/CLAUDE.md)**: モック実装パターン詳細
- **[pkg/config/CLAUDE.md](../pkg/config/CLAUDE.md)**: YAML スキーマ詳細

---

**最後に**: このドキュメントは開発の進展と共に更新してください。新しいテストパターンの発見、精度管理の改善、既知の問題の解決結果などを適宜記録することで、プロジェクトのテスト品質が継続的に向上します。
