# internal/GEMINI.md

このディレクトリには、外部に公開しないアプリケーションの内部ロジックおよびライブラリが含まれています。

## パッケージ構成

### Core / Sandbox Layer (AgentRunner)

- **`core/`**: タスク実行ステートマシン (FSM)、コンテキスト管理。
- **`worker/`**: Docker サンドボックス制御、Worker CLI (Codex 等) の実行ラッパー。
- **`meta/`**: LLM (Meta-agent) との通信、プロンプト制御、YAML プロトコル処理。
- **`note/`**: タスク実行後の Markdown ノート生成ロジック。

### Orchestration Layer (Multiverse)

- **`orchestrator/`**:
  - **`Executor`**: `agent-runner` プロセスの起動と監視。
  - **`TaskStore`**: タスクメタデータと実行履歴の永続化 (`$HOME/.multiverse` 以下)。
  - **`Scheduler` / `IPC`**: タスクキュー管理とプロセス間通信。
  - **`TaskGraphManager`**: 依存グラフ管理、トポロジカルソート。
  - **`ExecutionOrchestrator`**: 自律実行ループ、一時停止/再開機能。
  - **`BacklogStore`**: 問題・検討材料管理。
  - **`RetryPolicy`**: 指数バックオフリトライ、永続化対応。

### Chat Layer

- **`chat/`**:
  - **`ChatHandler`**: チャット →Meta-agent→ タスク分解。
  - **`ChatSessionStore`**: セッション・メッセージ履歴の JSONL 永続化。

### IDE Layer

- **`ide/`**:
  - Wails アプリケーション固有のバックエンドロジック。
  - Workspace の検出・管理ロジック。

### Utility Layer

- **`cli/`**: コマンドラインフラグ定義・解析。
- **`logging/`**: 構造化ログ（log/slog）、Trace ID 伝播。
- **`mock/`**: テスト用モック実装（Function Field Injection パターン）。
