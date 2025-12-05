# 仕様ドキュメント

このディレクトリには AgentRunner の確定仕様が含まれています。

## ドキュメント一覧

### [core-specification.md](core-specification.md)

AgentRunner のコア仕様を定義します。

- **対象読者**: 実装者、レビュアー
- **内容**:
  - Task YAML スキーマ
  - TaskContext 構造
  - タスク状態機械（FSM）
  - Task Note フォーマット
  - CLI インターフェース

### [meta-protocol.md](meta-protocol.md)

Meta-agent との通信プロトコル仕様を定義します。

- **対象読者**: Meta-agent 実装者、プロトコル設計者
- **内容**:
  - `plan_task` プロトコル
  - `next_action` プロトコル
  - `completion_assessment` プロトコル
  - YAML メッセージフォーマット
  - エラーハンドリング

### [worker-interface.md](worker-interface.md)

Worker 実行とサンドボックス環境の仕様を定義します。

- **対象読者**: Worker 実装者、インフラ担当者
- **内容**:
  - Worker 実行インターフェース
  - Docker サンドボックス仕様
  - 環境変数とマウント仕様
  - 実行結果フォーマット
  - タイムアウトとエラーハンドリング

### [orchestrator-spec.md](orchestrator-spec.md)

Orchestrator のタスク管理・永続化・IPC 仕様を定義します。

- **対象読者**: Orchestrator 実装者、IDE バックエンド開発者
- **内容**:
  - Task Scheduler / Executor / Store
  - IPC（ファイルベースキュー・結果）
  - データモデル（Task, Attempt）
  - 拡張計画

### [logging-specification.md](logging-specification.md)

統一ロギングシステムの仕様を定義します。

- **対象読者**: 開発者、インフラ担当者
- **内容**:
  - Trace ID 伝播
  - 構造化ログ（log/slog）
  - ログレベル定義
  - JSON/Text フォーマット

### [testing-strategy.md](testing-strategy.md)

Backend/Frontend のテスト戦略を定義します。

- **対象読者**: テスター、開発者
- **内容**:
  - テスト配置とディレクトリ構成
  - テスト実行方法
  - 検証範囲と受け入れ基準

## 仕様の読み方

1. まず [core-specification.md](core-specification.md) でシステムの基本仕様を理解
2. Meta-agent を実装する場合は [meta-protocol.md](meta-protocol.md) を参照
3. Worker を実装する場合は [worker-interface.md](worker-interface.md) を参照

## 仕様の更新ルール

- 仕様変更は必ず設計レビューを経てから反映
- バージョン管理は Git のタグで管理
- 後方互換性を破る変更は明示的にマーク
