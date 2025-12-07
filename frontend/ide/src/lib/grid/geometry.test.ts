import { describe, it, expect } from "vitest";
import { grid } from "../../design-system";
import type { Viewport } from "../../types/grid";
import {
  applyZoomToSize,
  getEdgeEndpointsInScreen,
  getEdgeEndpointsInWorld,
  getNodeRectInScreen,
  getNodeRectInWorld,
  gridToScreen,
  gridToWorld,
  worldToScreen,
} from "./geometry";

const viewport: Viewport = { zoom: 2, panX: 10, panY: -20 };

describe("geometry helpers", () => {
  it("converts grid to world coordinates", () => {
    const p = gridToWorld({ col: 1, row: 2 });
    expect(p).toEqual({
      x: 1 * (grid.cellWidth + grid.gap),
      y: 2 * (grid.cellHeight + grid.gap),
    });
  });

  it("converts world to screen with zoom and pan", () => {
    const p = worldToScreen({ x: 100, y: 50 }, viewport);
    expect(p).toEqual({
      x: 100 * viewport.zoom + viewport.panX,
      y: 50 * viewport.zoom + viewport.panY,
    });
  });

  it("returns node rect in world and screen space", () => {
    const world = getNodeRectInWorld({ col: 0, row: 0 });
    expect(world).toEqual({
      x: 0,
      y: 0,
      width: grid.cellWidth,
      height: grid.cellHeight,
    });

    const screen = getNodeRectInScreen({ col: 0, row: 0 }, viewport);
    expect(screen).toEqual({
      x: viewport.panX,
      y: viewport.panY,
      width: grid.cellWidth * viewport.zoom,
      height: grid.cellHeight * viewport.zoom,
    });
  });

  it("computes edge endpoints consistently (world and screen)", () => {
    const from = { col: 0, row: 0 };
    const to = { col: 1, row: 0 };

    const world = getEdgeEndpointsInWorld(from, to);
    expect(world.start.x).toBe(grid.cellWidth);
    expect(world.start.y).toBe(grid.cellHeight / 2);
    expect(world.end.x).toBe(grid.cellWidth + grid.gap);
    expect(world.end.y).toBe(grid.cellHeight / 2);

    const screen = getEdgeEndpointsInScreen(from, to, viewport);
    const startScreen = worldToScreen(world.start, viewport);
    const endScreen = worldToScreen(world.end, viewport);
    expect(screen.start).toEqual(startScreen);
    expect(screen.end).toEqual(endScreen);
  });

  it("gridToScreen matches worldToScreen(gridToWorld)", () => {
    const gridPos = { col: 2, row: 3 };
    expect(gridToScreen(gridPos, viewport)).toEqual(
      worldToScreen(gridToWorld(gridPos), viewport)
    );
  });

  it("scales size by zoom factor", () => {
    expect(applyZoomToSize(100, viewport)).toBe(100 * viewport.zoom);
  });
});
