# AgentRunner ドキュメント

このディレクトリには AgentRunner プロジェクトの設計・仕様・開発ガイドが含まれています。

## ドキュメント構成

### 📋 [specifications/](specifications/) - 仕様ドキュメント

確定した仕様を定義するドキュメント群です。実装の基準となります。

- [core-specification.md](specifications/core-specification.md) - コア仕様（YAML、TaskContext、FSM、Task Note）
- [meta-protocol.md](specifications/meta-protocol.md) - Meta-agent プロトコル仕様
- [worker-interface.md](specifications/worker-interface.md) - Worker 実行仕様
- [orchestrator-spec.md](specifications/orchestrator-spec.md) - Orchestrator 仕様（Task 永続化・スケジューリング・IPC）
- [logging-specification.md](specifications/logging-specification.md) - ロギング仕様（Trace ID・構造化ログ）
- [testing-strategy.md](specifications/testing-strategy.md) - テスト戦略（Backend/Frontend E2E）

### 🏗️ [design/](design/) - 設計ドキュメント

システムの設計思想と実装方針を説明するドキュメント群です。

- [architecture.md](design/architecture.md) - システムアーキテクチャ
- [implementation-guide.md](design/implementation-guide.md) - 実装ガイド（Go 固有）
- [data-flow.md](design/data-flow.md) - データフロー設計

### 📖 [guides/](guides/) - 開発ガイド

開発者向けの実践的なガイドです。

- [testing.md](guides/testing.md) - テスト戦略とベストプラクティス
- [codex-integration.md](guides/codex-integration.md) - Codex 統合テスト実行ガイド

### 🔧 その他

- [CLAUDE.md](CLAUDE.md) - ドキュメント整理ルールと管理方針

## ドキュメントの読み方

### 初めての方

1. [design/architecture.md](design/architecture.md) でシステム全体像を把握
2. [specifications/core-specification.md](specifications/core-specification.md) でコア仕様を理解
3. [design/implementation-guide.md](design/implementation-guide.md) で実装方針を確認

### 実装者向け

1. [specifications/](specifications/) で仕様を確認
2. [design/implementation-guide.md](design/implementation-guide.md) で実装パターンを学習
3. [guides/testing.md](guides/testing.md) でテスト方法を確認

### アーキテクト向け

1. [design/architecture.md](design/architecture.md) でシステム設計を確認
2. [design/data-flow.md](design/data-flow.md) でデータフローを理解
3. [specifications/](specifications/) で仕様詳細を確認

## ドキュメント管理

ドキュメントの整理ルールと更新方針については [CLAUDE.md](CLAUDE.md) を参照してください。
