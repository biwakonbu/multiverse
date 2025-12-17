# タスク実行と多軸グルーピング設計（Planning → Execution）

最終更新: 2025-12-17

## 1. 背景 / 問題

### 1.1 「タスクは作られるが実行されない」

- `ExecutionOrchestrator` は `StartExecution()` を呼ぶまで `IDLE` のまま（`internal/orchestrator/execution_orchestrator.go:79`）。
- 【解消】`SendChatMessage` 完了後に Chat Autopilot が `StartExecution()` を冪等に呼び、直後に `Scheduler.ScheduleReadyTasks()` を 1 回実行して自走を開始する（`app.go:532`、`app.go:546`）。
- 【補足】UI からの明示的な開始/停止はフォールバック（強制介入）であり、通常フローの必須操作にはしない（3 章）。

### 1.2 「タスクがフラットで、分類/可視化が雑になる」

- Frontend の WBS ツリーは `milestone -> phase -> task` でツリー化する設計（`frontend/ide/src/stores/wbsStore.ts:161-240`）。
- backend の `ListTasks()` は `design/wbs.json` + `design/nodes/*.json` + `state/tasks.json` を join して `dependencies/phaseName/milestone/wbsLevel` を返す（`app.go:279`）。
  - これにより UI では `phaseName/milestone` が空扱いにならず、WBS が 1 グループに潰れにくい。
- `design/state` 側も、TaskState.Kind が全タスクで `"implementation"` 固定になっており（`internal/chat/handler.go:579-596`）、作業種別（仕様/ドキュメント/設計/実装/検証など）という分類軸を表現できない。

## 2. ゴール / 非ゴール

### 2.1 ゴール

1. **Planning → Execution の遷移を明示**し、「いつまでタスク生成が続くのか分からない」を解消する。
2. **複数軸（Facet）での可視化**を可能にする。
   - 例: `phaseName`, `milestone`, `workType`, `domain/component`, `status`, `text search`
3. 既存ワークスペースの `design/state/tasks` の互換性を壊さない。

### 2.2 非ゴール（当面）

- 高度なクエリ言語やサーバーサイド検索インデックス。
- リモート実行/分散ワーカープール最適化。

## 3. 設計方針（結論）

- **分類メタデータ（Facet）は `design/` を正**とし、`state/` と `TaskStore(tasks/*.jsonl)` は表示/実行のために同期する。
- **UI は “Group By” と “Filters” を同じ Facet 概念で扱う**（WBS も Graph も同一フィルタで絞り込み可能にする）。
- 実行は **「チャット駆動（Autopilot）」を基本**とし、UI の実行ボタンはフォールバック（停止・一時停止等の非常用）として扱う。

## 4. データモデル（Facet）

### 4.1 Facet の定義（最小）

| フィールド | 例 | 用途 |
| --- | --- | --- |
| `phaseName` | `概念設計/実装設計/実装/検証` | フェーズ別グルーピング |
| `milestone` | `M1-Feature-Design` | 機能/エピック単位のまとまり |
| `wbsLevel` | `1/2/3` | 粗い工程区分 |
| `workType` | `spec/docs/design/implementation/test` | 「仕様/ドキュメント/設計/実装/検証」軸 |
| `domain` | `orchestrator/frontend/meta/...` | 機能カテゴリ（コンポーネント） |
| `tags[]` | `["ux","refactor"]` | 任意ラベル |

### 4.2 永続化先

#### A. `design/nodes/*.json`（推奨: 正）

- `persistence.NodeDesign` に以下を追加する想定:
  - `phase_name`, `milestone`, `wbs_level`, `work_type`, `domain`, `tags`

#### B. `state/tasks.json`（実行/表示用の複製）

- `persistence.TaskState.Inputs`（柔軟）に `facet.*` を複製する（例: `inputs["facet.phase_name"] = "実装"`）。
- これにより Scheduler/Executor が **design を読まなくても最低限の分類**を参照できる。

#### C. `tasks/*.jsonl`（IDE 表示の後方互換）

- `orchestrator.Task` にも同等の Facet を持たせ、IDE の一覧/Graph/WBS 表示で利用する。

## 5. Facet の生成規則（優先順位）

1. **明示指定（将来）**: Meta plan_patch が `work_type/domain/tags` を返す場合、それを正とする。
2. **推定（当面）**: 既存フィールドから決定論で推定する。
   - `phaseName == "概念設計"` → `workType=spec`（ただしタイトル/説明に「ドキュメント/README」が強く含まれる場合は `docs`）
   - `phaseName == "実装設計"` → `workType=design`
   - `phaseName == "実装"` → `workType=implementation`
   - `phaseName == "検証"` または「テスト」が強く含まれる → `workType=test`
   - `domain` は `suggestedImpl.filePaths` のパス接頭辞（例: `internal/orchestrator/...`）から推定する（推定不能なら空）。

## 6. Planning → Execution（実行制御）

### 6.1 UI 導線（フォールバック）

- 実行制御（Start/Pause/Resume/Stop）は、**ユーザーが強制介入するためのフォールバック**として UI に提供する。
  - 配置候補: Toolbar 右端、または TaskBar に “Run/Pause/Stop” を追加。

### 6.2 Chat Autopilot（基本）

- ユーザーは「計画して」「実行して」などの役割分担を要求されない。
- Chat の「タスク永続化」完了後に以下を実行する:
  1. `ExecutionOrchestrator` が `IDLE` なら `StartExecution()`（`internal/orchestrator/execution_orchestrator.go:79`、`app.go:601`）
  2. 直後に `Scheduler.ScheduleReadyTasks()` を 1 回呼び、開始直後から進むことを保証（2 秒ポーリング待ちを削減）

### 6.3 自然言語での介入（必須）

- ユーザーはチャットで自然に介入できる（例: 「止めて」「一旦止めて」「続けて」「状況教えて」）。
- 実装は 2 系統を許容する:
  - **決定論（安全側）**: 明確な制御語（stop/pause/resume/status）だけは LLM を経由せず即時に `StopExecution/PauseExecution/ResumeExecution/GetExecutionState` にマップする。
  - **Meta 主導（柔軟）**: それ以外は Meta-agent に渡し、計画更新（plan_patch）や優先度付けを含めて判断させる。

### 6.4 人間への質問（Backlog → Chat）

- Meta-agent が人間に確認すべき事項は **チャットに質問として出る**ことを基本 UX とする。
- 既存のバックログ通知は `backlog:added` としてイベント化済み（`internal/orchestrator/events.go:38`、`internal/orchestrator/execution_orchestrator.go:646`）。
- 設計方針:
  - `BacklogTypeQuestion` を活用し、質問は Backlog に永続化しつつ、チャットにも「質問メッセージ」として表示する。
  - 未解決の質問がある間は、実行を `PAUSED` にして待つ（ユーザー回答後に自動再開）。
  - 回答は `ResolveBacklogItem` で保存し（`app.go:563`）、回答内容は次回の Meta plan_patch/実行コンテキストに含める。

#### 6.4.1 質問の生成源（2案）

- **案A: 計画時（plan_patch）に質問を返す**
  - `plan_patch` レスポンスに `questions[]`（blocking/optional）を追加し、ChatHandler が質問をチャットに表示して待つ。
  - メリット: 実装が単純。タスク実行前に不明点を回収できる。
- **案B: 実行時（agent-runner の next_action）で `ask_human` を扱う**
  - 現状の AgentRunner Core は `run_worker/mark_complete` 以外を Unknown として即 `FAILED` 扱いにしている（`internal/core/runner.go:317-320`）。
  - `ask_human` を正式に扱うには、`NextActionResponse` に質問ペイロードを追加し、Runner が「質問→中断→再試行（回答を Inputs に入れて再実行）」を実装する必要がある。

## 7. Backend API / UI 反映

### 7.1 `ListTasks()` の責務

- IDE が必要とする `phaseName/milestone/wbsLevel/dependencies` と Facet を必ず返す。
- 実装方式は 2 案:
  - **案1（最短）**: TaskStore（`tasks/*.jsonl`）から読み出す（既に Phase/Milestone を持つ）
  - **案2（正攻法）**: `design/nodes` + `state/tasks` を join して DTO を組み立てる（Facet の正を `design` に置く）

### 7.2 フロント（可視化）

- `facetStore`（derived）で以下を提供:
  - `availableFacets`: milestone/phase/workType/domain の集合と件数
  - `activeFilters`: 選択中の条件
  - `groupBy`: 現在の grouping 軸（例: milestone→phase, workType→domain など）
- `UnifiedFlowCanvas` は `taskList`（フィルタ済み）を受け取れるので、Graph 側は `taskList` を差し替えることで絞り込みできる（`frontend/ide/src/lib/flow/UnifiedFlowCanvas.svelte:42-75`）。
- WBS 側は `wbsStore` の入力（tasks）をフィルタ済みにした派生ストアを使う。

## 8. 移行（既存ワークスペース）

- 既存の `design/nodes` に新フィールドが無い場合は空として扱う（Go の JSON Unmarshal では unknown/missing フィールドは安全に扱える）。
- 互換のため、最初の段階では TaskStore に存在する `phaseName/milestone/wbsLevel/dependencies` を読み、`design/state` へ補完する「オンデマンド補正」を提供する（明示的マイグレーションは不要）。

## 9. 実装ステップ（最短ルート）

1. **ListTasks の修正**: `phaseName/milestone/dependencies/wbsLevel` を返す（案1で即効性優先）。
2. **Chat Autopilot**: `SendChatMessage` 完了後に `StartExecution + ScheduleReadyTasks` を呼び、チャットだけで「計画→実行」に遷移させる。
3. **質問 UX**: `backlog:added` をチャットにブリッジし、質問（BacklogTypeQuestion）を会話として扱う。
4. **Kind/WorkType**: `internal/chat/handler.go` の TaskState.Kind をフェーズに応じて設定し、Facet を `state/tasks.json` に複製。
5. **Facet UI**: group-by + filter を追加し、Graph/WBS の両方に適用。
