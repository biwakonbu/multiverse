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

## 設計の読み方

1. [architecture.md](architecture.md) でシステム全体像を把握
2. [data-flow.md](data-flow.md) で実行フローを理解
3. [implementation-guide.md](implementation-guide.md) で実装方針を確認

## 設計の更新ルール

- 設計変更は実装前に文書化
- 設計判断の理由を明記
- 代替案と選択理由を記録
