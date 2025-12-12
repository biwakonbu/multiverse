# Meta Package - Meta-agent 通信層

このパッケージはMeta-agent（LLM）との通信を管理します。OpenAI API、YAML プロトコル、モック実装を含みます。

## 概要

- **client.go**: Meta-agent通信クライアント、LLM呼び出し、リアル/モック実装の切り替え
- **protocol.go**: YAML メッセージプロトコル定義（構造体）

## YAML プロトコル仕様

Meta-agentとの通信はすべて YAML形式で行われます。

### 共通メッセージラッパー

```yaml
type: <message_type>
version: 1
payload: <type-specific-payload>
```

```go
type MetaMessage struct {
    Type    string      `yaml:"type"`    // "plan_task" または "next_action"
    Version int         `yaml:"version"` // 1
    Payload interface{} `yaml:"payload"` // 型によって異なる
}
```

### 1. plan_task - タスク計画フェーズ

**用途**: PRDからAcceptanceCriteria（AC）を生成

**Requestフロー**:
```
Runner.PlanTask(ctx, prdText string)
  → Client.PlanTask() 実装選択
  → callLLM() 呼び出し
```

**System Prompt**:
```
You are a Meta-agent that plans software development tasks.
Your goal is to read a PRD and break it down into Acceptance Criteria.
Output MUST be a YAML block with the following structure:
type: plan_task
version: 1
payload:
  task_id: "TASK-..."
  acceptance_criteria:
    - id: "AC-1"
      description: "..."
      type: "e2e"
      critical: true
```

**Expected Response YAML**:
```yaml
type: plan_task
version: 1
payload:
  task_id: "TASK-001"
  acceptance_criteria:
    - id: "AC-1"
      description: "ユーザーが新規アカウントを作成できる"
      type: "e2e"
      critical: true
    - id: "AC-2"
      description: "ログイン機能が動作する"
      type: "unit"
      critical: true
```

**Response構造体**:
```go
type PlanTaskResponse struct {
    TaskID             string                `yaml:"task_id"`
    AcceptanceCriteria []AcceptanceCriterion `yaml:"acceptance_criteria"`
}

type AcceptanceCriterion struct {
    ID          string `yaml:"id"`          // AC一意識別子（AC-1, AC-2等）
    Description string `yaml:"description"` // AC説明
    Type        string `yaml:"type"`        // "e2e", "unit", "integration"等
    Critical    bool   `yaml:"critical"`    // 重要度（true=必須、false=オプション）
    Passed      bool   `yaml:"passed"`      // 完了状態（context summaryで使用）
}
```

**パース処理** (client.go:96-143):
```go
// 1. LLM呼び出し
resp, err := c.callLLM(ctx, systemPrompt, userPrompt)

// 2. YAML パース to MetaMessage
var msg MetaMessage
yaml.Unmarshal([]byte(resp), &msg) // msg.Type = "plan_task"

// 3. Payload を PlanTaskResponse に再パース
payloadBytes, _ := yaml.Marshal(msg.Payload)
var plan PlanTaskResponse
yaml.Unmarshal(payloadBytes, &plan)

return &plan, nil
```

**エラーハンドリング**:
- LLM呼び出しエラー → fmt.Errorf("...") でRunner に返す
- YAML パースエラー → Runner が StateFailed に遷移
- モック時 → 常に成功（テスト用）

### 2. next_action - 実行ループフェーズ

**用途**: 現在のタスク状態から次のアクション（Worker実行 or 完了）を決定

**Requestフロー**:
```
Runner.Run() 実行ループ内
  → Meta.NextAction(ctx, taskSummary)
  → Decision + WorkerCall（if "run_worker"）を取得
```

**TaskSummary** (プロトコル.go:56-62):
```go
type TaskSummary struct {
    Title              string                  // タスクタイトル
    State              string                  // "PENDING", "PLANNING", "RUNNING"等
    AcceptanceCriteria []AcceptanceCriterion   // 現在のAC状態（PassedフラグはMeta-agentで更新）
    WorkerRunsCount    int                     // これまでのWorker実行回数
}
```

**System Prompt**:
```
You are a Meta-agent that orchestrates a coding task.
Decide the next action based on the current context.
Output MUST be a YAML block with type: next_action.
```

**Expected Response YAML**:
```yaml
type: next_action
version: 1
payload:
  decision:
    action: "run_worker"  # または "mark_complete", "ask_human", "abort"
    reason: "AC-1を実装するため、Worker実行が必要"
  worker_call:
    worker_type: "codex-cli"
    mode: "exec"
    prompt: "以下のACを実装してください:\n- ユーザー登録機能を実装\n..."
```

**Response構造体**:
```go
type NextActionResponse struct {
    Decision   Decision   `yaml:"decision"`           // 決定内容
    WorkerCall WorkerCall `yaml:"worker_call,omitempty"` // 決定="run_worker"時のみ
}

type Decision struct {
    Action string `yaml:"action"` // "run_worker" | "mark_complete" | "ask_human" | "abort"
    Reason string `yaml:"reason"` // 決定理由（デバッグ用）
}

type WorkerCall struct {
    WorkerType string `yaml:"worker_type"` // "codex-cli"
    Mode       string `yaml:"mode"`        // "exec"（現在のみ）
    Prompt     string `yaml:"prompt"`      // Worker に渡すプロンプト
}
```

**パース処理** (client.go:145-190):
```go
// LLM呼び出し、MetaMessage パース、Payload再パース
// plan_taskと同様の流れ
```

**Decision別の動作** (runner.go:120-141):
```
- "run_worker": Worker.RunWorker(prompt) を実行、結果をWorkerRuns[]に記録
- "mark_complete": ループ脱出、タスク完了
- "ask_human" / "abort" / その他: エラー（未実装）
```

## LLM 通信詳細（callLLM）

**実装** (client.go:56-94):

```go
func (c *Client) callLLM(ctx context.Context, systemPrompt, userPrompt string) (string, error)
```

### OpenAI Chat Completion API

**エンドポイント**: `https://api.openai.com/v1/chat/completions`

**リクエスト構造**:
```go
type chatRequest struct {
    Model    string    `json:"model"`    // "gpt-4-turbo"（デフォルト）
    Messages []message `json:"messages"`
}

type message struct {
    Role    string `json:"role"`    // "system" / "user" / "assistant"
    Content string `json:"content"`
}
```

**メッセージ送信順序**:
1. System Message: `role="system"`, Content=systemPrompt
   - Meta-agentの役割と出力フォーマット指示
2. User Message: `role="user"`, Content=userPrompt
   - PRDテキストまたはタスク状態要約

**レスポンス処理**:
```go
type chatResponse struct {
    Choices []struct {
        Message message `json:"message"`
    } `json:"choices"`
}
```

**エラーハンドリング**:
```go
if resp.StatusCode != 200 {
    // OpenAI APIエラー
    return "", fmt.Errorf("OpenAI API error: %s %s", resp.Status, body)
}
if len(result.Choices) == 0 {
    return "", fmt.Errorf("no choices returned from LLM")
}
```

**タイムアウト**: 60秒（client.go:30）

## モック実装

**目的**: テスト時にLLM APIを呼び出さずに動作確認

**有効化方法**:
```yaml
runner:
  meta:
    kind: "mock"  # "openai-chat" ではなく "mock"
    model: "..."  # 無視される
```

### plan_task モック (client.go:97-104)

```go
if c.kind == "mock" {
    return &PlanTaskResponse{
        TaskID: "TASK-MOCK",
        AcceptanceCriteria: []AcceptanceCriterion{
            {ID: "AC-1", Description: "Mock AC 1", Type: "mock", Critical: true},
        },
    }, nil
}
```

**特徴**:
- 常に固定のAC（AC-1）を返す
- APIエラーなし

### next_action モック (client.go:146-161)

```go
if c.kind == "mock" {
    if taskSummary.WorkerRunsCount == 0 {
        // 初回: Worker実行
        return &NextActionResponse{
            Decision: Decision{Action: "run_worker", Reason: "Mock run"},
            WorkerCall: WorkerCall{
                WorkerType: "codex-cli",
                Mode:       "exec",
                Prompt:     "echo 'Hello from Mock Worker'",
            },
        }, nil
    }
    // 2回目以降: 完了
    return &NextActionResponse{
        Decision: Decision{Action: "mark_complete", Reason: "Mock complete"},
    }, nil
}
```

**特徴**:
- WorkerRunsCount=0: run_workerを返す
- WorkerRunsCount≥1: mark_completeを返す
- 常に2ループで完了

**使用シーン**:
- 統合テスト (test/integration/run_flow_test.go)
- ローカル開発でAPIキー不要
- LLM APIコスト削減

## クライアント初期化

```go
client := meta.NewClient(kind, apiKey, model)
```

**パラメータ**:
- `kind`: "openai-chat" または "mock"
- `apiKey`: OpenAI API キー（kind="openai-chat"時は必須）
- `model`: LLMモデル（デフォルト: "gpt-4-turbo"）

## YAML パース戦略

### LLMレスポンスがMarkdown code blockで包含される場合への対応

**現在の課題**:
- LLMが `\`\`\`yaml ... \`\`\`` で応答することがある
- 直接 yaml.Unmarshal() するとパース失敗

**対応方法** (TODO client.go:128):
```
// TODO: Implement robust extraction.
```

**推奨実装**:
```go
// LLMレスポンスからYAMLをリバースエンジニア
// 方法1: バックティック除去
resp = strings.Trim(resp, "`yaml\n")
resp = strings.Trim(resp, "\n```")

// 方法2: 正規表現で抽出
re := regexp.MustCompile(`\`\`\`yaml\s*([\s\S]*?)\`\`\``)
matches := re.FindStringSubmatch(resp)
if len(matches) > 1 {
    resp = matches[1]
}
```

## エラー回復戦略

### APIエラーが発生した場合
- **Runner側**: StateFailed に遷移（再試行なし）
- **推奨**: 将来的に retry logic を実装（exponential backoff等）

### YAML パースエラー
- **原因**: LLMが期待フォーマットで応答していない
- **診断**: エラーメッセージに response: %s を含めるため、ログから LLMレスポンスを確認可能
- **対応**: System Promptを改善またはモデルを変更

## Codex CLI プロバイダ（cli_provider.go）

### 概要

CodexCLIProvider は Codex CLI を使用して Meta-agent 機能を提供する。
agenttools パッケージを使用してコマンドを構築・実行する。

### タイムアウト設定

```go
// Meta-agent 用のデフォルトタイムアウト（10分）
const DefaultMetaAgentTimeout = 10 * time.Minute
```

LLM の処理時間は予測困難なため、十分な時間を確保する。
このタイムアウトは `agenttools.Execute()` で使用され、親コンテキストから独立して動作する。

### YAML 抽出処理（extractYAML）

Codex CLI の出力にはヘッダー情報が含まれるため、`extractYAML` 関数で YAML 部分を抽出する。

**Codex CLI 出力例:**
```
OpenAI Codex v0.65.0 (research preview)
--------
workdir: /path/to/project
model: gpt-5.2
provider: openai
--------
user
プロンプト内容...
codex
type: decompose
version: 1
payload:
  understanding: "..."
```

**抽出方法（優先順）:**
1. Markdown code block（```yaml ... ```）
2. 汎用 code block（``` ... ```）
3. バックティック除去
4. `type:` で始まる行から末尾まで抽出（Codex CLI ヘッダー対応）

```go
// Method 4: Codex CLI 出力から "type:" で始まる YAML を抽出
reTypeYAML := regexp.MustCompile(`(?m)^type:\s+\w+`)
loc := reTypeYAML.FindStringIndex(response)
if loc != nil {
    return strings.TrimSpace(response[loc[0]:])
}
```

## パフォーマンス考慮

### API呼び出し回数
- plan_task: タスクごとに1回
- next_action: ループごとに1回（最大10回）
- **典型的なタスク**: 3～5ループ → 4～6回のAPIコール

### トークン使用料金
- システムプロンプト: 約80～100 tokens
- 各ユーザープロンプト: 300～1000 tokens（タスク複雑度による）
- **コスト削減**: モック実装を使用

## 既知の制限事項と改善案

### 1. AC状態更新の未実装
- 現在、plan_taskで受け取ったACの Passed フラグは更新されない
- 改善案: next_actionレスポンスに `updated_acceptance_criteria` フィールドを追加

### 2. Markdown抽出ロジックの未実装
- LLMが code block で応答する場合、パース失敗
- 改善案: 正規表現による robust な抽出（上記参照）

### 3. 限定されたDecision Action
- 現在サポート: "run_worker", "mark_complete"
- 将来対応: "ask_human", "abort", "retry_with_context"

### 4. エラー時の再試行ロジック
- API呼び出しエラーでは即座にFAILED
- 改善案: exponential backoff による再試行

## 関連ドキュメント

- [core/CLAUDE.md](../core/CLAUDE.md): Runner FSM と TaskContext
- [worker/CLAUDE.md](../worker/CLAUDE.md): Worker実行の詳細
- `/docs/AgentRunner-architecture.md`: アーキテクチャ全体
- `/docs/agentrunner-spec-v1.md`: プロトコル仕様詳細
