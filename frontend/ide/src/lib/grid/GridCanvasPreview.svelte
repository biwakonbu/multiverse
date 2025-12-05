<!--
  GridCanvasPreview - Storybook 用の GridCanvas プレビューコンポーネント

  ストア依存を排除し、props のみで動作するバージョン。
  Storybook での単独表示・テスト用途。
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import GridNodePreview from './GridNodePreview.svelte';
  import type { TaskStatus } from '../../types';

  // ノードデータの型
  interface NodeData {
    id: string;
    title: string;
    status: TaskStatus;
    poolId: string;
    col: number;
    row: number;
  }

  // サンプルノードデータ
  export let nodes: NodeData[] = [];

  // ビューポート状態
  export let zoom = 1;
  export let panX = 32;
  export let panY = 32;

  // 選択状態
  export let selectedId: string | null = null;

  // ズーム設定
  const zoomConfig = {
    min: 0.25,
    max: 3.0,
    step: 0.1,
    wheelFactor: 0.1
  };

  let containerRef: HTMLDivElement;
  let isDragging = false;
  let dragStartX = 0;
  let dragStartY = 0;
  let dragStartPanX = 0;
  let dragStartPanY = 0;

  // CSS transform 文字列
  $: canvasTransform = `translate(${panX}px, ${panY}px) scale(${zoom})`;
  $: zoomPercent = Math.round(zoom * 100);

  // ホイールズーム
  function handleWheel(event: WheelEvent) {
    event.preventDefault();

    if (event.ctrlKey || event.metaKey) {
      // Ctrl/Cmd + ホイールでズーム
      const rect = containerRef.getBoundingClientRect();
      const mouseX = event.clientX - rect.left;
      const mouseY = event.clientY - rect.top;

      // マウス位置を中心にズーム
      const delta = event.deltaY > 0 ? -zoomConfig.wheelFactor : zoomConfig.wheelFactor;
      const newZoom = Math.max(zoomConfig.min, Math.min(zoomConfig.max, zoom + delta));

      if (newZoom !== zoom) {
        const scale = newZoom / zoom;
        panX = mouseX - (mouseX - panX) * scale;
        panY = mouseY - (mouseY - panY) * scale;
        zoom = newZoom;
      }
    } else {
      // 通常のホイールでパン
      panX -= event.deltaX;
      panY -= event.deltaY;
    }
  }

  // ドラッグ開始
  function handlePointerDown(event: PointerEvent) {
    if (event.button === 1 || (event.button === 0 && event.shiftKey)) {
      event.preventDefault();
      containerRef.setPointerCapture(event.pointerId);
      isDragging = true;
      dragStartX = event.clientX;
      dragStartY = event.clientY;
      dragStartPanX = panX;
      dragStartPanY = panY;
    }
  }

  // ドラッグ中
  function handlePointerMove(event: PointerEvent) {
    if (!isDragging) return;

    const deltaX = event.clientX - dragStartX;
    const deltaY = event.clientY - dragStartY;

    panX = dragStartPanX + deltaX;
    panY = dragStartPanY + deltaY;
  }

  // ドラッグ終了
  function handlePointerUp(event: PointerEvent) {
    if (isDragging) {
      containerRef.releasePointerCapture(event.pointerId);
      isDragging = false;
    }
  }

  // キーボードショートカット
  function handleKeydown(event: KeyboardEvent) {
    if (event.key === '+' || event.key === '=') {
      event.preventDefault();
      zoom = Math.min(zoomConfig.max, zoom + zoomConfig.step);
    } else if (event.key === '-') {
      event.preventDefault();
      zoom = Math.max(zoomConfig.min, zoom - zoomConfig.step);
    } else if (event.key === '0') {
      event.preventDefault();
      zoom = 1;
      panX = 32;
      panY = 32;
    }
  }

  // ノードクリック
  function handleNodeClick(id: string) {
    selectedId = selectedId === id ? null : id;
  }
</script>

<!-- svelte-ignore a11y-no-noninteractive-tabindex a11y-no-noninteractive-element-interactions -->
<div
  class="canvas-container"
  bind:this={containerRef}
  on:wheel={handleWheel}
  on:pointerdown={handlePointerDown}
  on:pointermove={handlePointerMove}
  on:pointerup={handlePointerUp}
  on:pointercancel={handlePointerUp}
  on:keydown={handleKeydown}
  role="application"
  aria-label="タスクグリッド"
  tabindex="0"
>
  <!-- グリッド背景パターン -->
  <div class="grid-background" style="transform: {canvasTransform};">
    <svg class="grid-pattern" width="100%" height="100%">
      <defs>
        <pattern
          id="grid-dots"
          width="200"
          height="140"
          patternUnits="userSpaceOnUse"
        >
          <circle cx="100" cy="70" r="1.5" fill="var(--mv-color-border-subtle, #333333)" />
        </pattern>
      </defs>
      <rect width="100%" height="100%" fill="url(#grid-dots)" />
    </svg>
  </div>

  <!-- ノードレイヤー -->
  <div class="nodes-layer" style="transform: {canvasTransform};">
    {#each nodes as node (node.id)}
      <div
        on:click={() => handleNodeClick(node.id)}
        on:keydown={(e) => e.key === 'Enter' && handleNodeClick(node.id)}
        role="button"
        tabindex="-1"
      >
        <GridNodePreview
          id={node.id}
          title={node.title}
          status={node.status}
          poolId={node.poolId}
          col={node.col}
          row={node.row}
          zoomLevel={zoom}
          selected={selectedId === node.id}
        />
      </div>
    {/each}
  </div>

  <!-- ズームインジケーター -->
  <div class="zoom-indicator">
    {zoomPercent}%
  </div>

  <!-- 操作ヒント -->
  <div class="controls-hint">
    <span>Scroll: パン</span>
    <span>Ctrl+Scroll: ズーム</span>
    <span>Shift+ドラッグ: パン</span>
    <span>+/-/0: ズーム操作</span>
  </div>
</div>

<style>
  .canvas-container {
    position: relative;
    width: 100%;
    height: 500px;
    overflow: hidden;
    background: var(--mv-color-surface-app, #1a1a1a);
    cursor: grab;
    touch-action: none;
    border-radius: var(--mv-radius-md, 8px);
  }

  .canvas-container:active {
    cursor: grabbing;
  }

  .canvas-container:focus {
    outline: 2px solid var(--mv-color-border-focus, #4caf50);
    outline-offset: -2px;
  }

  .grid-background {
    position: absolute;
    inset: -1000px;
    width: calc(100% + 2000px);
    height: calc(100% + 2000px);
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
  }

  .nodes-layer > div {
    display: contents;
  }

  .zoom-indicator {
    position: absolute;
    bottom: var(--mv-spacing-md, 16px);
    right: var(--mv-spacing-md, 16px);
    background: var(--mv-color-surface-secondary, #252525);
    color: var(--mv-color-text-secondary, #aaaaaa);
    padding: var(--mv-spacing-xxs, 4px) var(--mv-spacing-xs, 8px);
    border-radius: var(--mv-radius-sm, 4px);
    font-size: var(--mv-font-size-xs, 10px);
    font-family: var(--mv-font-mono, monospace);
    pointer-events: none;
    opacity: 0.8;
  }

  .controls-hint {
    position: absolute;
    bottom: var(--mv-spacing-md, 16px);
    left: var(--mv-spacing-md, 16px);
    display: flex;
    gap: var(--mv-spacing-sm, 12px);
    font-size: var(--mv-font-size-xs, 10px);
    color: var(--mv-color-text-muted, #888888);
    pointer-events: none;
  }

  .controls-hint span {
    background: var(--mv-color-surface-secondary, #252525);
    padding: 2px 6px;
    border-radius: var(--mv-radius-sm, 4px);
  }
</style>
