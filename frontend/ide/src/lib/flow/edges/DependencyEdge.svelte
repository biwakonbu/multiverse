<script lang="ts">
  import { type EdgeProps } from "@xyflow/svelte";
  import { grid } from "../../../design-system";

  interface Props extends EdgeProps {
    data?: {
      satisfied?: boolean;
    };
  }

  let {
    id,
    sourceX,
    sourceY,
    targetX,
    targetY,
    sourcePosition,
    targetPosition,
    style = "",
    markerEnd,
    data,
  }: Props = $props();

  // Custom Circuit Path Logic ported from ConnectionLine.svelte
  function calculatePath(
    sx: number,
    sy: number,
    tx: number,
    ty: number
  ): string {
    const dist = tx - sx;

    if (dist > 40) {
      // 順方向：素直なS字サーキット
      const controlDist = Math.min(dist * 0.4, 60);
      return `M ${sx} ${sy} C ${sx + controlDist} ${sy}, ${tx - controlDist} ${ty}, ${tx} ${ty}`;
    } else {
      // 逆方向/近接：テクニカルな迂回
      const loopWidth = 40;
      const verticalBypass =
        ty > sy ? grid.cellHeight + 30 : -(grid.cellHeight + 30);

      return `M ${sx} ${sy}
              Q ${sx + loopWidth} ${sy} ${sx + loopWidth} ${sy + verticalBypass / 2}
              L ${sx + loopWidth} ${sy + verticalBypass}
              L ${tx - loopWidth} ${sy + verticalBypass}
              Q ${tx - loopWidth} ${ty} ${tx} ${ty}`;
    }
  }

  let edgePath = $derived(calculatePath(sourceX, sourceY, targetX, targetY));

  let strokeColor = $derived(
    data?.satisfied
      ? "var(--mv-color-status-succeeded-border)"
      : "var(--mv-color-status-blocked-border)"
  );

  let edgeClass = $derived(data?.satisfied ? "satisfied" : "unsatisfied");
  let markerEndId = $derived(
    `marker-terminal-${data?.satisfied ? "satisfied" : "unsatisfied"}`
  );
</script>

<g class="connection-line {edgeClass}">
  <!-- 背景パス（ヒット判定拡大用） -->
  <path
    class="path-hit"
    d={edgePath}
    fill="none"
    stroke="transparent"
    stroke-width="16"
  />

  <!-- メインパス -->
  <path
    class="path-main"
    d={edgePath}
    fill="none"
    stroke={strokeColor}
    marker-end="url(#{markerEndId})"
    marker-start="url(#marker-source)"
  />

  <!-- シグナルパルス（データフロー） -->
  {#if !data?.satisfied}
    <rect width="4" height="4" fill="var(--mv-primitive-aurora-purple)" rx="1">
      <animateMotion
        dur="1.5s"
        repeatCount="indefinite"
        path={edgePath}
        keyPoints="0;1"
        keyTimes="0;1"
        calcMode="linear"
      />
    </rect>
  {/if}
</g>

<style>
  .connection-line {
    pointer-events: none;
  }

  .path-hit {
    stroke-width: 20;
    cursor: pointer;
    pointer-events: stroke;
  }

  .path-main {
    stroke-width: 1.5;
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

  /* ホバー時 logic would rely on parent or JS state, 
     Svelte Flow selected state handles some, but hover might need 'interactive' handling
     or CSS on g:hover if pointer-events allowed. */
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
