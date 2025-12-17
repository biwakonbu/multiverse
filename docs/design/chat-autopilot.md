# Chat Autopilot 設計（会話だけで「計画→実行→質問→継続」）

最終更新: 2025-12-17

## 1. 目的

ユーザーが「計画して」「実行して」などの操作/役割分担を意識せず、**自然な会話だけ**で開発が前進する状態を作る。

本設計は以下を満たす:

- チャット入力を起点に、Meta-agent が計画/実行/再計画を柔軟に判断し、必要なら自走でタスクを実行する。
- 不明点が出たら、エージェントがチャットで質問し、人間の回答を取り込んで継続する。
- IDE の実行ボタン（Start/Pause/Stop）はフォールバック（強制介入）であり、必須操作にしない。

## 2. 以前のギャップ（一次ソース） ※解消済み

### 2.1 チャットは「分解→保存」で止まる

- `ChatHandler.HandleMessage` は `Meta.PlanPatch` → 永続化（create/update/delete/move の適用）まで実行し、そこで完了する（`internal/chat/handler.go`）。
- 【解消】`SendChatMessage` 完了後に Chat Autopilot が `StartExecution()` を冪等に呼び、直後に `Scheduler.ScheduleReadyTasks()` を 1 回実行して自走を開始する（`app.go:532`、`app.go:546`）。

### 2.2 “人間に質問する” が実行ループに無い

- AgentRunner Core の `Runner` は Meta の `next_action` を `run_worker/mark_complete` しか扱わず、その他は unknown として `FAILED` で終了する（`internal/core/runner.go:317-320`）。
- つまり `ask_human` を実行ループに入れるには Core 側の実装拡張が必要。

### 2.3 可視化グルーピングが崩れる

- Frontend は `milestone -> phase -> task` を前提に WBS を構築する（`frontend/ide/src/stores/wbsStore.ts:161`）。
- 【解消】`ListTasks()` は `NodeDesign` 由来の `phaseName/milestone/wbsLevel/dependencies` を返す（`app.go:279`、`app.go:400`）。

## 3. 設計方針（結論）

1. **Chat Autopilot をバックエンドに実装**し、チャットの完了時点で実行ループ起動とスケジューリングを行う。
2. 自然言語の “介入” を許容するが、危険操作（停止/再開/対象変更）は **決定論で解釈**できる範囲を先に持つ（誤作動を避ける）。
3. 人間への質問はまず **plan_patch 由来の質問**（計画時の未確定事項）として実装し、将来的に Core の `ask_human` 対応へ拡張する。
4. 分類/可視化（facet）は `design/` を正として、IDE の表示は `ListTasks()` が必要な情報を必ず返す。

## 4. コンポーネント

### 4.1 Chat Autopilot（新規）

バックエンド側に導入する論理コンポーネント（実装は `app.go` または `internal/chat` に配置）。

責務:

- チャット入力の解釈（制御語の検出 + それ以外は meta へ）
- `Meta.PlanPatch` 実行と永続化（既存の `ChatHandler` を利用）
- 計画が更新されたら **実行開始/スケジューリング**を自動で行う
- 未解決の質問があれば停止して待つ

### 4.2 ExecutionOrchestrator（既存）

- 実行ループは `Start()` を呼ぶと 2 秒ポーリングでキューを処理する（`internal/orchestrator/execution_orchestrator.go:79`）。
- Ready タスクの enqueue は `Scheduler.ScheduleReadyTasks()` が担う（`internal/orchestrator/execution_orchestrator.go:245`、`internal/orchestrator/scheduler.go:112`）。

### 4.3 Backlog（既存・拡張）

- バックログは永続化され、`backlog:added` を IDE に通知できる（`internal/orchestrator/execution_orchestrator.go:646`、`frontend/ide/src/stores/backlogStore.ts:97`）。
- `BacklogTypeQuestion` が定義済み（`internal/orchestrator/backlog.go:21`）だが、現状の生成経路は主に failure 由来。

## 5. 主要フロー

### 5.1 チャット入力 → 計画生成 → 自動実行開始（Autopilot 基本）

1. IDE → `SendChatMessage(sessionId, message)`
2. ChatHandler が `Meta.PlanPatch` → `design/state/task_store` へ差分永続化（`internal/chat/handler.go`）
3. Autopilot が以下を実行（追加）
   - `GetExecutionState()` が `IDLE` なら `StartExecution()`（`app.go:633`、`app.go:601`）
   - 直後に `Scheduler.ScheduleReadyTasks()` を 1 回呼び、開始直後から進むことを保証
4. ExecutionOrchestrator がジョブを処理して `Executor` を起動し、`agent-runner` を実行する（`internal/orchestrator/execution_orchestrator.go:256`、`internal/orchestrator/executor.go:63`）

補足: `StartExecution()` は “already running” を返し得る（`internal/orchestrator/execution_orchestrator.go:82-85`）。Autopilot 側は **冪等**に扱う。

### 5.2 自然言語での介入（最小セット）

Autopilot は以下の制御語を LLM を経由せず解釈する（決定論・安全側）:

- 「止めて/停止」→ `StopExecution()`
- 「一旦止めて/一時停止」→ `PauseExecution()`
- 「続けて/再開」→ `ResumeExecution()`
- 「状況/ステータス」→ `GetExecutionState()` + タスクサマリ提示

それ以外の入力は meta に渡して `plan_patch`（再計画/整理）を行い、計画更新後は 5.1 の自動実行フローに接続する。

### 5.3 人間への質問（MVP: plan_patch 由来）

課題: Core の `ask_human` は未対応（`internal/core/runner.go:317-320`）。よって MVP は plan_patch に質問を含める。

案:

- `plan_patch` の payload に `questions[]` を追加し、ChatHandler がチャットに表示する。
- blocking な質問が残っている間は Autopilot が `PauseExecution()` し、回答を受けたら再度 `plan_patch` を走らせて計画を更新する。

質問の永続化は Backlog と統合する:

- 質問は `BacklogTypeQuestion` として保存し、未解決を IDE に見せる。
- 回答は `ResolveBacklogItem(id, resolution)` に保存し（`app.go:563`）、次回の plan_patch コンテキストに含める。

## 6. API / イベント（追加・整理）

### 6.1 既存 API（利用する）

- `StartExecution/PauseExecution/ResumeExecution/StopExecution/GetExecutionState`（`app.go:601`、`frontend/ide/wailsjs/go/main/App.d.ts:54`）
- `SendChatMessage`（`app.go:532`）
- `GetBacklogItems/ResolveBacklogItem`（`app.go:645`、`app.go:673`）

### 6.2 既存イベント（利用する）

- `chat:progress`（`internal/orchestrator/events.go:36`）
- `execution:stateChange`（`internal/orchestrator/events.go:33`）
- `task:created` / `task:stateChange`（`internal/orchestrator/events.go:32`）
- `backlog:added`（`internal/orchestrator/events.go:38`）

### 6.3 追加イベント（提案）

Autopilot の挙動が見えるように `chat:progress` に以下の step を追加する:

- `AutopilotStartingExecution`
- `AutopilotScheduling`
- `AutopilotPausedForQuestion`

（既存の `ChatProgressEvent` の枠で表現可能: `internal/orchestrator/events.go:58`）

## 7. データ（分類/グルーピングと Autopilot の相互作用）

分類設計は `docs/design/task-execution-and-visual-grouping.md` に従う。

Autopilot が前提とする最低要件:

- `ListTasks()` が `phaseName/milestone/wbsLevel/dependencies` を返す（WBS/Graph のグルーピングが壊れない）
- 失敗や質問の状態が IDE に表示される（Backlog/Chat で可視化）

## 8. 実装チェックリスト（PRD と同期）

- PRD の “チャットだけで計画→実行へ遷移” を満たす（`PRD.md` の 7.2 に対応）
- `SendChatMessage` の完了後に `StartExecution + ScheduleReadyTasks` を実行し、実行開始の導線を不要にする
- 失敗時の Backlog を “質問” としても扱えるようにし、会話に出す
- `ListTasks` の返却値を修正して WBS/Graph の分類が成立するようにする
