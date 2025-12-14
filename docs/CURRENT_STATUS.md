# Current System Status Report

最終更新: 2025-12-14

## 1. コンポーネント開発ステータス

システム全体の主要コンポーネントの実装状況と、仕様との一致度まとめています。

### Project Structure

- **Backend (Go)**: `internal/core`, `internal/orchestrator`, `internal/worker` など主要パッケージは実装済み。
- **Frontend (Svelte/Wails)**: IDE フロントエンドは `frontend/ide` に実装済み。基本的な UI (Task Graph, Chat, Backlog) は構築完了。

### Status Matrix

| コンポーネント       | ステータス | 詳細                                                                                                     | 備考                        |
| -------------------- | ---------- | -------------------------------------------------------------------------------------------------------- | --------------------------- |
| **AgentRunner Core** | 🟢 Stable  | Task FSM, Meta-agent 通信, Docker Sandbox 制御は安定稼働。                                               | `runner.go`, `sandbox.go`   |
| **Worker Executor**  | 🟢 Stable  | Docker コンテナ内でのコマンド実行、環境変数注入は実装済み。                                              | `worker/executor.go`        |
| **Orchestrator**     | 🟡 Beta    | 基本的なタスク実行ループは動作。Force Stop, Retry は実装済みだが、IPC がファイルベースなど拡張余地あり。 | `execution_orchestrator.go` |
| **Task Store**       | 🟢 Stable  | ファイルベース (`~/.multiverse`) でのタスク永続化は実装済み。                                            | `task_store.go`             |
| **IPC**              | 🟡 Beta    | ファイルポーリングベースの実装。WebSocket 化は未着手。                                                   | `ipc/filesystem_queue.go`   |
| **IDE Frontend**     | 🟡 Beta    | タスク作成、監視フローは実装済み。E2E テスト調整中。                                                     | Svelte + Wails              |

## 2. Orchestrator 実装詳細

仕様書 (`docs/specifications/orchestrator-spec.md`) 以降に追加・詳細化された重要機能です。

### Execution State Machine

`ExecutionOrchestrator` は以下の状態を持ち、明示的な停止・再開が可能です。

- `IDLE`: 停止中。タスクを処理しない。
- `RUNNING`: 稼働中。キューをポーリングし、タスクを実行する。
- `PAUSED`: 一時停止中。新規タスクの開始を保留する。

### Force Stop & Cleanup

- **`Stop()` メソッド**: オーケストレーターのループを停止し、現在実行中のタスク（Attempt）があれば `context.CancelFunc` を通じて強制終了します。
- **Graceful Shutdown**: 実行中の `agent-runner` プロセスはコンテキストキャンセルによりシグナルを受け取り、Docker コンテナの停止（Cleanup）を試みます。

### Reliability (Retry & Backlog)

`HandleFailure` メソッドにより、タスク失敗時の挙動を制御します。

1. **Retry Strategy**:

   - `AttemptCount` に基づいてバックオフ時間を計算（Exponential Backoff）。
   - タスクステータスを `RETRY_WAIT` に変更し、将来の再実行をスケジュール。
   - 再実行待ちのタスクは、時間が来ると `PENDING` にリセットされ再スケジュールされる。

2. **Backlog Strategy**:
   - リトライ上限を超えた、あるいは致命的なエラーの場合、タスクを **Backlog** (`BacklogStore`) に移動。
   - ユーザーによるレビュー待ち状態とする。

## 3. 既知の制約 (Known Limitations)

### Executor / AgentRunner

- **デフォルト設定**: `Executor` が生成する Task YAML はデフォルトで `runner.max_loops: 5` と `runner.worker.kind: "codex-cli"` を設定します（`state/tasks.json` の `inputs.runner_max_loops` / `inputs.runner_worker_kind` で上書き可能）。
- **Meta-agent プロバイダ（IDE）**: IDE 経由の計画/分解は `openai-chat`（HTTP/OpenAI API）または `codex-cli`（Codex CLI）を使用します。デフォルトは `openai-chat` ですが、`OPENAI_API_KEY` 未設定かつ `codex` が利用可能な場合は `codex-cli` に自動フォールバックします（`app.go` の `newMetaClientFromConfig()`、`internal/meta/client.go` の `Decompose()` 参照）。
- **Worker 環境**: 現在は Docker ベースの実行のみサポート。

### IPC

- **ポーリング負荷**: ファイルシステムポーリング (2 秒間隔) を行っているため、大量のジョブがある場合の性能に懸念あり。将来的に WebSocket または gRPC への移行推奨。

### Testing

- **E2E テスト**: Frontend の E2E テスト (`pnpm test:e2e`) は存在するが、タイムアウト等の問題で調整が必要な状態。

## 4. Recent Hardening (vNext P0)

MVP 完了後の品質強化（Quality Hardening）として、以下の重要項目が完了しています（2025-12-14 時点）。

1.  **QH-001 (Structure Context)**: `plan_patch` への入力コンテキスト決定論化（Task 200 件 / WBS Node 200 件のトリミング仕様の確立）。
2.  **QH-003 (Delete Semantics)**: `cascade=false` 時の splice 挙動による WBS 破壊防止（孤児ノード防止）。
3.  **QH-004 (Recoverability)**: History Append -> State Update の順序保証と、失敗時の `history_failed` レコード記録。
4.  **QH-005 (Mock Stability)**: テスト用 Mock の構造ベース移行により、外部ネットワーク依存なしで `go test ./...` が安定通過。
