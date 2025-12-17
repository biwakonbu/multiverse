# CLI サブスクリプション設定ガイド

AgentRunner はローカルの CLI セッションを利用してタスクを実行します。アプリ内に API キーを保持せず、既存サブスクリプションをそのまま利用できます。

## 対応プロバイダ

- **Codex CLI**: `codex`
- **Claude Code**: `claude` / `claude-code`
- **Gemini CLI**: `gemini`
- **Cursor CLI**: `cursor`

## セットアップ手順

### 1. Codex CLI

1. Codex CLI をインストール
2. ログイン:
   ```bash
   codex login
   ```
   `~/.codex/auth.json` が作成されます。
3. AgentRunner は `~/.codex/auth.json` をサンドボックスコンテナへ自動マウントします（ReadOnly）。

#### モデル/価格（参照 URL）

- https://platform.openai.com/docs/pricing

#### このプロジェクトのデフォルト/推奨モデル

- Meta-agent: `gpt-5.2`（実装: `internal/agenttools/codex.go`）
- Worker: `gpt-5.1-codex`（実装: `internal/agenttools/codex.go`）
- Worker（高速）: `gpt-5.1-codex-mini`（ショートハンド: `5.1-codex-mini`、実装: `internal/agenttools/openai_models.go`）

### 2. Claude Code

1. Claude Code をインストール:
   ```bash
   npm install -g @anthropic-ai/claude-code
   ```
2. ログイン:
   ```bash
   claude login
   ```
3. `claude` コマンドが PATH 上にあることを確認

#### モデル一覧（参照 URL）

- https://platform.claude.com/docs/en/about-claude/models/overview

#### このプロジェクトのデフォルトモデル

- `claude-haiku-4-5-20251001`（実装: `internal/agenttools/claude.go`）
- 公式ドキュメント上に現れたモデル ID は `KnownClaudeModels` として実装に同梱（`internal/agenttools/claude_models.go`）

### 3. Gemini CLI

Gemini CLI の詳細は `docs/guides/gemini-cli.md` を参照してください。

### 4. Cursor CLI

Cursor CLI が PATH 上にあることを確認してください。

## Multiverse IDE 側の設定

1. **Settings** -> **LLM** を開く
2. Provider を選択（例: `codex-cli`, `claude-code`）
3. "Test Connection" で疎通確認

## トラブルシュート

- **Session not found**: 各 CLI の login を実行し、認証情報が作成されていることを確認してください。
- **Permission denied（macOS）**: Docker/Terminal に Full Disk Access が必要になる場合があります。
