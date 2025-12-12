# TODO: Multiverse IDE 実装計画

### Phase 1: Data Model Enhancements (SuggestedImpl & Artifacts) - **DONE**

- [x] **Backend**: Update `Task` struct in `internal/orchestrator/task_store.go`
  - Add `SuggestedImpl` (Language, FilePaths, Constraints)
  - Add `Artifacts` (Files, Logs)
- [x] **Meta-Layer**: Update `DecomposedTask` in `internal/meta/protocol.go`
- [x] **Chat**: Update `persistTasks` in `internal/chat/handler.go` to map these fields
- [x] **Prompt**: Update `decomposeSystemPrompt` in `internal/meta/client.go`
- [x] **Verification**: Add unit tests for persistence and mapping

## Phase 2: エージェントワークフローの深化

「チャットで指示 → 正確にコード生成」の精度を上げる。

- [/] **Planner (Chat Handler) の改善**
  - [x] Meta-agent プロンプトの調整: `SuggestedImpl` を出力するように指示
  - [x] ファイルパスの検証ロジック（存在しないファイルへの変更指示などを事前に警告）
- [ ] **Executor の `SuggestedImpl` 対応**
  - [x] `agent-runner` への入力 YAML に `suggested_impl` 情報を渡す処理
  - [ ] Worker 側での当該情報の利用実装

### Phase 3: IDE Frontend Integration - **DONE**

- [x] **Types**: Update TypeScript definitions (`frontend/ide/src/types` & `schemas`)
- [x] **UI**: Add `SuggestedImpl` cues to `TaskNode` / `WBSNode`
- [x] **UI**: Create `TaskPropPanel` to display details
- [x] **Storybook**: Update stories to verify `SuggestedImpl` rendering

## Phase 4: リカバリと安全性 (Future)

- [x] **スナップショット機能**
  - [x] `state/` ディレクトリ全体のバックアップとリストア
- [ ] **承認フロー**
  - [ ] 重要なファイル変更に対する「ユーザー承認」ステップの導入
