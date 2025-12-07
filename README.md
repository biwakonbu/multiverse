# agent-runner

AI エージェント（Meta-agent と Worker agent）を組み合わせて、開発タスクを自動化するメタエージェント・オーケストレーションツール。

## 概要

**agent-runner** は、PRD（要件定義）を入力として、Meta-agent（LLM）による計画・評価と、Worker agent（Codex CLI 等）によるコード編集・テスト実行を組み合わせることで、開発タスクを自律的に完遂します。

```
Meta-agent (LLM)
    ↕ YAML プロトコル ↕
AgentRunner Core (状態管理)
    ↕ Docker Sandbox ↕
Worker Agents (Codex CLI等)
```

- 🔒 **安全**: Docker サンドボックス環境での実行
- 📝 **追跡可能**: 全ての実行履歴を Markdown Task Note として記録
- 🔌 **拡張可能**: インターフェース設計によるモック対応

## 技術スタック

- **言語**: Go 1.24.0+
- **外部依存**: Docker、OpenAI API（Meta-agent 用）
- **テスト**: `gopter`（プロパティベーステスト）

## クイックスタート

### 環境構築

```bash
# リポジトリをクローン
git clone https://github.com/biwakonbu/agent-runner.git
cd agent-runner

# 依存関係をインストール
go mod download

# 環境変数を設定
export OPENAI_API_KEY="sk-..."
export CODEX_API_KEY="..."  # または ~/.codex/auth.json を使用
```

### ビルド

```bash
# バイナリをビルド
go build -o agent-runner ./cmd/agent-runner

# または直接実行
go run cmd/agent-runner/main.go < sample_task_go.yaml
```

### テスト実行

```bash
# ユニットテスト（依存なし）
go test ./...

# 統合テスト（Mock 使用）
go test ./test/integration/...

# Docker Sandbox テスト
go test -tags=docker -timeout=10m ./test/sandbox/...

# Codex 統合テスト
go test -tags=codex -timeout=10m ./test/codex/...

# 全テストを実行（推奨）
go test -tags=docker,codex -timeout=15m ./...

# カバレッジレポート生成
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# ゴールデンテスト（一気通しテスト）
## Backend (GT-1, GT-2)
go test -v ./test/integration/... -run TestTaskBuilder_Golden
go test -v -tags=codex ./test/codex/... -run TestCodex_TableDriven/golden_todo.yaml

## Frontend E2E (GT-3)
cd frontend/ide && pnpm test:e2e tests/golden_flow.spec.ts && cd ../..
```

### サンドボックス環境のセットアップ

```bash
# Codex worker のランタイムをビルド
docker build -t agent-runner-codex:latest sandbox/

# 軽量テスト用イメージをビルド
docker build -t agent-runner-test:latest test/sandbox/ -f test/sandbox/Dockerfile.test

# Docker Sandbox テストを実行
go test -tags=docker -timeout=10m ./test/sandbox/...

# Codex 統合テスト
go test -tags=codex -timeout=10m ./test/codex/...
```

## 使用方法

### タスク定義ファイル（YAML）

```yaml
version: 1

task:
  id: "TASK-001"
  title: "新機能を実装"
  repo: "/path/to/repo" # 絶対パスを推奨

  prd:
    text: |
      # 要件定義
      - 機能 A を実装
      - テストを作成
      - ドキュメント更新

  test:
    command: "go test ./..."
    cwd: "./"

runner:
  meta:
    kind: "openai-chat"
    model: "gpt-5.1-codex-max-high" # または --meta-model フラグで指定

  worker:
    kind: "codex-cli"
    docker_image: "agent-runner-codex:latest"
    max_run_time_sec: 1800
```

### 実行

```bash
./agent-runner < task.yaml
```

実行完了後、`.agent-runner/task-TASK-001.md` に詳細な履歴が記録されます。

## アーキテクチャ

### 三層構造

| レイヤー             | 役割                           | 主要コンポーネント        |
| -------------------- | ------------------------------ | ------------------------- |
| **Meta-agent**       | タスク計画・評価               | OpenAI API クライアント   |
| **AgentRunner Core** | 状態管理・オーケストレーション | Runner、TaskContext、FSM  |
| **Worker Agents**    | 実装・テスト実行               | Codex CLI、Docker Sandbox |

### パッケージ構成

```
internal/
├── core/          # タスク FSM、状態管理
├── meta/          # LLM 通信、YAML プロトコル
├── worker/        # Worker 実行、Docker 管理
├── note/          # Task Note 生成
└── mock/          # テスト用モック実装

pkg/
└── config/        # 設定構造体

cmd/
└── agent-runner/  # エントリポイント
```

詳細は [docs/AgentRunner-architecture.md](docs/AgentRunner-architecture.md) を参照してください。

---

## multiverse IDE (Desktop Application)

**multiverse IDE** は、agent-runner Core をデスクトップアプリケーションから操作するための GUI ツールです。Wails + Svelte + TypeScript で構築されています。

### ビルド方法

```bash
# 前提条件: Wails CLI のインストール
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# フロントエンドの依存関係をインストール
cd frontend/ide && pnpm install && cd ../..

# デスクトップアプリをビルド
wails build

# 生成されたアプリ
# macOS: build/bin/multiverse.app
```

### 起動方法

```bash
# macOS
open build/bin/multiverse.app

# または直接実行
./build/bin/multiverse.app/Contents/MacOS/multiverse
```

### 使い方

1. **Workspace 選択**: アプリ起動時にプロジェクトルートを選択
2. **Task 作成**: 「New Task」ボタンでタスクを作成（Title と Pool ID を入力）
3. **Task 実行**: Task 詳細画面で「Run」ボタンをクリック
4. **ステータス確認**: ポーリングで自動更新（2 秒間隔）

### 運用上の注意点

| 項目                  | 説明                                       |
| --------------------- | ------------------------------------------ |
| **データ保存先**      | `~/.multiverse/workspaces/<workspace-id>/` |
| **Task ファイル**     | `tasks/<task-id>.jsonl` (JSONL 形式)       |
| **Attempt ファイル**  | `attempts/<attempt-id>.json` (JSON 形式)   |
| **agent-runner パス** | 現在は `./agent-runner` を前提（改善予定） |

### 既知の制限事項

1. **agent-runner バイナリの配置**: IDE は `./agent-runner` バイナリが同じディレクトリにあることを前提としています

   ```bash
   # 対策: agent-runner を先にビルドしてコピー
   go build -o build/bin/multiverse.app/Contents/MacOS/agent-runner ./cmd/agent-runner
   ```

2. **Worker CLI の設定**: 現在は `codex` CLI をハードコードしています（将来的に設定可能にする予定）

3. **TypeScript 設定の警告**: `tsconfig.json` に警告が出ますが、ビルドには影響しません

### テストコマンド

```bash
# Backend テスト
go test -v ./internal/ide/...
go test -v ./internal/orchestrator/...

# フロントエンドビルド確認
cd frontend/ide && pnpm run build && cd ../..

# 全体ビルド確認
wails build
```

### トラブルシューティング

| 問題                       | 原因                   | 対策                                            |
| -------------------------- | ---------------------- | ----------------------------------------------- |
| Wails build が失敗         | Node.js 未インストール | `brew install node`                             |
| フロントエンドビルドエラー | 依存関係不足           | `cd frontend/ide && pnpm install`               |
| Task 実行が失敗            | agent-runner がない    | `go build -o ./agent-runner ./cmd/agent-runner` |
| Workspace が見つからない   | パーミッション         | `~/.multiverse/` の権限を確認                   |

### ディレクトリ構成

```
cmd/
├── multiverse/        # IDE バックエンド（Wails バインディング）
└── multiverse-orchestrator/  # Orchestrator CLI（将来用）

internal/
├── ide/                   # Workspace 管理
└── orchestrator/          # Task/Attempt 永続化、スケジューラ、Executor

frontend/
└── ide/                   # Svelte + TypeScript フロントエンド
    └── src/
        ├── App.svelte
        └── lib/
            ├── WorkspaceSelector.svelte
            ├── TaskList.svelte
            ├── TaskDetail.svelte
            └── TaskCreate.svelte
```

---

## 開発ガイド

### コード規約

- **言語**: コメントは日本語、関数・変数名は英語
- **テスト**: 依存性注入でモック化、プロパティベーステストで不変条件検証
- **ロギング**: 現在 `fmt.Printf` を使用（今後 `slog` への移行を検討）

### テスト戦略

- **ユニットテスト**: `internal/mock` でモック実装を注入、個別パッケージの機能検証
- **プロパティベーステスト**: `gopter` で状態遷移の不変条件を検証
- **Mock 統合テスト**: 複数コンポーネントの連携確認（外部依存なし）
- **Docker Sandbox テスト**: 実際の Docker API とコンテナ管理の動作検証（`-tags=docker`）
- **Codex 統合テスト**: 実際の Codex CLI による end-to-end テスト（`-tags=codex`）

詳細は [TESTING.md](TESTING.md) を参照してください。

### 既知の課題

**相対パスの解決**

- タスク設定で相対パス `"."` を使用すると Docker マウントエラーが発生します
- 対応：絶対パスを使用するか、`worker/executor.go` で `filepath.Abs` を使用してください

## ドキュメント

- **[CLAUDE.md](CLAUDE.md)** - Claude Code 開発ガイド
- **[TESTING.md](TESTING.md)** - テストベストプラクティス
- **[docs/AgentRunner-architecture.md](docs/AgentRunner-architecture.md)** - アーキテクチャ詳細仕様
- **[docs/agentrunner-spec-v1.md](docs/agentrunner-spec-v1.md)** - MVP/v1 仕様書
- **[docs/AgentRunner-impl-design-v1.md](docs/AgentRunner-impl-design-v1.md)** - Go 実装設計

## 貢献

このプロジェクトはオープンソースです。バグ報告、機能提案、プルリクエストを歓迎します。

### PR を作成する前に

1. fork してブランチを作成
2. 変更内容をテストで検証（`go test ./...`）
3. TESTING.md のガイドラインに従ってテストを追加
4. コミットメッセージは日本語で記載
5. PR を作成

## ライセンス

MIT License - [LICENSE](LICENSE) を参照してください。

## 参考リンク

- [OpenAI API ドキュメント](https://platform.openai.com/docs)
- [Docker ドキュメント](https://docs.docker.com/)
- [Codex CLI](https://github.com/openai/codex)

# Test comment for pre-commit hook
