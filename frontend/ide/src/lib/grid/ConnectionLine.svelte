<script lang="ts">
  import { grid, gridToCanvas } from "../../design-system";
  import type { TaskEdge } from "../../stores/taskStore";
  import { taskNodes } from "../../stores";

  // Props
  export let edge: TaskEdge;

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
    if (!from || !to) {
      // #region agent log
      fetch(
        "http://127.0.0.1:7242/ingest/e0c5926c-4256-4f95-83f1-ee92ab435f0c",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            sessionId: "debug-session",
            runId: "pre-fix",
            hypothesisId: "B",
            location: "ConnectionLine.svelte:24",
            message: "edge missing node position",
            data: {
              edge,
              hasFrom: Boolean(from),
              hasTo: Boolean(to),
            },
            timestamp: Date.now(),
          }),
        }
      ).catch(() => {});
      // #endregion agent log
      return "";
    }

    const fromCanvas = gridToCanvas(from.col, from.row);
    const toCanvas = gridToCanvas(to.col, to.row);

    // テクニカルな接続点計算
    // 始点: ノードの右端中央
    const startX = fromCanvas.x + grid.cellWidth;
    const startY = fromCanvas.y + grid.cellHeight / 2;

    // 終点: ノードの左端中央
    const endX = toCanvas.x;
    const endY = toCanvas.y + grid.cellHeight / 2;

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
    stroke-width: 1.5;
    transition:
      stroke var(--mv-duration-normal),
      stroke-width var(--mv-duration-fast),
      filter var(--mv-duration-fast);
    vector-effect: non-scaling-stroke;
  }

  /* 満たされた依存 - より洗練されたグロー */
  .satisfied .path-main {
    stroke-opacity: 0.6;
    filter: drop-shadow(0 0 2px var(--mv-color-status-succeeded-border));
  }

  /* 未満の依存（アクティブ）- ダッシュラインとグロー */
  .unsatisfied .path-main {
    stroke-opacity: 0.85;
    stroke-dasharray: 3 5;
    stroke-linecap: round;
    filter: drop-shadow(0 0 3px var(--mv-color-status-blocked-border));
  }

  /* ホバー時 - 強調されたグロー */
  .connection-line:hover .path-main {
    stroke-width: 2.5;
    stroke-opacity: 1;
    stroke-dasharray: none;
  }

  .satisfied:hover .path-main {
    stroke: var(--mv-color-status-succeeded-text);
    filter: drop-shadow(0 0 4px var(--mv-color-status-succeeded-text))
      drop-shadow(0 0 8px var(--mv-glow-frost-2-border));
  }

  .unsatisfied:hover .path-main {
    stroke: var(--mv-color-status-blocked-text);
    filter: drop-shadow(0 0 4px var(--mv-color-status-blocked-text))
      drop-shadow(0 0 8px rgba(180, 142, 173, 0.4));
  }
</style>
