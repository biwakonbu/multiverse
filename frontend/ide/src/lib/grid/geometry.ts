import { grid } from "../../design-system";
import type { Viewport } from "../../types/grid";

export interface GridPosition {
  col: number;
  row: number;
}

export interface Point {
  x: number;
  y: number;
}

export interface Rect extends Point {
  width: number;
  height: number;
}

/** グリッド座標（col,row）をワールド座標に変換 */
export function gridToWorld({ col, row }: GridPosition): Point {
  return {
    x: col * (grid.cellWidth + grid.gap),
    y: row * (grid.cellHeight + grid.gap),
  };
}

/** ワールド座標をスクリーン座標へ（ズーム・パン適用） */
export function worldToScreen({ x, y }: Point, viewport: Viewport): Point {
  return {
    x: x * viewport.zoom + viewport.panX,
    y: y * viewport.zoom + viewport.panY,
  };
}

/** グリッド座標をスクリーン座標へ（ズーム・パン適用） */
export function gridToScreen(pos: GridPosition, viewport: Viewport): Point {
  return worldToScreen(gridToWorld(pos), viewport);
}

/** ノードのワールド座標での矩形を取得 */
export function getNodeRectInWorld(pos: GridPosition): Rect {
  const topLeft = gridToWorld(pos);
  return {
    x: topLeft.x,
    y: topLeft.y,
    width: grid.cellWidth,
    height: grid.cellHeight,
  };
}

/** ノードのスクリーン座標での矩形を取得（ズーム・パン後） */
export function getNodeRectInScreen(
  pos: GridPosition,
  viewport: Viewport
): Rect {
  const rect = getNodeRectInWorld(pos);
  const topLeft = worldToScreen({ x: rect.x, y: rect.y }, viewport);
  return {
    x: topLeft.x,
    y: topLeft.y,
    width: rect.width * viewport.zoom,
    height: rect.height * viewport.zoom,
  };
}

/** エッジ両端（ワールド座標）: 右端中央→左端中央 */
export function getEdgeEndpointsInWorld(
  from: GridPosition,
  to: GridPosition
): { start: Point; end: Point } {
  const fromRect = getNodeRectInWorld(from);
  const toRect = getNodeRectInWorld(to);

  return {
    start: {
      x: fromRect.x + fromRect.width,
      y: fromRect.y + fromRect.height / 2,
    },
    end: {
      x: toRect.x,
      y: toRect.y + toRect.height / 2,
    },
  };
}

/** エッジ両端（スクリーン座標） */
export function getEdgeEndpointsInScreen(
  from: GridPosition,
  to: GridPosition,
  viewport: Viewport
): { start: Point; end: Point } {
  const endpoints = getEdgeEndpointsInWorld(from, to);
  return {
    start: worldToScreen(endpoints.start, viewport),
    end: worldToScreen(endpoints.end, viewport),
  };
}

/** ビューポートを矩形に適用した際のズーム適用幅 */
export function applyZoomToSize(
  size: number,
  viewport: Viewport
): number {
  return size * viewport.zoom;
}
