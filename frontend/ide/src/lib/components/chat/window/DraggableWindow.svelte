<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let initialPosition = { x: 20, y: 20 };
  export let title: string = "";
  export let controls = { minimize: true, close: true };

  let position = { ...initialPosition };
  let isDragging = false;
  let windowEl: HTMLElement;
  let isMinimized = false;

  const dispatch = createEventDispatcher<{
    close: void;
    minimize: { minimized: boolean };
  }>();

  function startDrag(e: MouseEvent) {
    if (e.button !== 0) return;
    if ((e.target as HTMLElement).closest(".window-controls")) return;
    if (!windowEl) return;

    isDragging = true;
    window.addEventListener("mouseup", stopDrag);
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
    window.removeEventListener("mouseup", stopDrag);
  }

  function toggleMinimize() {
    isMinimized = !isMinimized;
    dispatch("minimize", { minimized: isMinimized });
  }

  function closeWindow() {
    dispatch("close");
  }
</script>

<svelte:window on:mousemove={onMouseMove} />

<!-- svelte-ignore a11y-no-static-element-interactions -->
<div
  class="floating-window"
  class:minimized={isMinimized}
  style="top: {position.y}px; left: {position.x}px;"
  bind:this={windowEl}
>
  <div class="header" on:mousedown={startDrag}>
    <div class="header-content">
      <slot name="header">
        <span class="title">{title}</span>
      </slot>
    </div>
    <div class="window-controls">
      {#if controls.minimize}
        <button
          class="control-btn"
          on:click|stopPropagation={toggleMinimize}
          aria-label="Minimize"
          type="button"
        >
          _
        </button>
      {/if}
      {#if controls.close}
        <button
          class="control-btn close"
          on:click|stopPropagation={closeWindow}
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
      <slot />
    </div>

    {#if $$slots.footer}
      <div class="footer">
        <slot name="footer" />
      </div>
    {/if}
  {/if}
</div>

<style>
  .floating-window {
    position: fixed;
    width: var(--mv-floating-window-width);
    height: var(--mv-floating-window-height);

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
    z-index: 1000;
    overflow: hidden;
    transition: height 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .floating-window.minimized {
    height: var(--mv-size-floating-header);
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
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
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
