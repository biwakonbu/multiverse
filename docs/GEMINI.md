# docs/GEMINI.md

このディレクトリには、プロジェクトの仕様書、設計書、ガイドが含まれています。

## ドキュメント構成

### 1. `specifications/` (仕様書)

システムの振る舞い、インターフェース、データ構造の定義。

- `core-specification.md`: AgentRunner Core の詳細仕様。
- `meta-protocol.md`: LLM との通信プロトコル (YAML) 定義。
- `orchestrator-spec.md`: **[新規]** Orchestrator のアーキテクチャと IPC 仕様。
- `testing-strategy.md`: **[新規]** E2E テストの設計と方針。

### 2. `design/` (設計書)

アーキテクチャの決定理由、データフロー、実装指針。

- `architecture.md`: システム全体の構成図。
- `data-flow.md`: データと状態遷移の流れ。

### 3. `guides/` (ガイド)

開発者向けの手順書。

- `cli-subscription.md`: **[新規]** CLI サブスクリプション運用ガイド（Codex 等の利用方法）。

## ドキュメントの読み方

1. **全体像**: ルートの `GEMINI.md` と `PRD.md` を参照。
2. **詳細仕様**: 作成・修正するコンポーネントに対応する `specifications/` 内のファイルを参照。
3. **実装詳細**: `internal/` 以下のコードと `GEMINI.md` を参照。
