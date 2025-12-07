# agenttools パッケージ

## 責務

外部 AI エージェントツール（Codex CLI、Claude Code など）の実行計画（ExecPlan）を生成し、
統一されたインターフェースで実行する。

## 主要コンポーネント

### ExecPlan（types.go）

CLI 実行計画を表す構造体。

```go
type ExecPlan struct {
    Command string            // 実行コマンド（例: "codex"）
    Args    []string          // 引数
    Env     map[string]string // 環境変数
    Workdir string            // 作業ディレクトリ
    Timeout time.Duration     // タイムアウト（親コンテキストから独立）
    Stdin   string            // 標準入力内容
}
```

### Execute（exec.go）

ExecPlan を実行し、結果を返す。

```go
func Execute(ctx context.Context, plan ExecPlan) ExecResult
```

**タイムアウト処理:**
- `plan.Timeout > 0` の場合、親コンテキストから独立した新しいコンテキストを作成
- 親コンテキストのキャンセルに影響されずに実行できる

```go
if plan.Timeout > 0 {
    execCtx, cancel = context.WithTimeout(context.Background(), plan.Timeout)
    defer cancel()
}
```

**Graceful Shutdown（Go 1.20+）:**
- コンテキストキャンセル時は SIGTERM を送信
- 5秒待機後に SIGKILL

```go
const GracefulShutdownDelay = 5 * time.Second

cmd.Cancel = func() error {
    return cmd.Process.Signal(syscall.SIGTERM)
}
cmd.WaitDelay = GracefulShutdownDelay
```

### CodexProvider（codex.go）

Codex CLI 用の ExecPlan を生成する。

**主要フラグ:**
- `--dangerously-bypass-approvals-and-sandbox`: Docker 内実行時に使用
- `-C <workdir>`: 作業ディレクトリ指定
- `-m <model>`: モデル指定
- `-c reasoning_effort=<level>`: 思考の深さ

**stdin 使用時:**
```go
if req.UseStdin {
    plan.Args = append(plan.Args, "-")  // プロンプト = stdin
    plan.Stdin = req.Prompt
}
```

## 設計原則

### Provider パターン

各エージェントツール（Codex、Claude Code など）は `AgentToolProvider` インターフェースを実装。

```go
type AgentToolProvider interface {
    Kind() string
    Capabilities() Capability
    Build(ctx context.Context, req Request) (ExecPlan, error)
}
```

### Registry

`Register()` で Provider を登録、`Build()` で ExecPlan を生成。

```go
Register("codex-cli", func(cfg ProviderConfig) (AgentToolProvider, error) {
    return NewCodexProvider(cfg), nil
})

plan, err := Build(ctx, "codex-cli", cfg, req)
```

## 使用例

### Meta-agent からの呼び出し

```go
req := agenttools.Request{
    Prompt:          fullPrompt,
    Model:           "gpt-5.1",
    Timeout:         10 * time.Minute,  // 親コンテキストから独立
    UseStdin:        true,
    ToolSpecific: map[string]interface{}{
        "docker_mode": false,
        "json_output": false,
    },
}

plan, _ := agenttools.Build(ctx, "codex-cli", cfg, req)
result := agenttools.Execute(ctx, plan)
```

### Worker からの呼び出し

Worker では Docker コンテナ内で実行するため、`Sandbox.Exec()` を使用。
`agenttools.Execute()` はホスト上で直接実行する場合に使用。

## テスト戦略

- Provider の Build テスト: フラグ生成の正確性
- Execute のモックテスト: コマンド実行のシミュレーション
- タイムアウトテスト: 長時間実行のハンドリング

## 既知の制限事項

### 1. Codex CLI ヘッダー出力
Codex CLI は標準出力にヘッダー情報を含む。YAML 抽出は `meta/client.go` の `extractYAML()` で対応。

### 2. stdin のみサポート
現在は stdin 経由のプロンプト送信のみサポート。ファイル入力は未対応。
