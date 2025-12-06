<script lang="ts">
  import { grid, gridToCanvas } from '../../design-system';
  import type { TaskEdge } from '../../stores/taskStore';
  import { taskNodes } from '../../stores';

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
    if (!from || !to) return '';

    const fromCanvas = gridToCanvas(from.col, from.row);
    const toCanvas = gridToCanvas(to.col, to.row);

    // ノードの中心から接続点を計算
    // 始点: ノードの右端中央
    const startX = fromCanvas.x + grid.cellWidth;
    const startY = fromCanvas.y + grid.cellHeight / 2;

    // 終点: ノードの左端中央
    const endX = toCanvas.x;
    const endY = toCanvas.y + grid.cellHeight / 2;

    // ベジェ曲線の制御点
    const midX = (startX + endX) / 2;
    const controlOffset = Math.min(Math.abs(endX - startX) / 2, 80);

    // 水平方向が十分にある場合はスムーズなカーブ
    // そうでない場合は迂回するパス
    if (endX > startX + 40) {
      // 通常のベジェ曲線
      return `M ${startX} ${startY} C ${startX + controlOffset} ${startY}, ${endX - controlOffset} ${endY}, ${endX} ${endY}`;
    } else {
      // 迂回パス（左から右へ戻る場合）
      const loopOffset = 60;
      const verticalOffset =
        endY > startY ? grid.cellHeight / 2 + 20 : -(grid.cellHeight / 2 + 20);

      return `M ${startX} ${startY}
              C ${startX + loopOffset} ${startY},
                ${startX + loopOffset} ${startY + verticalOffset},
                ${midX} ${startY + verticalOffset}
              C ${endX - loopOffset} ${startY + verticalOffset},
                ${endX - loopOffset} ${endY},
                ${endX} ${endY}`;
    }
  }

  // 線のスタイルクラス
  $: lineClass = edge.satisfied ? 'satisfied' : 'unsatisfied';
</script>

{#if pathData}
  <g class="connection-line {lineClass}">
    <!-- 背景パス（ホバー用） -->
    <path
      class="path-background"
      d={pathData}
      fill="none"
      stroke="transparent"
      stroke-width="12"
    />

    <!-- メインパス -->
    <path
      class="path-main"
      d={pathData}
      fill="none"
      marker-end="url(#arrowhead-{edge.satisfied ? 'satisfied' : 'unsatisfied'})"
    />
  </g>
{/if}

<style>
  .connection-line {
    pointer-events: stroke;
  }

  .path-main {
    stroke-width: 2;
    transition:
      stroke var(--mv-transition-hover),
      stroke-width var(--mv-transition-hover);
  }

  /* 満たされた依存 */
  .satisfied .path-main {
    stroke: var(--mv-color-status-succeeded-border);
    stroke-dasharray: none;
  }

  /* 未満の依存 */
  .unsatisfied .path-main {
    stroke: var(--mv-color-status-blocked-border);
    stroke-dasharray: 8 4;
    animation: dash-flow 1s linear infinite;
  }

  @keyframes dash-flow {
    to {
      stroke-dashoffset: -12;
    }
  }

  /* ホバー時 */
  .connection-line:hover .path-main {
    stroke-width: 3;
  }

  .satisfied:hover .path-main {
    stroke: var(--mv-color-status-succeeded-text);
  }

  .unsatisfied:hover .path-main {
    stroke: var(--mv-color-status-blocked-text);
  }
</style>
