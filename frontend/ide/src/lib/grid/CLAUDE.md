# Grid - Factorio風タスクグリッドUI

## 責務

タスクを2D俯瞰ビューで表示するFactorio風グリッドUIを提供する。
100個以上のノードを一覧表示し、ズーム/パンで効率的にナビゲーション可能。

## コンポーネント構成

### メインコンポーネント（ストア依存）

| ファイル | 役割 |
|---------|------|
| GridCanvas.svelte | グリッド全体のキャンバス。ズーム/パン、ノードレイヤー管理 |
| GridNode.svelte | 個別タスクノード。ステータス表示、選択状態管理 |

### Storybook用プレビューコンポーネント

| ファイル | 役割 |
|---------|------|
| GridCanvasPreview.svelte | ストア非依存のキャンバス。props でデータを受け取る |
| GridNodePreview.svelte | ストア非依存のノード。単体テスト・Storybook 表示用 |
| StatusIndicator.svelte | ステータスを示すドット + ラベル。running 時パルスアニメーション |

## 設計思想

### なぜ2種類のコンポーネントがあるか

**本番用（GridCanvas, GridNode）**:
- Svelte ストア（`selectedTaskId`, `taskNodes` 等）に依存
- アプリケーション状態と密結合

**プレビュー用（*Preview）**:
- props のみで動作、ストア依存なし
- Storybook で単独表示・テスト可能
- デザインの確認・イテレーションに使用

### ズームレベルによる情報表示

```
zoom < 0.4:  ステータスドットのみ
zoom >= 0.4: タイトル表示
zoom >= 1.2: 詳細情報（poolId等）表示
```

### ステータスカラー体系

7種類のステータスに対応した色分け:

| ステータス | 背景色 | ボーダー | 意味 |
|-----------|--------|---------|------|
| PENDING | グレー | グレー | 待機中 |
| READY | 青系 | 青 | 実行可能 |
| RUNNING | 緑系 | 緑 + パルス | 実行中 |
| SUCCEEDED | 暗緑 | 緑 | 成功 |
| FAILED | 赤系 | 赤 | 失敗 |
| CANCELED | グレー | グレー | キャンセル |
| BLOCKED | 黄系 | 黄 | ブロック |

## 操作方法

- **スクロール**: パン（移動）
- **Ctrl/Cmd + スクロール**: ズーム
- **Shift + ドラッグ** / **中クリック + ドラッグ**: パン
- **+/-キー**: ズームイン/アウト
- **0キー**: ズームリセット
- **ノードクリック**: 選択/選択解除

## Storybook

```bash
pnpm storybook    # http://localhost:6006 で起動
```

ストーリー:
- Grid/StatusIndicator: ステータスインジケーターの各状態
- Grid/GridNode: ノードの各ステータス、ズームレベル変化
- Grid/GridCanvas: キャンバス全体、複数ノード配置、操作デモ

## CSS変数

グリッドコンポーネントは以下のCSS変数を使用:

```css
/* グリッドサイズ */
--mv-grid-cell-width: 160px;
--mv-grid-cell-height: 100px;
--mv-spacing-cell-gap: 40px;

/* ステータス色 */
--mv-color-status-{status}-bg
--mv-color-status-{status}-border
--mv-color-status-{status}-text

/* アニメーション */
--mv-animation-pulse-duration: 2000ms;
--mv-animation-glow-size: 8px;
--mv-animation-glow-color: rgba(68, 187, 68, 0.4);
```

## 拡張時の注意

- 新しいステータスを追加する場合は `types/task.ts` と CSS 両方を更新
- パフォーマンス考慮: 大量ノード（100+）では仮想化を検討
- アクセシビリティ: `role`, `aria-label`, `tabindex` を適切に設定
