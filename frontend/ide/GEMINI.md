# frontend/ide/GEMINI.md

このディレクトリは Multiverse IDE の Web フロントエンドを提供します。

## 技術スタック

- **Svelte 5**: リアクティブ UI フレームワーク
- **TypeScript 5**: 型安全な JavaScript
- **Vite 5**: 高速ビルドツール
- **Wails v2**: Go ↔ Web IPC
- **oxlint**: 高速リンター
- **Vitest**: ユニットテストフレームワーク
- **Storybook 8.6**: コンポーネントカタログ
- **Playwright**: E2E テスト

## パッケージマネージャー

**pnpm** を使用します。npm や yarn は使用しないでください。

## 開発コマンド

```bash
pnpm install          # 依存パッケージインストール
pnpm dev              # 開発サーバー起動
pnpm build            # 本番ビルド
pnpm check            # Svelte 型チェック
pnpm lint             # ESLint (oxlint) チェック
pnpm lint:css         # Stylelint チェック
pnpm check:all        # 全チェック（型 + lint + knip）
pnpm storybook        # Storybook 起動（http://localhost:6006）
pnpm test:e2e         # Playwright E2E テスト
```

## ディレクトリ構成

- **`src/`**: ソースコード
  - `design-system/`: デザイントークン・基底コンポーネント
  - `stores/`: Svelte Store（状態管理）
  - `types/`: TypeScript 型定義
  - `lib/`: UI コンポーネント
    - `grid/`: 2D 俯瞰グリッドビュー
    - `wbs/`: WBS リスト・グラフビュー
    - `toolbar/`: ツールバー・ズームコントロール
    - `backlog/`: バックログパネル
    - `brand/`: ブランドロゴ・テキスト
    - `welcome/`: ウェルカム画面
    - `components/`: 共有 UI（FloatingChatWindow など）
- **`wailsjs/`**: Wails 自動生成バインディング
- **`tests/`**: E2E テスト（Playwright）

## デザインシステム

### テーマ: Nord Deep

**コンセプト**: 深い背景にパステル UI が輝くデザイン

- **ベース**: Nord パレットをベースに深い背景色を拡張
- **UI**: Aurora パレットをパステル化したステータス色
- **グロー**: 控えめな効果（IDE としての実用性重視）

### スタイル: Glassmorphism

**Phantom Glass** スタイルを採用:

- **背景**: 半透明の暗いガラス効果（`--mv-glass-bg`）
- **ボーダー**: 微細な白いハイライト（`--mv-glass-border-subtle`）
- **シャドウ**: アンビエントシャドウとインナーハイライト

### UI コンセプト: Crystal HUD

ツールバー、パネル、チャットウィンドウなどに適用:

- **透明感**: 背景が透けて見える洗練されたパネル
- **グロー効果**: Frost 青系のアクセントグロー
- **フォント**: Orbitron（ブランド）、Rajdhani（ディスプレイ）

詳細は `src/design-system/CLAUDE.md` および `CLAUDE.md` を参照してください。
