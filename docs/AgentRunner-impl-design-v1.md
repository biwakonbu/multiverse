# AgentRunner v1 実装設計（Go / Sandbox / Meta プロンプト）

## 1. Go プロジェクト構成・パッケージ分割

### 1-1. ルート構成

```text
agent-runner/
  cmd/
    agent-runner/
      main.go

  internal/
    cli/
    config/
    task/
    meta/
    sandbox/
    worker/
    tasknote/
    llm/
    logging/

  go.mod
  go.sum
  README.md
```

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
  - `plan_task` / `next_action` 用の呼び出し I/F
  - プロンプトテンプレートの組み立て
  - YAML 応答のパース・検証

```go
type Service interface {
    PlanTask(ctx context.Context, tc *task.TaskContext) (*PlanTaskResult, error)
    NextAction(ctx context.Context, tc *task.TaskContext) (*NextActionResult, error)
}
```

実際の LLM 呼び出しは `internal/llm` に委譲。

#### internal/llm

- OpenAI などの具体的な LLM クライアント。
- 主な責務
  - ChatCompletion API 呼び出し
  - 将来的にプロバイダを差し替えるための抽象化

```go
type Client interface {
    Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
}
```

#### internal/sandbox

- サンドボックス実行（Docker / 将来の Kubernetes 等）の抽象化。
- 主な責務
  - `SandboxExecutor` インターフェース
  - `DockerLocalExecutor` 実装（v1）
  - Request/Result/Errors の定義

（詳細は後述の §2 で定義）

#### internal/worker

- Worker ごとの実行ロジック。
- v1 では `codex-cli` のみ。
- 主な責務
  - Meta の `worker_call` 情報をもとに SandboxExecutor を呼び出す
  - 実行結果（stdout/stderr/exit code）を `TaskContext` に反映

```go
type Runner interface {
    RunWorker(ctx context.Context, tc *task.TaskContext, call WorkerCall) (*WorkerResult, error)
}
```

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
package sandbox

import (
    "context"
    "time"
)

type Mount struct {
    Source   string // ホスト側パス（絶対パス）
    Target   string // コンテナ側パス
    ReadOnly bool
}

type ResourceLimit struct {
    CPUQuota    string        // 例: "0.5"（0.5 vCPU）や ""（未指定）
    MemoryLimit string        // 例: "1g"（1 GiB）や ""（未指定）
    Timeout     time.Duration // 実行タイムアウト
}

type Request struct {
    Image   string            // 必須。例: "co-routine/agent-worker-codex:latest"
    Command []string          // 必須。例: {"codex", "exec", "--non-interactive", "..."}
    Env     map[string]string // 任意。KEY=VALUE

    Workdir string  // コンテナ内の作業ディレクトリ。例: "/workspace/project"
    Mounts  []Mount // リポジトリ等のマウント

    Resource ResourceLimit
}

type Result struct {
    ExitCode  int
    Stdout    []byte
    Stderr    []byte
    StartedAt time.Time
    EndedAt   time.Time
}

type ErrorKind string

const (
    ErrorUnknown        ErrorKind = "unknown"
    ErrorTimeout        ErrorKind = "timeout"
    ErrorInfra          ErrorKind = "infra"           // Docker daemon エラー等
    ErrorInvalidRequest ErrorKind = "invalid_request" // Request の不備
)

type Error struct {
    Kind  ErrorKind
    Op    string // 例: "docker_run"
    Cause error
}

func (e *Error) Error() string {
    // Kind / Op / Cause を含めたメッセージ
    return ...
}

func (e *Error) Unwrap() error {
    return e.Cause
}

type Executor interface {
    Run(ctx context.Context, req *Request) (*Result, error)
}
```

- 実装側の方針
  - **正常系**:
    - コンテナが起動し `ExitCode` が得られたら `error == nil` で返す（ExitCode != 0 でも）。
    - 呼び出し側は ExitCode を見て Worker 成功/失敗を判断する。
  - **異常系**:
    - Docker 自体の異常（Timeout / daemon 死亡 / ネットワークエラーなど）は `*sandbox.Error` を返す。
    - `ErrorKind` により上位がメッセージ・再試行可否を判断可能。

### 2-2. DockerLocalExecutor 実装

```go
type DockerLocalExecutor struct {
    Logger        *slog.Logger
    DockerBinPath string // 例: "docker"
    DefaultImage  string // Request.Image が空のときのデフォルト
}
```

`Run` の実装方針：

1. `Request` のバリデーション
   - `Image` が空なら `DefaultImage` を使用。
   - `Command` が空なら `ErrorInvalidRequest` でエラー。
2. `docker run` 用の引数を組み立てる
   - ベース:

     ```sh
     docker run --rm
       --workdir <Workdir>
       -v <src>:<dst>:ro|rw
       -e KEY=VALUE
       --network=none
       <image> <command...>
     ```

   - ResourceLimit
     - `Timeout` は Go 側の `context` で管理（`exec.CommandContext`）
     - `CPUQuota` / `MemoryLimit` は `--cpus` / `--memory` などにマッピング（必要なら）

3. `exec.CommandContext` で実行
   - `cmd := exec.CommandContext(ctx, e.DockerBinPath, args...)`
   - `stdout` / `stderr` は `bytes.Buffer` にバッファリング（v1 はストリーミングしない）
4. 実行結果のマッピング
   - `cmd.Run()` の戻り値
     - `ctx` の締め切りで `context.DeadlineExceeded` -> `ErrorTimeout`
     - それ以外の `*exec.ExitError` は「正常系」（ExitCode に反映）
     - その他のエラーは `ErrorInfra`
   - `Result` に
     - `ExitCode`
     - `Stdout` / `Stderr`
     - `StartedAt` / `EndedAt`

ログ方針：

- `sandbox` パッケージ内では
  - DEBUG レベルで `docker run` の引数概要をログ
  - ERROR レベルで `ErrorInfra` / `ErrorInvalidRequest` の詳細をログ
- `stdout` / `stderr` の中身はログには書かない（サイズ膨張を防ぐため）。呼び出し側が Task Note に載せる。

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

```text
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
```

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

```text
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
```

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

## 4. 次にやるべきこと（実装タスク候補）

この設計を前提に、他のコーディングエージェントに渡すタスクとしては:

1. **internal/sandbox 実装**
   - 上記インターフェースどおり `Executor` / `DockerLocalExecutor` / `Error` を実装。
2. **internal/meta 実装**
   - `planTaskSystemPrompt` / `nextActionSystemPrompt` を定数として埋め込み。
   - `TaskSpec` / `TaskContext` からテンプレートを埋める関数を実装。
   - 応答 YAML をパースし、構造体にマッピング。
3. **internal/cli 実装**
   - YAML 読み込み → `TaskContext` 初期化 → Meta → Worker → TaskNote → exit code 返却のひととおりのフロー。

ここまで実装すれば、最低限の「ワンショット AgentRunner CLI」が動く状態になります。
