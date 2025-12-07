<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { Button } from "../../design-system";

  interface Props {
    scale?: number;
  }

  let { scale = 1.0 }: Props = $props();

  const dispatch = createEventDispatcher<{
    zoomIn: void;
    zoomOut: void;
    reset: void;
  }>();
</script>

<div class="zoom-controls">
  <Button
    variant="secondary"
    size="small"
    on:click={() => dispatch("zoomOut")}
    label="-"
    title="ズームアウト"
  />
  <span class="scale-label">{Math.round(scale * 100)}%</span>
  <Button
    variant="secondary"
    size="small"
    on:click={() => dispatch("zoomIn")}
    label="+"
    title="ズームイン"
  />
  <Button
    variant="ghost"
    size="small"
    on:click={() => dispatch("reset")}
    label="Reset"
    title="リセット"
  />
</div>

<style>
  .zoom-controls {
    position: absolute;
    bottom: var(--mv-spacing-lg);
    left: calc(var(--mv-spacing-lg) * 2 + var(--mv-icon-size-xxxl));
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
    background: var(--mv-color-surface-primary);
    padding: var(--mv-spacing-xs);
    border-radius: var(--mv-radius-md);
    box-shadow: var(--mv-shadow-card);
    z-index: 100;
  }

  .scale-label {
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-secondary);
    min-width: var(--mv-zoom-label-min-width);
    text-align: center;
  }
</style>
