/**
 * グリッドビュー関連の型定義
 */

// ビューポート状態
export interface Viewport {
  // ズームレベル（0.25 ~ 3.0）
  zoom: number;
  // パン位置（キャンバス座標）
  panX: number;
  panY: number;
}

// ドラッグ状態
export interface DragState {
  isDragging: boolean;
  startX: number;
  startY: number;
  startPanX: number;
  startPanY: number;
}
