# AgentRunner 仕様書（v1 / MVP）

最終更新: 2025-11-20

---

## 0. 位置づけ

本ドキュメントは、AgentRunner の「最新仕様」として、これまでの議論内容を統合・整理したものです。  
過去の仕様書は本ドキュメントに置き換える前提とします。

- 実装方針:
  - まずは **MVP としての最小スライス** を実装し、課題を炙り出しながら精度を上げる。
  - 構成はシンプルかつ拡張可能に保つ。
- 入力インターフェース:
  - **CLI から 1 枚の YAML を stdin で受け取る** ことだけを前提とする。
  - CLI オプションによる細かい設定は行わない。
- 永続化:
  - タスクの実行中状態はメモリ上で管理。
  - タスク完了後に **Task Note を Markdown でリポジトリ内に出力**する。
  - DB は利用しない（v1）。

---

## 1. 目的とスコープ

### 1.1 目的

AgentRunner は、以下を目的とする「エージェント・オーケストレーション・ランナー」です。

- PRD（要件定義）を入力として受け取り、
- Meta エージェントによる計画・評価と、
- Worker エージェント（例: Codex CLI）によるコード編集・テスト実行を組み合わせて、
- 1 つの開発タスクを **半自動的〜自動的に完遂させる**。

最終成果物としては:

- リポジトリへのコード変更（コミットは別レイヤの責務）
- 実行ログを整理した **Task Note（Markdown）**

を生成する。

### 1.2 スコープ（v1 / MVP）

v1 では以下の範囲に限定して実装する。

- Meta:
  - OpenAI Chat モデル（例: gpt-5.1）を 1 つだけ利用。
  - Meta からの I/F は `plan_task` / `next_action` / `completion_assessment` の 3 種類の YAML。
- Worker:
  - Worker 種別は `codex-cli` の 1 種のみを想定。
  - Codex CLI への具体的なオプションは実 CLI ドキュメントに基づいて実装時に調整する（後述 1.3）。
- 実行フロー:
  - 1 タスクにつき 1 回以上の Worker 実行を許容する（ループ対応済み）。
  - `max_loops` 設定により、Meta エージェントの判断で複数回の Worker 実行と再計画が可能。

### 1.3 前提・外部依存（事実 / 仮説の分離）

- 事実:
  - AgentRunner 自体の I/F（本仕様）は、本ドキュメントで定義した内容に従う。
  - 入力は YAML、出力は Markdown + リポジトリ変更。
- 仮説 / 要実機検証:
  - `codex-cli` のコマンドラインオプション例（`codex exec --sandbox ...` 等）は、実際の CLI 仕様を参照して調整が必要。
  - Docker イメージ名や、コンテナ内パスは **設計上のデフォルト案** であり、プロジェクト固有に変える可能性がある。

実装時には Codex CLI 等の外部ツールのドキュメントを参照し、本仕様との整合を確認した上でコマンドテンプレートを確定させる。

---

## 2. 全体アーキテクチャ

### 2.1 コンポーネント構成

AgentRunner は概ね以下のコンポーネントから構成される。

- CLI:
  - `agent-runner run`
  - stdin から YAML を受け取り、Core を起動するエントリポイント。
- Core:
  - 主オーケストレータ。
  - YAML のパース / TaskContext の構築 / Meta 呼び出し / Worker 実行 / FSM 管理 / Task Note 出力を担当。
- Meta Adapter:
  - Meta エージェント（LLM）とやり取りする層。
  - Core → Meta へのプロンプト構築、Meta 応答の YAML パースを行う。
- Worker Executor:
  - Docker サンドボックス内で Worker を実行する層。
  - Codex CLI などの実 CLI を子プロセスとして起動し、結果を収集する。
- Sandbox Manager:
  - Worker 実行のための Docker コンテナの起動 / 終了・ボリュームマウントなどの設定を行う。
  - v1 では Core 内の一モジュールとして実装してもよい（外部サービスにしない）。
- Task Note Writer:
  - TaskContext の情報をもとに、Markdown の Task Note を生成しファイルに出力する。

### 2.2 データフロー概要

1. ユーザは Task YAML を作成し、`agent-runner run < task.yaml` を実行する。
2. CLI は stdin の YAML を読み込み、Core に渡す。
3. Core は YAML を `TaskConfig` にパースし、デフォルト値を補完して `TaskContext` を構築する。
4. Core は PRD テキストを読み込み（ファイル or インライン）、Meta Adapter 経由で `plan_task` を実行する。
5. Meta から受け取った `acceptance_criteria` を TaskContext に保存する。
6. Core は Meta に対して `next_action` を問い合わせ、`run_worker` か `mark_complete` かを判断する。
7. `run_worker` の場合、Core は Worker Executor に Worker 実行を依頼し、Docker サンドボックス内で Codex CLI を起動する。
8. Worker 実行が完了したら、結果（exit code / 実行ログ / 要約など）を TaskContext に記録する。
9. （MVP ではシンプルなルールに基づいて）テスト実行 / 成否判定を行う。
10. タスク完了と判定されたら、Task Note Writer が TaskContext から Markdown を生成し、リポジトリ直下の `.agent-runner/task-<task_id>.md` に書き出す。
11. プロセス終了。

### 2.3 FSM（ステートマシン）概要（v1）

v1 の Task FSM は以下の最小パスを想定する。

- `PENDING`:
  - YAML を読み込み TaskContext を構築した直後の状態。
- `PLANNING`:
  - Meta に `plan_task` を投げている状態。
- `RUNNING`:
  - Worker 実行中。
- `VALIDATING`:
  - テスト実行および完了判定を行っている状態。
- `COMPLETE`:
  - 成功終了。
- `FAILED`:
  - Meta or Worker or テストが致命的に失敗し、タスクを終了した状態。

v1 ではループを最小限とし、以下のような直線的な遷移から始める。

`PENDING` → `PLANNING` → `RUNNING` → `VALIDATING` → `COMPLETE/FAILED`

将来的に `RUNNING` ↔ `PLANNING` の再計画ループや `ON_HOLD` 等を拡張可能とするが、MVP 実装では必須ではない。

---

## 3. CLI / YAML 仕様

### 3.1 CLI 仕様

- コマンド:
  - `agent-runner run`
- 入力:
  - stdin から 1 枚の Task YAML を受け取る。
- オプション:
  - v1 ではコマンドラインオプションは導入しない。
- 出力:
  - stdout:
    - 実行ログ（人間が読む用の簡易ログ。JSON 等ではなくテキストでよい）
  - ファイル出力:
    - Task Note: `<repo>/.agent-runner/task-<task_id>.md`

使用例:

```bash
agent-runner run < task.yaml
```

### 3.2 Task YAML スキーマ（v1）

#### 3.2.1 全体構造

```yaml
version: 1

task:
  id: "TASK-123" # 任意。未指定ならランナー側で採番
  title: "ユーザ登録 API の実装" # 任意。Task Note のヘッダなどに利用
  repo: "." # 任意。作業対象リポジトリのパス（デフォルト "."）

  prd:
    path: "./docs/TASK-123.md" # PRD をファイルから読む場合
    # text: |                    # もしくは PRD 本文を直接埋め込む場合
    #   ここに PRD 本文...

  test:
    command: "npm test" # 任意。自動テストを走らせるコマンド
    # cwd: "./"                  # 任意。テスト実行ディレクトリ（デフォルトは repo）

runner:
  meta:
    kind: "openai-chat" # v1 は固定想定
    model: "gpt-5.1" # 任意。未指定ならデフォルトモデルを利用
    # system_prompt: |           # 任意。Meta 用 system prompt を上書きしたい場合

  worker:
    kind: "codex-cli" # v1 は "codex-cli" 固定
    # docker_image: ...          # 任意。デフォルトイメージを上書きする場合
    # max_run_time_sec: 1800     # 任意。1 回の Worker 実行タイムアウト
    # env:
    #   CODEX_API_KEY: "env:CODEX_API_KEY"  # "env:" 接頭辞ならホスト環境変数を参照
```

#### 3.2.2 必須フィールド

v1 で「実際に必須とする」のは次の項目のみとする。

- `version`
  - 値: `1`
- `task.prd` のいずれか:
  - `task.prd.path` または
  - `task.prd.text`

それ以外は未指定の場合にデフォルトで補完される。

#### 3.2.3 デフォルト補完ルール

- `task.id`
  - 未指定の場合、UUID 等で自動採番（形式は実装依存）。
- `task.title`
  - 未指定の場合、`task.id` をタイトルとして利用してよい。
- `task.repo`
  - 未指定の場合 `"."`（カレントディレクトリ）とみなす。
- `task.test`
  - 未指定の場合、テスト自動実行は行わず、Meta / Worker に任せる。
- `runner.meta.kind`
  - 未指定の場合 `"openai-chat"`。
- `runner.meta.model`
  - 未指定の場合、設定ファイル or 環境変数経由のデフォルトモデルを用いる。
- `runner.worker.kind`
  - 未指定の場合 `"codex-cli"`。
- `runner.worker.docker_image`
  - 未指定の場合、事前に設定されたデフォルト Docker イメージ（例: `ghcr.io/<org>/agent-runner-codex:latest`）を利用。
- `runner.worker.max_run_time_sec`
  - 未指定の場合、実装側のデフォルト（例: 1800 秒）。

---

## 4. Meta インターフェース仕様

Meta は LLM ベースのエージェントで、Core との間で「YAML メッセージ」を交換する。

### 4.1 役割

- PRD を解析し、タスクの受け入れ条件 (`acceptance_criteria`) を定義する。
- 実行ステップごとに「次に何をするべきか」（`next_action`）を決定する。
- （必要に応じて）完了評価 (`completion_assessment`) を生成する。

### 4.2 呼び出し単位

Meta とのやり取りは次の 3 種類のリクエスト/レスポンスで構成される。

1. `plan_task`:
   - 入力: PRD テキスト / タスクメタ情報。
   - 出力: `acceptance_criteria` のリスト。
2. `next_action`:
   - 入力: TaskContext（これまでの Worker 実行履歴など）。
   - 出力: 次のアクション (`run_worker` or `mark_complete`)。
3. `completion_assessment`（v1 では必須ではない）:
   - 入力: 最終状態の TaskContext。
   - 出力: 完了評価コメントや、残課題の一覧など。

#### 4.2.1 LLM エラー再試行ロジック（実装済み）

v1 実装では、LLM API 呼び出しの信頼性を向上させるため、以下の再試行ロジックを実装している：

- **再試行対象エラー**: HTTP 5xx、タイムアウト、Rate Limit（429）
- **再試行回数**: 最大 3 回
- **Exponential Backoff**: 1 秒 → 2 秒 → 4 秒
- **非再試行エラー**: HTTP 4xx（400, 401, 403 など）

これにより、一時的なネットワークエラーや API の過負荷による失敗を自動的に回復し、タスク実行の成功率を向上させている。

### 4.3 `plan_task` 出力 YAML

#### 4.3.1 概要

Meta は以下のような YAML を返す。

```yaml
type: plan_task
acceptance_criteria:
  - id: "AC-1"
    description: "ユーザー登録APIが正常系で 201 を返すこと"
  - id: "AC-2"
    description: "必須項目のバリデーションエラー時に 400 を返すこと"
```

#### 4.3.2 フィールド定義

- `type`:
  - 固定値: `"plan_task"`
- `acceptance_criteria`:
  - 配列。各要素は:
    - `id`: string（任意だが生成を推奨）
    - `description`: string（人間が理解できる自然言語）

### 4.4 `next_action` 出力 YAML

#### 4.4.1 概要

Meta は以下の 2 パターンのいずれかを返す。

1. Worker 実行を要求する場合:

```yaml
type: next_action
decision:
  action: "run_worker"
  reason: "まだ実装が行われていないため"

worker_call:
  worker_type: "codex-cli"
  mode: "exec"
  prompt: |
    ここに Codex に渡すべき指示文（自然言語 + 手順）が入る
```

2. タスク完了と判断する場合:

```yaml
type: next_action
decision:
  action: "mark_complete"
  reason: "全ての受け入れ条件が満たされ、テストも成功したため"
```

#### 4.4.2 フィールド定義

- `type`:
  - 固定値: `"next_action"`
- `decision`:
  - `action`: `"run_worker"` or `"mark_complete"`
  - `reason`: string（人間向けの理由）
- `worker_call`:
  - `action` が `"run_worker"` の場合のみ必須。
  - フィールド:
    - `worker_type`: string（v1 では `"codex-cli"` を想定）
    - `mode`: string（v1 では `"exec"` 固定）
    - `prompt`: string（Worker に渡すべき自然言語の指示文）

### 4.5 `completion_assessment` 出力 YAML

v1 で実装済み。Meta エージェントはタスク完了時に以下のような評価を出力する。

```yaml
type: completion_assessment
summary: |
  ユーザー登録APIの実装は完了しており、以下の受け入れ条件を満たしています。
details:
  passed_criteria:
    - "AC-1"
    - "AC-2"
  remaining_risks:
    - "性能テストは未実施"
```

---

## 5. Worker 実行仕様

### 5.1 役割

Worker Executor は Meta の `worker_call` に従い、Sandbox 上で Worker（Codex CLI 等）を実行し、その結果を Core に返す。

### 5.2 v1 Worker 種別: `codex-cli`

- `runner.worker.kind` が `"codex-cli"` の場合、
  - 事前に定義された Docker イメージ内で Codex CLI を起動する。
- Codex CLI の具体的なオプションや JSON 出力形式は、**実ツール仕様に依存**するため、実装時に調整が必要。

本ドキュメントでは以下の程度までを「設計上のテンプレ」として示す（仮の例）:

```bash
codex exec   --sandbox workspace-write   --json   --cwd /workspace/project   "<Meta から渡された prompt>"
```

#### 5.2.1 コンテナライフサイクル最適化（実装済み）

v1 実装では、パフォーマンス最適化のため、以下のコンテナライフサイクル管理を採用している：

- **タスク開始時**: 1 回だけコンテナを起動（`WorkerExecutor.Start()`）
- **Worker 実行時**: 既存コンテナ内で `docker exec` を実行（`WorkerExecutor.RunWorker()`）
- **タスク完了時**: コンテナを停止（`WorkerExecutor.Stop()`）

この設計により、Worker 実行ごとにコンテナを起動・停止する場合と比較して、5-10 倍の高速化を実現している。

### 5.3 サンドボックス（Docker）仕様（案）

- デフォルト Docker イメージ:
  - 例: `ghcr.io/<org>/agent-runner-codex:latest`
  - v1 実装: `ghcr.io/biwakonbu/agent-runner-codex:latest`
- コンテナ内パス:
  - プロジェクト: `/workspace/project`
- マウント:
  - ホストの `task.repo` パスを `/workspace/project` にマウント。
  - Codex 認証情報や設定ディレクトリ（例: `~/.codex`）を read-only / read-write でマウントする。
- 環境変数:
  - `runner.worker.env` の値をコンテナ内に注入。
  - `"env:XXX"` 形式はホストの環境変数 `XXX` を参照して値を設定。

#### 5.3.1 ImagePull 自動実行（実装済み）

v1 実装では、Docker イメージが存在しない場合、自動的に `docker pull` を実行する。これにより、初回実行時のエラーを防ぎ、ユーザーが手動でイメージを取得する必要がなくなる。

#### 5.3.2 Codex 認証自動マウント（実装済み）

v1 実装では、以下の順序で Codex 認証情報を自動的に検出・設定する：

1. `~/.codex/auth.json` が存在する場合、read-only でコンテナ内の `/root/.codex/auth.json` にマウント
2. `~/.codex/auth.json` が存在しない場合、環境変数 `CODEX_API_KEY` をコンテナ内に注入

これにより、ユーザーは認証情報を手動で設定する必要がなく、既存の Codex CLI 環境をそのまま利用できる。

### 5.4 Worker 実行結果（内部表現）

Core が扱う Worker 実行結果の内部表現を以下のように定義する。

```go
type WorkerRunResult struct {
    ID          string    // ラン毎の ID（UUID 等）
    StartedAt   time.Time
    FinishedAt  time.Time
    ExitCode    int
    RawOutput   string    // CLI の stdout/stderr（まとめてもよい）
    Summary     string    // 必要に応じて要約
    Error       error     // 実行エラー（起動失敗など）
}
```

TaskContext には `[]WorkerRunResult` として記録しておく。

---

## 6. TaskContext と Task Note

### 6.1 TaskContext 概要

Core がタスク実行中に保持する状態を `TaskContext` と呼ぶ。

#### 6.1.1 フィールド例

```go
type TaskContext struct {
    ID        string        // task.id
    Title     string        // task.title
    RepoPath  string        // task.repo の絶対パス
    State     TaskState     // FSM の現状態

    PRDText   string        // PRD 本文（ファイル読み込み or text の統合結果）

    AcceptanceCriteria []AcceptanceCriterion // Meta plan_task の結果
    MetaCalls          []MetaCallLog         // Meta 呼び出し履歴
    WorkerRuns         []WorkerRunResult     // Worker 実行履歴

    TestConfig *TestSpec   // task.test
    TestResult *TestResult // 実行した場合

    StartedAt  time.Time
    FinishedAt time.Time
}
```

`AcceptanceCriterion`, `MetaCallLog`, `TestResult` 等は実装側で構造を定義する。

### 6.2 Task Note 出力仕様

#### 6.2.1 出力パス

- パス:
  - `<repo>/.agent-runner/task-<task_id>.md`
- ディレクトリ:
  - `.agent-runner` ディレクトリが存在しない場合は作成する。

#### 6.2.2 Markdown テンプレート（例）

以下は v1 で利用する Task Note のテンプレ例である。必要に応じてフィールドを増減する。

````markdown
# Task Note - {{ .ID }} {{ if .Title }}- {{ .Title }}{{ end }}

- Task ID: {{ .ID }}
- Title: {{ .Title }}
- Started At: {{ .StartedAt }}
- Finished At: {{ .FinishedAt }}
- State: {{ .State }}

---

## 1. 概要

{{ .Summary }}

---

## 2. PRD 概要

{{ .PRDSummary }}

<details>
<summary>PRD 原文</summary>

```text
{{ .PRDText }}
```
````

</details>

---

## 3. 受け入れ条件 (Acceptance Criteria)

{{ range .AcceptanceCriteria }}

- [{{ if .Passed }}x{{ else }} {{ end }}] {{ .ID }}: {{ .Description }}
  {{ end }}

---

## 4. 実行ログ (Meta / Worker)

### 4.1 Meta Calls

{{ range .MetaCalls }}

#### {{ .Type }} at {{ .Timestamp }}

```yaml
{ { .RequestYAML } }
```

```yaml
{ { .ResponseYAML } }
```

{{ end }}

### 4.2 Worker Runs

{{ range .WorkerRuns }}

#### Run {{ .ID }} (ExitCode={{ .ExitCode }}) at {{ .StartedAt }} - {{ .FinishedAt }}

```text
{{ .RawOutput }}
```

{{ end }}

---

## 5. テスト結果

{{ if .TestResult }}

- Command: `{{ .TestResult.Command }}`
- ExitCode: {{ .TestResult.ExitCode }}
- Summary: {{ .TestResult.Summary }}

```text
{{ .TestResult.RawOutput }}
```

{{ else }}
テストは自動実行されませんでした。
{{ end }}

---

## 6. メモ / 残課題

{{ .Notes }}

```

実装では `text/template` 等を用いて TaskContext からこのテンプレートを埋めて出力する。

---

## 7. MVP v1 の制約と実装スコープ

### 7.1 実装スコープ（必須）

- CLI:
  - `agent-runner run` のみ。
  - stdin YAML 読み込み → TaskConfig パース → TaskContext 構築。
- Meta:
  - `plan_task` と `next_action` を呼び出すコードパス。
  - `completion_assessment` による完了判定と AC 評価。
  - `acceptance_criteria` と `next_action` のパース。
- Worker:
  - Docker 上で `codex-cli`（仮）を実行する Worker Executor。
  - 実行結果（exit code, stdout/stderr）を TaskContext に記録。
  - コンテナライフサイクル最適化（タスク開始時に起動、終了時に停止）。
- テスト:
  - v1 では `task.test.command` が指定されていない場合は自動テストを行わない。
  - 指定されている場合も、まずは Worker 内での実行（Meta の prompt で指示）に任せる想定とし、Core 側で別途テストプロセスを起動するかどうかは後続で検討。
- Task Note:
  - `.agent-runner/task-<task_id>.md` の生成。

### 7.2 後続拡張の候補（v1 では未実装）

- Core 側でテストコマンドを独立して実行し、その結果を Meta にフィードバックするループ。
- 複数 Worker サポート（Cursor CLI, Claude Code 等）。
- Web UI / ダッシュボードからのタスク起動・モニタリング。
- TaskContext の永続化（DB or KV）と `resume` 機能。

---

## 8. 実装開始のための最小タスク

1. Task YAML 用の struct（`TaskConfig`）を定義し、stdin から YAML を読み込んでパースする。
2. デフォルト補完ロジックを実装し、`TaskContext` を構築する。
3. PRD テキストの読み込みロジック（`prd.path` or `prd.text`）を実装する。
4. Meta Adapter を実装し、`plan_task` / `next_action` のやり取り（LLM 呼び出し + YAML パース）を行う。
5. Worker Executor を実装し、Docker 上で codex-cli を 1 回実行できるようにする。
6. TaskContext から Task Note を生成するテンプレートと、ファイル出力ロジックを実装する。
7. 簡易なログ出力（stdout）とエラー処理を実装する。

この範囲が完了すれば、YAML 一枚から実際に Worker（Codex）を動かし、Task Note を残すところまで「一通りの体験」が成立する。

---
```
