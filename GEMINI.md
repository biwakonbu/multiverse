# GEMINI.md

## プロジェクト概要

`multiverse` は、AI ネイティブな開発環境を実現するための統合プラットフォームです。以下のコンポーネント群で構成され、ローカル環境での自律的なソフトウェア開発タスクの実行を可能にします。

1.  **AgentRunner Core (Engine)**: AI エージェントを Docker サンドボックス内で安全に実行・管理するコアエンジン。
2.  **Multiverse Orchestrator (Backend)**: 複数のタスクと Worker を管理し、IDE からのリクエストを処理するオーケストレーション層。
3.  **Multiverse IDE (Frontend)**: 開発者がタスクの作成、実行、監視を行うためのデスクトップアプリケーション (Wails + Svelte)。

## アーキテクチャ

システム全体は 4 層構造になっています。

- **Frontend Layer (IDE)**: ユーザーインターフェース。チャット・タスクグラフ・WBS を表示。Wails + Svelte で実装。
- **Orchestration Layer**: ChatHandler、TaskGraphManager、ExecutionOrchestrator、タスク永続化。
- **Core Layer (AgentRunner)**: Meta-agent（LLM）による計画・評価と FSM ベースの状態管理。
- **Execution Layer (Worker)**: Docker コンテナ内での実際のコード生成やテスト実行。

## デザインシステム

### テーマ: Nord Deep

**コンセプト**: 深い背景にパステル UI が輝くデザイン

- **ベース**: Nord パレットをベースに深い背景色を拡張
- **UI**: Aurora パレットをパステル化したステータス色
- **グロー**: 控えめな効果（IDE としての実用性重視）
- **スタイル**: SF/Sci-Fi 風の洗練された UI

### スタイル: Glassmorphism（ガラスモーフィズム）

**Phantom Glass** スタイルを採用:

- **背景**: 半透明の暗いガラス効果（`rgba(22, 24, 30, 0.4)`）
- **ボーダー**: 微細な白いハイライト（`rgba(255, 255, 255, 0.05)`）
- **シャドウ**: アンビエントシャドウとインナーハイライト
- **ブラー**: 背景のぼかし効果

### UI コンセプト: Crystal HUD

ツールバー、パネル、チャットウィンドウなどに適用:

- **透明感**: 背景が透けて見える洗練されたパネル
- **グロー効果**: Frost 青系のアクセントグロー
- **フォント**: Orbitron（ブランド）、Rajdhani（ディスプレイ）

### CSS 変数プレフィックス

- `--mv-primitive-*`: 生の色値（Nord パレット + 拡張）
- `--mv-color-*`: セマンティックカラー（用途別）
- `--mv-glass-*`: ガラスモーフィズム用
- `--mv-spacing-*`, `--mv-font-*`, `--mv-shadow-*`: レイアウト・タイポグラフィ

詳細は `frontend/ide/src/design-system/CLAUDE.md` を参照。

## ビルドと実行

### パッケージマネージャー

このプロジェクトでは **pnpm** を使用します。npm や yarn は使用しないでください。

### 全体ビルド (IDE)

```bash
wails build
```

### コンポーネント別ビルド

- **AgentRunner Core**:
  ```bash
  go build ./cmd/agent-runner
  ```
- **Orchestrator CLI**:
  ```bash
  go build ./cmd/multiverse-orchestrator
  ```

### フロントエンド開発

```bash
cd frontend/ide
pnpm install
pnpm dev          # 開発サーバー起動
pnpm check        # Svelte 型チェック
pnpm lint         # ESLint (oxlint) チェック
pnpm lint:css     # Stylelint チェック
pnpm test:e2e     # Playwright E2E テスト
pnpm storybook    # Storybook 起動
```

## 開発の規約

- **言語**: コミュニケーションは常に **日本語** で行います。
- **ドキュメント**:
  - `docs/`: 仕様書と設計書。
  - `GEMINI.md`: 各ディレクトリの役割とコンテキスト（本ファイルおよびサブディレクトリ内のファイル）。
  - `CLAUDE.md`: AI アシスタント向けのガイドライン（共存）。

## ディレクトリ構成

- `cmd/`: 各コンポーネントのエントリポイント (`agent-runner`, `multiverse`, `multiverse-orchestrator`)
- `internal/`: 内部ロジック (`core`, `meta`, `worker`, `orchestrator`, `ide`, `chat`, `logging`)
- `frontend/`: IDE のフロントエンドコード (Svelte)
- `docs/`: プロジェクト全体のドキュメント

## 重要：作業前の確認事項

各ディレクトリの `CLAUDE.md` および `GEMINI.md` を必ず確認してください。
