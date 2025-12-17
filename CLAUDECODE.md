# Claude Code 統合設計（CLAUDECODE.md）

最終更新: 2025-12-14

## 目次

1. [目的](#1-目的)
2. [現状実装（一次ソース）](#2-現状実装一次ソース)
   - [2.1 AgentTools（CLI 実行計画）](#21-agenttoolscli-実行計画)
   - [2.2 Worker（セッション検証とイメージ選択）](#22-workerセッション検証とイメージ選択)
   - [2.3 Sandbox（認証情報のマウント）](#23-sandbox認証情報のマウント)
   - [2.4 Orchestrator（Pre-flight Check）](#24-orchestratorpre-flight-check)
3. [設計方針](#3-設計方針)
   - [3.1 認証の扱い（安全性）](#31-認証の扱い安全性)
   - [3.2 CLI インターフェースの前提](#32-cli-インターフェースの前提)
   - [3.3 stdin 運用の統一](#33-stdin-運用の統一)
4. [運用・トラブルシュート](#4-運用トラブルシュート)
   - [4.1 典型的な失敗と対処](#41-典型的な失敗と対処)
   - [4.2 macOS / Windows の差分](#42-macos--windows-の差分)
5. [今後の改善候補（設計上の宿題）](#5-今後の改善候補設計上の宿題)
6. [実装ステータス](#6-実装ステータス)

---

## 1. 目的

本ドキュメントは、Multiverse/AgentRunner の「WorkerKind=`claude-code`」を **Docker Sandbox 上で実行できる**ようにするための設計と、現状実装の一次ソースを整理する。

対象レイヤ:

- `internal/agenttools`: CLI 実行計画（ExecPlan）生成
- `internal/worker`: Docker コンテナ起動・セッション検証
- `internal/worker/sandbox.go`: 認証情報のマウント
- `internal/orchestrator`: Pre-flight Check

参照（モデル一覧）:

- https://platform.claude.com/docs/en/about-claude/models/overview

## 2. 現状実装（一次ソース）

### 2.1 AgentTools（CLI 実行計画）

【事実】`claude-code` の ExecPlan は `internal/agenttools/claude.go` が生成する。

- Kind 名: `claude-code`（`internal/agenttools/claude.go:33`）
- デフォルトモデル: `claude-haiku-4-5-20251001`（`internal/agenttools/claude.go:10`）
- 非対話の前提: `-p` を使用（`internal/agenttools/claude.go:64`）
- stdin 使用時: `-p -` とし、`plan.Stdin` に prompt を入れる（`internal/agenttools/claude.go:84`）

### 2.2 Worker（セッション検証とイメージ選択）

【事実】Worker Executor 起動時に、tool kind に応じてセッション検証とデフォルトイメージを分岐する。

- `claude-code` の場合は `verifyClaudeSession()`（`internal/worker/executor.go:306-344`）
- `codex-cli` の場合は `verifyCodexSession()`（`internal/worker/executor.go:278-304`）
- デフォルト Docker イメージ
  - `claude-code`: `ghcr.io/biwakonbu/agent-runner-claude:latest`（`internal/worker/executor.go:224-228`）
  - それ以外: `ghcr.io/biwakonbu/agent-runner-codex:latest`（`internal/worker/executor.go:227`）

### 2.3 Sandbox（認証情報のマウント）

【事実】Sandbox 起動時に、ホストの `~/.config/claude` が存在する場合は ReadOnly でコンテナへ bind mount する。

- host: `~/.config/claude`（`internal/worker/sandbox.go:92`）
- container: `/root/.config/claude`（`internal/worker/sandbox.go:98`）

### 2.4 Orchestrator（Pre-flight Check）

【事実】Orchestrator Executor が `agent-runner` 起動前に、WorkerKind に応じてセッション検証を行う。

- `claude-code` の場合: `~/.config/claude` の存在確認（`internal/orchestrator/executor.go:481-513`）
- `codex-cli` の場合: `CODEX_SESSION_TOKEN`/`CODEX_API_KEY` または `~/.codex/auth.json` の確認（`internal/orchestrator/executor.go:455-479`）
- 検証失敗時: UI に `process:metaUpdate` イベントでエラー通知を送信

## 3. 設計方針

### 3.1 認証の扱い（安全性）

【事実】認証情報はホストからコンテナへ ReadOnly でマウントする（`internal/worker/sandbox.go:98`）。

【提案】ログ出力・イベントに認証ファイルの中身やパスの詳細を含めない（ファイル存在有無や `--version` 出力程度に留める）。

### 3.2 CLI インターフェースの前提

【事実】本実装は「`claude` コマンドが `-p`（print/single-shot）をサポートする」ことを前提にしている（`internal/agenttools/claude.go:64`）。

【仮説】Claude Code CLI のバージョン/配布形態によっては引数仕様が変わる可能性があるため、将来的に `ProviderConfig.Flags` と `ToolSpecific` で調整余地を残すのが安全。

### 3.3 stdin 運用の統一

【事実】Codex CLI は stdin 時に `-` を prompt として渡し、`plan.Stdin` で中身を渡す（`internal/agenttools/codex.go:158`）。

【提案】Claude 側も同じ規約（`-p -` + `stdin`）に統一し、巨大 prompt でも引数長制限に当たりにくい構成にする（現実装は統一済み、`internal/agenttools/claude.go:84`）。

## 4. 運用・トラブルシュート

### 4.1 典型的な失敗と対処

- 【事実】`~/.config/claude` が存在しない場合、Worker は `claude login` を促すエラーで開始に失敗する（`internal/worker/executor.go:264`）。
- 【提案】CI/E2E では `WorkerKind` を `codex-cli`（mock runner）に寄せ、`claude-code` は手元検証ジョブに切り出す。

### 4.2 macOS / Windows の差分

【不明】Claude Code の認証保存場所が OS/バージョンで固定かどうかは、このリポジトリ内の一次ソースだけでは確証できない。

【提案】認証ディレクトリは config で上書き可能にし、OS 別の既定値（例: XDG/Library/AppData）を段階的に追加する。

## 5. 今後の改善候補（設計上の宿題）

- Docker イメージのビルド定義（`ghcr.io/biwakonbu/agent-runner-claude:latest`）の再現性確保（Dockerfile/README の整備）。
- `claude` の CLI パス（`ProviderConfig.CLIPath`）を Worker 側の実行環境にも反映できるようにする（コンテナ内の実体と一致させる）。

## 6. 実装ステータス

| 項目                          | 状態    | 実装箇所                                    |
| ----------------------------- | ------- | ------------------------------------------- |
| ExecPlan 生成                 | ✅ 完了 | `internal/agenttools/claude.go`             |
| Worker セッション検証         | ✅ 完了 | `internal/worker/executor.go:306-344`       |
| Orchestrator Pre-flight Check | ✅ 完了 | `internal/orchestrator/executor.go:481-513` |
| Docker イメージ選択           | ✅ 完了 | `internal/worker/executor.go:224-228`       |
| 認証情報マウント              | ✅ 完了 | `internal/worker/sandbox.go:92-98`          |
| UI エラー通知                 | ✅ 完了 | `internal/orchestrator/executor.go:486-497` |
