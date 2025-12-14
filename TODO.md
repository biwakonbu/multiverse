# Pragmatic MVP（完了）+ Quality Hardening（vNext）実装手順

最終更新: 2025-12-14

## 0. 前提

- MVP のゴールは `PRD.md` の「MVP 完了条件」参照。
- リアルタイム UX は最小限（Wails Events による `chat:progress`/`task:stateChange`/`task:log` は実装済み、ログはフロントで最大 1000 行に制限）。UI のカクつき防止と体感性能を優先する（`internal/orchestrator/events.go:34`、`internal/orchestrator/executor.go:121`、`frontend/ide/src/stores/logStore.ts:16`）。

## 0.1 まず読む（PRD と並行で参照する一次ドキュメント）

### 必読（この順で読む）

1. `PRD.md`（到達点と主要フローの定義）
2. `docs/COMPLETE_DOCUMENTATION.md`（全体像・用語・仕様索引）
3. `docs/design/chat-autopilot.md`（会話だけで「計画 → 実行 → 質問 → 継続」する設計）
4. `docs/design/task-execution-and-visual-grouping.md`（タスク実行と多軸グルーピング設計）

### 仕様（必要になった時点で読む）

- `docs/specifications/orchestrator-spec.md`（Orchestrator の仕様）
- `docs/specifications/meta-protocol.md`（plan_task/next_action/completion_assessment/decompose の仕様）
- `docs/specifications/worker-interface.md`（WorkerCall/実行インターフェース）
- `docs/specifications/testing-strategy.md`（テスト戦略）

### 実装参照（設計の根拠として読む）

- `docs/design/data-flow.md`（Core の plan→next_action→assessment ループ）
- `docs/task-builder-and-golden-test-design.md`（MVP の一気通しパイプライン）
- `docs/CURRENT_STATUS.md`（現状の制約・既知課題）

### 「この TODO を更新する時」のルール

- `PRD.md` と矛盾する変更は、先に `PRD.md` を更新してから TODO を直す。
- Chat Autopilot / grouping の設計変更は、先に `docs/design/chat-autopilot.md` と `docs/design/task-execution-and-visual-grouping.md` を更新してから TODO を直す。

## 1. 状態

- MVP は完了（`PRD.md` の 9 章）。
- vNext は “Quality Hardening（100 点化）” を最優先で実施する（`PRD.md` の 10〜12 章）。
- `ISSUE.md` は運用上のバックログ置き場として残すが、**設計/DoD/優先度の真実源は `PRD.md`** とする。

### 1.1 再レビュー（2025-12-14）

- 【事実】`go test ./...` はローカル実行で成功（品質ゲートは復旧）。
- 【評価】総合: **100/100**（PRD.md 13.3 の P0 タスク全完了）。
- 【達成】QH-001〜005 を全て実装完了（詳細は下記 3.2 参照）。

---

## 2. 反省の継承（この TODO を実行する前提）

この章は “実装手順に必ず反映するべき反省点” をチェックリスト化する。詳細は `PRD.md` 10 章を正とする。

- 【完了】plan_patch の入力コンテキストは "落とさず渡す" まで改善し、既存タスクは決定論ソートで 200 件に丸める。WBS `node_index` は BFS で 200 ノード上限にトリミング（`internal/meta/utils.go:211-216`、`internal/meta/utils.go:254-297`）。
- 【完了】delete(cascade=false) は splice 実装 + 不変条件テスト追加済み。move の回帰防止テストも 6 件追加（`internal/chat/plan_patch_wbs_test.go:201-385`）。
- 【完了】history は append 先行、append 失敗時は failure action を記録。失敗注入テストも追加（`internal/chat/plan_patch_history_test.go`）。
- 【完了】meta mock は構造ベース（JSON パース + switch 文）に移行（`internal/meta/mock_adapter.go:44-83`）。

---

## 3. vNext: Quality Hardening（100 点化）実行順序

### 3.1 事前チェック（必須）

- `go test ./...` を実行し、green を維持する（現状は成功、詳細は `PRD.md` 13.1）。
- WBS の不変条件（`PRD.md` 11.2）を読み、delete/move の期待動作をチームで合意する。
- テストの穴を機械的に確認する（例: `rg -n 'moveNodeInWBS\\(' internal/chat/*_test.go`）。

### 3.2 P0（最優先・これが終わらない限り他へ進まない）

#### 3.2.1 QH-001: plan_patch プロンプトに構造化コンテキストを完全継承 ✅ 完了

- 目的/受け入れ条件/検証は `PRD.md` 12.1 を正とする。
- 【完了】`existing_tasks` の facet/依存・`existing_wbs.node_index`・`conversation_history` はプロンプトに含まれている（`internal/meta/utils.go:155`）。
- 【完了】既存タスクは status 優先 + ID 昇順で決定論的に 200 件に丸める（`internal/meta/utils.go:154`）。
- 【完了】WBS `node_index` は BFS で 200 ノード上限にトリミング（`internal/meta/utils.go:211-216`、`trimWBSNodesBFS` 関数追加）。テスト: `internal/meta/utils_test.go`。

#### 3.2.2 QH-003: delete(cascade=false) のセマンティクス確定 + 実装 + テスト ✅ 完了

- 決め打ち（案 A/案 B）は `PRD.md` 12.2 を参照し、PRD に採用方針を明記してから実装する。
- WBS 不変条件テストを必須化する（孤児/重複 children を検知）。
- 【完了】案 A（splice）は実装済みで、delete 周りの不変条件テストも追加済み（`internal/chat/plan_patch.go:511`、`internal/chat/plan_patch_wbs_test.go:15`）。
- 【完了】move の回帰防止テストを 6 件追加（`internal/chat/plan_patch_wbs_test.go:201-385`）。基本 move、before/after/index position、同一親内 reorder、重複防止テストを含む。

#### 3.2.3 QH-004: history→design/state の順序保証（擬似トランザクション） ✅ 完了

- `PRD.md` 12.3 の受け入れ条件を満たすまで完了扱いにしない。
- 【完了】順序は改善し、append 失敗時は failure action を記録する（`internal/chat/plan_patch.go:380`、`internal/chat/plan_patch.go:398`）。
- 【完了】失敗注入テストを追加（`internal/chat/plan_patch_history_test.go`）。`TestHistoryAppendFailure_RecordsFailureAction` 等 3 件のテストで failure action 記録を検証。

#### 3.2.4 QH-005: meta テストの無外部依存化（品質ゲート復旧） ✅ 完了

- `go test ./...` がネットワーク無しで通ることを "Done" の必須条件にする（`PRD.md` 11.3）。
- 【完了】`go test ./...` は成功。mock は構造ベース（JSON パース + switch 文）に移行完了（`internal/meta/mock_adapter.go:44-83`）。system/user プロンプトを構造的に抽出し判定。

### 3.3 P1（運用品質）

#### 3.3.1 QH-006: 実行ログ UI の運用可能化

- `ISSUE.md:15` の残タスクを “最小導線” に落とす（タスク別フィルタ/クリア/常時表示）。

#### 3.3.2 QH-007: Codex CLI セッション検証と注入仕様

- `ISSUE.md:21` を “実装 + ドキュメント” まで閉じる。

### 3.4 P2（将来拡張：仕様/テストを先に固めてから実装）

以下は `PRD.md` 12.6 を正とし、着手順は運用上の都合で入れ替えてよい。ただし P0 を完了してから着手する。

- QH-008: Artifacts.Files の自動抽出/反映（`ISSUE.md:38`）
- QH-009: Meta Protocol のバージョニング導入（`ISSUE.md:52`）
- QH-010: 追加 Worker 種別のサポート（`ISSUE.md:64`）
- QH-011: IPC の WebSocket / gRPC 化（`ISSUE.md:78`）
- QH-012: Frontend E2E の安定化（`ISSUE.md:90`）
- QH-013: Task Note 保存の圧縮（`ISSUE.md:102`）

---

## 4. 完了判定（この TODO の Definition of Done）

以下がすべて満たされない限り “100 点” としてクローズしない。DoD の正は `PRD.md` 11 章。

- 【事実】`go test ./...` がネットワーク不要で安定して通る。
- 【事実】plan_patch の create/update/move/delete（cascade true/false）で WBS 不変条件を満たすことをテストで担保できる。
- 【事実】history の順序が設計通りで、失敗時の扱いが実装・テスト・ドキュメントで一致している。
