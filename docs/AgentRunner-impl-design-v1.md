# AgentRunner v1 実装設計（Go / Sandbox / Meta プロンプト）

## 1. Go プロジェクト構成・パッケージ分割

### 1-1. ルート構成

```text
agent-runner/
  cmd/
    agent-runner/
      main.go

  internal/
    core/          # Core オーケストレータ（実装済み）
    meta/          # Meta エージェント Client（実装済み）
    worker/        # Worker Executor + Sandbox Manager（実装済み）
    note/          # Task Note Writer（実装済み）
    mock/          # テスト用 Mock 実装（実装済み）

  pkg/
    config/        # YAML スキーマ定義（実装済み）

  test/
    integration/   # 統合テスト
    sandbox/       # Docker Sandbox テスト
    codex/         # Codex 統合テスト

  go.mod
  go.sum
  README.md
```

**注**: 設計時の `internal/cli`, `internal/task`, `internal/sandbox`, `internal/llm`, `internal/logging` は、実装時に以下のように統合されました：

- `internal/cli` → `cmd/agent-runner/main.go` に統合
- `internal/task` → `internal/core/context.go` に統合
- `internal/sandbox` → `internal/worker/sandbox.go` に統合
- `internal/llm` → `internal/meta/client.go` に統合
- `internal/logging` → `cmd/agent-runner/main.go` で直接 `slog` を使用

各ディレクトリの責務は以下。

#### cmd/agent-runner

- エントリポイント。
- 主な責務
  - `main()` 関数
  - CLI 引数・フラグ（v1 ではほぼ無し）のパース
  - 標準入力からの YAML 読み込み
  - `internal/cli` に処理を委譲

```go
// cmd/agent-runner/main.go（イメージ）
func main() {
    ctx := context.Background()

    logger := logging.New() // slog 等

    if err := cli.Run(ctx, os.Stdin, os.Stdout, os.Stderr, logger); err != nil {
        // エラー種別ごとに exit code を変えてもよい
        os.Exit(1)
    }
}
```

#### internal/cli

- 「1 回の CLI 実行」をまとめてハンドリングするファサード。
- 主な責務
  - Task YAML の読み込み・バリデーション
  - `task.Context` の初期化
  - `meta` / `worker` / `tasknote` を呼び出すオーケストレーション
  - 実行結果にもとづく exit code 決定

```go
package cli

func Run(
    ctx context.Context,
    stdin io.Reader,
    stdout, stderr io.Writer,
    logger *slog.Logger,
) error {
    // 1. YAML 読み込み
    // 2. TaskContext 初期化
    // 3. Meta.plan → Worker.run → Test → TaskNote 生成
    // 4. 状態に応じて error / nil を返す
}
```

#### internal/config

- 設定値の集約。
- 主な責務
  - 環境変数や将来の設定ファイルからの読み込み
  - OpenAI API Key / モデル名
  - Codex CLI のパス / デフォルトイメージ名 など

```go
type Config struct {
    OpenAIAPIKey     string
    OpenAIModel      string
    CodexDockerImage string
    // ... 他必要に応じて
}
```

#### internal/task

- Task YAML のスキーマ定義・バリデーション・内部 `TaskContext` の定義。
- 主な責務
  - 外部 YAML の構造体（`TaskSpec`）
  - 実行時状態を持つ `TaskContext`
  - ステートマシン（`State` enum）

```go
type TaskSpec struct {
    Version int `yaml:"version"`
    Task struct {
        ID    string `yaml:"id"`
        Title string `yaml:"title"`
        Repo  string `yaml:"repo"`
        PRD   struct {
            Path string `yaml:"path"`
            Text string `yaml:"text"`
        } `yaml:"prd"`
        Test struct {
            Command string `yaml:"command"`
        } `yaml:"test"`
    } `yaml:"task"`

    Runner struct {
        Meta struct {
            Kind  string `yaml:"kind"`
            Model string `yaml:"model"`
        } `yaml:"meta"`
        Worker struct {
            Kind         string            `yaml:"kind"`
            DockerImage  string            `yaml:"docker_image"`
            MaxRunTime   int               `yaml:"max_run_time_sec"`
            EnvFromHost  map[string]string `yaml:"env"`
        } `yaml:"worker"`
    } `yaml:"runner"`
}

type State string

const (
    StatePending    State = "PENDING"
    StatePlanning   State = "PLANNING"
    StateRunning    State = "RUNNING"
    StateValidating State = "VALIDATING"
    StateComplete   State = "COMPLETE"
    StateFailed     State = "FAILED"
)

type TaskContext struct {
    Spec   TaskSpec
    State  State
    Repo   string // 実際の絶対パス
    PRD    string // 実際の PRD テキスト
    // Meta・Worker・Test の実行ログ、要約など
}
```

#### internal/meta

- Meta（LLM）呼び出し専用パッケージ。
- 主な責務
  - `plan_task` / `next_action` / `completion_assessment` 用の呼び出し I/F
  - プロンプトテンプレートの組み立て
  - YAML 応答のパース・検証
  - **LLM エラー再試行ロジック（Exponential Backoff）**

```go
type Service interface {
    PlanTask(ctx context.Context, tc *task.TaskContext) (*PlanTaskResult, error)
    NextAction(ctx context.Context, tc *task.TaskContext) (*NextActionResult, error)
    CompletionAssessment(ctx context.Context, tc *task.TaskContext) (*CompletionAssessmentResult, error)
}
```

実際の LLM 呼び出しは `internal/meta/client.go` 内で完結しており、`internal/llm` パッケージは存在しません。

#### internal/meta/client.go (旧 internal/llm)

- OpenAI などの具体的な LLM クライアント。
- 主な責務
  - ChatCompletion API 呼び出し
  - **Exponential Backoff による再試行**
    - 対象: 5xx, Timeout, RateLimit
    - 最大 3 回、1s -> 2s -> 4s

```go
type Client interface {
    Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
}
```

#### internal/worker

- Worker ごとの実行ロジック + Sandbox 管理。
- v1 では `codex-cli` のみ。
- 主な責務
  - Meta の `worker_call` 情報をもとに SandboxExecutor を呼び出す
  - **コンテナライフサイクル管理（Start / Exec / Stop）**
  - 実行結果（stdout/stderr/exit code）を `TaskContext` に反映

```go
type Runner interface {
    RunWorker(ctx context.Context, tc *task.TaskContext, call WorkerCall) (*WorkerResult, error)
}
```

#### internal/worker/sandbox.go (旧 internal/sandbox)

- サンドボックス実行（Docker）の抽象化。
- 主な責務
  - `SandboxExecutor` インターフェース
  - `DockerLocalExecutor` 実装（v1）
  - **ImagePull 自動実行**
  - **Codex 認証自動マウント**

#### internal/tasknote

- Task Note（Markdown）の生成。
- 主な責務
  - `TaskContext` から Markdown 文字列を組み立て
  - `<repo>/.agent-runner/task-<id>.md` に保存

#### internal/logging

- ロガーの初期化（例: `slog`）。
- 主な責務
  - ログフォーマット（JSON/テキスト）の統一
  - log level の設定

---

## 2. SandboxExecutor / DockerLocalExecutor 設計

### 2-1. インターフェース定義

```go
package worker

import (
    "context"
)

type SandboxProvider interface {
    StartContainer(ctx context.Context, image string, repoPath string, env map[string]string) (string, error)
    Exec(ctx context.Context, containerID string, cmd []string) (int, string, error)
    StopContainer(ctx context.Context, containerID string) error
}
```

- 実装側の方針
  - **StartContainer**:
    - コンテナを起動し、ID を返す。
    - **ImagePull 自動実行**: イメージが存在しない場合、自動的に pull する。
    - **Codex 認証自動マウント**: `~/.codex/auth.json` があればマウント、なければ `CODEX_API_KEY` 環境変数を注入。
  - **Exec**:
    - 既存コンテナ内でコマンドを実行する。
    - ExitCode と Output (Stdout + Stderr) を返す。
  - **StopContainer**:
    - コンテナを強制停止・削除する。

### 2-2. DockerLocalExecutor 実装

`SandboxManager` として実装。

```go
type SandboxManager struct {
    cli *client.Client
}
```

`StartContainer` の実装方針：

1. `ImageInspect` でイメージ確認 → なければ `ImagePull`。
2. マウント準備:
   - `repoPath` → `/workspace/project`
   - `~/.codex/auth.json` (存在すれば) → `/root/.codex/auth.json` (ReadOnly)
3. 環境変数準備:
   - 引数の `env` を注入
   - `CODEX_API_KEY` (存在すれば) を注入
4. `ContainerCreate` & `ContainerStart`:
   - `Cmd`: `["tail", "-f", "/dev/null"]` (Keep Alive)
   - `WorkingDir`: `/workspace/project`

`Exec` の実装方針：

1. `ContainerExecCreate`:
   - `AttachStdout`, `AttachStderr`: true
2. `ContainerExecAttach`:
   - 出力を `stdcopy.StdCopy` でバッファリング
3. `ContainerExecInspect`:
   - ExitCode 取得

`StopContainer` の実装方針：

1. `ContainerStop` (Timeout: 0) で強制停止。

---

## 3. Meta 用プロンプトテンプレート（plan_task / next_action）

ここでは「LLM に渡すメッセージ構造」と「実際のテンプレート文言」を定義します。  
想定は OpenAI Chat API（`gpt-5.1`）ですが、プロンプト自体は他プロバイダでも流用可能な形にしてあります。

### 3-1. 共通：前提

- 応答フォーマットは **YAML のみ**。
- **プロンプト中に明示的に「他のテキストは書かない」** と指示する。
- 言語ポリシー
  - PRD が日本語なら Acceptance Criteria / reason も日本語優先。
  - PRD が英語なら英語優先。
  - 混在している場合は PRD に合わせて柔軟に。

---

### 3-2. plan_task 用テンプレート

#### 3-2-1. 想定メッセージ構造

```go
messages := []llm.Message{
    {Role: "system", Content: planTaskSystemPrompt},
    {Role: "user",   Content: renderPlanTaskUserPrompt(taskSpecYAML, prdText)},
}
```

#### 3-2-2. system メッセージ案

````text
あなたはソフトウェア開発プロジェクトのテックリードです。

- あなたの役割は、与えられたタスク仕様と PRD（要件定義）から、
  「このタスクが完了したとみなすための Acceptance Criteria（受け入れ条件）」を設計することです。
- Acceptance Criteria は、実装者やレビュアーがタスクの完了可否を判断できる程度に具体的である必要があります。
- 過剰に細かい実装手順を書かず、「何ができればよいか」にフォーカスしてください。

出力フォーマットについて:

- 出力は必ず 1 つの YAML ドキュメントのみとします。
- コードブロック（```）や解説文は一切書かないでください。
- YAML のスキーマは次のとおりです:

  type: "plan_task"
  acceptance_criteria:
    - id: string          # 例: "AC-1"
      description: string # 具体的な受け入れ条件の説明
      rationale: string   # なぜそれが必要なのか・意図（任意だが推奨）

- フィールド名は必ず上記のとおりにしてください（追加フィールドは定義しないでください）。
- 言語は PRD の言語に合わせてください（PRD が日本語なら日本語、英語なら英語）。

以上のルールを厳守し、YAML だけを出力してください。
````

#### 3-2-3. user メッセージテンプレート案

```text
以下に、このタスクの入力情報を示します。

1. Task YAML（実行設定）

---
{{TASK_SPEC_YAML}}
---

2. PRD（要件定義テキスト）

---
{{PRD_TEXT}}
---

お願いしたいこと:

- 上記をもとに、このタスクの完了条件を表す Acceptance Criteria を設計してください。
- できるだけ 3〜10 個程度の粒度にまとめてください（多すぎても少なすぎてもレビューが困難です）。
- 出力は前述のスキーマに従った YAML のみとしてください。
```

---

### 3-3. next_action 用テンプレート

#### 3-3-1. 想定メッセージ構造

```go
messages := []llm.Message{
    {Role: "system", Content: nextActionSystemPrompt},
    {Role: "user",   Content: renderNextActionUserPrompt(taskContextSummaryYAML)},
}
```

ここで `taskContextSummaryYAML` は、Go 側で `TaskContext` から以下のような「要約 YAML」を生成して渡す想定です（イメージ）:

```yaml
task:
  id: "TASK-123"
  title: "Implement API endpoint X"
  prd_summary: "..."
acceptance_criteria:
  - id: "AC-1"
    description: "..."
last_worker_result:
  exists: true
  exit_code: 0
  stdout_tail: "..."
  stderr_tail: "..."
test_result:
  executed: true
  exit_code: 0
state: "RUNNING"
notes: []
```

#### 3-3-2. system メッセージ案

````text
あなたはソフトウェア開発タスクを管理するテックリード兼オーケストレータです。

- 与えられたタスクコンテキスト（TaskContext）にもとづき、
  次に何をすべきかを決定する役割を担います。
- あなたが選択できるアクションは次の 2 つです:

  1. "run_worker":
     - コード変更や追加の作業を行うために、Worker（例: コーディングエージェント CLI）を実行する。
  2. "mark_complete":
     - すでに Acceptance Criteria が満たされており、タスクを完了したと判断する。

出力フォーマットについて:

- 出力は必ず 1 つの YAML ドキュメントのみとします。
- コードブロック（```）や解説文は一切書かないでください。
- YAML のスキーマは次のとおりです:

  type: "next_action"
  decision:
    action: "run_worker" | "mark_complete"
    reason: string
  worker_call:
    worker_type: string # 例: "codex-cli"
    mode: string        # 例: "exec"
    prompt: string      # Worker に渡す指示文全体
  # 注意: action が "mark_complete" の場合、worker_call は null でもよい

- "reason" には、Acceptance Criteria と TaskContext を踏まえてなぜその判断をしたかを簡潔に書いてください。
- "prompt" には、Worker が 1 回の実行で行うべき作業を、実装者に指示するイメージで具体的に記述してください。
  - タスクの目的
  - 現在の状態・既存コードの前提
  - 行うべき変更内容（優先度の高いものに絞る）
  - 必要であればテスト実行方針
- ただし、Worker は 1 回で巨大な作業を行うべきではありません。
  - 1 回の "run_worker" で実施する範囲を適切に絞り込んでください。

- フィールド名は必ず上記のとおりにしてください（追加フィールドは定義しないでください）。
- 言語はタスクの PRD の言語に合わせてください。

以上のルールを厳守し、YAML だけを出力してください。
````

#### 3-3-3. user メッセージテンプレート案

```text
以下に、このタスクの現在のコンテキスト（TaskContext の要約）を YAML で示します。

---
{{TASK_CONTEXT_SUMMARY_YAML}}
---

お願いしたいこと:

- このタスクの状態を評価し、
  - 追加で Worker を実行すべきか（action = "run_worker"）
  - もしくはタスクを完了済みとみなしてよいか（action = "mark_complete"）
  を決定してください。
- action が "run_worker" の場合は、Worker に与えるべき "prompt" の内容を具体的に設計してください。
- action が "mark_complete" の場合は、なぜ Acceptence Criteria が満たされていると判断できるのかを "reason" に明記してください。
- 出力は前述のスキーマに従った YAML のみとしてください。
```

---

### 3-4. completion_assessment 用テンプレート

#### 3-4-1. 想定メッセージ構造

```go
messages := []llm.Message{
    {Role: "system", Content: completionAssessmentSystemPrompt},
    {Role: "user",   Content: renderCompletionAssessmentUserPrompt(taskSummary)},
}
```

#### 3-4-2. system メッセージ案

````text
あなたはタスクの完了状況を評価する Meta-agent です。
Acceptance Criteria と Worker の実行結果を確認し、タスクが完了したかどうかを判定してください。

出力フォーマットについて:

- 出力は必ず 1 つの YAML ドキュメントのみとします。
- コードブロック（```）や解説文は一切書かないでください。
- YAML のスキーマは次のとおりです:

  type: "completion_assessment"
  version: 1
  payload:
    all_criteria_satisfied: boolean
    summary: string
    by_criterion:
      - id: string      # AC-ID
        status: string  # "passed" | "failed"
        comment: string # 判定理由

- フィールド名は必ず上記のとおりにしてください。
- 言語はタスクの PRD の言語に合わせてください。

以上のルールを厳守し、YAML だけを出力してください。
````

#### 3-4-3. user メッセージテンプレート案

```text
以下に、このタスクの Acceptance Criteria と実行結果の要約を示します。

{{TASK_SUMMARY_YAML}}

お願いしたいこと:

- 各 Acceptance Criteria が満たされているか判定してください。
- 全ての Criteria が満たされている場合、all_criteria_satisfied を true にしてください。
- 判定結果を前述のスキーマに従った YAML のみで出力してください。
```

---

## 4. 実装状況（2025-11-21 時点）

以下の実装が完了しています:

1. **internal/sandbox 実装**
   - `Executor` / `DockerLocalExecutor` 実装済み。
   - ImagePull 自動実行、Codex 認証自動マウント実装済み。
   - `env:` プレフィックスによるホスト環境変数の解決実装済み。
2. **internal/meta 実装**
   - `plan_task` / `next_action` / `completion_assessment` 実装済み。
   - Exponential Backoff による再試行ロジック実装済み。
   - `system_prompt` オーバーライド機能実装済み。
3. **internal/cli 実装**
   - YAML 読み込み → `TaskContext` 初期化 → Meta → Worker → TaskNote → exit code 返却のフロー実装済み。
   - ループ実行と完了判定ロジック実装済み。

これにより、AgentRunner CLI の基本機能は動作可能な状態です。
