# 製品要件定義書 (PRD): Multiverse IDE Agent Workflow & Architecture

## 1. はじめに

本 PRD は、AI ネイティブな開発環境「Multiverse IDE」の中核となる、エージェントワークフローと永続化アーキテクチャを定義します。
従来の「静的な設計書」と「動的な実行状態」の厳密な分離（v2 理想像）から、**「IDE がワークスペースの状態を統合管理し、エージェントが自律的にタスクを消化する」** 実用的かつ堅牢なアーキテクチャ（Pragmatic MVP）への移行を定めます。

## 2. ゴールと目的

- **IDE 主導のワークフロー**: ユーザーは IDE のチャットや GUI を通じて指示を出し、IDE がそれをタスクグラフに変換して管理する。
- **自律的な実行**: バックエンドの Orchestrator が依存関係を解決しながら、複数のエージェント（Worker）を並列に稼働させる。
- **完全な可観測性**: タスクの生成、実行、結果、チャット履歴がすべて永続化され、IDE 上でリアルタイムに可視化される。
- **実用的な永続化**: 複雑さを排除した「統合タスクストア」により、データの整合性と開発速度を両立する。

## 3. システムアーキテクチャ

### 3.1 全体構成

```mermaid
graph TD
    User[Developer] -->|Chat/GUI| IDE[Multiverse IDE (Frontend)]
    IDE -->|Wails Events| BE[Orchestrator (Backend)]

    subgraph Backend
        Chat[Chat Handler]
        Graph[Task Graph Manager]
        Store[Task Store (JSONL)]
        Sched[Scheduler]
        Exec[Executor]
    end

    subgraph Workers
        Agent1[Code Agent]
        Agent2[Analysis Agent]
    end

    Chat -->|Task Generation| Store
    Store -->|Events| IDE
    Graph -->|Ready Tasks| Sched
    Sched -->|Job| Exec
    Exec -->|Run| Workers
    Workers -->|Result| Exec
    Exec -->|Update| Store
```

### 3.2 データモデル (Unified Task Model)

設計情報（WBS）と実行状態（State）を単一の **Task** エンティティとして管理し、`JSONL` で永続化します。

- **Task**: 作業の最小単位。

  - `ID`: 一意な識別子 (UUID)。
  - `Title`, `Description`: タスクの内容。
  - `Status`: `PENDING`, `RUNNING`, `SUCCEEDED`, `FAILED` 等。
  - `Dependencies`: 依存するタスク ID のリスト (DAG 構築用)。
  - `PhaseName`: WBS フェーズ（概念設計, 実装, 検証 等）。
  - `SuggestedImpl`: **[実装済]** エージェントへの具体的な実装指示（言語、変更ファイルパス、制約事項）。
  - `Artifacts`: **[実装済]** 生成された成果物（ファイルパス、ログファイル等）。

- **Work Breakdown Structure (WBS)**:
  - 固定的な `wbs.json` ファイルは持たず、IDE がフラットなタスクリストから動的にツリー構造（マイルストーン > フェーズ > タスク）を導出する。

## 4. 機能要件

### 4.1 エージェントワークフロー

1. **チャットからのタスク生成**:
   - ユーザーのチャット入力を Meta-agent (LLM) が分析。
   - 依存関係を含む一連のタスク (`Task` オブジェクト) を生成。
   - `SuggestedImpl` フィールドに、変更すべきファイルや技術スタックの制約を含める（`Meta` プロトコル v1.1）。
2. **タスクグラフ管理**:
   - バックエンドはタスクの `Dependencies` を解析し、実行可能なタスク (`Ready`) を特定。
   - 循環参照を検出し、エラーとして報告。
3. **実行とスケジューリング**:
   - 利用可能な Worker プール（Codex, Test 等）に対してタスクをディスパッチ。
   - 実行結果（成功/失敗、生成ファイル）を `Artifacts` としてタスクに記録。

### 4.2 永続化と履歴

- **Unified Task Store**:
  - `~/.multiverse/workspaces/<id>/tasks/<task-id>.jsonl`
  - 各行がタスクの状態スナップショットを表す（追記のみ）。最後の行が最新状態。
- **Chat Session Store**:
  - チャット履歴を `sessions/<session-id>.jsonl` に保存。タスク生成の文脈を保持。

### 4.3 IDE インテグレーション

- **リアルタイム同期**:
  - バックエンドの状態変化は Wails イベントを通じて即座にフロントエンドに通知。
- **可視化**:
  - **Graph View**: 依存関係と実行状況をノードグラフ (`TaskNode`) として表示。「IP」インジケーターにより実装ヒントの有無を可視化。
  - **WBS View**: タスクをフェーズごとにグループ化して進捗バーと共に表示 (`WBSNode`)。
  - **Property Panel**: 選択したタスクの詳細（`SuggestedImpl`、`Artifacts` 含む）を表示。

## 5. 次期開発フェーズ (Next Steps)

1. **Snapshot 機能**: 特定時点のワークスペース状態を保存・復元する機能（「プラン A 失敗時の巻き戻し」用）。
2. **IPC / WebSocket 強化**: 大量のログやターミナル出力をリアルタイムに IDE にストリーミングする。
3. **Multi-Agent Orchestration**: Code Agent と Review Agent の協調動作フローの実装。
