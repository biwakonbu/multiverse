# TODO: Pragmatic MVP 実装手順

最終更新: 2025-12-12

## 0. 前提

- MVP のゴールは `PRD.md` の「MVP 完了条件」参照。
- リアルタイム UX（ログストリーミング等）は後回し。UI のカクつき防止と体感性能を優先する。

## 1. 計画生成の橋渡し（Chat → design/state）

### 1.1 Decompose → WBS/NodeDesign 永続化（DONE）

【目的】Scheduler が参照する `design/` を、チャット入力から生成できる状態にする。

- 対象ファイル:
  - `internal/chat/handler.go`
  - `internal/orchestrator/persistence/repo.go`
  - `internal/orchestrator/persistence/models.go`
- 実装タスク:
  1. Handler に `persistence.WorkspaceRepository` を注入し、チャット処理から design/state へ書き込めるようにする。
  2. `decomposeResp` から WBS ルートを作成/更新し `design/wbs.json` に保存する。
  3. 各 `DecomposedTask` を 1:1 で `NodeDesign` に写像し `design/nodes/<node-id>.json` に保存する。
  4. `DecomposedTask.dependencies` を `node_id` に解決し、`NodeDesign.Dependencies` に格納する。
- 完了条件:
  - `~/.multiverse/workspaces/<id>/design/wbs.json` と `design/nodes/*.json` が生成される。

### 1.2 Decompose → NodesRuntime/TasksState 永続化（DONE）

【目的】ExecutionOrchestrator が読む `state/` を plan と同期させ、実行可能にする。

- 対象ファイル:
  - `internal/chat/handler.go`
  - `internal/orchestrator/persistence/repo.go`
- 実装タスク:
  1. NodeDesign 作成時に `state/nodes-runtime.json` に `NodeRuntime{status:"planned"}` を追加（既存なら更新）。
  2. 各 NodeDesign に対応する `TaskState{kind:"implementation", status:"pending"}` を `state/tasks.json` に追加。
  3. `TaskState.NodeID` と `NodeDesign.NodeID` を一致させる。
- 完了条件:
  - `state/nodes-runtime.json` と `state/tasks.json` に新規エントリが作成される。

### 1.3 TaskStore との同期（DONE）

【目的】IDE 表示用の TaskStore と design/state の整合を取る。

- 対象ファイル:
  - `internal/orchestrator/task_store.go`
  - `internal/chat/handler.go`
- 実装タスク:
  1. TaskStore の Task.ID を NodeDesign/TaskState と同一にする（既存 UUID マッピングを整理）。
  2. `dependencies / wbsLevel / phaseName / suggestedImpl / acceptanceCriteria` を同期して保存する。
- 完了条件:
  - IDE の Task 一覧/グラフが従来どおり表示できる。

## 2. 実行結果の反映（state/design の整合）

### 2.1 AttemptCount / Retry の整理（DONE）

【目的】RetryPolicy が確実に動くよう、試行回数を永続化する。

- 対象ファイル:
  - `internal/orchestrator/execution_orchestrator.go`
  - `internal/orchestrator/persistence/models.go`
- 実装タスク:
  1. `TaskState` に `AttemptCount` フィールドを追加するか、`Inputs["attempt_count"]` を正式仕様として扱う。
  2. `processJob` 開始時に attempt_count をインクリメントして `state/tasks.json` に保存する。
- 完了条件:
  - 連続失敗時に backoff が段階的に伸びる。

### 2.2 Task 成功時の NodesRuntime 更新（DONE）

【目的】ノード依存が解決され、後続タスクが READY になるようにする。

- 対象ファイル:
  - `internal/orchestrator/execution_orchestrator.go`
- 実装タスク:
  1. `attempt.Status == SUCCEEDED` の場合、該当 `node_id` の `NodeRuntime.Status` を `implemented` に更新（無ければ作成）。
  2. 更新後に `state/nodes-runtime.json` を保存する。
- 完了条件:
  - 依存ノード完了後、次ノードのタスクが自動で READY になり実行される。

### 2.3 TaskStore / IDE イベント反映（DONE）

【目的】IDE 表示と実行状態を同期する。

- 対象ファイル:
  - `internal/orchestrator/executor.go`
  - `internal/orchestrator/execution_orchestrator.go`
- 実装タスク:
  1. TaskStore の Task.Status と Artifacts を更新する。
  2. `EventTaskStateChange` など既存イベントで IDE に通知する。
- 完了条件:
  - IDE に SUCCEEDED/FAILED が反映される。

## 3. Executor YAML の最小改善（ハードコード排除）

### 3.1 max_loops / worker kind の受け渡し（DONE）

【目的】MVP でも最低限 Runner 設定を差し替えられるようにする。

- 対象ファイル:
  - `internal/orchestrator/executor.go`
- 実装タスク:
  1. `TaskState.Inputs` などから `runner.meta.max_loops`/`runner.worker.kind` を読んで YAML に反映する。
  2. 値が無い場合は現状デフォルトにフォールバックする。
- 完了条件:
  - 設定変更が破壊的変更なしに反映される。

## 4. IDE UX/性能の改善（優先）

### 4.1 グラフ再レイアウトのバッチ化（DONE）

【目的】大量タスク生成・状態変化時のカクつきを抑える。

- 対象ファイル:
  - `frontend/ide/src/stores/*`
  - `frontend/ide/src/lib/flow/UnifiedFlowCanvas.svelte`
- 実装タスク:
  1. stateChange イベントの連続受信時に layout 計算をまとめる。
- 完了条件:
  - タスク大量生成時に明確な UI の遅延が出ない。

※ `UnifiedFlowCanvas.svelte` の `$effect` 内でタイマーによるバッチ化を実装し、連続更新時のレイアウト計算回数を抑制済み。

### 4.2 タスクノード表示の可読性改善（DONE）

【目的】ノードが Task ID だけを表示して「何のタスクか分からない」問題を解消する。

- 対象ファイル（候補）:
  - `frontend/ide/src/lib/flow/UnifiedFlowCanvas.svelte`
  - `frontend/ide/src/stores/*`（タスク/ノード表示用 Store）
- 実装タスク:
  1. ノードの主表示を `Task.Title` / `NodeDesign.Name` 優先に変更し、ID は補助情報（ツールチップ/コピー用）へ退避。
  2. `implementation:` などの prefix 表示を簡素化し、Phase/Status との視認性を両立する。
  3. 依存関係の解釈ができるよう、ノード内に簡易サマリ（例: 1 行 description 先頭）を任意表示。
- 完了条件:
  - ノード一覧/グラフだけで「何をするタスクか」判断できる。

※ タイトルの視認性（フォント/コントラスト）改善と、タイトル供給経路（`App.ListTasks` で `Inputs.title`/`NodeDesign.Name` を優先）まで完了。  
　prefix 整理や description 先頭のサマリ表示は将来的な改善候補。

### 4.3 グラフの拡大縮小/パン操作の復旧（DONE）

【目的】画面サイズを大きくしないと見れない/ズームが効かない問題を解消する。

- 対象ファイル（候補）:
  - `frontend/ide/src/lib/flow/UnifiedFlowCanvas.svelte`
  - `frontend/ide/src/lib/flow/*`（Flow/Canvas 操作）
- 実装タスク:
  1. ホイール/トラックパッドでのズーム、ドラッグでのパンのイベントハンドリングを調査し修正。
  2. ズームレベルの上限/下限と初期 fit-to-view を調整し、タスク生成直後でも全体が見えるようにする。
  3. 画面右下のミニマップ/ビューリセットの導線を追加/修正（ある場合は動作保証）。
- 完了条件:
  - 標準的な画面サイズでズーム/パンが一貫して動作し、全体把握が可能。

### 4.4 チャット送信の即時反映（Optimistic UI/DONE）

【目的】自分の入力メッセージがチャット UI に反映されるまでの遅延を無くす。

- 対象ファイル（候補）:
  - `frontend/ide/src/stores/*`（Chat/Session Store）
  - `frontend/ide/src/components/*`（Chat UI）
- 実装タスク:
  1. 送信ボタン押下時点でローカルに user message を append し、UI に即時表示。
  2. バックエンド保存/Meta 処理が失敗した場合のロールバック/エラー表示を追加。
  3. 送信中の入力欄のロック/解除とフォーカス維持を最適化。
- 完了条件:
  - 入力 → 送信後、100ms 程度以内にユーザーメッセージが表示される。

### 4.5 チャット進捗イベントのシームレス表示（DONE）

【目的】decompose/永続化/スケジュール等の待ち時間が長い時に、進捗が途切れず分かるようにする。

- 対象ファイル（候補）:
  - `internal/chat/handler.go`（`EventChatProgress` の発火）
  - `frontend/ide/src/stores/*` / `frontend/ide/src/components/*`（Progress 表示）
- 実装タスク:
  1. `EventChatProgress` をフロントで購読し、チャット内の system/progress メッセージとして逐次表示。
  2. 「Processing/Persisting/Completed/Failed」など段階表示を UI で整理（スピナー/タイムライン）。
  3. 初回ロード時の ChatHistory 読み込み中もプレースホルダを即時表示し、体感待ちを減らす。
- 完了条件:
  - Meta 処理に数十秒かかっても、ユーザーが「止まっている」と感じない。

※ `internal/chat/handler.go` が `chat:progress` を段階発火し、フロントは `chat.ts` で購読。Generalタブでは最新progressを system メッセージとして表示、Logタブは詳細ログを保持。

### 4.6 WBS オーバーレイ表示の再設計（DONE）

【目的】WBS をオーバーレイした結果見にくい問題を解消する。

- 対象ファイル（候補）:
  - `frontend/ide/src/components/*`（WBS/Graph UI）
  - `frontend/ide/src/lib/flow/UnifiedFlowCanvas.svelte`
- 実装タスク:
  1. オーバーレイを廃止し、サイドパネル/ツリー表示など「独立領域」に切り替える。
  2. Graph と WBS の同期（選択/ハイライト/スクロール連動）を最小限で実装。
  3. WBS の表示密度（階層折りたたみ、文字サイズ、余白）を調整。
- 完了条件:
  - WBS と Graph の両方が同時に読み取れる。

※ WBSListView を Graph から分離した左サイドパネルとして表示するよう変更し、オーバーレイによる視認性問題を解消。同期は既存の選択/展開状態ストアに委譲。

### 4.7 潜在コンフリクト表示の信頼性改善（DONE）

【目的】Meta decompose が返す `potential_conflicts` の虚偽/過剰検出によりユーザーが混乱する問題を減らす。

- 対象ファイル（候補）:
  - `internal/chat/handler.go`
  - `internal/meta/client.go`
  - `docs/specifications/meta-protocol.md`
- 実装タスク:
  1. ChatHandler 側で `potential_conflicts` の `file` が実在するか検証し、存在しない場合は非表示 or 「新規想定」ラベルで弱く表示。
  2. 将来的に、Workspace の実在ファイル一覧を Meta decompose の Context に渡して精度を上げる。
  3. 仕様として「`potential_conflicts` はヒューリスティックで false positive を含む」旨を明記。
- 完了条件:
  - 実在しないファイルのコンフリクトが強い警告として出ない。

## 5. ゴールデンパス検証

### 5.1 自動ゴールデンパス（DONE）

【目的】IDE の入力をモックし、バックエンドの「Chat→ 計画生成 → 依存解決 → 実行 → 同期」までを自動で検証する。

- 実装:
  - `test/e2e/golden_pass_test.go`
- 検証内容:
  - ChatHandler が `design/` と `state/` を生成すること。
  - ExecutionOrchestrator が依存順にタスクを実行し、`state/` と TaskStore を更新すること。

### 5.2 手動一気通し（必要なら）

- 手順:
  1. IDE 起動 → ワークスペース選択
  2. チャットで簡単な要求を入力
  3. 生成されたタスク（またはノード）を Run
  4. 完了ステータスと成果物を確認
- 観測ポイント:
  - `design/`・`state/`・`tasks/` の 3 層が整合する。
  - 依存順に実行される。

### 5.3 テスト追加（既存パターンに沿う/DONE）

- 対象ファイル:
  - `internal/chat/handler_test.go`
  - `internal/orchestrator/*_test.go`
- 追加タスク:
  1. `decompose → design/state 保存` の単体テスト。
  2. `processJob` が `NodesRuntime` を更新するテスト。

※ `internal/chat/handler_test.go` に design/state 永続化の単体テスト、`internal/orchestrator/execution_orchestrator_test.go` に成功時Runtime更新テストを追加済み。

## 6. 将来拡張

### 6.1 Artifacts.Files の自動抽出/反映

【目的】実行したタスクが「どのファイルを生成/変更したか」を IDE で追跡できるようにする。

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

### 6.2 Meta Protocol のバージョニング導入

【目的】Meta-agent と Core 間のプロトコル互換性を将来にわたって維持する。

- 対象ファイル（候補）:
  - `internal/core/meta/*`
  - `docs/specifications/meta-protocol.md`
- 実装タスク:
  1. YAML メッセージに `protocol_version`（または同等）を追加し、Core 側で解釈する。
  2. バージョン不一致時のフォールバック/警告/拒否方針を定義する。
- 完了条件:
  - プロトコル更新時に旧クライアントが安全に扱える。

### 6.3 追加 Worker 種別のサポート

【目的】`codex-cli` 以外の CLI エージェントを Worker として選択可能にする。

- 対象ファイル（候補）:
  - `internal/orchestrator/executor*.go`
  - `internal/worker/*`
  - `docs/cli-agents/*`
- 実装タスク:
  1. Worker kind と Docker イメージ/起動コマンドの対応表を追加する。
  2. `runner.worker.kind` に応じた選択とエラーハンドリングを実装する。
  3. 各 CLI のナレッジ（`docs/cli-agents/<kind>/...`）とテストを追加する。
- 完了条件:
  - `gemini-cli` / `claude-code-cli` / `cursor-cli` を指定して実行できる。

### 6.4 IPC の WebSocket / gRPC 化

【目的】ファイルポーリング IPC の性能/拡張性の制約を解消する。

- 対象ファイル（候補）:
  - `internal/orchestrator/ipc/*`
  - `frontend/ide/src/*`
- 実装タスク:
  1. Queue/イベント通知を WebSocket か gRPC に置き換える設計を確定する。
  2. 既存 file-based IPC と並行稼働できる移行パスを用意する。
- 完了条件:
  - 大量ジョブ時のポーリング負荷が解消される。

### 6.5 Frontend E2E の安定化

【目的】IDE フロントの E2E が CI で継続的に回る状態にする。

- 対象ファイル（候補）:
  - `frontend/ide/*`
  - `docs/guides/testing.md`
- 実装タスク:
  1. タイムアウト/待機条件/テストデータを見直し安定化する。
  2. 失敗時ログの拡充とリトライ方針を整備する。
- 完了条件:
  - `pnpm test:e2e` が安定して完走する。

### 6.6 Task Note 保存の圧縮

【目的】大きな Task Note/履歴の保存サイズを抑え、読み書き性能を維持する。

- 対象ファイル（候補）:
  - `internal/note/*`
  - `internal/orchestrator/persistence/*`
- 実装タスク:
  1. Task Note の圧縮形式（gzip 等）と保存/読み込み API を定義する。
  2. 既存データとの後方互換を確保する。
- 完了条件:
  - Task Note の保存容量が有意に削減される。

## 7. 仕上げ（自己レビュー起点の軽量改善）

### 7.1 チャットローディング状態の単一責務化（DONE）

【目的】`isChatLoading` の二重制御による状態揺れ/レースを避け、保守性を上げる。

- 対象ファイル:
  - `frontend/ide/src/stores/chat.ts`
- 実装タスク:
  1. `chat:progress` 側での `isChatLoading` 変更をやめ、`sendMessage` に責務を集約。
  2. progress 表示は `chatLog` の更新のみで行う。
- 完了条件:
  - 送信→完了まで loading 状態が一貫し、フリッカーや競合が起きない。

### 7.2 テストの冗長読み込み整理（DONE）

【目的】テストの可読性と速度を維持しつつ、無駄な state 読み込みを削る。

- 対象ファイル:
  - `internal/chat/handler_test.go`
- 実装タスク:
  1. `TestHandler_HandleMessage_PersistsDesignAndState` で `LoadNodesRuntime`/`LoadTasks` をループ外に出して検証を簡潔化。
- 完了条件:
  - テストの意図が明確で、無駄な I/O がない。

### 7.3 小さな冗長コメント削除（DONE）

【目的】軽微な冗長を減らし、読みやすさを上げる。

- 対象ファイル:
  - `app.go`
- 実装タスク:
  1. `ListTasks` の重複コメントを削除。
- 完了条件:
  - 同一コメントの重複がない。

### 7.4 LLM デフォルト更新と Codex CLI の一時廃止整理（DONE）

【目的】IDE 経由の計画/分解が Codex CLI 非依存で安定して動く状態にする。

- 対象ファイル:
  - `internal/ide/llm_config.go`
  - `app.go`
  - `internal/meta/client.go`
  - `docs/design/architecture.md`
  - `docs/cli-agents/codex/CLAUDE.md`
  - `docs/CURRENT_STATUS.md`
- 実装タスク:
  1. Meta-agent のデフォルト Kind を `openai-chat` に切り替え、`codex-cli` 指定時はフォールバック。
  2. Meta のデフォルトモデルを `gpt-5.2` に更新し、Worker（Codex CLI）のデフォルトは `gpt-5.1-codex` に据え置き（必要に応じて上書き可能）。
  3. 仕様/ガイドへ「Meta での codex-cli は当面無効化」旨を追記。
- 完了条件:
  - IDE のチャットからの計画/分解が Codex CLI 無しで実行できる。

### 7.5 デバッグ/冗長コメントの整理（DONE）

【目的】読み手のノイズとなる試行錯誤コメントを除去し、コードの意図を明確にする。

- 対象ファイル:
  - `app.go`
  - `internal/meta/client.go`
  - `internal/ide/llm_config.go`
- 実装タスク:
  1. 途中経過のメモ・仮説コメントを削除し、必要な説明のみ残す。
  2. gofmt 相当のインデント整形で可読性を維持する。
- 完了条件:
  - 主要エントリポイント/設定周りが簡潔に読める。
