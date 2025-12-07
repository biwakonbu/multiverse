# PRD v3.0: multiverse - チャット駆動 AI 開発支援プラットフォーム

## 1. プロダクトビジョン

### 1.1 ビジョンステートメント

**multiverse** は、チャットインターフェースを通じて開発者の意図を理解し、
Meta-agent が自律的にタスクを分解・実行・評価する AI 開発支援プラットフォームです。

**コアコンセプト:**

- チャットウィンドウが全ての入力経路（AI との対話）
- Meta-agent による徹底的なタスク分解
  - 概念設計 → 実装設計 → 実装計画 → タスクマネジメント → アサイン
- 2D 俯瞰 UI でタスクグラフを視覚化（有向グラフ）
- WBS はリリースマイルストーンとして別枠管理
- 自律実行（計画 → 実行まで全自動、一時停止機能あり）

### 1.2 解決する課題

| 現状の課題                   | multiverse v3.0 での解決                     |
| ---------------------------- | -------------------------------------------- |
| タスク作成が手動・煩雑       | チャットから自然言語でタスク生成             |
| タスク間依存関係の管理が困難 | 有向グラフで依存関係を可視化                 |
| 達成判定が曖昧               | 細分化されたタスクで個別・シンプルな達成判定 |
| 人間の介入が頻繁に必要       | 自律実行ループで人間待ち不要                 |
| 問題・検討材料の散逸         | バックログで一元管理                         |
| LLM がモックのまま           | **本番 LLM 接続で実際のタスク処理**          |

### 1.3 ターゲットユーザー

- ソフトウェア開発者（個人・チーム）
- AI アシスタントと協調して開発を進めたいエンジニア
- 複数の並行タスクを俯瞰的に管理したい開発リーダー

---

## 2. 実装フェーズ概要

### Phase 1-3: 完了済み ✅

| フェーズ | 内容                         | ステータス |
| -------- | ---------------------------- | ---------- |
| Phase 1  | チャット → タスク生成（MVP） | ✅ 完了    |
| Phase 2  | 依存関係グラフ・WBS 表示     | ✅ 完了    |
| Phase 3  | 自律実行ループ               | ✅ 完了    |

**Phase 1-3 で実装済みの機能:**

- FloatingChatWindow によるチャット入力
- ChatHandler によるメッセージ処理
- Meta-agent decompose プロトコル（モック実装）
- Task 構造体（依存関係、WBS、生成元情報）
- GridCanvas でのタスク表示
- ConnectionLine による依存関係矢印
- WBS ビュー
- TaskGraphManager
- ExecutionOrchestrator（開始・一時停止・再開・停止）
- BacklogStore
- EventEmitter（リアルタイム通知）
- リトライポリシー

---

## 3. Phase 4: CLI セッション統合と実タスク実行【現在のフェーズ】

### 3.1 概要

Phase 1-3 で構築した基盤を活用し、モック LLM から CLI セッション（Codex CLI 等）ベースの実タスク実行を実現する。
チャットメッセージから生成されたタスクを、実際に agent-runner で実行できるようにする。

**重要方針:**
- API キーは不要。Codex / Claude Code / Gemini / Cursor など **CLI サブスクリプションセッションを優先利用**する
- Meta 層も CLI セッション前提に置き換え、API キー依存を排除する

### 3.2 機能要件

#### FR-P4-001: CLI プロバイダ接続

**設計方針:**

- LLMConfigStore で CLI プロバイダを管理（`codex-cli`, `claude-code`, `gemini-cli`, `cursor-cli` 等）
- デフォルトは `mock`（開発用）、本番は `codex-cli` を想定
- CLI セッションの検証と引き継ぎを実装（環境変数、ソケット、マウント等）

**実装方針（最新版反映）:**

| 項目                    | 内容                                                                                  |
| ----------------------- | ------------------------------------------------------------------------------------- |
| CLI プロバイダ基盤       | `internal/agenttools` に ProviderConfig/Request/ExecPlan/registry を実装               |
| Codex プロバイダ         | `internal/agenttools/codex.go`（exec/chat、model/temperature/max-tokens/flags/env）   |
| 他プロバイダ             | Gemini / Claude Code / Cursor は stub 登録のみ（未実装エラーを明示）                  |
| Worker 実行              | `internal/worker/executor.go` が WorkerCall→ExecPlan 変換後に Sandbox.Exec で実行     |
| WorkerCall 拡張          | model/flags/env/tool_specific/use_stdin/workdir 等を許容（Meta/Orchestrator から指定可） |
| stdin サポート           | まだ Sandbox.Exec では未対応。UseStdin 指定時はエラーで弾く                          |
| セッション検証           | Codex CLI セッションは既存の `verifyCodexSession`（auth.json or CODEX_API_KEY）を利用 |

```go
// internal/agenttools/types.go
type Request struct {
    Prompt       string
    Mode         string
    Model        string
    Temperature  *float64
    MaxTokens    *int
    Workdir      string
    Timeout      time.Duration
    ExtraEnv     map[string]string
    Flags        []string
    ToolSpecific map[string]interface{}
    UseStdin     bool
}
```

#### FR-P4-002: 実行制御 UI の完成

**ツールバー ExecutionControls:**

現在の実装（`app.go`）には API が存在するが、フロントエンドの UI が未完成。

```svelte
<!-- frontend/ide/src/lib/toolbar/ExecutionControls.svelte -->
<script lang="ts">
    import { executionState } from '../../stores/executionStore';
    import { StartExecution, PauseExecution, ResumeExecution, StopExecution } from '../../../wailsjs/go/main/App';
    import Button from '../design-system/Button.svelte';
    import StatusIndicator from '../design-system/StatusIndicator.svelte';

    async function handleStart() {
        try {
            await StartExecution();
        } catch (e) {
            // TODO: Toast 通知
            console.error('Failed to start execution:', e);
        }
    }

    async function handlePause() {
        await PauseExecution();
    }

    async function handleResume() {
        await ResumeExecution();
    }

    async function handleStop() {
        await StopExecution();
    }
</script>

<div class="execution-controls">
    <StatusIndicator status={$executionState === 'RUNNING' ? 'running' : $executionState === 'PAUSED' ? 'paused' : 'idle'} />

    {#if $executionState === 'IDLE'}
        <Button variant="primary" size="sm" on:click={handleStart}>
            ▶ 開始
        </Button>
    {:else if $executionState === 'RUNNING'}
        <Button variant="warning" size="sm" on:click={handlePause}>
            ⏸ 一時停止
        </Button>
        <Button variant="danger" size="sm" on:click={handleStop}>
            ⏹ 停止
        </Button>
    {:else if $executionState === 'PAUSED'}
        <Button variant="primary" size="sm" on:click={handleResume}>
            ▶ 再開
        </Button>
        <Button variant="danger" size="sm" on:click={handleStop}>
            ⏹ 停止
        </Button>
    {/if}
</div>
```

#### FR-P4-003: タスク実行ログのリアルタイム表示

**設計方針:**

- agent-runner の stdout/stderr をリアルタイムでフロントエンドに転送
- Wails Events を使用してログ行を逐次送信
- ログビューワーで実行中タスクの進捗を確認可能

```go
// internal/orchestrator/executor.go 拡張

// StreamingExecuteTask はログをリアルタイムストリーミングしながらタスクを実行する
func (e *Executor) StreamingExecuteTask(ctx context.Context, task *Task, emitter EventEmitter) (*Attempt, error) {
    // ... (既存の ExecuteTask ロジックを拡張)

    // stdout/stderr をリアルタイムで送信
    stdoutPipe, _ := cmd.StdoutPipe()
    stderrPipe, _ := cmd.StderrPipe()

    go func() {
        scanner := bufio.NewScanner(stdoutPipe)
        for scanner.Scan() {
            emitter.Emit(EventTaskLog, TaskLogEvent{
                TaskID: task.ID,
                Stream: "stdout",
                Line:   scanner.Text(),
                Time:   time.Now(),
            })
        }
    }()

    // ... (stderr も同様)
}
```

**フロントエンド TaskLogView:**

```svelte
<!-- frontend/ide/src/lib/components/TaskLogView.svelte -->
<script lang="ts">
    import { taskLogs } from '../../stores/taskLogStore';
    import { EventsOn } from '../../../wailsjs/runtime/runtime';
    import { onMount } from 'svelte';

    export let taskId: string;

    $: logs = $taskLogs[taskId] ?? [];

    onMount(() => {
        EventsOn('task:log', (event: { taskId: string; stream: string; line: string }) => {
            if (event.taskId === taskId) {
                taskLogs.addLog(taskId, event);
            }
        });
    });
</script>

<div class="task-log-view">
    {#each logs as log}
        <pre class="log-line {log.stream}">{log.line}</pre>
    {/each}
</div>
```

#### FR-P4-004: 設定画面

**LLM 設定:**

```svelte
<!-- frontend/ide/src/lib/settings/LLMSettings.svelte -->
<script lang="ts">
    import { GetLLMConfig, SetLLMConfig, TestLLMConnection } from '../../../wailsjs/go/main/App';
    import Input from '../design-system/Input.svelte';
    import Button from '../design-system/Button.svelte';

    let config = {
        kind: 'mock',
        apiKey: '',
        model: 'gpt-4o',
    };
    let testResult = '';
    let testing = false;

    async function loadConfig() {
        config = await GetLLMConfig();
    }

    async function saveConfig() {
        await SetLLMConfig(config);
    }

    async function testConnection() {
        testing = true;
        try {
            testResult = await TestLLMConnection();
        } catch (e) {
            testResult = `エラー: ${e}`;
        } finally {
            testing = false;
        }
    }
</script>

<div class="llm-settings">
    <h3>LLM 設定</h3>

    <label>
        プロバイダ
        <select bind:value={config.kind}>
            <option value="mock">モック（開発用）</option>
            <option value="openai-chat">OpenAI</option>
        </select>
    </label>

    {#if config.kind === 'openai-chat'}
        <Input
            label="API キー"
            type="password"
            bind:value={config.apiKey}
            placeholder="sk-..."
        />
        <Input
            label="モデル"
            bind:value={config.model}
            placeholder="gpt-4o"
        />
    {/if}

    <div class="actions">
        <Button on:click={saveConfig}>保存</Button>
        <Button variant="secondary" on:click={testConnection} disabled={testing}>
            {testing ? '接続テスト中...' : '接続テスト'}
        </Button>
    </div>

    {#if testResult}
        <div class="test-result">{testResult}</div>
    {/if}
</div>
```

**バックエンド API:**

```go
// app.go に追加

// LLMConfig は LLM 設定
type LLMConfig struct {
    Kind   string `json:"kind"`
    APIKey string `json:"apiKey"`
    Model  string `json:"model"`
}

// GetLLMConfig は現在の LLM 設定を取得する
func (a *App) GetLLMConfig() LLMConfig {
    return LLMConfig{
        Kind:   os.Getenv("MULTIVERSE_META_KIND"),
        APIKey: "********", // マスク表示
        Model:  os.Getenv("MULTIVERSE_META_MODEL"),
    }
}

// SetLLMConfig は LLM 設定を保存する
func (a *App) SetLLMConfig(config LLMConfig) error {
    // 設定ファイルに保存（環境変数ではなく永続化）
    // TODO: セキュアな保存方法（OS keychain）
}

// TestLLMConnection は LLM 接続をテストする
func (a *App) TestLLMConnection() (string, error) {
    metaClient := newMetaClientFromConfig(a.getLLMConfig())
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // 簡単なテストプロンプトを送信
    _, err := metaClient.Decompose(ctx, &meta.DecomposeRequest{
        UserInput: "テスト接続",
        Context:   meta.DecomposeContext{},
    })
    if err != nil {
        return "", err
    }
    return "接続成功", nil
}
```

### 3.3 受け入れ条件

| ID       | 条件                                                  |
| -------- | ----------------------------------------------------- |
| AC-P4-01 | CLI プロバイダ（Codex CLI 等）を設定画面から選択・保存できる |
| AC-P4-02 | 設定画面から CLI セッション検証を実行できる               |
| AC-P4-03 | チャットメッセージが CLI プロバイダで処理される             |
| AC-P4-04 | 生成されたタスクが実際に agent-runner で実行される    |
| AC-P4-05 | タスク実行ログがリアルタイムで表示される              |
| AC-P4-06 | 実行コントロール（開始/一時停止/再開/停止）が機能する |
| AC-P4-07 | Docker サンドボックスで Codex CLI セッションが引き継がれる |

---

## 4. Phase 5: 高度な機能【将来】

### 4.1 マルチ LLM プロバイダ対応

| プロバイダ            | 対応状況       |
| --------------------- | -------------- |
| OpenAI                | Phase 4 で対応 |
| Anthropic (Claude)    | Phase 5        |
| Google (Gemini)       | Phase 5        |
| ローカル LLM (Ollama) | Phase 5        |

### 4.2 高度なタスク管理

- タスクの手動編集・削除
- タスクの優先度変更
- タスクのマージ・分割
- カスタム依存関係の追加

### 4.3 チーム機能

- マルチユーザー対応
- タスクのアサイン
- レビューワークフロー

---

## 5. アーキテクチャ

### 5.1 現在の 4 層構造

```
┌─────────────────────────────────────────────────────┐
│  multiverse (Desktop UI)                            │
│  - ChatWindow → タスク生成                           │
│  - GridCanvas → 依存グラフ表示                       │
│  - WBSView → マイルストーン表示                      │
│  - BacklogPanel → バックログ管理                     │
│  - ExecutionControls → 実行制御 (Phase 4 強化)       │
│  - LLMSettings → LLM 設定 (Phase 4 新規)            │
└──────────────┬──────────────────────────────────────┘
               │ Wails IPC + Events
┌──────────────▼──────────────────────────────────────┐
│  Orchestrator Layer                                 │
│  - ChatHandler                                      │
│  - TaskGraphManager                                 │
│  - ExecutionOrchestrator                            │
│  - BacklogStore                                     │
│  - TaskStore / Scheduler                            │
│  - LLMConfigStore (Phase 4 新規)                    │
└──────────────┬──────────────────────────────────────┘
               │
┌──────────────▼──────────────────────────────────────┐
│  AgentRunner Core + Meta-agent                      │
│  - FSM（状態遷移）                                   │
│  - decompose プロトコル                              │
│  - OpenAI API 接続 (Phase 4 本番化)                 │
└──────────────┬──────────────────────────────────────┘
               │
┌──────────────▼──────────────────────────────────────┘
│  Worker (Docker Sandbox)                            │
│  - Codex CLI 実行                                   │
└─────────────────────────────────────────────────────┘
```

### 5.2 Phase 4 で追加するコンポーネント

| コンポーネント    | 場所                             | 責務             |
| ----------------- | -------------------------------- | ---------------- |
| LLMConfigStore    | internal/ide/llm_config.go       | LLM 設定の永続化 |
| ExecutionControls | frontend/ide/src/lib/toolbar/    | 実行制御 UI      |
| LLMSettings       | frontend/ide/src/lib/settings/   | LLM 設定 UI      |
| TaskLogView       | frontend/ide/src/lib/components/ | タスクログ表示   |
| taskLogStore      | frontend/ide/src/stores/         | ログ状態管理     |

---

## 6. 技術スタック

### 6.1 バックエンド（維持）

| カテゴリ     | 技術       | バージョン           |
| ------------ | ---------- | -------------------- |
| 言語         | Go         | 1.23+                |
| デスクトップ | Wails      | v2                   |
| コンテナ     | Docker     | -                    |
| LLM          | OpenAI API | gpt-4o / gpt-4o-mini |

### 6.2 フロントエンド（維持）

| カテゴリ       | 技術       | バージョン |
| -------------- | ---------- | ---------- |
| フレームワーク | Svelte     | 4          |
| 型安全         | TypeScript | 5          |
| ビルド         | Vite       | 5          |
| パッケージ管理 | pnpm       | -          |

### 6.3 LLM プロバイダ（Phase 4-5）

| プロバイダ     | 実行方式                    | セッション管理              |
| -------------- | --------------------------- | --------------------------- |
| Codex CLI      | `codex chat` コマンド       | CLI サブスクリプションセッション |
| Claude Code CLI| `claude-code chat` コマンド  | CLI サブスクリプションセッション |
| Gemini CLI     | `gemini chat` コマンド       | CLI サブスクリプションセッション |
| Cursor CLI     | `cursor chat` コマンド       | CLI サブスクリプションセッション |
| Mock           | モック実装                  | なし                        |

---

## 7. マイルストーン

### Phase 4 マイルストーン（2 週間）

**Week 1:**

- [ ] LLM 設定 UI の実装
- [ ] LLM 接続テスト機能
- [ ] 環境変数のセキュア保存
- [ ] ExecutionControls UI の完成

**Week 2:**

- [ ] タスク実行ログのリアルタイム表示
- [ ] 本番 LLM での E2E テスト
- [ ] ドキュメント更新
- [ ] リリース準備

---

## 8. 技術的リスクと対策

| リスク               | 影響度 | 対策                                           |
| -------------------- | ------ | ---------------------------------------------- |
| CLI セッション未認証 | 高     | 起動時検証、明示エラー表示                     |
| Docker 内セッション引き継ぎ失敗 | 高     | 環境変数/ボリュームマウントでセッション情報を伝播 |
| CLI コマンド実行エラー | 中     | エラーハンドリング、リトライポリシー           |
| LLM 応答の不安定さ   | 中     | プロンプトエンジニアリング、バリデーション強化 |

---

## 9. 用語集

| 用語         | 説明                                               |
| ------------ | -------------------------------------------------- |
| Meta-agent   | CLI セッション（Codex CLI 等）を使ってタスク分解・評価を行うエージェント     |
| Decompose    | ユーザー入力からタスクを分解するプロトコル         |
| agent-runner | Docker 内でタスクを実行するコアエンジン            |
| Worker       | 実際のコード生成・テスト実行を行う CLI（Codex 等） |
| CLI セッション | Codex / Claude Code / Gemini / Cursor 等の CLI サブスクリプションセッション |
