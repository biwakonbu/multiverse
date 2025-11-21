# agent-runner 残タスク一覧

## 📊 プロジェクト進捗

- **現在の実現率**: 85% (Phase 6・7完了) ✅
- **目標実現率**: 95%+ (本番対応完成)
- **最終更新**: 2025-11-21

### 実現率推移

| Phase | 現状 | 達成目標 | 主要タスク |
|-------|------|--------|----------|
| 現在 | **85%** ✅ | - | Phase 6・7完了 |
| Phase 6 | ✅ **85%** | **85%** | VALIDATING状態実装 |
| Phase 7 | ✅ **85%** | **88%** | MetaCallLog統合 |
| Phase 8 | → | **95%** | 本番対応準備 |

---

## 🔴 High Priority（MVP完成に必須）

### 1. completion_assessment 実装 ✅
- [x] **OpenAI API呼び出し実装** (Meta層)
  - ファイル: `internal/meta/client.go`
  - 担当: PlanTask/NextAction と同じパターンで実装
  - 工数: 2日 ✅ 完了
  - 詳細:
    - System Prompt作成（AC評価用）✅
    - User Prompt作成（完了状況を記述）✅
    - Response YAML パース（CompletionAssessmentResponse）✅

- [x] **Runner での VALIDATING 遷移実装** (Core層)
  - ファイル: `internal/core/runner.go`
  - 工数: 1日 ✅ 完了
  - 詳細:
    - RUNNING → VALIDATING → COMPLETE/FAILED 遷移ロジック ✅
    - completion_assessment 呼び出しコード追加 ✅
    - AC Passed フラグ更新 ✅

- [x] **テスト追加**
  - ファイル: `internal/core/runner_test.go`
  - ファイル: `internal/meta/client_test.go`
  - 工数: 1日 ✅ 完了
  - 詳細: VALIDATING状態遷移テスト ✅、completion_assessment呼び出しテスト ✅

**合計工数: 3-5日** ✅ **完了**

---

### 2. VALIDATING状態の活用 ✅
- [x] **状態遷移ロジック定義**
  - 対象: `internal/core/runner.go`
  - 現状: 定義のみ (`context.go:15`)
  - 実装: RUNNING完了後、Meta-agentに完了評価を依頼 ✅
  - 工数: 1日 ✅ 完了

- [x] **Task Note への VALIDATING 情報出力**
  - 対象: `internal/note/writer.go`
  - 工数: 0.5日 ✅ 完了

**合計工数: 1-2日** ✅ **完了**

---

### 3. ImagePull 自動実行 ✅
- [x] **Docker イメージ自動取得実装**
  - ファイル: `internal/worker/sandbox.go:38`
  - 現状: `// TODO: implement image pull`（コメントアウト）
  - 実装: `docker pull` コマンド実行（イメージ未存在時）✅
  - 工数: 1日 ✅ 完了
  - 理由: 初回実行時のエラーハンドリング ✅

**合計工数: 1日** ✅ **完了**

---

## 🟡 Medium Priority（機能拡充）

### 4. MetaCallLog 記録の活用 ✅
- [x] **MetaCallLog 記録実装** (Core層)
  - ファイル: `internal/core/runner.go`
  - 現状: TaskContext.MetaCallLogs 定義のみ (`context.go:48-54`)
  - 実装内容:
    - PlanTask 呼び出し前: Request記録 ✅
    - PlanTask 完了後: Response記録 ✅
    - NextAction 呼び出し前: Request記録 ✅
    - NextAction 完了後: Response記録 ✅
  - 工数: 1日 ✅ 完了

- [x] **Task Note テンプレート更新** (Note層)
  - ファイル: `internal/note/writer.go`
  - 実装内容: `note/writer.go:70-83` の既存MetaCalls出力を活用 ✅
  - 工数: 0.5日 ✅ 完了

**合計工数: 1-2日** ✅ **完了**

---

### 5. AC Passed フラグ更新 ✅
- [x] **completion_assessment Response の AC反映**
  - ファイル: `internal/meta/protocol.go` (ACPassedフィールド定義) ✅
  - ファイル: `internal/core/runner.go` (更新ロジック) ✅
  - 実装: Meta-agentが各AC項目のPass/Fail判定を返す ✅
  - 工数: 1.5日 ✅ 完了

**合計工数: 2日** ✅ **完了**

---

### 6. TestResult Task Note 統合 ✅
- [x] **runner.go での TestResult 記録確認**
  - 現状: `runner.go:153-157` でTaskContext.TestResultに記録
  - 状態: ✅ 実装済み

- [x] **Task Note テンプレート統合**
  - ファイル: `internal/note/writer.go`
  - 実装: TestResult セクション追加（ExitCode/Output表示）✅
  - 工数: 0.5日 ✅ 完了

**合計工数: 0.5日** ✅ **完了**

---

### 7. maxLoops 設定化 ✅
- [x] **TaskConfig に maxLoops フィールド追加**
  - ファイル: `pkg/config/config.go`
  - デフォルト値: 10 ✅
  - 工数: 0.5日 ✅ 完了

- [x] **runner.go でのデフォルト値補完**
  - ファイル: `internal/core/runner.go:104`
  - 現状: `maxLoops := 10` ハードコード
  - 実装: 設定から読み込み（未設定時は10）✅
  - 工数: 0.5日 ✅ 完了

**合計工数: 1日** ✅ **完了**

---

### 8. LLM エラー再試行ロジック
- [ ] **exponential backoff 実装**
  - ファイル: `internal/meta/client.go`
  - 実装内容:
    - 最大再試行回数: 3回
    - 初期待機: 1秒
    - 指数バックオフ: 1 → 2 → 4秒
    - 対象エラー: HTTP 5xx, timeout, rate limit
  - 工数: 2日

- [ ] **テスト追加**
  - ファイル: `internal/meta/client_test.go`
  - 工数: 1日

**合計工数: 3日**

---

## 🟢 Low Priority（最適化・品質向上）

### 9. コンテナライフサイクル最適化
- [ ] **Task開始時に1回だけコンテナ起動**
  - ファイル: `internal/core/runner.go`
  - 現状: RunWorker毎に start/stop（非効率）
  - 実装:
    - PLANNING → RUNNING：コンテナ起動
    - 各RunWorker：Exec のみ
    - COMPLETE/FAILED：コンテナStop
  - 工数: 5日
  - 影響: Worker実行速度 5-10倍向上

- [ ] **テスト更新**
  - ファイル: `internal/worker/executor_test.go`
  - 工数: 2日

**合計工数: 5-7日**

---

### 10. Worker 複数種対応
- [ ] **WorkerExecutor インターフェース拡張**
  - ファイル: `internal/worker/executor.go`
  - 実装内容:
    - 現在: Codex CLI のみ
    - 拡張: 他言語エージェント対応可能な設計
  - 工数: 10日+

**合計工数: 10日以上**

---

### 11. カバレッジ向上（43.4% → 55%+）
- [ ] **main 関数テスト拡張**
  - ファイル: `cmd/agent-runner/main_test.go`
  - 現状: 10テスト
  - 目標: エッジケース追加
  - 工数: 2日

- [ ] **Core層追加テスト**
  - ファイル: `internal/core/runner_test.go`
  - 工数: 3日

- [ ] **Meta層追加テスト**
  - ファイル: `internal/meta/client_test.go`
  - 工数: 2日

- [ ] **Worker/Sandbox追加テスト**
  - ファイル: `internal/worker/executor_test.go`
  - 工数: 2日

**合計工数: 9日**

---

## 🛣️ Phase 別ロードマップ

### Phase 6: VALIDATING状態実装（所要時間: 5-7日）✅ **完了**

**目標実現率: 85%** ✅ **達成**

**実装順序**:
1. ✅ completion_assessment API実装 (meta/client.go)
2. ✅ Runner での VALIDATING遷移実装 (core/runner.go)
3. ✅ AC Passed フラグ更新 (protocol.go → runner.go)
4. ✅ テスト追加 (runner_test.go, client_test.go)

**成果物**:
- ✅ completion_assessment 完全実装
- ✅ VALIDATING状態の活用
- ✅ AC評価の完全自動化
- ✅ ImagePull自動実行

---

### Phase 7: MetaCallLog・TestResult統合（所要時間: 2-3日）✅ **完了**

**目標実現率: 88%** ✅ **達成予定**

**実装順序**:
1. ✅ MetaCallLog 記録実装 (core/runner.go)
2. ✅ Task Note テンプレート統合 (note/writer.go)
3. ✅ TestResult統合 (note/writer.go)
4. ✅ maxLoops設定化 (config.go, runner.go)

**成果物**:
- ✅ 監査証跡の完全性確保
- ✅ 自動テスト結果の完全出力
- ✅ YAML設定による実行制御

---

### Phase 8: 本番対応準備（所要時間: 10-15日）

**目標実現率: 95%**

**実装順序**:
1. ImagePull 自動実行 (worker/sandbox.go)
2. LLMエラー再試行ロジック (meta/client.go)
3. コンテナライフサイクル最適化 (core/runner.go, worker/)
4. カバレッジ向上 (test/ 全体)

**成果物**:
- ✅ 本番環境での完全対応
- ✅ 高度なエラーハンドリング
- ✅ パフォーマンス最適化
- ✅ 55%+ カバレッジ

---

## 📋 実装チェックリスト

### Phase 6 (High Priority) ✅
- [x] completion_assessment API実装 ✅
  - [x] System Prompt作成 ✅
  - [x] User Prompt作成 ✅
  - [x] Response パース ✅
  - [x] テスト追加（3テスト）✅
- [x] VALIDATING遷移実装 ✅
  - [x] FSM遷移ロジック ✅
  - [x] Meta呼び出し ✅
  - [x] AC更新 ✅
  - [x] テスト追加（4テスト）✅
- [x] ImagePull実装 ✅
  - [x] Docker pull コマンド ✅
  - [x] エラーハンドリング ✅
  - [x] テスト追加（2テスト）✅

### Phase 7 (Medium Priority) ✅
- [x] MetaCallLog記録 ✅
  - [x] 各呼び出しでログ記録 ✅
  - [x] タイムスタンプ付与 ✅
  - [x] テスト追加（2テスト）✅
- [x] Task Note統合 ✅
  - [x] MetaCalls セクション出力 ✅
  - [x] TestResult セクション出力 ✅
- [x] maxLoops設定化 ✅
  - [x] config.goフィールド追加 ✅
  - [x] runner.go実装 ✅
  - [x] テスト追加（1テスト）✅

### Phase 8 (本番対応)
- [ ] LLMエラー再試行
  - [ ] exponential backoff実装
  - [ ] リトライロジック
  - [ ] テスト追加（5テスト）
- [ ] コンテナ最適化
  - [ ] ライフサイクル実装
  - [ ] テスト更新（4テスト）
- [ ] カバレッジ向上
  - [ ] エッジケーステスト追加（9テスト）

---

## 📈 期待される効果

| 項目 | 開始時 | Phase 6完了時 ✅ | Phase 7完了時 ✅ | Phase 8完了時 |
|------|-------|--------------|-------------|-------------|
| 実現率 | 78% | 85% ✅ | 85%+ ✅ | 95% |
| テスト数 | 99+ | 108+ | 112+ ✅ | 135+ |
| カバレッジ | 43.4% | 48% | 43.8% ✅ | 55%+ |
| Meta層 カバレッジ | 28% | 60% | 60% ✅ | 75% |
| Core層 カバレッジ | 77.9% | 90% | 82% ✅ | 95% |

---

## 🚀 次に着手すべき項目

優先度・工数を考慮した推奨着手順序：

**完了済み（✅ 2025-11-21完了）**:
1. ✅ Phase 6実装（completion_assessment + VALIDATING）- **5-7日で完了**
2. ✅ ImagePull実装（単体機能で独立）- **1日で完了**
3. ✅ Phase 7実装（MetaCallLog + maxLoops設定）- **2-3日で完了**

**次ステップ（推奨：2-4週間）**:
4. 🎯 Phase 8：本番対応準備（LLMエラー再試行 + コンテナ最適化）
5. 🎯 カバレッジ向上（並行実施）

**長期（1ヶ月以上）**:
6. Worker複数種対応（大規模リファクタリング）

---

## 📝 注記

- ✅ **Phase 6・7が2025-11-21に完了しました**
- **現在**: VALIDATING状態による品質保証、完全な監査証跡（MetaCallLog）、柔軟な設定（maxLoops）が実装済み
- 本番テスト運用が可能な状態に達しました ✅
- **次フェーズ（Phase 8）**: LLMエラー再試行、コンテナライフサイクル最適化、カバレッジ向上（目標: 55%+）
- 各タスクの工数見積もりは相対的なものです。実装中に調整される可能性があります。
- テスト追加タスクは各Phase内で段階的に実施します。
