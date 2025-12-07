/**
 * ビューポート状態管理ストア
 *
 * ズーム・パン状態をグローバルに管理
 * 
 * 座標変換の説明:
 * - レンダリング: style="zoom: {zoom}; transform: translate({panX / zoom}px, {panY / zoom}px);"
 * - スクリーン座標 = ワールド座標 * zoom + panX
 * - ワールド座標 = (スクリーン座標 - panX) / zoom
 * 
 * マウス位置を起点にズームするには:
 * 1. マウス位置のワールド座標を計算
 * 2. ズーム後、同じワールド座標が同じスクリーン座標に留まるようにpanを調整
 */

import { writable, derived } from 'svelte/store';
import type { Viewport, DragState } from '../types';
import { zoom as zoomConfig } from '../design-system';

// 初期ビューポート状態
const initialViewport: Viewport = {
  zoom: zoomConfig.default,
  panX: 0,
  panY: 0,
};

// 初期ドラッグ状態
const initialDragState: DragState = {
  isDragging: false,
  startX: 0,
  startY: 0,
  startPanX: 0,
  startPanY: 0,
};

/**
 * マウス位置を起点にズームするためのパン調整を計算
 * 
 * @param mouseX スクリーン座標のマウスX位置
 * @param mouseY スクリーン座標のマウスY位置
 * @param oldZoom 変更前のズームレベル
 * @param newZoom 変更後のズームレベル
 * @param oldPanX 変更前のパンX
 * @param oldPanY 変更前のパンY
 * @returns 新しいパン位置
 */
function calculateZoomPan(
  mouseX: number,
  mouseY: number,
  oldZoom: number,
  newZoom: number,
  oldPanX: number,
  oldPanY: number
): { panX: number; panY: number } {
  // マウス位置のワールド座標を計算 (スクリーン座標からワールド座標へ)
  // スクリーン座標 = ワールド座標 * zoom + panX
  // ワールド座標 = (スクリーン座標 - panX) / zoom
  const worldX = (mouseX - oldPanX) / oldZoom;
  const worldY = (mouseY - oldPanY) / oldZoom;
  
  // ズーム後、同じワールド座標が同じスクリーン座標に留まるようにpanを計算
  // mouseX = worldX * newZoom + newPanX
  // newPanX = mouseX - worldX * newZoom
  const newPanX = mouseX - worldX * newZoom;
  const newPanY = mouseY - worldY * newZoom;
  
  console.log('[Zoom Debug]', {
    mouse: { x: mouseX, y: mouseY },
    world: { x: worldX, y: worldY },
    oldZoom,
    newZoom,
    oldPan: { x: oldPanX, y: oldPanY },
    newPan: { x: newPanX, y: newPanY },
  });
  
  return { panX: newPanX, panY: newPanY };
}

// ビューポートストア
function createViewportStore() {
  const { subscribe, set, update } = writable<Viewport>(initialViewport);

  return {
    subscribe,

    // ズームイン（オプションでマウス位置を指定）
    zoomIn: (mouseX?: number, mouseY?: number) => {
      update((v) => {
        const newZoom = Math.min(v.zoom + zoomConfig.step, zoomConfig.max);
        
        // マウス位置が指定されている場合はそこを起点にズーム
        if (mouseX !== undefined && mouseY !== undefined) {
          const { panX, panY } = calculateZoomPan(
            mouseX, mouseY, v.zoom, newZoom, v.panX, v.panY
          );
          return { zoom: newZoom, panX, panY };
        }
        
        return { ...v, zoom: newZoom };
      });
    },

    // ズームアウト（オプションでマウス位置を指定）
    zoomOut: (mouseX?: number, mouseY?: number) => {
      update((v) => {
        const newZoom = Math.max(v.zoom - zoomConfig.step, zoomConfig.min);
        
        // マウス位置が指定されている場合はそこを起点にズーム
        if (mouseX !== undefined && mouseY !== undefined) {
          const { panX, panY } = calculateZoomPan(
            mouseX, mouseY, v.zoom, newZoom, v.panX, v.panY
          );
          return { zoom: newZoom, panX, panY };
        }
        
        return { ...v, zoom: newZoom };
      });
    },

    // 特定のズームレベルに設定（マウス位置起点オプション付き）
    setZoom: (zoom: number, mouseX?: number, mouseY?: number) => {
      update((v) => {
        const newZoom = Math.max(zoomConfig.min, Math.min(zoom, zoomConfig.max));
        
        if (mouseX !== undefined && mouseY !== undefined) {
          const { panX, panY } = calculateZoomPan(
            mouseX, mouseY, v.zoom, newZoom, v.panX, v.panY
          );
          return { zoom: newZoom, panX, panY };
        }
        
        return { ...v, zoom: newZoom };
      });
    },

    // ホイールズーム（マウス位置を考慮）
    wheelZoom: (delta: number, mouseX: number, mouseY: number) => {
      update((v) => {
        const zoomDelta = -delta * zoomConfig.wheelFactor * 0.01;
        const newZoom = Math.max(
          zoomConfig.min,
          Math.min(v.zoom + zoomDelta, zoomConfig.max)
        );

        const { panX, panY } = calculateZoomPan(
          mouseX, mouseY, v.zoom, newZoom, v.panX, v.panY
        );

        return { zoom: newZoom, panX, panY };
      });
    },

    // パン位置を設定
    setPan: (x: number, y: number) => {
      update((v) => ({
        ...v,
        panX: x,
        panY: y,
      }));
    },

    // リセット
    reset: () => set(initialViewport),
  };
}

// ドラッグ状態ストア
function createDragStore() {
  const { subscribe, set } = writable<DragState>(initialDragState);

  return {
    subscribe,

    // ドラッグ開始
    startDrag: (x: number, y: number, panX: number, panY: number) => {
      set({
        isDragging: true,
        startX: x,
        startY: y,
        startPanX: panX,
        startPanY: panY,
      });
    },

    // ドラッグ終了
    endDrag: () => {
      set(initialDragState);
    },
  };
}

export const viewport = createViewportStore();
export const drag = createDragStore();

// CSS Transform用の派生ストア
// 座標変換: スクリーン座標 = ワールド座標 * zoom + pan
// transform の適用順序は右から左なので、scale → translate の順で適用される
// translate を先に書くことで pan がズームの影響を受けないようにする
export const canvasTransform = derived(viewport, ($viewport) => {
  return `translate3d(${$viewport.panX}px, ${$viewport.panY}px, 0) scale(${$viewport.zoom})`;
});

// ズームパーセント表示用
export const zoomPercent = derived(
  viewport,
  ($viewport) => Math.round($viewport.zoom * 100)
);
