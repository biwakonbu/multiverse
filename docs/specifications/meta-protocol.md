# Meta-agent プロトコル仕様

最終更新: 2025-11-22

## 概要

本ドキュメントは Meta-agent と AgentRunner Core 間の通信プロトコルを定義します。Meta-agent は LLM ベースのエージェントで、YAML メッセージを介して Core とやり取りします。

## 1. Meta-agent の役割

Meta-agent は以下の責務を持ちます：

1. **計画**: PRD から Acceptance Criteria（受け入れ条件）を設計
2. **判断**: 次のアクション（Worker 実行 or 完了）を決定
3. **評価**: タスク完了状況を評価

## 2. プロトコル概要

### 2.1 呼び出し単位

Meta とのやり取りは 3 種類のリクエスト/レスポンスで構成されます：

| プロトコル              | 入力         | 出力                | 用途       |
| ----------------------- | ------------ | ------------------- | ---------- |
| `plan_task`             | PRD テキスト | Acceptance Criteria | タスク計画 |
| `next_action`           | TaskContext  | 次のアクション      | 実行判断   |
| `completion_assessment` | TaskContext  | 完了評価            | 完了判定   |

### 2.2 YAML フォーマット

すべてのメッセージは YAML 形式です。

**共通ルール**:

- 単一ドキュメント（`---` は 1 つまで）
- インデント: 半角スペース 2 個
- トップレベルに `type` フィールド必須

## 3. plan_task プロトコル

### 3.1 目的

PRD を解析し、タスクの受け入れ条件（Acceptance Criteria）を定義します。

### 3.2 入力

Core は以下の情報を Meta に渡します：

- Task YAML（タスク設定）
- PRD テキスト（要件定義）

### 3.3 出力 YAML

```yaml
type: plan_task
acceptance_criteria:
  - id: "AC-1"
    description: "ユーザー登録APIが正常系で 201 を返すこと"
  - id: "AC-2"
    description: "必須項目のバリデーションエラー時に 400 を返すこと"
```

### 3.4 フィールド定義

| フィールド                          | 型     | 必須 | 説明                            |
| ----------------------------------- | ------ | ---- | ------------------------------- |
| `type`                              | string | ✅   | 固定値: `"plan_task"`           |
| `acceptance_criteria`               | array  | ✅   | 受け入れ条件のリスト            |
| `acceptance_criteria[].id`          | string | 推奨 | 受け入れ条件の ID（例: "AC-1"） |
| `acceptance_criteria[].description` | string | ✅   | 受け入れ条件の説明              |

### 3.5 実装例

```go
type PlanTaskResponse struct {
    Type               string                  `yaml:"type"`
    AcceptanceCriteria []AcceptanceCriterion   `yaml:"acceptance_criteria"`
}

type AcceptanceCriterion struct {
    ID          string `yaml:"id"`
    Description string `yaml:"description"`
}
```

## 4. next_action プロトコル

### 4.1 目的

現在のタスク状態を評価し、次のアクション（Worker 実行 or 完了）を決定します。

### 4.2 入力

Core は TaskContext の要約を Meta に渡します：
The transport format between AgentRunner and LLM is **JSON**.
Internally, the `MetaClient` converts this JSON into YAML to maintain compatibility with legacy processing logic before unmarshaling into Go structs.

- **Request**: JSON sent to LLM (via prompts).
- **Response**: JSON string received from LLM.
- **Conversion**: JSON string -> YAML string -> Go Struct.

All structs in `internal/meta/protocol.go` are tagged with both `yaml` and `json` to support this flow.

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
state: "RUNNING"
```

### 4.3 出力 YAML

#### 4.3.1 Worker 実行を要求する場合

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

#### 4.3.2 タスク完了と判断する場合

```yaml
type: next_action
decision:
  action: "mark_complete"
  reason: "全ての受け入れ条件が満たされ、テストも成功したため"
```

### 4.4 フィールド定義

| フィールド                  | 型     | 必須     | 説明                                    |
| --------------------------- | ------ | -------- | --------------------------------------- |
| `type`                      | string | ✅       | 固定値: `"next_action"`                 |
| `decision.action`           | string | ✅       | `"run_worker"` または `"mark_complete"` |
| `decision.reason`           | string | ✅       | 判断理由                                |
| `worker_call`               | object | 条件付き | `action` が `"run_worker"` の場合必須   |
| `worker_call.worker_type`   | string | ✅       | Worker 種別（v1: `"codex-cli"`）        |
| `worker_call.mode`          | string | ✅       | 実行モード（v1: `"exec"`）              |
| `worker_call.prompt`        | string | ✅       | Worker への指示文                       |
| `worker_call.model`         | string | 任意     | 使用するモデル ID                       |
| `worker_call.flags`         | array  | 任意     | CLI フラグのリスト                      |
| `worker_call.env`           | map    | 任意     | 環境変数のマップ                        |
| `worker_call.tool_specific` | map    | 任意     | ツール固有の設定                        |
| `worker_call.use_stdin`     | bool   | 任意     | 標準入力を使用するかどうか              |

### 4.5 実装例

```go
type NextActionResponse struct {
    Type       string              `yaml:"type"`
    Decision   Decision            `yaml:"decision"`
    WorkerCall *WorkerCall         `yaml:"worker_call,omitempty"`
}

type Decision struct {
    Action string `yaml:"action"`
    Reason string `yaml:"reason"`
}

type WorkerCall struct {
    WorkerType string `yaml:"worker_type"`
    Mode       string `yaml:"mode"`
    Prompt     string `yaml:"prompt"`
}
```

## 5. completion_assessment プロトコル

### 5.1 目的

タスク完了時に、Acceptance Criteria の達成状況を評価します。

### 5.2 入力

Core は最終状態の TaskContext を Meta に渡します。

### 5.3 出力 YAML

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

### 5.4 フィールド定義

| フィールド                | 型     | 必須 | 説明                               |
| ------------------------- | ------ | ---- | ---------------------------------- |
| `type`                    | string | ✅   | 固定値: `"completion_assessment"`  |
| `summary`                 | string | ✅   | 完了評価のサマリ                   |
| `details.passed_criteria` | array  | 推奨 | 満たされた受け入れ条件の ID リスト |
| `details.remaining_risks` | array  | 推奨 | 残存リスクのリスト                 |

### 5.5 実装例

```go
type CompletionAssessmentResponse struct {
    Type    string                       `yaml:"type"`
    Summary string                       `yaml:"summary"`
    Details CompletionAssessmentDetails  `yaml:"details"`
}

type CompletionAssessmentDetails struct {
    PassedCriteria  []string `yaml:"passed_criteria"`
    RemainingRisks  []string `yaml:"remaining_risks"`
}
```

## 6. エラーハンドリング

### 6.1 LLM エラー再試行ロジック

v1 実装では、LLM API 呼び出しの信頼性を向上させるため、以下の再試行ロジックを実装しています：

| 項目                    | 設定                                      |
| ----------------------- | ----------------------------------------- |
| **再試行対象エラー**    | HTTP 5xx、タイムアウト、Rate Limit（429） |
| **再試行回数**          | 最大 3 回                                 |
| **Exponential Backoff** | 1 秒 → 2 秒 → 4 秒                        |
| **非再試行エラー**      | HTTP 4xx（400, 401, 403 など）            |

### 6.2 YAML パースエラー

Meta が不正な YAML を返した場合：

1. エラーログを出力
2. Meta に再試行を要求（最大 3 回）
3. 3 回失敗した場合、タスクを FAILED に遷移

### 6.3 タイムアウト

Meta 呼び出しのタイムアウト設定は、使用するプロバイダによって異なります。

#### OpenAI Chat プロバイダ

- デフォルト: 60 秒
- 環境変数 `META_TIMEOUT_SEC` で変更可能

#### Codex CLI プロバイダ

LLM の処理は時間がかかるため、より長いタイムアウトを設定しています。

| 層          | デフォルト値 | 説明                                                |
| ----------- | ------------ | --------------------------------------------------- |
| ChatHandler | 15 分        | `chat/handler.go` の `metaTimeout`                  |
| Meta-agent  | 10 分        | `meta/cli_provider.go` の `DefaultMetaAgentTimeout` |
| agenttools  | 10 分        | `ExecPlan.Timeout` で指定                           |

**タイムアウト階層**:

```
ChatHandler (15分)
  └→ Meta.Decompose()
       └→ CodexCLIProvider (10分)
            └→ agenttools.Execute() (親コンテキストから独立)
```

`agenttools.Execute()` は `ExecPlan.Timeout` が設定されている場合、親コンテキストから独立した新しいコンテキストを作成します。これにより、外部プロセス（Codex CLI）の実行時間を正確に制御できます。

#### Graceful Shutdown

タイムアウト発生時、プロセスは以下の順序で終了されます：

1. **SIGTERM** 送信（graceful shutdown のチャンス）
2. **5 秒待機**（`GracefulShutdownDelay`）
3. **SIGKILL** 送信（強制終了）

これにより、Codex CLI は可能な限りクリーンに終了できます

## 7. プロンプト設計

### 7.1 System Prompt

Meta には以下の System Prompt が設定されます：

````text
あなたはソフトウェア開発タスクを管理するテックリード兼オーケストレータです。

- 与えられたタスクコンテキスト（TaskContext）にもとづき、
  次に何をすべきかを決定する役割を担います。
- 出力は必ず 1 つの YAML ドキュメントのみとします。
- コードブロック（```）や解説文は一切書かないでください。
````

### 7.2 System Prompt のカスタマイズ

Task YAML で `runner.meta.system_prompt` を指定することで、System Prompt を上書きできます：

```yaml
runner:
  meta:
    system_prompt: |
      カスタム System Prompt
```

## 8. 実装状況

### 8.1 実装済み機能

- ✅ `plan_task` プロトコル
- ✅ `next_action` プロトコル
- ✅ `completion_assessment` プロトコル
- ✅ LLM エラー再試行ロジック（Exponential Backoff）
- ✅ System Prompt カスタマイズ
- ✅ YAML パースエラーハンドリング

### 8.2 制約事項

- v1 では OpenAI Chat API のみサポート
- プロトコルバージョニングは未実装（将来拡張予定）

## 9. decompose プロトコル (v2.0)

### 9.1 目的

チャットインターフェースからの自然言語入力（ユーザーの要望）に基づき、タスクをフェーズに分解し、具体的な実行タスク（Acceptance Criteria 含む）を生成します。

### 9.2 入力

Core は以下の情報を Meta に渡します：

- ユーザーの入力メッセージ
- 既存タスクの要約（依存関係解決のため）
- 会話履歴（コンテキスト維持のため）

### 9.3 出力 YAML

```yaml
type: decompose
understanding: |
  ユーザーは「商品一覧ページ」の実装を希望しています。
  既存の API 定義に基づき、フロントエンドの実装が必要です。
phases:
  - name: "実装設計"
    milestone: "design"
    tasks:
      - id: "temp-task-1"
        title: "コンポーネント設計"
        description: "商品カードコンポーネントの Props と State を設計する"
        wbs_level: 2
        estimated_effort: "small"
        acceptance_criteria:
          - "Figma デザインと一致する Props が定義されていること"
        suggested_impl:
          language: "typescript"
          file_paths: ["src/components/ProductCard.svelte"]

  - name: "実装"
    milestone: "implementation"
    tasks:
      - id: "temp-task-2"
        title: "コンポーネント実装"
        description: "設計に基づきコードを実装する"
        dependencies: ["temp-task-1"]
        wbs_level: 3
        estimated_effort: "medium"
potential_conflicts:
  - file: "src/routes/products/+page.svelte"
    tasks: ["TASK-001"]
    warning: "他のタスクで変更中の可能性があります"
```

### 9.4 フィールド定義

| フィールド            | 型     | 必須 | 説明                     |
| :-------------------- | :----- | :--- | :----------------------- |
| `type`                | string | ✅   | 固定値: `"decompose"`    |
| `understanding`       | string | ✅   | ユーザー意図の理解・要約 |
| `phases`              | array  | ✅   | フェーズ別タスクリスト   |
| `potential_conflicts` | array  | 任意 | 潜在的なコンフリクト情報 |

#### 9.4.1 Phase & Task 構造

**Phase**:

| フィールド  | 型     | 必須 | 説明                                 |
| :---------- | :----- | :--- | :----------------------------------- |
| `name`      | string | ✅   | フェーズ名（例: "概念設計", "実装"） |
| `milestone` | string | ✅   | マイルストーン ID                    |
| `tasks`     | array  | ✅   | タスクリスト                         |

**DecomposedTask**:

| フィールド            | 型     | 必須 | 説明                              |
| :-------------------- | :----- | :--- | :-------------------------------- |
| `id`                  | string | ✅   | 一時 ID（依存関係定義用）         |
| `title`               | string | ✅   | タスクタイトル                    |
| `description`         | string | ✅   | 詳細説明                          |
| `acceptance_criteria` | array  | ✅   | 達成条件リスト (string)           |
| `dependencies`        | array  | 任意 | 依存するタスク ID（一時 ID 可）   |
| `wbs_level`           | int    | ✅   | WBS 階層 (1=概念, 2=設計, 3=実装) |
| `estimated_effort`    | string | ✅   | 推定工数 (small/medium/large)     |
| `suggested_impl`      | object | 任意 | 実装ヒント                        |

**SuggestedImpl**:

| フィールド    | 型     | 必須 | 説明             |
| :------------ | :----- | :--- | :--------------- |
| `language`    | string | 任意 | 推奨言語         |
| `file_paths`  | array  | 任意 | 関連ファイルパス |
| `constraints` | array  | 任意 | 実装上の制約     |

### 9.5 実装例

```go
type DecomposeResponse struct {
    Understanding      string              `yaml:"understanding"`
    Phases             []DecomposedPhase   `yaml:"phases"`
    PotentialConflicts []PotentialConflict `yaml:"potential_conflicts"`
}

type DecomposedPhase struct {
    Name      string           `yaml:"name"`
    Milestone string           `yaml:"milestone"`
    Tasks     []DecomposedTask `yaml:"tasks"`
}

type DecomposedTask struct {
    ID                 string         `yaml:"id"`
    Title              string         `yaml:"title"`
    Description        string         `yaml:"description"`
    AcceptanceCriteria []string       `yaml:"acceptance_criteria"`
    Dependencies       []string       `yaml:"dependencies"`
    WBSLevel           int            `yaml:"wbs_level"`
    EstimatedEffort    string         `yaml:"estimated_effort"`
    SuggestedImpl      *SuggestedImpl `yaml:"suggested_impl,omitempty"`
}
```
