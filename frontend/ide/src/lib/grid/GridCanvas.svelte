<script lang="ts">
  import { onMount } from 'svelte';
  import GridNode from './GridNode.svelte';
  import { viewport, drag, canvasTransform, zoomPercent } from '../../stores';
  import { taskNodes, gridBounds } from '../../stores';
  import { zoom as zoomConfig } from '../../design-system';

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
      drag.startDrag(event.clientX, event.clientY, $viewport.panX, $viewport.panY);
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
    if (event.key === '+' || event.key === '=') {
      event.preventDefault();
      viewport.zoomIn();
    } else if (event.key === '-') {
      event.preventDefault();
      viewport.zoomOut();
    } else if (event.key === '0') {
      event.preventDefault();
      viewport.reset();
    }
  }

  onMount(() => {
    // グローバルキーボードイベント
    window.addEventListener('keydown', handleKeydown);
    return () => {
      window.removeEventListener('keydown', handleKeydown);
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
          id="grid-dots"
          width="200"
          height="140"
          patternUnits="userSpaceOnUse"
        >
          <circle cx="100" cy="70" r="1.5" fill="var(--mv-color-border-subtle)" />
        </pattern>
      </defs>
      <rect width="100%" height="100%" fill="url(#grid-dots)" />
    </svg>
  </div>

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

  .nodes-layer {
    position: absolute;
    top: 0;
    left: 0;
    transform-origin: 0 0;
    padding: var(--mv-spacing-xl);
  }

  .zoom-indicator {
    position: absolute;
    bottom: var(--mv-spacing-md);
    right: var(--mv-spacing-md);
    background: var(--mv-color-surface-secondary);
    color: var(--mv-color-text-secondary);
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    border-radius: var(--mv-radius-sm);
    font-size: var(--mv-font-size-xs);
    font-family: var(--mv-font-mono);
    pointer-events: none;
    opacity: 0.8;
  }
</style>
