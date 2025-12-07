<script lang="ts">
  import { grid } from "../../design-system";
  import type { TaskEdge } from "../../stores/taskStore";
  import { taskNodes } from "../../stores";
  import { getEdgeEndpointsInWorld } from "./geometry";

  // Props
  export let edge: TaskEdge = { from: "", to: "", satisfied: false };

  // ノード位置のマップを取得
  $: nodePositions = new Map(
    $taskNodes.map((n) => [n.task.id, { col: n.col, row: n.row }])
  );

  // 始点と終点の座標を計算
  $: fromPos = nodePositions.get(edge.from);
  $: toPos = nodePositions.get(edge.to);

  // SVGパスを計算
  $: pathData = calculatePath(fromPos, toPos);

  function calculatePath(
    from: { col: number; row: number } | undefined,
    to: { col: number; row: number } | undefined
  ): string {
    if (!from || !to) return "";

    const { start, end } = getEdgeEndpointsInWorld(from, to);
    const startX = start.x;
    const startY = start.y;
    const endX = end.x;
    const endY = end.y;

    // 水平距離
    const dist = endX - startX;

    // Circuit Style: Tight Bezier for "Wire" look
    // 制御点を近くして、直線部分を長くする

    if (dist > 40) {
      // 順方向：素直なS字サーキット
      const controlDist = Math.min(dist * 0.4, 60);
      return `M ${startX} ${startY} C ${startX + controlDist} ${startY}, ${endX - controlDist} ${endY}, ${endX} ${endY}`;
    } else {
      // 逆方向/近接：テクニカルな迂回
      const loopWidth = 40;
      const verticalBypass =
        endY > startY ? grid.cellHeight + 30 : -(grid.cellHeight + 30);
      const midX = (startX + endX) / 2;

      // 直線的な動きを意識したカーブ
      return `M ${startX} ${startY}
              Q ${startX + loopWidth} ${startY} ${startX + loopWidth} ${startY + verticalBypass / 2}
              L ${startX + loopWidth} ${startY + verticalBypass}
              L ${endX - loopWidth} ${startY + verticalBypass}
              Q ${endX - loopWidth} ${endY} ${endX} ${endY}`;
    }
  }

  // 線のスタイルクラス
  $: lineClass = edge.satisfied ? "satisfied" : "unsatisfied";
  $: strokeColor = edge.satisfied
    ? "var(--mv-color-status-succeeded-border)"
    : "var(--mv-color-status-blocked-border)";
</script>

{#if pathData}
  <g class="connection-line {lineClass}">
    <!-- 背景パス（ヒット判定拡大用） -->
    <path
      class="path-hit"
      d={pathData}
      fill="none"
      stroke="transparent"
      stroke-width="16"
    />

    <!-- メインパス（細く、鋭く） -->
    <path
      class="path-main"
      d={pathData}
      fill="none"
      stroke={strokeColor}
      marker-end="url(#marker-terminal-{edge.satisfied
        ? 'satisfied'
        : 'unsatisfied'})"
      marker-start="url(#marker-source)"
    />

    <!-- シグナルパルス（データフロー） -->
    {#if !edge.satisfied}
      <rect
        width="4"
        height="4"
        fill="var(--mv-primitive-aurora-purple)"
        rx="1"
      >
        <animateMotion
          dur="1.5s"
          repeatCount="indefinite"
          path={pathData}
          keyPoints="0;1"
          keyTimes="0;1"
          calcMode="linear"
        />
      </rect>
    {/if}
  </g>
{/if}

<style>
  .connection-line {
    pointer-events: stroke;
    transition: opacity var(--mv-duration-normal);
  }

  .path-main {
    stroke-width: 1.5; /* Technical thin line */
    transition:
      stroke var(--mv-duration-normal),
      stroke-width var(--mv-duration-fast);
    vector-effect: non-scaling-stroke;
  }

  /* 満たされた依存 */
  .satisfied .path-main {
    stroke-opacity: 0.5;
  }

  /* 未満の依存（アクティブ） */
  .unsatisfied .path-main {
    stroke-opacity: 0.9;
    stroke-dasharray: 2 4;
    stroke-linecap: square;
  }

  /* ホバー時 */
  .connection-line:hover .path-main {
    stroke-width: 2.5;
    stroke-opacity: 1;
    stroke-dasharray: none;
  }

  .satisfied:hover .path-main {
    stroke: var(--mv-color-status-succeeded-text);
  }

  .unsatisfied:hover .path-main {
    stroke: var(--mv-color-status-blocked-text);
  }
</style>
