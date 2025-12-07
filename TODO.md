# TODO: multiverse v3.0 - Phase 4 Implementation

Based on PRD v3.0 - Codex CLI 統合と実タスク実行

---

## 現在のステータス

| フェーズ    | 内容                             | ステータス |
| ----------- | -------------------------------- | ---------- |
| Phase 1     | チャット → タスク生成            | ✅ 完了    |
| Phase 2     | 依存関係グラフ・WBS 表示         | ✅ 完了    |
| Phase 3     | 自律実行ループ                   | ✅ 完了    |
| **Phase 4** | **Codex CLI 統合と実タスク実行** | 🚧 進行中  |

---

## 設計方針（重要・現状差分あり）

> [!IMPORTANT]
> API キーは不要。Codex / Claude Code / Gemini / Cursor など **CLI サブスクリプションセッションを優先利用**する。Meta 層も CLI セッション前提に置き換え、API キー依存を排除する。

**現在のデータフロー（実装ベース）:**

```
Chat → Meta-agent (openai-chat via HTTP + OPENAI_API_KEY) → Task 生成
                                                            ↓
ExecutionOrchestrator → agent-runner → Docker Sandbox → codex CLI（既存セッション想定）
```

---

## 現在の実装メモ（2025-12-07 時点）

### バックエンド
- [x] **LLMConfigStore** (`internal/ide/llm_config.go`)
  - Kind/Model/BaseURL/SystemPrompt を `~/.multiverse/config/llm.json` に永続化
  - 環境変数オーバーライドあり（API キー保存は不要にする方針）
- [x] **App API** (`app.go`)
  - `GetLLMConfig` / `SetLLMConfig` / `TestLLMConnection` を追加
  - ただし **ChatHandler 生成は `newMetaClientFromEnv()` 固定**で LLMConfigStore の設定が Meta 層に反映されない
  - `TestLLMConnection` は OpenAI API キー前提の HTTP 呼び出し（API キー不要の CLI セッション検証に置換予定）
- [x] **AgentToolProvider 基盤** (`internal/agenttools`)
  - 共通 Request/ExecPlan/ProviderConfig と registry を追加
  - Codex CLI プロバイダ実装（exec/chat、model/temperature/max-tokens/flags/env を透過）
  - Gemini / Claude Code / Cursor は stub プロバイダで登録（未実装アラートのみ）
- [x] **Worker Executor**
  - `RunWorker` → `RunWorkerCall` に内部委譲し、AgentToolProvider 経由で ExecPlan を構築して Sandbox.Exec 実行
  - `meta.WorkerCall` に model/flags/env/tool_specific/use_stdin などを拡張し、CLI 切替の土台を用意
  - stdin 実行は未サポート（現在はエラーにする）

### フロントエンド
- [x] **LLMSettings** (`frontend/ide/src/lib/settings/LLMSettings.svelte`)
  - プロバイダ選択、モデル/エンドポイント入力、接続テスト UI
  - API キーは「環境変数に設定済みか」を表示するのみ（保存不可）
- [x] **Toolbar 設定ボタン & モーダル** (`Toolbar.svelte`, `App.svelte`)
  - 設定モーダルから LLMSettings を呼び出し

### ビルド検証
- [x] `go build .`
- [x] `pnpm build`（警告 5 件、エラー 0）
- [x] `pnpm check`

---

## 残りのタスク（優先度順）

### 完了済み（Phase4 実装要点）
- [x] Meta/LLM: LLMConfigStore 経由で `codex-cli` 初期化、接続テストを CLI セッション検証に変更
- [x] Worker: コンテナ起動前に Codex セッション検証を強制し、未ログインなら IDE へエラー通知して中断
- [x] Orchestrator: 実行ログを `task:log` イベントでストリーミング
- [x] UI: LLMSettings を CLI セッション表示に対応（codex-cli 選択可）
- [x] Doc: PRD/TODO/Golden テスト設計を CLI 前提に更新

### 残タスク（フォローアップ）
- [ ] CLI サブスクリプション運用手順を GEMINI.md / CLAUDE.md / guides に追記
- [ ] E2E: CLI セッション未設定時の IDE 通知を含む回帰テストを追加
- [ ] Sandbox Exec で stdin 入力をサポートし、AgentToolProvider の UseStdin を有効化
- [ ] Gemini / Claude Code / Cursor の実プロバイダを実装し、registry stub を置換
- [ ] Meta 層からの WorkerCall 生成で新フィールド（model/flags/env/tool_specific）を活用する経路を整備

---

## 設計上の注意点

### Codex / CLI 統合（現状）
1. **Meta-agent (decompose)**: `internal/meta/client.go` が HTTP で OpenAI Chat Completion を呼び出す（`OPENAI_API_KEY` 必須）。CLI サブスクリプション非対応。
2. **Worker (codex-cli)**: `internal/worker/executor.go` が Docker サンドボックス内で `codex exec ...` を実行。CLI セッション引き継ぎ方法は未整備。

### セッション/環境（現状）

| 項目                    | 用途                                       | 備考                         |
| ----------------------- | ------------------------------------------ | ---------------------------- |
| `MULTIVERSE_META_KIND`  | Meta-agent の種別                          | 現状: mock / openai-chat     |
| `MULTIVERSE_META_MODEL` | Meta-agent のモデル                        | 現状: gpt-5.1 |
| CLI セッション          | Codex / Claude Code / Gemini / Cursor 等   | **API キー不要。要セッション** |

---

## 次のアクション

1. Meta 層を CLI セッション対応に変更する設計・実装方針を決定（AgentToolProvider と整合）
2. `agent-runner` + worker へ CLI セッションを確実に引き継ぐ仕組みを確認（env/マウント/cli path）
3. `go test ./internal/ide/...` 実行で LLMConfigStore の回帰確認
4. ストリーミングログと CLI ベース接続の E2E テストを追加

---

## 追加で必要な対応（漏れ防止メモ）
- [ ] CLI サブスクリプション運用手順のドキュメント化（auth.json / env / codex login）
- [ ] CLI 未ログイン時の IDE 通知と再試行 UX の改善（案内リンク・ボタン）
