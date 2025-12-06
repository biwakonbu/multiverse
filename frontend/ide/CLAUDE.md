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
├── wailsjs/                 # Wails 自動生成バインディング（コミット対象）
│   └── go/
│       ├── main/
│       │   ├── App.js       # Go メソッドの JavaScript ラッパー
│       │   └── App.d.ts     # TypeScript 型定義
│       └── models.ts        # Go 構造体の TypeScript 型
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

### テーマ: Nord Deep

**コンセプト**: 深い背景にパステル UI が輝くデザイン

- **ベース**: Nord パレットを基に深い背景色を拡張
- **UI**: Aurora パレットをパステル化したステータス色
- **グロー**: 控えめな効果（IDE としての実用性重視）
- **スタイル**: ゲーム的な UI（Factorio 風 2D 俯瞰タイル配置）

### デザイントークン階層（3層構造）

**ハードコード禁止**: コンポーネントに生の色値（`#ff0000`, `rgba(...)`）を直接書かない。
全ての色は CSS 変数経由で指定し、トークン階層を通じて一元管理する。

```
1. プリミティブカラー    # 生の色値（Nord パレット + 拡張）
       ↓
2. セマンティックカラー  # 用途別（var() でプリミティブを参照）
       ↓
3. コンポーネント        # var(--mv-color-*) を使用
```

#### プリミティブカラー（`--mv-primitive-*`）

生の色値を定義。直接使用せず、セマンティック層から参照される。

| カテゴリ | 用途 | 変数例 |
|---------|------|--------|
| Polar Night | 深い背景（Nord） | `--mv-primitive-polar-night-0` |
| Snow Storm | 明るいテキスト | `--mv-primitive-snow-storm-0` |
| Frost | 青系アクセント | `--mv-primitive-frost-0` ~ `--mv-primitive-frost-3` |
| Aurora | ステータス色 | `--mv-primitive-aurora-*` |
| Deep | 深い背景（拡張） | `--mv-primitive-deep-0` ~ `--mv-primitive-deep-5` |
| Pastel | パステル化した色 | `--mv-primitive-pastel-*` |

#### セマンティックカラー（`--mv-color-*`）

用途で命名し、`var()` でプリミティブを参照:

```css
/* 例: セマンティック層の定義 */
--mv-color-surface-app: var(--mv-primitive-deep-0);
--mv-color-border-focus: var(--mv-primitive-frost-1);
--mv-color-status-running-text: var(--mv-primitive-pastel-green);
```

| カテゴリ | 変数例 | 用途 |
|---------|--------|------|
| `--mv-color-status-*` | `-running-bg`, `-failed-text` | タスクステータス表現 |
| `--mv-color-surface-*` | `-app`, `-node`, `-overlay` | 背景・パネル |
| `--mv-color-border-*` | `-default`, `-focus` | 境界線 |
| `--mv-color-text-*` | `-primary`, `-muted` | テキスト |
| `--mv-color-glow-*` | `-focus`, `-selected` | グロー効果 |
| `--mv-color-shadow-*` | `-elevated` | シャドウ |

### コンポーネントでの使用

CSS 変数のみを使用。フォールバック値は不要（変数は必ず定義される）。

```css
/* ✅ 正しい使用法 */
.node {
  background: var(--mv-color-surface-node);
  border-color: var(--mv-color-border-default);
  box-shadow: 0 0 0 3px var(--mv-color-glow-focus);
}

/* ❌ 禁止: ハードコード */
.node {
  background: #272b36;
  box-shadow: 0 0 0 3px rgba(136, 192, 208, 0.3);
}

/* ❌ 禁止: フォールバック値 */
.node {
  background: var(--mv-color-surface-node, #272b36);
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
| プリミティブ | `--mv-primitive-{palette}-{index}` | 生の色値 |
| ステータス色 | `--mv-color-status-{status}-{bg/border/text}` | 7種のタスク状態 |
| サーフェス | `--mv-color-surface-{app/primary/node/...}` | 背景色 |
| グロー | `--mv-color-glow-{focus/selected/error}` | グロー効果 |
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

## Wails バインディング（wailsjs/）

### コミット対象

`wailsjs/go/` 以下は **コミット対象**。Go バックエンドのメソッドを TypeScript から呼び出すための型定義・ラッパーを含む。

| ファイル | 内容 |
|---------|------|
| `go/main/App.js` | Go の `App` 構造体メソッドを呼び出す JavaScript 関数 |
| `go/main/App.d.ts` | 上記の TypeScript 型定義 |
| `go/models.ts` | Go 構造体（Task, Workspace 等）の TypeScript 型 |

### 生成タイミング

- `wails dev` または `wails build` 実行時に自動生成
- Go バックエンドのメソッドシグネチャ変更時に再生成される

### 使用例

```typescript
import { GetTasks, CreateTask } from '../wailsjs/go/main/App';
import { main } from '../wailsjs/go/models';

// Go メソッド呼び出し
const tasks: main.Task[] = await GetTasks(workspaceId);
await CreateTask(workspaceId, title, poolId);
```

### 注意事項

- `wailsjs/wailsjs/` のような二重ディレクトリが生成された場合は削除する（.gitignore で除外済み）
- `wailsjs/runtime/` は Wails が実行時に注入するため、コミット不要

## 設計原則

### デザインシステム駆動

- **ハードコード禁止**: 生の色値（`#rrggbb`, `rgba()`）をコンポーネントに書かない
- **3層トークン階層**: プリミティブ → セマンティック → CSS変数
- **セマンティック命名**: 色名でなく用途で命名（`surface-node`, `status-running`）
- **フォールバック不要**: CSS変数は必ず定義されているため `var(--x, fallback)` の fallback は書かない
- **新カラー追加時**: `variables.css` と `tokens/colors.ts` の両方を更新

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
