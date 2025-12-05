# Frontend IDE - Svelte + TypeScript フロントエンド

このディレクトリは multiverse IDE の Web フロントエンドを提供します。

## 責務

- **UI レンダリング**: Svelte コンポーネントによる UI
- **バックエンド通信**: Wails IPC 経由で Go バックエンドを呼び出し
- **状態管理**: Svelte Store によるグローバル状態管理
- **デザインシステム**: 視覚的一貫性を保証するトークン管理

## UI方針: Factorio風タイル配置

従来のリスト形式ではなく、2D俯瞰のタイル配置UIを採用。
100個以上のAIエージェント/タスクを一瞥で把握できる設計。

詳細は `.claude/skills/factorio-ui/SKILL.md` を参照。

## ディレクトリ構成

```
frontend/ide/
├── src/
│   ├── main.ts              # エントリポイント
│   ├── app.css              # グローバルスタイル（CSS変数インポート）
│   ├── App.svelte           # ルートコンポーネント
│   │
│   ├── design-system/       # デザインシステム
│   │   ├── CLAUDE.md        # デザインシステム設計指針
│   │   ├── index.ts         # エクスポート集約
│   │   ├── variables.css    # CSS変数定義（162行）
│   │   └── tokens/          # デザイントークン
│   │       ├── colors.ts    # カラーパレット（7ステータス色）
│   │       ├── spacing.ts   # グリッド・スペーシング
│   │       ├── animation.ts # アニメーション・ズーム設定
│   │       └── typography.ts # タイポグラフィ
│   │   └── components/      # デザインシステムコンポーネント
│   │       ├── Button.svelte     # ボタン（4バリアント）
│   │       ├── Badge.svelte      # ステータスバッジ
│   │       ├── Card.svelte       # カード（3バリアント）
│   │       ├── Input.svelte      # テキスト入力
│   │       └── *.stories.ts      # Storybook ストーリー
│   │
│   ├── stores/              # Svelte Store
│   │   ├── taskStore.ts     # タスクデータ・選択状態
│   │   ├── viewportStore.ts # ズーム・パン状態
│   │   └── index.ts         # エクスポート集約
│   │
│   ├── types/               # 型定義
│   │   ├── task.ts          # Task, TaskStatus, TaskNode
│   │   ├── grid.ts          # Viewport, DragState, Point
│   │   └── index.ts         # エクスポート集約
│   │
│   └── lib/                 # UI コンポーネント
│       ├── WorkspaceSelector.svelte  # Workspace選択画面
│       ├── TaskCreate.svelte         # タスク作成フォーム
│       ├── TaskList.svelte           # 旧実装（非推奨）
│       ├── TaskDetail.svelte         # 旧実装（非推奨）
│       │
│       ├── grid/            # グリッドビュー
│       │   ├── CLAUDE.md             # グリッドコンポーネント設計指針
│       │   ├── GridCanvas.svelte     # ズーム・パン対応キャンバス
│       │   ├── GridCanvasPreview.svelte  # Storybook用（ストア非依存）
│       │   ├── GridNode.svelte       # タスクノード（ステータス色・アニメーション）
│       │   ├── GridNodePreview.svelte    # Storybook用（ストア非依存）
│       │   ├── StatusIndicator.svelte    # ステータスインジケーター
│       │   ├── *.stories.ts          # Storybook ストーリー
│       │   └── index.ts
│       │
│       ├── panel/           # パネル
│       │   ├── DetailPanel.svelte    # 選択タスクの詳細表示
│       │   └── index.ts
│       │
│       └── toolbar/         # ツールバー
│           ├── Toolbar.svelte        # ズームコントロール・ステータスサマリ
│           └── index.ts
│
├── wailsjs/                 # Wails 自動生成バインディング
│   └── go/main/App.js       # Go API の TypeScript ラッパー
├── index.html               # HTML テンプレート
├── package.json             # 依存パッケージ
├── vite.config.ts           # Vite ビルド設定
└── tsconfig.json            # TypeScript 設定
```

## コンポーネント階層

```
App.svelte
├── WorkspaceSelector.svelte      # Workspace 未選択時
└── (Workspace 選択後)
    ├── Toolbar.svelte            # 上部ツールバー
    ├── GridCanvas.svelte         # メイン：2Dグリッドビュー
    │   └── GridNode.svelte       # 各タスクノード
    ├── DetailPanel.svelte        # 右サイド：詳細パネル
    └── TaskCreate.svelte         # モーダル：タスク作成
```

## 主要コンポーネント

### GridCanvas.svelte

2Dグリッドキャンバス。ズーム・パン操作を提供。

**機能**:
- ホイールズーム（Ctrl/Cmd + スクロール）
- パン操作（中クリック or Shift + 左クリック + ドラッグ）
- キーボードショートカット（+/-/0）

**使用ストア**: `viewport`, `taskNodes`

### GridNode.svelte

タスクノードの視覚表現。

**Props**:
- `task`: Task オブジェクト
- `col`, `row`: グリッド位置
- `zoomLevel`: 現在のズームレベル

**特徴**:
- ステータス別の色分け（CSS変数経由）
- 実行中はパルスアニメーション
- ズームレベルに応じた情報量調整

### Toolbar.svelte

上部ツールバー。

**機能**:
- アプリタイトル
- 新規タスク作成ボタン
- ステータスサマリ（実行中/待機/失敗のカウント）
- ズームコントロール（+/-/リセット）

### DetailPanel.svelte

選択タスクの詳細表示。

**機能**:
- タスク情報表示（タイトル、ステータス、Pool、日時）
- タスク実行ボタン
- スライドイン/アウトアニメーション

## 状態管理

### viewportStore

```typescript
import { viewport, zoomPercent, canvasTransform } from './stores';

// ズーム操作
viewport.zoomIn();
viewport.zoomOut();
viewport.setZoom(1.5);
viewport.wheelZoom(delta, mouseX, mouseY);

// パン操作
viewport.setPan(x, y);

// リセット
viewport.reset();
```

### taskStore

```typescript
import { tasks, selectedTaskId, selectedTask, taskNodes } from './stores';

// タスク管理
tasks.setTasks(taskList);
tasks.addTask(task);
tasks.updateTask(id, updates);

// 選択管理
selectedTaskId.select(taskId);
selectedTaskId.clear();

// 派生ストア
$taskNodes      // グリッド配置されたTaskNode[]
$selectedTask   // 選択中のTask | null
```

## デザインシステム

### CSS変数の使用

全コンポーネントでCSS変数を使用:

```css
.node {
  width: var(--mv-grid-cell-width);
  height: var(--mv-grid-cell-height);
  background: var(--mv-color-surface-node);
  border-radius: var(--mv-radius-md);
  transition: var(--mv-transition-hover);
}

.node.status-running {
  background: var(--mv-color-status-running-bg);
  border-color: var(--mv-color-status-running-border);
  animation: mv-pulse var(--mv-duration-pulse) infinite;
}
```

### TypeScriptトークン

```typescript
import { colors, spacing, gridToCanvas, zoom } from './design-system';

// グリッド座標変換
const { x, y } = gridToCanvas(col, row);

// ズーム設定
zoom.min    // 0.25
zoom.max    // 3.0
zoom.default // 1.0
```

### CSS変数一覧

| カテゴリ | 変数例 | 用途 |
|---------|--------|------|
| ステータス色 | `--mv-color-status-{status}-{bg/border/text}` | 7種のタスク状態 |
| サーフェス | `--mv-color-surface-{app/primary/node/...}` | 背景色 |
| スペーシング | `--mv-spacing-{xxs~xxl}` | 4px基準スケール |
| グリッド | `--mv-grid-{cell-width/cell-height/gap}` | ノードサイズ |
| タイポグラフィ | `--mv-font-{size/weight}-*` | テキストスタイル |
| アニメーション | `--mv-duration-{fast/normal/pulse}` | タイミング |

## 開発コマンド

```bash
# 依存パッケージインストール（pnpm を使用）
pnpm install

# 開発サーバー起動（Wails dev と連携）
pnpm dev

# 本番ビルド
pnpm build

# 型チェック
pnpm check

# 全チェック（型 + lint + knip）
pnpm check:all

# Storybook 起動（http://localhost:6006）
pnpm storybook

# Storybook ビルド
pnpm build-storybook
```

## Storybook

コンポーネントカタログとして Storybook 8 を導入。

### 起動方法

```bash
pnpm storybook    # http://localhost:6006 で起動
```

### コンポーネント構成

| カテゴリ | コンポーネント | 説明 |
|---------|---------------|------|
| Design System | Button | 4バリアント（primary/secondary/ghost/danger） |
| Design System | Badge | 7ステータス対応、パルスアニメーション |
| Design System | Card | 3バリアント（default/elevated/outlined） |
| Design System | Input | テキスト入力、エラー状態対応 |
| Grid | StatusIndicator | ステータスドット + ラベル |
| Grid | GridNode | タスクノード、ズームレベル対応 |
| Grid | GridCanvas | 2D俯瞰キャンバス、ズーム/パン操作 |

### Storybook用プレビューコンポーネント

本番コンポーネント（GridCanvas, GridNode）はSvelteストアに依存するため、
Storybook用にストア非依存のプレビューコンポーネントを用意:

- `GridCanvasPreview.svelte` - props でノードデータを受け取る
- `GridNodePreview.svelte` - props でタスク情報を受け取る

これにより単体でのデザイン確認・イテレーションが可能。

## 技術スタック

- **Svelte 4**: リアクティブ UI フレームワーク
- **TypeScript 5**: 型安全な JavaScript
- **Vite 5**: 高速ビルドツール
- **Wails v2**: Go ↔ Web IPC
- **oxlint**: 高速リンター

## 設計原則

### デザインシステム駆動

- 全スタイルはCSS変数経由
- 生の色値・サイズ値を直接書かない
- デザイントークンで一貫性を保証

### Store ベースの状態管理

- `svelte/store` でグローバル状態を管理
- 派生ストア（derived）で計算済み値を提供
- コンポーネント間の結合度を下げる

### ポーリングによる状態同期

- WebSocket 等のリアルタイム通信は未実装
- 2 秒間隔のポーリングで実用的な更新頻度を確保

## 拡張予定

### 短期

- [ ] タスク削除ボタン
- [ ] ローディング表示
- [ ] エラー表示改善

### 長期

- [ ] ログビューア
- [ ] Worker Pool 設定 UI
- [ ] ConnectionLine（タスク間依存関係の矢印）
- [ ] テーマ切り替え（ダーク/ライト）

## 関連ドキュメント

- [../../cmd/multiverse-ide/CLAUDE.md](../../cmd/multiverse-ide/CLAUDE.md): Go バックエンド
- [../../internal/orchestrator/CLAUDE.md](../../internal/orchestrator/CLAUDE.md): Task データモデル
- [src/design-system/CLAUDE.md](src/design-system/CLAUDE.md): デザインシステム詳細
- [src/lib/grid/CLAUDE.md](src/lib/grid/CLAUDE.md): グリッドコンポーネント詳細
- [../../wails.json](../../wails.json): Wails プロジェクト設定
