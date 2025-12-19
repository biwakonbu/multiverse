# Gemini CLI ナレッジ

## 概要

Gemini CLI は Google が提供する AI エージェントの CLI。multiverse IDE では Worker エージェントとして Docker コンテナ内で実行される。

## 対応状況

- **ステータス**: ✅ 対応済み
- **WorkerKind**: `gemini-cli`

## 実装（一次ソース）

### 実行コマンド

- 実行バイナリは `gemini`（デフォルト）:
  - `internal/agenttools/gemini.go` の `cliPath` 既定値が `gemini`

### 実行インターフェース

- 非対話実行は `-p` を使用する:
  - `internal/agenttools/gemini.go` が `--model <id> --output-format json --yolo -p <prompt>` を構築
- stdin を使用する場合は `-p -` とし、プロンプト本体は stdin に流す:
  - `internal/agenttools/gemini.go` が `-p -` + `plan.Stdin` を設定

### デフォルトモデル

- デフォルトモデルは `gemini-3-flash-preview`:
  - `internal/agenttools/gemini.go` の `DefaultGeminiModel`

### 対応モデル（抜粋）

- `gemini-3-flash-preview`
- `gemini-3-pro-preview`

### モデル一覧（参照）

- `docs/guides/gemini-cli.md`

## 認証とサンドボックス

### 認証ファイル/ディレクトリ

- ホストの `~/.gemini` が存在する場合、コンテナへ ReadOnly で bind mount される:
  - コンテナ側: `/root/.gemini`
  - 実装: `internal/worker/sandbox.go`

### Pre-flight / セッション検証

- Orchestrator は `GEMINI_API_KEY` / `GOOGLE_API_KEY` / `~/.gemini/.env` を確認する:
  - 実装: `internal/orchestrator/executor.go`
- Worker は同様の確認 + `gemini --version` を併用して検証する:
  - 実装: `internal/worker/executor.go`

### 環境変数

- `GEMINI_API_KEY` / `GOOGLE_API_KEY` / `GOOGLE_GENAI_USE_VERTEXAI` / `GOOGLE_CLOUD_PROJECT` をコンテナへ注入する:
  - 実装: `internal/worker/executor.go`, `internal/worker/sandbox.go`

## Docker イメージ

- デフォルトイメージは `ghcr.io/biwakonbu/agent-runner-gemini:latest`:
  - 実装: `internal/worker/executor.go`

## タスク YAML 例

```yaml
runner:
  worker:
    kind: "gemini-cli"
    model: "gemini-3-flash-preview"
    max_run_time_sec: 300
    env:
      GEMINI_API_KEY: "env:GEMINI_API_KEY"
```

## 注意事項

- Gemini CLI のバージョンは固定していない（最新安定版前提）。
- 実際の対応フラグ/挙動は CLI のバージョン差分に影響されるため、必要に応じて `agenttools.ProviderConfig.Flags` で調整する。
