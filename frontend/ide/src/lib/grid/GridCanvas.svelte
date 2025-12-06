<script lang="ts">
  import { get } from "svelte/store";
  import { onMount } from "svelte";
  import GridNode from "./GridNode.svelte";
  import ConnectionLine from "./ConnectionLine.svelte";
  import { viewport, drag, canvasTransform, zoomPercent } from "../../stores";
  import { taskNodes, gridBounds, taskEdges } from "../../stores";
  import { zoom as zoomConfig } from "../../design-system";

  let containerRef: HTMLDivElement;

  // ホイールズーム
  function handleWheel(event: WheelEvent) {
    event.preventDefault();

    if (event.ctrlKey || event.metaKey) {
      // Ctrl/Cmd + ホイールでズーム
      const rect = containerRef.getBoundingClientRect();
      const mouseX = event.clientX - rect.left;
      const mouseY = event.clientY - rect.top;
      viewport.wheelZoom(event.deltaY, mouseX, mouseY);
    } else {
      // 通常のホイールでパン
      viewport.setPan(
        $viewport.panX - event.deltaX,
        $viewport.panY - event.deltaY
      );
    }
  }

  // ドラッグ開始
  function handlePointerDown(event: PointerEvent) {
    // 中クリックまたはスペース+左クリックでパン開始
    if (event.button === 1 || (event.button === 0 && event.shiftKey)) {
      event.preventDefault();
      containerRef.setPointerCapture(event.pointerId);
      drag.startDrag(
        event.clientX,
        event.clientY,
        $viewport.panX,
        $viewport.panY
      );
    }
  }

  // ドラッグ中
  function handlePointerMove(event: PointerEvent) {
    if (!$drag.isDragging) return;

    const deltaX = event.clientX - $drag.startX;
    const deltaY = event.clientY - $drag.startY;

    viewport.setPan($drag.startPanX + deltaX, $drag.startPanY + deltaY);
  }

  // ドラッグ終了
  function handlePointerUp(event: PointerEvent) {
    if ($drag.isDragging) {
      containerRef.releasePointerCapture(event.pointerId);
      drag.endDrag();
    }
  }

  // キーボードショートカット
  function handleKeydown(event: KeyboardEvent) {
    // ズームショートカット
    if (event.key === "+" || event.key === "=") {
      event.preventDefault();
      viewport.zoomIn();
    } else if (event.key === "-") {
      event.preventDefault();
      viewport.zoomOut();
    } else if (event.key === "0") {
      event.preventDefault();
      viewport.reset();
    }
  }

  onMount(() => {
    const snapshot = {
      nodes: get(taskNodes).length,
      edges: get(taskEdges).length,
      panX: get(viewport).panX,
      panY: get(viewport).panY,
      zoom: get(viewport).zoom,
    };

    // #region agent log
    fetch("http://127.0.0.1:7242/ingest/e0c5926c-4256-4f95-83f1-ee92ab435f0c", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        sessionId: "debug-session",
        runId: "pre-fix",
        hypothesisId: "C",
        location: "GridCanvas.svelte:onMount",
        message: "grid canvas mounted",
        data: snapshot,
        timestamp: Date.now(),
      }),
    }).catch(() => {});
    // #endregion agent log

    // グローバルキーボードイベント
    window.addEventListener("keydown", handleKeydown);
    return () => {
      window.removeEventListener("keydown", handleKeydown);
    };
  });
</script>

<!-- svelte-ignore a11y-no-noninteractive-tabindex -->
<div
  class="canvas-container"
  bind:this={containerRef}
  on:wheel={handleWheel}
  on:pointerdown={handlePointerDown}
  on:pointermove={handlePointerMove}
  on:pointerup={handlePointerUp}
  on:pointercancel={handlePointerUp}
  role="application"
  aria-label="タスクグリッド"
  tabindex="0"
>
  <!-- グリッド背景パターン -->
  <div class="grid-background" style="transform: {$canvasTransform};">
    <svg class="grid-pattern" width="100%" height="100%">
      <defs>
        <pattern
          id="grid-cross"
          width="200"
          height="140"
          patternUnits="userSpaceOnUse"
        >
          <path
            d="M96 70H104M100 66V74"
            stroke="var(--mv-primitive-aurora-yellow)"
            stroke-width="1"
            opacity="0.15"
          />
        </pattern>
      </defs>
      <rect width="100%" height="100%" fill="url(#grid-cross)" />
    </svg>
  </div>

  <!-- 接続線レイヤー（ノードの下に表示） -->
  <svg class="connections-layer" style="transform: {$canvasTransform};">
    <defs>
      <!-- Technical Markers -->

      <!-- Source Port (Hollow Circle) -->
      <marker
        id="marker-source"
        markerWidth="8"
        markerHeight="8"
        refX="4"
        refY="4"
        orient="auto"
        markerUnits="userSpaceOnUse"
      >
        <circle
          cx="4"
          cy="4"
          r="2.5"
          fill="var(--mv-color-surface-app)"
          stroke="var(--mv-color-text-muted)"
          stroke-width="1"
        />
      </marker>

      <!-- Terminal: Satisfied (Solid Square) -->
      <marker
        id="marker-terminal-satisfied"
        markerWidth="10"
        markerHeight="10"
        refX="5"
        refY="5"
        orient="auto"
        markerUnits="userSpaceOnUse"
      >
        <rect
          x="2"
          y="2"
          width="6"
          height="6"
          fill="var(--mv-color-status-succeeded-border)"
        />
      </marker>

      <!-- Terminal: Unsatisfied (Solid Diamond) -->
      <marker
        id="marker-terminal-unsatisfied"
        markerWidth="12"
        markerHeight="12"
        refX="6"
        refY="6"
        orient="auto"
        markerUnits="userSpaceOnUse"
      >
        <path
          d="M6 1 L11 6 L6 11 L1 6 Z"
          fill="var(--mv-color-status-blocked-border)"
        />
      </marker>
    </defs>
    {#each $taskEdges as edge (`${edge.from}-${edge.to}`)}
      <ConnectionLine {edge} />
    {/each}
  </svg>

  <!-- ノードレイヤー -->
  <div class="nodes-layer" style="transform: {$canvasTransform};">
    {#each $taskNodes as node (node.task.id)}
      <GridNode
        task={node.task}
        col={node.col}
        row={node.row}
        zoomLevel={$viewport.zoom}
      />
    {/each}
  </div>

  <!-- ズームインジケーター -->
  <div class="zoom-indicator">
    {$zoomPercent}%
  </div>
</div>

<style>
  .canvas-container {
    position: relative;
    flex: 1;
    overflow: hidden;
    background: var(--mv-color-surface-app);
    cursor: grab;
    touch-action: none;
  }

  .canvas-container:active {
    cursor: grabbing;
  }

  .grid-background {
    position: absolute;
    inset: calc(var(--mv-canvas-overflow-size) * -1);
    width: calc(100% + var(--mv-canvas-overflow-size) * 2);
    height: calc(100% + var(--mv-canvas-overflow-size) * 2);
    transform-origin: 0 0;
    pointer-events: none;
  }

  .grid-pattern {
    width: 100%;
    height: 100%;
  }

  .connections-layer {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    transform-origin: 0 0;
    pointer-events: none;
    overflow: visible;
    padding: var(--mv-spacing-xl);
  }

  .nodes-layer {
    position: absolute;
    top: 0;
    left: 0;
    transform-origin: 0 0;
    padding: var(--mv-spacing-xl);
  }

  /* ズームインジケーター - ガラスモーフィズムスタイル */
  .zoom-indicator {
    position: absolute;
    bottom: var(--mv-spacing-md);
    right: var(--mv-spacing-md);

    /* Glass background */
    background: var(--mv-glass-bg-chat);
    backdrop-filter: blur(16px);

    /* Refined border */
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-strong);
    border-radius: var(--mv-radius-md);

    /* Styling */
    color: var(--mv-primitive-frost-1);
    padding: var(--mv-spacing-xs) var(--mv-spacing-sm);
    font-size: var(--mv-font-size-xs);
    font-family: var(--mv-font-display);
    font-weight: var(--mv-font-weight-semibold);
    letter-spacing: var(--mv-letter-spacing-wide);

    /* Shadow and glow */
    box-shadow: var(--mv-shadow-zoom-glow);
    text-shadow: var(--mv-text-shadow-zoom);

    pointer-events: none;
    transition: all var(--mv-duration-fast);
  }
</style>
