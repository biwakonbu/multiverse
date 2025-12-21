# Tooling Selection Design

## 概要

本ドキュメントは、Meta/Worker の実行ツール・モデルをカテゴリ別に選択できる "Tooling" 設計をまとめる。
IDE から詳細設定を編集し、AgentRunner と Orchestrator に反映することを目的とする。

前提: 各 CLI (Codex CLI / Claude Code / Gemini CLI) は最新安定版の利用を想定する。
具体バージョンは運用で固定し、各ガイドに従う (例: `docs/guides/gemini-cli.md`)。

## 目的

- すべてのカテゴリ (meta/task/plan/execution/worker) で詳細なツール・モデル選択を可能にする。
- 率 (weight) による候補配分と、Rate Limit 時の自動切替を提供する。
- 強制モードで "全カテゴリを指定ツール・モデルで実行" を可能にする。
- IDE から設定を調整し、永続化する。

## 主要概念

### Tooling Config

- `runner.tooling` として Task YAML に埋め込む。
- IDE 側は `~/.multiverse/config/tooling.json` に保存する。

一次ソース:
- `pkg/config/tooling.go`
- `pkg/config/config.go`
- `internal/ide/tooling_config.go`
- `internal/orchestrator/executor.go`

### Profile

- `profiles[]` の 1 つを `activeProfile` で選択する。
- `profiles[0]` を暗黙のデフォルトにする。

一次ソース: `internal/tooling/selector.go`

### Category

- `meta`, `plan`, `task`, `execution`, `worker` の 5 種。
- 未定義のカテゴリは `meta` の設定へフォールバックする。

一次ソース: `internal/tooling/selector.go`

### Candidate

- `tool` + `model` の組を候補とする。
- optional: `cliPath`, `flags`, `env`, `toolSpecific`, `systemPrompt` を上書き可能。

一次ソース:
- `pkg/config/tooling.go`
- `internal/meta/cli_provider.go`
- `internal/core/runner.go`

### Force Mode

- `force.enabled=true` の場合、全カテゴリでこの候補を使用する。

一次ソース: `internal/tooling/selector.go`

## 選択アルゴリズム

### ルール

1. `force.enabled` が true の場合は Force Candidate を返す。
2. 指定カテゴリの候補が存在する場合は `strategy` に従う。
3. カテゴリ未定義の場合は `meta` 設定へフォールバック。
4. Rate Limit 判定時に `fallback_on_rate_limit=true` なら候補をクールダウンし再選択。

一次ソース:
- `internal/tooling/selector.go`
- `internal/tooling/rate_limit.go`
- `internal/meta/tooling_client.go`
- `internal/core/runner.go`

### availability 判定

- `openai-chat` は `OPENAI_API_KEY` の有無で判定。
- CLI 系は `exec.LookPath` で存在判定。

一次ソース: `internal/tooling/selector.go`

## デフォルトプロファイル

IDE 既定値は `balanced`。
`balanced` / `fast` の 2 プロファイルを用意する。

一次ソース: `internal/ide/tooling_config.go`

### balanced

- Meta/Plan/Task/Execution: Codex/Claude/Gemini を重みで配分
- Worker: Codex を主軸、Claude/Gemini を補助

### fast

- Meta/Plan/Task/Execution: 低レイテンシ寄りの配分
- Worker: Codex + Gemini Flash を中心

## 統合ポイント

### Meta (IDE / AgentRunner)

- IDE: `newMetaClientFromConfig()` が ToolingConfig を読み込み、ToolingClient を組成する。
- AgentRunner: `cmd/agent-runner` が Task YAML の `runner.tooling` を読み、ToolingClient を使用する。

一次ソース:
- `app.go`
- `cmd/agent-runner/main.go`
- `internal/meta/tooling_client.go`

### Worker (AgentRunner)

- `internal/core/runner.go` が `worker` カテゴリの候補を選択し、`WorkerCall` を上書きする。

一次ソース: `internal/core/runner.go`

### Orchestrator

- `Executor.generateTaskYAML()` が `runner.tooling` を YAML に埋め込む。

一次ソース: `internal/orchestrator/executor.go`

### IDE UI

- TaskBar から Tooling Settings を開く。
- JSON 直接編集 + Force Mode + Active Profile を操作する。

一次ソース:
- `frontend/ide/src/lib/settings/ToolingSettingsPanel.svelte`
- `frontend/ide/src/lib/settings/ToolingSettingsWindow.svelte`
- `frontend/ide/src/lib/hud/TaskBar.svelte`

## 設定例 (JSON)

```json
{
  "activeProfile": "balanced",
  "profiles": [
    {
      "id": "balanced",
      "name": "Balanced",
      "categories": {
        "meta": {
          "strategy": "weighted",
          "fallbackOnRateLimit": true,
          "cooldownSec": 120,
          "candidates": [
            { "tool": "codex-cli", "model": "gpt-5.2", "weight": 40 },
            { "tool": "claude-code", "model": "claude-sonnet-4-5-20250929", "weight": 30 },
            { "tool": "gemini-cli", "model": "gemini-3-pro-preview", "weight": 20 },
            { "tool": "openai-chat", "model": "gpt-5.2", "weight": 10 }
          ]
        },
        "worker": {
          "strategy": "weighted",
          "fallbackOnRateLimit": true,
          "cooldownSec": 120,
          "candidates": [
            { "tool": "codex-cli", "model": "gpt-5.2-codex", "weight": 60 },
            { "tool": "claude-code", "model": "claude-haiku-4-5-20251001", "weight": 25 },
            { "tool": "gemini-cli", "model": "gemini-3-flash-preview", "weight": 15 }
          ]
        }
      }
    }
  ],
  "force": {
    "enabled": false,
    "tool": "",
    "model": ""
  }
}
```

## 既知の制約

- Rate Limit 判定は文字列ベースの簡易判定であり、精度は限定的。
  さらなるエラー型判定の追加は今後の改善余地がある。

一次ソース: `internal/tooling/rate_limit.go`

## テスト

- Selector の基本動作: `internal/tooling/selector_test.go`
- ToolingClient のフォールバック動作: `internal/meta/tooling_client_test.go`
- ToolingConfig 永続化: `internal/ide/tooling_config_test.go`
- Orchestrator の YAML 生成 (golden): `internal/orchestrator/executor_tooling_golden_test.go`

ゴールデンファイル:
- `internal/orchestrator/testdata/task_yaml_with_tooling.golden`
