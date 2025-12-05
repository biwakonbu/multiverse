/**
 * ビューポート状態管理ストア
 *
 * ズーム・パン状態をグローバルに管理
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

// ビューポートストア
function createViewportStore() {
  const { subscribe, set, update } = writable<Viewport>(initialViewport);

  return {
    subscribe,

    // ズームイン
    zoomIn: () => {
      update((v) => ({
        ...v,
        zoom: Math.min(v.zoom + zoomConfig.step, zoomConfig.max),
      }));
    },

    // ズームアウト
    zoomOut: () => {
      update((v) => ({
        ...v,
        zoom: Math.max(v.zoom - zoomConfig.step, zoomConfig.min),
      }));
    },

    // 特定のズームレベルに設定
    setZoom: (zoom: number) => {
      update((v) => ({
        ...v,
        zoom: Math.max(zoomConfig.min, Math.min(zoom, zoomConfig.max)),
      }));
    },

    // ホイールズーム（マウス位置を考慮）
    wheelZoom: (delta: number, mouseX: number, mouseY: number) => {
      update((v) => {
        const zoomDelta = -delta * zoomConfig.wheelFactor * 0.01;
        const newZoom = Math.max(
          zoomConfig.min,
          Math.min(v.zoom + zoomDelta, zoomConfig.max)
        );

        // マウス位置を中心にズーム
        const zoomRatio = newZoom / v.zoom;
        const newPanX = mouseX - (mouseX - v.panX) * zoomRatio;
        const newPanY = mouseY - (mouseY - v.panY) * zoomRatio;

        return {
          zoom: newZoom,
          panX: newPanX,
          panY: newPanY,
        };
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
export const canvasTransform = derived(viewport, ($viewport) => {
  return `scale(${$viewport.zoom}) translate(${$viewport.panX / $viewport.zoom}px, ${$viewport.panY / $viewport.zoom}px)`;
});

// ズームパーセント表示用
export const zoomPercent = derived(
  viewport,
  ($viewport) => Math.round($viewport.zoom * 100)
);
