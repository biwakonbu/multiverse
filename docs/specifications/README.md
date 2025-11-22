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

## 仕様の読み方

1. まず [core-specification.md](core-specification.md) でシステムの基本仕様を理解
2. Meta-agent を実装する場合は [meta-protocol.md](meta-protocol.md) を参照
3. Worker を実装する場合は [worker-interface.md](worker-interface.md) を参照

## 仕様の更新ルール

- 仕様変更は必ず設計レビューを経てから反映
- バージョン管理は Git のタグで管理
- 後方互換性を破る変更は明示的にマーク
