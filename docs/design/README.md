# 設計ドキュメント

このディレクトリには AgentRunner の設計思想と実装方針が含まれています。

## ドキュメント一覧

### [architecture.md](architecture.md)

システム全体のアーキテクチャを説明します。

- **対象読者**: アーキテクト、技術リード
- **内容**:
  - システム構成
  - コンポーネント設計
  - 役割分担
  - 設計思想と原則

### [implementation-guide.md](implementation-guide.md)

Go 言語での実装ガイドを提供します。

- **対象読者**: 実装者、コントリビューター
- **内容**:
  - パッケージ構成
  - 依存性注入パターン
  - インターフェース設計
  - 実装パターン
  - テスト戦略

### [data-flow.md](data-flow.md)

データフローと状態遷移を説明します。

- **対象読者**: 実装者、デバッガー
- **内容**:
  - タスク実行フロー
  - 状態遷移図
  - データ変換
  - エラーフロー

### [task-execution-and-visual-grouping.md](task-execution-and-visual-grouping.md)

タスクの「計画→実行」遷移と、IDE 上での多軸グルーピング/フィルタリング設計を説明します。

- **対象読者**: 実装者、UI/UX 設計者
- **内容**:
  - Planning と Execution の責務分離
  - 分類メタデータ（Facet）設計
  - Backend API / Frontend 表示方針
  - 既存ワークスペースの互換・移行方針

### [chat-autopilot.md](chat-autopilot.md)

自然な会話だけで「計画→実行→質問→継続」を回すための Chat Autopilot 設計です。

- **対象読者**: 実装者、プロダクト設計者
- **内容**:
  - Autopilot の責務とデータフロー
  - 自然言語での停止/再開/状況確認
  - 質問（Backlog）を会話に統合する方針
  - 既存 Orchestrator/Runner との整合

### [tooling-selection.md](tooling-selection.md)

Tooling (ツール/モデル選択) の設計を説明します。

- **対象読者**: アーキテクト、実装者
- **内容**:
  - ToolingConfig の構造
  - 選択アルゴリズムとフォールバック
  - IDE/Orchestrator/AgentRunner の統合点
  - 既知の制約とテスト

## 設計の読み方

1. [architecture.md](architecture.md) でシステム全体像を把握
2. [data-flow.md](data-flow.md) で実行フローを理解
3. [implementation-guide.md](implementation-guide.md) で実装方針を確認

## 設計の更新ルール

- 設計変更は実装前に文書化
- 設計判断の理由を明記
- 代替案と選択理由を記録
