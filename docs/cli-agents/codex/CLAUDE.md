# Codex CLI ナレッジ

## 概要

Codex CLI は OpenAI が提供する AI コーディングアシスタントのコマンドラインツール。
multiverse IDE では Worker エージェントとして Docker コンテナ内で実行される。
※ IDE の Meta-agent はデフォルト `openai-chat` だが、`OPENAI_API_KEY` 未設定かつ `codex` が利用可能な場合は `codex-cli` に自動フォールバックする（`app.go` の `newMetaClientFromConfig()` 参照）。

## 参照 URL（モデル/価格）

- https://platform.openai.com/docs/pricing

## 対応バージョン

- **現在**: 0.65.0
- **最終更新**: 2025-12-07

バージョン固有の詳細は [version-0.65.0.md](version-0.65.0.md) を参照。

## multiverse IDE での使用

### 統一実装

multiverse IDE では Worker 実行における Codex CLI 呼び出しを `internal/agenttools/codex.go` の `CodexProvider` で統一管理。
Meta-agent も `codex-cli` を選択可能（`openai-chat` は `OPENAI_API_KEY` 必須）。

### 実行モード

| モード | 用途 | docker_mode | json_output |
|--------|------|-------------|-------------|
| Worker | Docker 内タスク実行 | `true`（デフォルト） | `true`（デフォルト） |
| Meta-agent（参考） | ホスト上で計画・分解 | `false` | `false` |

### Worker 実行コマンド（Docker 内）

```bash
	codex exec \
	  --dangerously-bypass-approvals-and-sandbox \
	  -C /workspace/project \
	  --json \
	  -m gpt-5.1-codex \
	  -c reasoning_effort=medium \
	  "プロンプト..."
```

### Meta-agent 実行コマンド（ホスト上 / 参考）

※ `codex-cli` をホスト上で実行する場合の参考例（IDE 経由でも `codex-cli` を選択可能）。

```bash
	codex exec \
	  -m gpt-5.2 \
	  -c reasoning_effort=medium \
	  -
```

### 必須フラグ（Worker 実行時）

| フラグ | 説明 | 備考 |
|--------|------|------|
| `--dangerously-bypass-approvals-and-sandbox` | サンドボックス・承認を無効化 | Docker が外部サンドボックスとして機能 |
| `-C <DIR>` | 作業ディレクトリ | デフォルト: `/workspace/project` |
| `--json` | JSONL 出力 | 機械可読形式で結果を取得 |
| `-m <MODEL>` | モデル指定 | デフォルト: `gpt-5.1-codex` |

### モデル設定

| 用途 | モデル ID | 設定箇所 |
|------|----------|---------|
| Worker タスク実行 | `gpt-5.1-codex` | `agenttools.DefaultCodexModel` |
| Meta-agent（計画・思考） | `gpt-5.2` | `agenttools.DefaultMetaModel` |
| 高速実行（必要時） | `gpt-5.1-codex-mini`（ショートハンド: `5.1-codex-mini`） | `internal/agenttools/openai_models.go` |

### 思考の深さ（reasoning effort）

| レベル | 用途 | 設定方法 |
|--------|------|---------|
| `low` | 単純なタスク | `-c reasoning_effort=low` |
| `medium` | 通常のタスク（**デフォルト**） | `-c reasoning_effort=medium` |
| `high` | 複雑なタスク・リトライ時 | `-c reasoning_effort=high` |

### stdin 入力

プロンプトを stdin から読み取る場合、PROMPT に `-` を指定:

```bash
	echo "プロンプト内容" | codex exec \
	  --dangerously-bypass-approvals-and-sandbox \
	  -C /workspace/project \
	  --json \
	  -m gpt-5.1-codex \
	  -
```

### 設定オーバーライド (-c)

`-c` フラグで TOML 形式の設定をオーバーライド可能:

```bash
# 複数の設定を指定
codex exec \
  -c reasoning_effort=high \
  -c temperature=0.5 \
  -c max_tokens=4000 \
  "プロンプト..."
```

## 出力形式

### JSONL 出力（--json 使用時）

各行が独立した JSON オブジェクト:

```json
{"type": "message", "content": "..."}
{"type": "tool_call", "name": "...", "arguments": {...}}
{"type": "tool_result", "output": "..."}
{"type": "final", "summary": "..."}
```

## 注意事項

### 存在しないフラグ

以下のフラグは Codex CLI 0.65.0 に**存在しない**:

| 誤ったフラグ | 正しい方法 |
|-------------|-----------|
| `--cwd` | `-C` / `--cd` |
| `--temperature` | `-c temperature=X` |
| `--max-tokens` | `-c max_tokens=X` |
| `--stdin` | PROMPT に `-` を指定 |

### 存在しないサブコマンド

| 誤ったコマンド | 備考 |
|---------------|------|
| `codex chat` | 対話モードは `codex [PROMPT]`（サブコマンドなし） |

## 関連ドキュメント

- [サンドボックス方針](../../design/sandbox-policy.md)
- [バージョン 0.65.0 詳細](version-0.65.0.md)
- [CLI エージェント共通ガイド](../README.md)
