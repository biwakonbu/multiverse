<script lang="ts">
  

  interface Props {
    /**
   * Tooltip Component
   * Simple Tooltip implementation.
   * Note: In a real app, this might stick to a parent. Here we implement a simple wrapper or standalone box.
   * For this design system, we'll make it a wrapper that shows the tip on hover.
   */
    content: string;
    position?: "top" | "bottom" | "left" | "right";
    class?: string;
    children?: import('svelte').Snippet;
  }

  let {
    content,
    position = "top",
    class: className = "",
    children
  }: Props = $props();
  
</script>

<div class="tooltip-wrapper {className}">
  {@render children?.()}
  <div class="tooltip position-{position}">
    {content}
  </div>
</div>

<style>
  .tooltip-wrapper {
    position: relative;
    display: inline-block;
  }

  .tooltip {
    visibility: hidden;
    position: absolute;
    z-index: var(--mv-z-tooltip);
    padding: var(--mv-spacing-xs) var(--mv-spacing-sm);

    background: var(--mv-glass-bg-strong);
    backdrop-filter: blur(12px);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border);
    border-radius: var(--mv-radius-sm);

    color: var(--mv-color-text-primary);
    font-size: var(--mv-font-size-xs);
    white-space: nowrap;
    opacity: 0;
    transition:
      opacity 0.2s ease,
      transform 0.2s ease;
    pointer-events: none;
    box-shadow: var(--mv-shadow-md);
  }

  .tooltip-wrapper:hover .tooltip {
    visibility: visible;
    opacity: 1;
  }

  /* Positions */
  .position-top {
    bottom: var(--mv-size-full);
    left: var(--mv-size-half);
    transform: translateX(-50%) translateY(calc(-1 * var(--mv-spacing-xxs)));
  }
  .tooltip-wrapper:hover .position-top {
    transform: translateX(-50%) translateY(calc(-1 * var(--mv-spacing-xs)));
  }

  .position-bottom {
    top: var(--mv-size-full);
    left: var(--mv-size-half);
    transform: translateX(-50%) translateY(var(--mv-spacing-xxs));
  }
  .tooltip-wrapper:hover .position-bottom {
    transform: translateX(-50%) translateY(var(--mv-spacing-xs));
  }

  .position-right {
    top: var(--mv-size-half);
    left: var(--mv-size-full);
    transform: translateY(-50%) translateX(var(--mv-spacing-xxs));
  }
  .tooltip-wrapper:hover .position-right {
    transform: translateY(-50%) translateX(var(--mv-spacing-xs));
  }

  .position-left {
    top: var(--mv-size-half);
    right: var(--mv-size-full);
    transform: translateY(-50%) translateX(calc(-1 * var(--mv-spacing-xxs)));
  }
  .tooltip-wrapper:hover .position-left {
    transform: translateY(-50%) translateX(calc(-1 * var(--mv-spacing-xs)));
  }
</style>
