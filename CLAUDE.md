# CLAUDE.md

このファイルはこのリポジトリで Claude コードを操作する際のガイダンスを提供します。

## プロジェクト概要

**multiverse** は、開発タスクを自動実行する統合プラットフォームです。以下の4層で構成されています:

- **multiverse IDE**: デスクトップ UI（Wails + Svelte）でタスクを GUI 管理
- **Orchestrator**: タスクのスケジューリング・永続化・IPC キュー管理
- **AgentRunner Core**: Meta-agent（LLM）による計画・評価と FSM ベースの状態管理
- **Worker Agents**: Docker サンドボックス内で実際の開発作業（コーディング、テスト）を実行

詳細な要件は [PRD.md](PRD.md) を参照してください。

## プロジェクト構造

```
multiverse/
├── cmd/
│   ├── agent-runner/              # 既存 Core CLI エントリポイント
│   │   └── main.go
│   ├── multiverse-ide/            # Wails デスクトップアプリ
│   │   ├── main.go                # Wails 初期化・Asset 埋め込み
│   │   └── app.go                 # IDE バックエンドロジック
│   └── multiverse-orchestrator/   # Orchestrator CLI（実装予定）
│
├── internal/
│   ├── core/                      # タスク実行エンジン ★詳細は core/CLAUDE.md
│   │   ├── runner.go              # FSM オーケストレーション
│   │   └── context.go             # TaskContext・TaskState 定義
│   ├── meta/                      # Meta-agent 通信層 ★詳細は meta/CLAUDE.md
│   │   ├── client.go              # OpenAI API 通信
│   │   └── protocol.go            # YAML プロトコル定義
│   ├── worker/                    # Worker 実行・サンドボックス ★詳細は worker/CLAUDE.md
│   │   ├── executor.go            # Worker CLI 実行の抽象化
│   │   └── sandbox.go             # Docker API 管理
│   ├── note/                      # Task Note 生成
│   │   └── writer.go              # Markdown テンプレート出力
│   ├── mock/                      # テスト用モック実装
│   │   ├── meta.go
│   │   ├── worker.go
│   │   └── note.go
│   │
│   ├── orchestrator/              # Orchestrator ドメインロジック（新規）
│   │   ├── task_store.go          # Task/Attempt の JSONL/JSON 永続化
│   │   ├── scheduler.go           # タスクスケジューリング
│   │   └── ipc/
│   │       └── filesystem_queue.go # ファイルベース IPC キュー
│   │
│   └── ide/                       # IDE バックエンドロジック（新規）
│       └── workspace_store.go     # Workspace メタデータ管理
│
├── frontend/
│   └── ide/                       # Svelte + TypeScript フロントエンド
│       ├── src/
│       │   ├── App.svelte         # メインコンポーネント
│       │   ├── main.ts            # エントリポイント
│       │   └── lib/               # UI コンポーネント
│       │       ├── WorkspaceSelector.svelte
│       │       ├── TaskList.svelte
│       │       ├── TaskCreate.svelte
│       │       └── TaskDetail.svelte
│       ├── package.json
│       └── vite.config.ts
│
├── pkg/
│   └── config/                    # 公開パッケージ（YAML 設定）
│       └── config.go              # TaskConfig 構造体定義
│
├── test/                          # ★詳細は test/CLAUDE.md
│   ├── integration/               # Mock 統合テスト
│   ├── sandbox/                   # Docker テスト（-tags=docker）
│   └── codex/                     # Codex 統合テスト（-tags=codex）
│
├── docs/                          # 設計・仕様・開発ガイド
├── examples/                      # サンプルタスク・スクリプト
├── sandbox/                       # Worker Docker イメージ定義
│   └── Dockerfile
│
├── wails.json                     # Wails プロジェクト設定
├── PRD.md                         # multiverse IDE v0.1 要件書
├── go.mod
└── Makefile
```

### ディレクトリの役割

| ディレクトリ | 役割 | 詳細 |
|------------|------|------|
| `cmd/agent-runner/` | Core CLI エントリポイント | [CLAUDE.md](cmd/agent-runner/CLAUDE.md) |
| `cmd/multiverse-ide/` | Wails デスクトップアプリ | [CLAUDE.md](cmd/multiverse-ide/CLAUDE.md) |
| `cmd/multiverse-orchestrator/` | Orchestrator CLI | [CLAUDE.md](cmd/multiverse-orchestrator/CLAUDE.md) |
| `internal/core/` | タスク FSM とオーケストレーション | [CLAUDE.md](internal/core/CLAUDE.md) |
| `internal/meta/` | Meta-agent（LLM）との通信層 | [CLAUDE.md](internal/meta/CLAUDE.md) |
| `internal/worker/` | Worker 実行と Docker 管理 | [CLAUDE.md](internal/worker/CLAUDE.md) |
| `internal/note/` | タスク実行履歴の Markdown 出力 | [CLAUDE.md](internal/note/CLAUDE.md) |
| `internal/mock/` | テスト用モック実装 | [CLAUDE.md](internal/mock/CLAUDE.md) |
| `internal/orchestrator/` | Task 永続化・スケジューラ・IPC | [CLAUDE.md](internal/orchestrator/CLAUDE.md) |
| `internal/ide/` | Workspace メタデータ管理 | [CLAUDE.md](internal/ide/CLAUDE.md) |
| `internal/cli/` | CLI フラグ処理 | [CLAUDE.md](internal/cli/CLAUDE.md) |
| `frontend/ide/` | Svelte + TS フロントエンド | [CLAUDE.md](frontend/ide/CLAUDE.md) |
| `pkg/config/` | YAML 設定パース（再利用可能） | [CLAUDE.md](pkg/config/CLAUDE.md) |
| `test/` | 4 段階テスト戦略 | [CLAUDE.md](test/CLAUDE.md) |
| `docs/` | 設計・仕様・開発ガイド | [CLAUDE.md](docs/CLAUDE.md) |
| `examples/` | サンプルタスク・実行スクリプト | [CLAUDE.md](examples/CLAUDE.md) |

## よく使うコマンド

### ビルド

```bash
# AgentRunner Core CLI をビルド
go build ./cmd/agent-runner
go build -o agent-runner ./cmd/agent-runner
```

### multiverse IDE（Wails）

```bash
# 開発モード（ホットリロード）
wails dev

# 本番ビルド（単一バイナリ生成）
wails build
```

### フロントエンド開発

```bash
# 依存パッケージインストール
cd frontend/ide && pnpm install

# 開発サーバー起動
cd frontend/ide && pnpm run dev

# 本番ビルド
cd frontend/ide && pnpm run build
```

### テスト

**詳細は [test/CLAUDE.md](test/CLAUDE.md) を参照してください。**

```bash
# ユニットテスト（依存なし、高速）
go test ./...

# 全テスト実行（Docker + Codex CLI 必須）
go test -tags=docker,codex -timeout=15m ./...

# カバレッジレポート生成
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

### AgentRunner Core 実行

```bash
# 標準入力から YAML を読み込む
./agent-runner < task.yaml

# パイプで実行
cat task.yaml | go run cmd/agent-runner/main.go
```

### Docker（Worker 実行環境）

```bash
# Codex Worker ランタイムをビルド
docker build -t agent-runner-codex:latest sandbox/
```

## アーキテクチャ

### 4層構造

```
┌─────────────────────────────────────┐
│  multiverse-ide (Desktop UI)       │
│  - Wails + Go Backend              │
│  - Svelte + TS Frontend            │
└──────────────┬──────────────────────┘
               │ Wails IPC
┌──────────────▼──────────────────────┐
│  Orchestrator Layer                 │
│  - WorkspaceStore                   │
│  - TaskStore (JSONL)                │
│  - Scheduler + IPC Queue            │
└──────────────┬──────────────────────┘
               │ 呼び出し
┌──────────────▼──────────────────────┐
│  AgentRunner Core (CLI)             │
│  - FSM オーケストレーション          │
│  - Meta-agent 通信                  │
└──────────────┬──────────────────────┘
               │ Docker Exec
┌──────────────▼──────────────────────┘
│  Worker (Docker Sandbox)            │
│  - Codex CLI 等                     │
└─────────────────────────────────────┘
```

### AgentRunner Core の状態機械（FSM）

```
PENDING → PLANNING → RUNNING → VALIDATING → COMPLETE
                                              ↓
                                            FAILED
```

**主要な遷移**:

- `PENDING → PLANNING`: Meta-agent が PRD から受け入れ条件（AC）を生成
- `PLANNING → RUNNING`: Meta-agent が worker 実行を要求
- `RUNNING → VALIDATING`: Worker 実行完了、Meta-agent が完了を評価
- `VALIDATING → COMPLETE/FAILED`: タスク完了判定

### multiverse IDE の Task 状態

```
PENDING → READY → RUNNING → SUCCEEDED / FAILED / CANCELED / BLOCKED
```

## データモデル（multiverse IDE）

全てのデータは `$HOME/.multiverse/workspaces/<workspace-id>/` 以下に保存されます。

### Workspace

```
~/.multiverse/workspaces/<workspace-id>/
├── workspace.json          # Workspace メタデータ
├── config/
│   └── worker-pools.json   # WorkerPool 設定
├── tasks/
│   └── <task-id>.jsonl     # Task 状態履歴（最後の行が最新）
├── attempts/
│   └── <attempt-id>.json   # Attempt 詳細
├── ipc/
│   ├── queue/<pool-id>/    # Orchestrator → Worker
│   └── results/            # Worker → Orchestrator
└── logs/
```

**workspace-id** = `sha1(projectRoot)[:12]`（決定的 ID）

### workspace.json

```json
{
  "version": "1.0",
  "projectRoot": "/path/to/project",
  "displayName": "My Project",
  "createdAt": "2024-12-05T...",
  "lastOpenedAt": "2024-12-05T..."
}
```

### Task（JSONL 形式）

`tasks/<task-id>.jsonl` - 1 行 = 1 JSON オブジェクト、最後の行が最新状態

```jsonl
{"id":"task-1","title":"Feature A","status":"PENDING","poolId":"codegen","createdAt":"...","updatedAt":"..."}
{"id":"task-1","title":"Feature A","status":"RUNNING","poolId":"codegen","createdAt":"...","updatedAt":"...","startedAt":"..."}
{"id":"task-1","title":"Feature A","status":"SUCCEEDED","poolId":"codegen","createdAt":"...","updatedAt":"...","startedAt":"...","doneAt":"..."}
```

### Attempt（JSON 形式）

`attempts/<attempt-id>.json` - 1 Attempt 1 ファイル

```json
{
  "id": "attempt-1",
  "taskId": "task-1",
  "status": "SUCCEEDED",
  "startedAt": "2024-12-05T...",
  "finishedAt": "2024-12-05T...",
  "errorSummary": ""
}
```

### IPC Queue

`ipc/queue/<pool-id>/<job-id>.json` - Orchestrator → Worker へのジョブ投入

```json
{
  "id": "job-1",
  "taskId": "task-1",
  "poolId": "codegen",
  "payload": { ... }
}
```

## 主要パッケージ

### `/internal/core` - AgentRunner タスク実行エンジン

- **runner.go**: FSM オーケストレーション、TaskContext 管理
- **context.go**: TaskContext（PRD、AC、worker 実行、meta 呼び出し）と TaskState
- プロパティベーステスト（gopter）で状態不変条件を検証
- **詳細**: [internal/core/CLAUDE.md](internal/core/CLAUDE.md)

### `/internal/meta` - Meta-agent 通信層

- **client.go**: OpenAI API 通信、モック実装対応（`kind: "mock"`）
- **protocol.go**: YAML メッセージ構造
- **詳細**: [internal/meta/CLAUDE.md](internal/meta/CLAUDE.md)

### `/internal/worker` - Worker 実行とサンドボックス

- **executor.go**: Worker CLI 実行の抽象化
- **sandbox.go**: Docker API 管理（コンテナ作成、exec、クリーンアップ）
- **詳細**: [internal/worker/CLAUDE.md](internal/worker/CLAUDE.md)

### `/internal/orchestrator` - Orchestrator ドメインロジック

- **task_store.go**: Task/Attempt の永続化（JSONL/JSON）
- **scheduler.go**: タスクスケジューリング
- **ipc/filesystem_queue.go**: ファイルベース IPC キュー
- **詳細**: [internal/orchestrator/CLAUDE.md](internal/orchestrator/CLAUDE.md)

### `/internal/ide` - IDE バックエンドロジック

- **workspace_store.go**: Workspace メタデータ管理
- **詳細**: [internal/ide/CLAUDE.md](internal/ide/CLAUDE.md)

### `/internal/cli` - CLI フラグ処理

- **flags.go**: コマンドラインフラグ定義とパース
- **詳細**: [internal/cli/CLAUDE.md](internal/cli/CLAUDE.md)

### `/internal/note` - Task Note 生成

- **writer.go**: TaskContext から Markdown Task Note を生成
- **詳細**: [internal/note/CLAUDE.md](internal/note/CLAUDE.md)

### `/internal/mock` - テスト用モック実装

- Function Field Injection パターンで依存性注入
- **詳細**: [internal/mock/CLAUDE.md](internal/mock/CLAUDE.md)

### `/pkg/config` - 設定管理

- **config.go**: タスク設定構造体（YAML パース）
- **詳細**: [pkg/config/CLAUDE.md](pkg/config/CLAUDE.md)

## タスク YAML 形式（AgentRunner Core）

```yaml
version: 1

task:
  id: "TASK-123"
  title: "タスク名"
  repo: "."

  prd:
    path: "./docs/prd.md"
    # または text: で直接埋め込み

  test:
    command: "npm test"
    cwd: "./"

runner:
  meta:
    kind: "openai-chat"  # または "mock"
    model: "gpt-5.1-codex-max-high"

  worker:
    kind: "codex-cli"
    docker_image: "agent-runner-codex:latest"
    max_run_time_sec: 1800
    env:
      CODEX_API_KEY: "env:CODEX_API_KEY"
```

## 重要な設計パターン

### 依存性注入（Dependency Injection）

`Runner` 構造体はインターフェースを受け入れ、モックベーステストを可能にする:

```go
type Runner struct {
    Config *config.TaskConfig
    Meta   MetaClient        // インターフェース
    Worker WorkerExecutor    // インターフェース
    Note   NoteWriter        // インターフェース
}
```

### TaskContext の伝播

実行状態（PRD、AC、worker 結果、meta 呼び出し、タイミング）が単一の `TaskContext` で保持され、FSM を通じて伝播。完全な再現性と監査証跡（Task Note）を実現。

### Sandbox 隔離

- タスク 1 つあたり 1 つの Docker コンテナ
- `/workspace/project` にリポジトリをマウント
- 完了またはエラー時の自動クリーンアップ

### ファイルベース IPC

- Orchestrator ↔ Worker 間は JSON ファイルで通信
- IDE は IPC ファイルを直接扱わず、Task/Attempt の状態を読む

## 技術スタック

### バックエンド

- **Go 1.23+**: 全てのバックエンドコンポーネント
- **Docker**: Worker 実行環境
- **OpenAI API**: Meta-agent（LLM）通信

### フロントエンド（multiverse IDE）

- **Wails v2**: Go + Web フロントエンドのデスクトップアプリフレームワーク
- **Svelte 4**: リアクティブ UI フレームワーク
- **TypeScript 5**: 型安全なフロントエンド
- **Vite 5**: 高速ビルドツール

## 環境変数

```bash
# Meta-agent 用（OpenAI API）
export OPENAI_API_KEY="sk-..."

# Worker agent 用（auth.json 未使用の場合）
export CODEX_API_KEY="..."
```

## サブディレクトリメモリ体系

各パッケージ内 `CLAUDE.md` は標準フォーマット: **責務** → **主要概念** → **実装パターン** → **テスト戦略** → **拡張・既知問題**

### 実装層パッケージ

| 層 | パッケージ | 責務 | 詳細 |
|----|-----------|------|------|
| **IDE** | ide | Workspace メタデータ管理 | [CLAUDE.md](internal/ide/CLAUDE.md) |
| **IDE** | orchestrator | Task/Attempt 永続化・スケジューラ | [CLAUDE.md](internal/orchestrator/CLAUDE.md) |
| **Core** | core | FSM・TaskContext・状態遷移 | [CLAUDE.md](internal/core/CLAUDE.md) |
| **Core** | meta | LLM 通信・YAML プロトコル | [CLAUDE.md](internal/meta/CLAUDE.md) |
| **Core** | worker | CLI 実行・Docker サンドボックス | [CLAUDE.md](internal/worker/CLAUDE.md) |
| **Util** | cli | CLI フラグ処理 | [CLAUDE.md](internal/cli/CLAUDE.md) |
| **Util** | note | Task Note 生成・テンプレート | [CLAUDE.md](internal/note/CLAUDE.md) |
| **Util** | mock | テストダブル・FuncField 注入 | [CLAUDE.md](internal/mock/CLAUDE.md) |
| **Config** | pkg/config | YAML 設定スキーマ | [CLAUDE.md](pkg/config/CLAUDE.md) |

### コマンド層

| パッケージ | 責務 | 詳細 |
|-----------|------|------|
| cmd/agent-runner | AgentRunner Core CLI | [CLAUDE.md](cmd/agent-runner/CLAUDE.md) |
| cmd/multiverse-ide | Wails デスクトップアプリ | [CLAUDE.md](cmd/multiverse-ide/CLAUDE.md) |
| cmd/multiverse-orchestrator | Orchestrator CLI | [CLAUDE.md](cmd/multiverse-orchestrator/CLAUDE.md) |

### フロントエンド

| パッケージ | 責務 | 詳細 |
|-----------|------|------|
| frontend/ide | Svelte + TypeScript UI | [CLAUDE.md](frontend/ide/CLAUDE.md) |

### テスト戦略

**[test/CLAUDE.md](test/CLAUDE.md)** - 4 段階テスト戦略、ビルドタグ、精度管理、コマンドリファレンス

## ドキュメント体系

### プロジェクト定義

- **[PRD.md](PRD.md)** - multiverse IDE v0.1 要件書

### AI 開発者向け（メモリ・操作ガイド）

- **[CLAUDE.md](CLAUDE.md)**（このファイル）- プロジェクト全体ガイド
- **[docs/CLAUDE.md](docs/CLAUDE.md)** - ドキュメント整理ルール・命名規則
- **各 internal/\*/CLAUDE.md** - パッケージ実装ガイダンス
- **[test/CLAUDE.md](test/CLAUDE.md)** - テスト戦略・ビルドタグ・精度管理

### docs/ - 設計・仕様・開発ガイド

| ファイル | 対象 | 用途 |
|---------|------|------|
| [README.md](docs/README.md) | 全員 | ドキュメントナビゲーション |
| **specifications/** | 実装者 | 確定仕様 |
| **design/** | アーキテクト | 設計ドキュメント |
| **guides/** | 開発者 | 開発ガイド |

### エンドユーザー向け（不変）

- **[README.md](README.md)** - プロジェクト紹介（変更禁止）
- **[GEMINI.md](GEMINI.md)** - プロジェクト背景（変更禁止）

## 開発ノート

- **言語**: コメント・ドキュメントは日本語、コード（関数・変数名）は英語
- **ドキュメント管理**: [docs/CLAUDE.md](docs/CLAUDE.md) で命名規則・責務分担を一元管理
- **テスト**: 依存性注入とモック使用、プロパティベーステストで不変条件検証
- **テスト自動化**: ビルドタグ（-tags=docker,codex）による段階的実行
- **ロギング**: 構造化ログ（log/slog）
- **依存関係**: Go 1.23+、Docker、OpenAI API、Wails v2、Node.js

## プロジェクトメモリ管理ルール

**CLAUDE.md には静的な構造・設計情報のみを記載します。**

### 記載すべき内容

- プロジェクト構造とアーキテクチャ
- 設計パターンと実装ガイドライン
- よく使うコマンドと操作方法
- 既知の課題と対応方法
- サブディレクトリメモリ体系

### 記載しない内容

- Phase 履歴や開発進捗
- カバレッジ推移や改善履歴
- 完了時の成果テーブル
- 「完了」「次フェーズの予定」等の進捗状況

**理由**: 進捗情報は時間とともに陳腐化し、メモリの可読性を低下させます。履歴情報は Git コミットログで管理してください。
