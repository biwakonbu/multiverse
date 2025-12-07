<script lang="ts">
  import { createBubbler, preventDefault } from 'svelte/legacy';

  const bubble = createBubbler();
  import { onMount } from "svelte";
  import GridNode from "./GridNode.svelte";
  import ConnectionLine from "./ConnectionLine.svelte";
  import { viewport, drag, canvasTransform, zoomPercent } from "../../stores";
  import { taskNodes, gridBounds, taskEdges } from "../../stores";
  import { zoom as zoomConfig } from "../../design-system";

  let containerRef: HTMLDivElement | null = $state(null);

  // マウス位置追跡（ズームの起点として使用）
  let lastMousePosition = { x: 0, y: 0 };

  const DRAG_CLASS = "mv-canvas-dragging";
  let isPanning = false;
  let isCanvasPointerDown = false;
  let prevBodyUserSelect: string | null = null;
  let prevHtmlUserSelect: string | null = null;

  const preventSelection = (event: Event) => {
    if (isPanning || isCanvasPointerDown) {
      event.preventDefault();
    }
  };

  function enableGlobalNoSelect() {
    if (isCanvasPointerDown) return;
    prevBodyUserSelect = document.body.style.userSelect;
    prevHtmlUserSelect = document.documentElement.style.userSelect;
    document.body.style.userSelect = "none";
    document.documentElement.style.userSelect = "none";
    document.body.classList.add(DRAG_CLASS);
    document.addEventListener("selectionstart", preventSelection, true);
    document.addEventListener("dragstart", preventSelection, true);
    window.getSelection()?.removeAllRanges();
  }

  function disableGlobalNoSelect() {
    document.body.classList.remove(DRAG_CLASS);
    document.removeEventListener("selectionstart", preventSelection, true);
    document.removeEventListener("dragstart", preventSelection, true);
    if (prevBodyUserSelect !== null) {
      document.body.style.userSelect = prevBodyUserSelect;
      prevBodyUserSelect = null;
    } else {
      document.body.style.removeProperty("user-select");
    }
    if (prevHtmlUserSelect !== null) {
      document.documentElement.style.userSelect = prevHtmlUserSelect;
      prevHtmlUserSelect = null;
    } else {
      document.documentElement.style.removeProperty("user-select");
    }
  }

  function handleCanvasPointerDown(event: PointerEvent) {
    // 右クリック以外の押下で選択抑止を有効化（ドラッグ開始想定）
    if (event.button !== 2) {
      isCanvasPointerDown = true;
      enableGlobalNoSelect();
    }
  }

  function handleCanvasPointerUp() {
    if (isCanvasPointerDown && !isPanning) {
      isCanvasPointerDown = false;
      disableGlobalNoSelect();
    }
  }

  // マウス移動時に位置を記録
  function handleMouseMove(event: MouseEvent) {
    if (!containerRef) return;
    const rect = containerRef.getBoundingClientRect();
    lastMousePosition = {
      x: event.clientX - rect.left,
      y: event.clientY - rect.top,
    };
  }

  // ホイールズーム
  function handleWheel(event: WheelEvent) {
    event.preventDefault();

    if (event.ctrlKey || event.metaKey) {
      // Ctrl/Cmd + ホイールでズーム
      if (!containerRef) return;
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
    handleCanvasPointerDown(event);
    // 中クリックまたはスペース+左クリックでパン開始
    if (event.button === 1 || (event.button === 0 && event.shiftKey)) {
      event.preventDefault();
      if (!containerRef) return;
      containerRef.setPointerCapture(event.pointerId);
      isPanning = true;
      enableGlobalNoSelect();
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
    // マウス位置を常に追跡
    handleMouseMove(event);

    if (!$drag.isDragging) return;

    const deltaX = event.clientX - $drag.startX;
    const deltaY = event.clientY - $drag.startY;

    viewport.setPan($drag.startPanX + deltaX, $drag.startPanY + deltaY);
  }

  // ドラッグ終了
  function handlePointerUp(event: PointerEvent) {
    if ($drag.isDragging) {
      containerRef?.releasePointerCapture(event.pointerId);
      drag.endDrag();
      isPanning = false;
      disableGlobalNoSelect();
    } else {
      handleCanvasPointerUp();
    }
  }

  // キーボードショートカット
  function handleKeydown(event: KeyboardEvent) {
    // ズームショートカット（マウス位置を起点に）
    if (event.key === "+" || event.key === "=") {
      event.preventDefault();
      viewport.zoomIn(lastMousePosition.x, lastMousePosition.y);
    } else if (event.key === "-") {
      event.preventDefault();
      viewport.zoomOut(lastMousePosition.x, lastMousePosition.y);
    } else if (event.key === "0") {
      event.preventDefault();
      viewport.reset();
    }
  }

  onMount(() => {
    // グローバルキーボードイベント
    window.addEventListener("keydown", handleKeydown);
    window.addEventListener("pointerup", handleCanvasPointerUp, true);
    window.addEventListener("pointercancel", handleCanvasPointerUp, true);
    return () => {
      window.removeEventListener("keydown", handleKeydown);
      window.removeEventListener("pointerup", handleCanvasPointerUp, true);
      window.removeEventListener("pointercancel", handleCanvasPointerUp, true);
      disableGlobalNoSelect();
    };
  });
</script>

<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<div
  class="canvas-container"
  bind:this={containerRef}
  onwheel={handleWheel}
  onpointerdown={handlePointerDown}
  onmousemove={handleMouseMove}
  onpointermove={handlePointerMove}
  onpointerup={handlePointerUp}
  onpointercancel={handlePointerUp}
  ondragstart={preventDefault(bubble('dragstart'))}
  onselectstart={preventDefault(bubble('selectstart'))}
  role="application"
  aria-label="タスクグリッド"
  tabindex="0"
>
  <div
    class="canvas-viewport"
    style="transform: {$canvasTransform}; transform-origin: 0 0;"
  >
    <!-- グリッド背景パターン -->
    <div class="grid-background">
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
    <svg class="connections-layer">
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
    <div class="nodes-layer">
      {#each $taskNodes as node (node.task.id)}
        <GridNode
          task={node.task}
          col={node.col}
          row={node.row}
          zoomLevel={$viewport.zoom}
        />
      {/each}
    </div>
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
    user-select: none;
  }

  :global(body.mv-canvas-dragging) {
    user-select: none !important;
  }

  .canvas-container:active {
    cursor: grabbing;
  }

  .grid-background {
    position: absolute;
    inset: calc(var(--mv-canvas-overflow-size) * -1);
    width: calc(100% + var(--mv-canvas-overflow-size) * 2);
    height: calc(100% + var(--mv-canvas-overflow-size) * 2);
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
    pointer-events: none;
    overflow: visible;
  }

  .nodes-layer {
    position: absolute;
    top: 0;
    left: 0;
  }

  .canvas-viewport {
    position: absolute;
    inset: 0;
    transform-origin: 0 0;
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
