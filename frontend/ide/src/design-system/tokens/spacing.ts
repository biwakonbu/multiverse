/**
 * スペーシングトークン定義
 *
 * グリッドシステムとUIスペーシングを一元管理。
 */

// グリッドシステム - タイル配置の基盤
export const grid = {
  // ノードサイズ
  cellWidth: 160,
  cellHeight: 100,
  // ノード間の余白
  gap: 40,
} as const;

// UIスペーシング - 4pxベースのスケール
export const spacing = {
  xxs: 4,
  xs: 8,
  sm: 12,
  md: 16,
  lg: 24,
  xl: 32,
  xxl: 48,
} as const;

// レイアウト固定サイズ
export const layout = {
  // ツールバー高さ
  toolbarHeight: 56,
  // サイドパネル幅
  detailPanelWidth: 360,
  // ミニマップサイズ
  minimapWidth: 150,
  minimapHeight: 100,
} as const;

// ノード内部のパディング
export const nodePadding = {
  x: 12,
  y: 10,
} as const;

// ボーダー半径
export const borderRadius = {
  sm: 4,
  md: 8,
  lg: 12,
} as const;

// 座標変換ヘルパー
export function gridToCanvas(col: number, row: number) {
  return {
    x: col * (grid.cellWidth + grid.gap),
    y: row * (grid.cellHeight + grid.gap),
  };
}

export function canvasToGrid(x: number, y: number) {
  return {
    col: Math.round(x / (grid.cellWidth + grid.gap)),
    row: Math.round(y / (grid.cellHeight + grid.gap)),
  };
}

export const spacingTokens = {
  grid,
  spacing,
  layout,
  nodePadding,
  borderRadius,
} as const;
