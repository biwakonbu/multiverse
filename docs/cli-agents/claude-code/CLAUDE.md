# Claude Code ナレッジ

## 概要

Claude Code は Anthropic が提供する AI コーディングアシスタントの CLI。
multiverse IDE では Worker エージェントとして Docker コンテナ内で実行される。

## 対応状況

- **ステータス**: ✅ 対応済み
- **WorkerKind**: `claude-code`（互換: `claude-code-cli`）

## 実装（一次ソース）

### 実行コマンド

- 実行バイナリは `claude`（デフォルト）:
  - `internal/agenttools/claude.go` の `cliPath` 既定値が `claude`

### 実行インターフェース

- 単発実行（非対話）を前提に `-p` を使用する:
  - `internal/agenttools/claude.go` が `--model <id> -p <prompt>` を構築
- stdin を使用する場合は `-p -` とし、プロンプト本体は stdin に流す:
  - `internal/agenttools/claude.go` が `-p -` + `plan.Stdin` を設定

### デフォルトモデル

- デフォルトモデルは `claude-haiku-4-5-20251001`:
  - `internal/agenttools/claude.go` の `DefaultClaudeModel`

### モデル一覧（参照 URL）

- https://platform.claude.com/docs/en/about-claude/models/overview

実装側には「公式ドキュメント上に現れたモデル ID の抜粋」として `KnownClaudeModels` を用意している:

- `internal/agenttools/claude_models.go`（`KnownClaudeModels`, `ClaudeModelsDocURL`）

`KnownClaudeModels` には Claude 4.5 系（Haiku / Sonnet / Opus）の ID も含む。

## 認証とサンドボックス

### 認証ディレクトリ

- ホストの `~/.config/claude` が存在する場合、コンテナへ ReadOnly で bind mount される:
  - コンテナ側: `/root/.config/claude`
  - 実装: `internal/worker/sandbox.go`

### Pre-flight / セッション検証

- Orchestrator は `~/.config/claude` の存在確認を行う:
  - 実装: `internal/orchestrator/executor.go`
- Worker は `~/.config/claude` の存在確認 + `claude --version` を併用して検証する:
  - 実装: `internal/worker/executor.go`

## モデルの切り替え

### 推奨

- 実行時のモデルは `worker_call.model` で上書きする（Meta-agent が WorkerCall を生成する設計）。
- 既定値を変える場合は `internal/agenttools/claude.go` の `DefaultClaudeModel` を更新する。
