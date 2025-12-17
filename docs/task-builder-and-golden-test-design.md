# Task Builder & Golden Test 設計書

本ドキュメントは、Multiverse IDE における「チャット入力 → plan_patch（WBS/Node/TaskState の作成・更新） → TaskConfig YAML 生成 → AgentRunner 実行 → 結果反映」までの最小パイプラインと、ゴールデンテスト（`TODO アプリを作成して`）の仕様を定義する。

実装時の指示書として利用することを前提とする。

---

## 1. 背景・目的

- ユーザーは **チャット UI から自然文を入力**してタスクを起動する。
- 内部では、その自然文をもとに **TaskConfig YAML** を生成し、それを AgentRunner に渡す。
- AgentRunner は、タスク分析・実装・ファイル生成・検証（テスト等）までを実行し、その結果を IDE に返す。
- Phase 0 のゴールは、以下の 1 本のパイプラインが「ローカルで一気通しで動作すること」である。

> Chat（`TODO アプリを作成して`）  
> → Meta plan_patch により WBS/NodeDesign/TaskState を生成/更新（ChatHandler）  
> → Orchestrator が依存解決し Executor で TaskConfig YAML を生成 → AgentRunner 実行  
> → 結果が IDE に表示される

TODO アプリの仕様・技術スタック・テスト戦略などは **一切固定しない**。  
本ドキュメントの範囲は「パイプラインとしての契約と責務」のみを定義する。

---

## 2. コンポーネントと責務

### 2.1 IDE (Chat Layer)

- ユーザーと対話するフロントエンド。
- ユーザー入力（自然文）をバックエンドの ChatHandler に送信する。
- Task の一覧表示、ステータス表示、結果サマリの表示を行う（TaskStore / state の反映を受け取る）。

### 2.2 Orchestrator

- Workspace / TaskStore / IPC の管理を行うバックエンドコンポーネント。
- 主な責務:
  - `state/tasks.json` / `state/nodes-runtime.json` / `design/nodes/*.json` を読み、依存関係を解決して READY タスクを選ぶ（Scheduler）。
  - READY タスクを IPC queue にジョブとして登録し、ExecutionOrchestrator が消費する。
  - Executor がタスクから TaskConfig YAML を生成し、`agent-runner` に stdin で渡して実行する。
  - 実行結果を `state/` と TaskStore に反映し、IDE にイベントとして露出する。

### 2.3 Task Builder（バックログ）

Task Builder（`raw_prompt` → TaskConfig YAML）の導入は `ISSUE.md`（Deferred: 「Task Builder（raw_prompt → TaskConfig YAML）」）を正とする。

### 2.4 AgentRunner

- Meta / Worker エージェントのランタイム。
- 入力:
  - TaskConfig YAML（Executor の出力をそのまま受け取る）
- 処理:
  - タスク分析・プランニング
  - コード編集・新規ファイル生成
  - 可能な範囲での検証（テスト実行・ビルド・lint 等）
- 出力:
  - Task 実行結果の JSON（タスクサマリ・検証内容・ステータス等）。

### 2.5 TaskStore / Workspace

- ローカルファイルシステム上のメタデータ保存レイヤ。
- ディレクトリ構造（概要）:

```text
~/.multiverse/workspaces/<workspace-id>/
  workspace.json
  tasks/
    <task-id>.jsonl
  ipc/
    queue/
    results/
  logs/
```

---

## 3. データモデル

### 3.1 TaskStore: Task レコード

ChatHandler により作成されるタスクの最小レコード定義（実体は `orchestrator.Task` の JSONL 追記）。

```jsonc
// ~/.multiverse/workspaces/<workspace-id>/tasks/<task-id>.jsonl
{
  "id": "golden-todo-001",
  "title": "TODO アプリを作成して",
  "description": "TODO アプリを作成して。技術スタックや実装方針、検証方法はあなたの判断に任せます。",
  "status": "PENDING",
  "dependencies": [],
  "wbsLevel": 1,
  "phaseName": "Implementation",
  "milestone": "implementation",
  "acceptanceCriteria": ["アプリが起動すること"]
}
```

※ ユーザーの自然文入力そのものは TaskStore ではなく ChatSessionStore に保存される。

※ Phase 0 では `test_command` 等は持たない。検証戦略は AgentRunner 側に委譲する。

### 3.2 IPC Queue: Job JSON

IDE からの「実行してほしい」要求は、Orchestrator に対して IPC queue 経由で渡される。

```jsonc
// ~/.multiverse/workspaces/<workspace-id>/ipc/queue/<job-id>.json
{
  "workspace_id": "abcd1234ef56",
  "task_id": "golden-todo-001"
}
```

- Orchestrator は queue ディレクトリをポーリングし、Job を検出して処理する。

### 3.3 TaskConfig YAML（Executor 出力 / AgentRunner 入力）

Executor が生成し、AgentRunner に渡される YAML の最小スキーマを定義する（`pkg/config/config.go` に準拠）。

```yaml
version: 1

task:
  id: "golden-todo-001"
  title: "TODO アプリを作成して"
  repo: "."
  prd:
    text: |
      TODO アプリを作成して。

runner:
  max_loops: 5
  worker:
    kind: "codex-cli"
    # 必要に応じて docker_image / env 等を拡張
```

必須フィールド:

- `version`（値は `1`）
- `task.prd`（`path` または `text` のいずれか）

Executor 実装は、このスキーマを満たす YAML を決定的に生成する。

### 3.4 AgentRunner 結果 JSON

AgentRunner がタスク実行完了時に Orchestrator に返す結果 JSON の最小仕様。

```jsonc
{
  "task_id": "golden-todo-001",
  "status": "succeeded",   // "succeeded" | "failed"
  "summary": "TODO アプリを作成し、基本的な追加・削除・一覧機能と簡単な検証処理を実行しました。",
  "validation": {
    "overall": "passed",   // "passed" | "failed" | "unknown"
    "commands": [
      {
        "command": "npm test",
        "exit_code": 0,
        "duration_ms": 12345
      }
    ]
  },
  "duration_ms": 600000
}
```

- `status`
  - AgentRunner レベルでの成功/失敗。
- `summary`
  - 実装内容の自然文サマリ（IDE 表示用）。
- `validation`
  - AgentRunner 内で実施した検証（テスト / ビルド / lint 等）の概要。
  - Phase 0 では 1 コマンド / 0 コマンドでも可（`commands` は空配列を許容）。
- `duration_ms`
  - 全体の実行時間（任意だが、あると便利）。

Orchestrator は、本 JSON を TaskAttempt（JSONL）に埋め込み、IDE から参照可能にする。

---

## 4. 処理フロー

### 4.1 Chat → Task 作成

1. ユーザーが IDE のチャット欄に以下を入力する。

   > `TODO アプリを作成して`

2. バックエンドの ChatHandler が Meta plan_patch を呼び出し、Task 群を生成/更新する。
3. ChatHandler が以下を永続化する:
   - `design/wbs.json`, `design/nodes/*.json`
   - `state/tasks.json`, `state/nodes-runtime.json`
   - TaskStore の `tasks/<task-id>.jsonl`

4. ユーザーは Task 一覧画面で生成されたタスクを確認できる。

### 4.2 Task 実行要求 → Orchestrator

1. ユーザーが IDE 上で Task の「Run」ボタンを押下。
2. IDE はバックエンドに実行要求を送信し、Scheduler が IPC queue に Job JSON を作成する（3.2 参照）。
3. ExecutionOrchestrator が queue ディレクトリを監視し、Job を検出。

### 4.3 Executor による TaskConfig YAML 生成

1. ExecutionOrchestrator は `state/tasks.json` と `design/nodes/*.json` から Task をロードする。
2. Executor が Task から TaskConfig YAML を生成する（3.3 に準拠）。
3. ExecutionOrchestrator が YAML を `agent-runner` に stdin で渡して実行する。
   - YAML としてパース可能か
   - 必須フィールドが存在するか
5. 検証に失敗した場合、または CLI セッションが無い場合は、その時点で TaskAttempt を `failed` として記録し、結果を IDE に返す。

### 4.4 AgentRunner 実行

1. Orchestrator は検証済み TaskConfig YAML を AgentRunner に渡す（実装としては `agent-runner` サブプロセスの stdin 等）。
2. AgentRunner は内部で以下を行う（振る舞いは AgentRunner 側の設計に従う）:
   - タスク分析・プランニング
   - コード編集・ファイル生成
   - 可能な限りの自己検証（テスト / ビルド / lint 等）
3. 完了時、AgentRunner は 3.4 の JSON を stdout（またはファイル）として出力する。
4. Orchestrator はこの JSON を受け取り、TaskAttempt として TaskStore に追記し、IPC results にも書き出す。

### 4.5 IDE での結果表示

1. IDE は IPC results をポーリング or ファイル監視し、対象 Job の result JSON を検出。
2. Task 一覧画面:
   - 対象 Task のステータスを `SUCCEEDED` / `FAILED` に更新。
3. Task 詳細画面:
   - `status` / `summary` / `validation.overall` / `validation.commands` 等を表示する。

---

## 5. ゴールデンテスト仕様

### 5.1 前提

- ゴールデンテストのユーザー入力は **固定** とする。

  ```text
  TODO アプリを作成して
  ```

- TODO アプリの解釈・技術スタック・設計・テスト戦略に関するルールは **一切課さない**。
- 検証対象は「アプリとして妥当か」ではなく、「パイプラインとして正しく通るか」である。

### 5.2 GT-1: Chat → TaskConfig（Executor テスト）

目的:

- `TODO アプリを作成して` の plan_patch 結果から **有効な TaskConfig YAML** が生成されることを確認する。

前提条件:

- Meta plan_patch をモックできること（LLM 実行は不要）

テスト手順（ロジック）:

1. テスト用 Workspace を作成（空 or ほぼ空でよい）。
2. ChatHandler に `TODO アプリを作成して` を入力し、Task を生成。
3. Executor を起動し、TaskConfig YAML を取得。
4. アサーション:
   - YAML としてパース可能。
   - `task.id` が TaskStore の `id` と一致。
   - `task.title` が `TODO アプリを作成して` を含む。
   - `task.repo` が `"."`。
   - `task.prd.text` に `TODO アプリを作成して` と Acceptance Criteria が含まれる。
   - `runner.max_loops` と `runner.worker.kind` が存在。

### 5.3 GT-2: TaskConfig → AgentRunner（実行テスト）

目的:

- TaskConfig YAML を AgentRunner に渡した際、実装・ファイル生成・自己検証までの処理が完了し、結果 JSON が返ることを確認する。

前提条件:

- Codex CLI がインストールされ、有効なセッションが存在すること
- Docker が起動しており、Codex Worker イメージが利用可能であること

テスト手順（ロジック）:

1. GT-1 で取得した TaskConfig YAML をそのまま AgentRunner に入力。
2. AgentRunner を実行し、結果 JSON（3.4）を取得。
   - AgentRunner は Docker サンドボックス内で Codex CLI を実行
   - Codex CLI セッションが Docker コンテナ内で利用可能であることを確認
3. アサーション:
   - プロセスとして正常終了している（exit code = 0 が望ましいが、結果 JSON の `status` を見て判定）。
   - Workspace ディレクトリ内で 1 つ以上のファイルが新規作成 or 更新されている。
   - 結果 JSON に以下が含まれる:
     - `task_id`（TaskStore の id と一致）
     - `status`（"succeeded" or "failed"）
     - `summary`（非空の文字列）
     - `validation` オブジェクト（存在すればよい。`commands` が空でも許容）

※ Phase 0 の時点では、`status = failed` であっても、「パイプラインとして最後まで処理され、結果が返る」ことを成功条件としてよい。

### 5.4 GT-3: E2E（Chat → plan_patch → TaskConfig → AgentRunner → 結果）

目的:

- IDE チャット入力から結果表示まで、全パスが一気通しで動くことを確認する。

テスト手順（ロジック）:

1. IDE のテストモードで以下を実行する:
   - Chat に `TODO アプリを作成して` を入力し、Task 作成。
   - Task の「Run」ボタンを押下。
2. バックグラウンドで:
   - ChatHandler が plan_patch → design/state/task_store 永続化を実行。
   - Orchestrator が依存解決 → Executor による TaskConfig YAML 生成 → AgentRunner 実行 → 結果 JSON 生成。
3. IDE で Task 詳細画面を開き、以下を確認:
   - ステータスが `SUCCEEDED` または `FAILED` のいずれか。
   - summary が表示されている。
   - validation.overall が `passed` / `failed` / `unknown` のいずれか（存在すればよい）。

---

## 6. 実装順序（Phase 0 向け指針）

実装順序の推奨:

1. Workspace / design / state / TaskStore / IPC（queue/results）の基盤実装。
2. IDE:
   - Workspace 選択 UI
   - Chat 入力 UI と Task 表示 UI
   - Task 実行要求 UI（Run ボタン）
3. ChatHandler:
   - Meta plan_patch 呼び出し
   - `design/`・`state/`・TaskStore の永続化
4. Orchestrator:
   - Scheduler による依存解決と Job enqueue
   - ExecutionOrchestrator による Job 実行と状態更新
5. Executor / AgentRunner 連携:
   - TaskConfig YAML 生成（Executor）
   - `agent-runner` 実行と結果 JSON の保存
6. ゴールデンテスト（GT-1 / GT-2 / GT-3）の追加

本設計書は Phase 0 の最小スコープを対象とする。  
Phase 1 以降で、複数エージェント、WorkerPool、シナリオベースの L2 テスト等を拡張するが、それらは別途仕様書で定義する。
