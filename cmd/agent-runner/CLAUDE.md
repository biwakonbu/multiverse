# AgentRunner CLI - エントリポイント

このディレクトリは AgentRunner Core の CLI エントリポイントを提供します。

## 責務

- **アプリケーション起動**: CLI からのタスク実行
- **依存性注入**: 各コンポーネント（Meta, Worker, Note）の初期化と注入
- **設定統合**: CLI フラグと YAML 設定の統合

## ファイル構成

| ファイル | 役割 |
|---------|------|
| main.go | CLI エントリポイント、依存性注入、実行制御 |

## 実行フロー

```
1. CLI フラグパース (cli.ParseFlags)
   ↓
2. 標準入力から YAML 読み込み
   ↓
3. YAML パース (config.TaskConfig)
   ↓
4. コンポーネント初期化
   - MetaClient (OpenAI API)
   - WorkerExecutor (Docker/Codex)
   - NoteWriter (Markdown 出力)
   ↓
5. Runner 実行 (core.Runner.Run)
   ↓
6. 結果出力
```

## 使用方法

```bash
# 基本的な使用
./agent-runner < task.yaml

# パイプで実行
cat task.yaml | ./agent-runner

# モデル指定
./agent-runner -meta-model gpt-4o < task.yaml
```

## 依存性注入パターン

```go
// コンポーネント初期化
metaClient := meta.NewClient(cfg.Runner.Meta.Kind, apiKey, metaModel, cfg.Runner.Meta.SystemPrompt)
workerExecutor, _ := worker.NewExecutor(cfg.Runner.Worker, cfg.Task.Repo)
noteWriter := note.NewWriter()

// Runner 生成（DI パターン）
runner := core.NewRunner(&cfg, metaClient, workerExecutor, noteWriter)
```

## 環境変数

| 変数名 | 必須 | 説明 |
|-------|------|------|
| OPENAI_API_KEY | Yes* | Meta-agent 用 OpenAI API キー |
| CODEX_API_KEY | No | Worker agent 用 Codex API キー |

*mock モードでは不要

## 設定解決の優先順位

### Meta Model

1. CLI フラグ (`-meta-model`)
2. YAML 設定 (`runner.meta.model`)
3. デフォルト値 (`gpt-5.2`)

## ログ出力

構造化ログ（`log/slog`）を使用:

```go
logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
```

- `INFO`: タスク開始、モデル解決、完了
- `WARN`: API キー未設定（mock モード）
- `ERROR`: 致命的エラー

## エラーハンドリング

### 起動時エラー

- CLI フラグパース失敗 → 即座に終了
- YAML 読み込み失敗 → 即座に終了
- Worker 初期化失敗 → 即座に終了

### 実行時エラー

- Runner.Run() の戻り値で判定
- エラー時は exit code 1

## テスト戦略

### Run 関数の分離

```go
// main() から分離してテスト可能に
func Run(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer, logger *slog.Logger) error
```

- 標準入出力を注入可能
- コンテキストでキャンセル制御
- ロガーを注入可能

### 統合テスト

- `test/integration/` で end-to-end テスト
- モック実装を使用した高速テスト
- Docker テスト（`-tags=docker`）

## ビルド

```bash
# 開発ビルド
go build ./cmd/agent-runner

# 名前指定
go build -o agent-runner ./cmd/agent-runner
```

## 関連ドキュメント

- [../../internal/cli/CLAUDE.md](../../internal/cli/CLAUDE.md): CLI フラグ処理
- [../../internal/core/CLAUDE.md](../../internal/core/CLAUDE.md): タスク実行エンジン
- [../../internal/meta/CLAUDE.md](../../internal/meta/CLAUDE.md): Meta-agent 通信
- [../../internal/worker/CLAUDE.md](../../internal/worker/CLAUDE.md): Worker 実行
- [../../pkg/config/CLAUDE.md](../../pkg/config/CLAUDE.md): YAML 設定スキーマ
