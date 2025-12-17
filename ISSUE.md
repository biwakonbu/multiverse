# ISSUE Log (2025-12-07)

## Open Items

- [x] Meta 層が CLI サブスクリプション未対応（API キー不要方針と不整合）

  - `app.go` の `newMetaClientFromEnv()` と `chat.NewHandler` が HTTP クライアント (OPENAI_API_KEY 前提) を生成し、LLMConfigStore/設定画面を参照しない。Codex / Gemini / Claude Code / Cursor などの CLI セッション再利用方針と乖離。
  - 対応完了（2025-12-14）: Meta プロバイダを `CLIProvider` (汎用) と `OpenAIProvider` に分離し、API キーなしでの利用を可能にした。`app.go` は `claude` または `codex` を自動検出し、サブスクリプションベースの利用を優先するよう改修済み。

- [x] LLM 設定 UI が実行系に反映されず、API キー前提の表示が残存

  - `LLMSettings` は Kind/Model を保存するが、CLI セッション状態を表示できず、`TestLLMConnection` も OpenAI HTTP 前提で CLI セッションを検証できない。API キーは不要なので UI をセッション表示に置換する必要。
  - 対応完了（2025-12-14）: API キーの強制を解除し、`TestLLMConnection` がプロバイダ経由で CLI バージョンチェックを行うように変更済み。設定画面で API キーが空でも CLI があれば接続成功となる。

- [x] 実行ログ（stdout/stderr）のリアルタイム配信/表示の整備

  - バックエンド: stdout/stderr を逐次読み取り、`task:log` イベントを送出済み（`internal/orchestrator/executor.go:93`、`internal/orchestrator/executor.go:121`、`internal/orchestrator/events.go:39`）。
  - フロント: `task:log` を購読して store に蓄積する実装は存在（`frontend/ide/src/stores/logStore.ts:49`）。
  - 対応完了（2025-12-14）: UI 上でのタスク別フィルタ/クリア導線/常時表示など、運用可能な表示体験に仕上げた（`frontend/ide/src/lib/hud/LiveLogStream.svelte`、`frontend/ide/src/stores/logStore.ts`）。

- [x] Codex CLI セッションの存在確認・注入手段が未整備
  - Worker Executor は `codex exec ...` を呼び出すが、セッション有無の検証・警告やコンテナへのセッション注入方法（環境変数/ボリューム）が明確でない。
  - 対応完了（2025-12-14）: コンテナ起動時にセッション確認を行い、失敗時はユーザー向けエラーメッセージを返す。環境変数（`CODEX_SESSION_TOKEN`/`CODEX_API_KEY`/`ANTHROPIC_API_KEY`）をコンテナへ注入する（`internal/worker/executor.go`）。

## Deferred (moved from TODO.md, 2025-12-13)

- [ ] 手動一気通し（必要なら）

  - 手順:
    1. IDE 起動 → ワークスペース選択
    2. チャットで簡単な要求を入力
    3. （Chat Autopilot）自動で計画 → 実行に遷移することを確認（未実装の場合はギャップとして残る、`app.go:432-452`）
    4. 完了ステータスと成果物を確認
  - 観測ポイント:
    - `design/`・`state/`・`tasks/` の 3 層が整合する。
    - 依存順に実行される。

- [ ] Artifacts.Files の自動抽出/反映

  - 【目的】実行したタスクが「どのファイルを生成/変更したか」を IDE で追跡できるようにする。
  - 対象ファイル（候補）:
    - `internal/note/writer.go`
    - `internal/orchestrator/executor.go`
    - `internal/orchestrator/execution_orchestrator.go`
  - 実装タスク:
    1. AgentRunner の Task Note/JSON 出力から変更・生成ファイルを抽出する仕組みを定義。
    2. 抽出結果を `Artifacts.Files` に保存し、TaskStore と state を同期。
    3. IDE の TaskPropPanel で一覧表示。
  - 完了条件:
    - タスク完了後にファイル一覧が確認できる。
  - 【事実】最小要件（変更/生成ファイルの検出→イベント→IDE 表示）は先行実装済み（検出: `internal/worker/executor.go:420`、反映: `internal/orchestrator/executor.go:115`、永続化/同期: `internal/orchestrator/execution_orchestrator.go:424`、表示: `frontend/ide/src/lib/components/ui/TaskPropPanel.svelte:82`）。

- [ ] Meta Protocol のバージョニング導入

  - 【目的】Meta-agent と Core 間のプロトコル互換性を将来にわたって維持する。
  - 対象ファイル（候補）:
    - `internal/core/meta/*`
    - `docs/specifications/meta-protocol.md`
  - 実装タスク:
    1. YAML メッセージに `protocol_version`（または同等）を追加し、Core 側で解釈する。
    2. バージョン不一致時のフォールバック/警告/拒否方針を定義する。
  - 完了条件:
    - プロトコル更新時に旧クライアントが安全に扱える。
  - 【一次ソース】`docs/specifications/meta-protocol.md:318` の制約事項（「プロトコルバージョニングは未実装」）から移設。

- [ ] 追加 Worker 種別のサポート

  - 【目的】`codex-cli` 以外の CLI エージェントを Worker として選択可能にする。
  - 対象ファイル（候補）:
    - `internal/orchestrator/executor*.go`
    - `internal/worker/*`
    - `docs/cli-agents/*`
    - `docs/specifications/worker-interface.md`
  - 実装タスク:
    1. Worker kind と Docker イメージ/起動コマンドの対応表を追加する。
    2. `runner.worker.kind` に応じた選択とエラーハンドリングを実装する。
    3. 各 CLI のナレッジ（`docs/cli-agents/<kind>/...`）とテストを追加する。
  - 完了条件:
    - `gemini-cli` / `claude-code-cli` / `cursor-cli` を指定して実行できる。
  - 【一次ソース】`docs/specifications/worker-interface.md:28` の「将来的に cursor-cli」を移設。
  - 【一次ソース】`docs/cli-agents/README.md:5` の「未対応 CLI（gemini/cursor）」から移設。

- [ ] Task Builder（raw_prompt → TaskConfig YAML）

  - 【目的】`raw_prompt` から TaskConfig YAML を LLM で生成できるようにし、Executor の決定的生成から段階的に移行できるようにする。
  - 【一次ソース】`docs/task-builder-and-golden-test-design.md:43` の「Task Builder（CLI プロバイダ）」から移設。
  - 実装タスク:
    1. Task Builder の入出力（raw_prompt/context → TaskConfig）を仕様化する。
    2. 検証手段（ゴールデンテスト/ユニットテスト）を定義する。
    3. 既存 Executor 生成とのフォールバック（feature flag 等）を用意する。
  - 完了条件:
    - 同一入力で決定論的に動作し、`go test ./...` の品質ゲートを維持できる。

- [ ] 永続化レイヤー（DB / resume / 分散実行）

  - 【目的】TaskContext/実行履歴の永続化と、タスクの resume・複数ノード実行を可能にする。
  - 【一次ソース】`docs/design/architecture.md:398` の「永続化レイヤー」から移設。
  - スコープ例:
    - TaskContext を DB（例: PostgreSQL）に永続化
    - タスクの resume 機能
    - 複数ノードでの分散実行

- [ ] Web UI（タスク起動/モニタリング/ログ可視化）

  - 【目的】デスクトップ IDE 以外の導線として、タスク起動・モニタリング・履歴可視化・リアルタイムログ表示を提供する。
  - 【一次ソース】`docs/design/architecture.md:416` の「Web UI」から移設。

- [ ] IPC の WebSocket / gRPC 化

  - 【目的】ファイルポーリング IPC の性能/拡張性の制約を解消する。
  - 対象ファイル（候補）:
    - `internal/orchestrator/ipc/*`
    - `frontend/ide/src/*`
  - 実装タスク:
    1. Queue/イベント通知を WebSocket か gRPC に置き換える設計を確定する。
    2. 既存 file-based IPC と並行稼働できる移行パスを用意する。
  - 完了条件:
    - 大量ジョブ時のポーリング負荷が解消される。

- [ ] Frontend E2E の安定化

  - 【目的】IDE フロントの E2E が CI で継続的に回る状態にする。
  - 対象ファイル（候補）:
    - `frontend/ide/*`
    - `docs/guides/testing.md`
  - 実装タスク:
    1. タイムアウト/待機条件/テストデータを見直し安定化する。
    2. 失敗時ログの拡充とリトライ方針を整備する。
  - 完了条件:
    - `pnpm test:e2e` が安定して完走する。

- [ ] Task Note 保存の圧縮
  - 【目的】大きな Task Note/履歴の保存サイズを抑え、読み書き性能を維持する。
  - 対象ファイル（候補）:
    - `internal/note/*`
    - `internal/orchestrator/persistence/*`
  - 実装タスク:
    1. Task Note の圧縮形式（gzip 等）と保存/読み込み API を定義する。
    2. 既存データとの後方互換を確保する。
  - 完了条件:
    - Task Note の保存容量が有意に削減される。
