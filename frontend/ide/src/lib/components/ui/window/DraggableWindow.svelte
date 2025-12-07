<script lang="ts">
  import { stopPropagation } from 'svelte/legacy';

  interface Props {
    id?: string;
    initialPosition?: { x: number; y: number };
    initialSize?: { width: number; height: number };
    title?: string;
    controls?: { minimize: boolean; close: boolean };
    zIndex?: number;
    header?: import('svelte').Snippet;
    children?: import('svelte').Snippet;
    footer?: import('svelte').Snippet;
    // コールバックプロップ
    onclose?: () => void;
    onminimize?: (data: { minimized: boolean }) => void;
    onclick?: () => void;
    ondragend?: (data: { x: number; y: number }) => void;
  }

  let {
    id = "unknown",
    initialPosition = { x: 20, y: 20 },
    initialSize = undefined,
    title = "",
    controls = { minimize: true, close: true },
    zIndex = 100,
    header,
    children,
    footer,
    onclose,
    onminimize,
    onclick,
    ondragend,
  }: Props = $props();

  let position = $state({ ...initialPosition });
  let isDragging = false;
  let windowEl: HTMLElement | undefined = $state();
  let isMinimized = $state(false);
  let size = $state(initialSize);

  function startDrag(e: MouseEvent) {
    if (e.button !== 0) return;
    if ((e.target as HTMLElement).closest(".window-controls")) return;
    if (!windowEl) return;

    isDragging = true;
    (windowEl as HTMLElement).style.cursor = 'grabbing';
    window.addEventListener("mouseup", stopDrag);
    onclick?.();
  }

  function onMouseMove(e: MouseEvent) {
    if (!isDragging) return;

    position = {
      x: position.x + e.movementX,
      y: position.y + e.movementY,
    };
  }

  function stopDrag() {
    isDragging = false;
    if (windowEl) (windowEl as HTMLElement).style.cursor = '';
    window.removeEventListener("mouseup", stopDrag);
    ondragend?.(position);
  }

  function toggleMinimize() {
    isMinimized = !isMinimized;
    onminimize?.({ minimized: isMinimized });
  }

  function closeWindow() {
    onclose?.();
  }

  function onWindowClick() {
    onclick?.();
  }
</script>

<svelte:window onmousemove={onMouseMove} />

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  class="floating-window"
  class:minimized={isMinimized}
  style:top="{position.y}px"
  style:left="{position.x}px"
  style:z-index={zIndex}
  style:width={size ? `${size.width}px` : undefined}
  style:height={size && !isMinimized ? `${size.height}px` : undefined}
  bind:this={windowEl}
  onmousedown={onWindowClick}
>
  <div class="header" onmousedown={startDrag}>
    <div class="header-content">
      {#if header}{@render header()}{:else}
        <span class="title">{title}</span>
      {/if}
    </div>
    <div class="window-controls">
      {#if controls.minimize}
        <button
          class="control-btn"
          onclick={stopPropagation(toggleMinimize)}
          aria-label="Minimize"
          type="button"
        >
          _
        </button>
      {/if}
      {#if controls.close}
        <button
          class="control-btn close"
          onclick={stopPropagation(closeWindow)}
          aria-label="Close"
          type="button"
        >
          ×
        </button>
      {/if}
    </div>
  </div>

  {#if !isMinimized}
    <div class="content">
      {@render children?.()}
    </div>

    {#if footer}
      <div class="footer">
        {@render footer?.()}
      </div>
    {/if}
  {/if}
</div>

<style>
  .floating-window {
    position: fixed;

    /* Default dims if not provided */
    min-width: var(--mv-floating-window-min-width);
    min-height: var(--mv-floating-window-min-height);

    /* If width/height not set via style, use default checks, but we rely on style now */

    /* Crystal HUD: Slightly more assertive than Header */
    background: var(--mv-glass-bg-chat);
    backdrop-filter: blur(24px);

    /* Assertive Border */
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-strong);
    border-top: var(--mv-border-width-thin) solid var(--mv-glass-border-top);
    border-bottom: var(--mv-border-width-thin) solid
      var(--mv-glass-border-bottom);

    border-radius: var(--mv-radius-lg);

    /* Deep Shadow */
    box-shadow: var(--mv-shadow-floating-panel);

    display: flex;
    flex-direction: column;

    /* z-index is set via style */
    overflow: hidden;
    transition: height 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .floating-window.minimized {
    height: var(--mv-size-floating-header) !important; /* Force override style height */
    overflow: hidden;
    background: var(--mv-glass-bg-minimized);
  }

  /* Header Area */
  .header {
    height: var(--mv-size-floating-header);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 var(--mv-spacing-md);
    cursor: grab;
    user-select: none;
    flex-shrink: 0;
    background: transparent;
    border-bottom: var(--mv-border-width-thin) solid
      var(--mv-glass-border-subtle);
  }

  .header:active {
    cursor: grabbing;
  }

  .header-content {
    flex: 1;
    overflow: hidden;
  }

  .title {
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-bold);
    color: var(--mv-color-text-secondary);
  }

  .window-controls {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
    margin-left: var(--mv-spacing-sm);
  }

  .control-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: var(--mv-size-control-btn);
    height: var(--mv-size-control-btn);
    background: transparent;
    border: var(--mv-border-width-thin) solid transparent;
    border-radius: var(--mv-radius-sm);
    color: var(--mv-color-text-muted);
    cursor: pointer;
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-sm);
    padding: 0;
    transition: all var(--mv-duration-fast);
  }

  .control-btn:hover {
    background: var(--mv-glass-hover);
    color: var(--mv-color-text-primary);
  }

  .control-btn.close:hover {
    background: var(--mv-glass-close-bg);
    color: var(--mv-primitive-aurora-red);
    border-color: var(--mv-glass-close-border);
  }

  .content {
    flex: 1;
    min-height: 0; /* Allow flex item to shrink below content size */
    overflow: hidden; /* Let slotted content handle its own scrolling */

    /* padding is controlled by children if needed, or default */
    display: flex;
    flex-direction: column;
  }

  .footer {
    flex-shrink: 0;
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
    border-top: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
    background: var(--mv-glass-footer);
  }
</style>
