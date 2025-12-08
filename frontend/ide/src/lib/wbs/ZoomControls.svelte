<script lang="ts">
  import { Button } from "../../design-system";

  interface Props {
    scale?: number;
    onzoomin?: () => void;
    onzoomout?: () => void;
    onreset?: () => void;
  }

  let { scale = 1.0, onzoomin, onzoomout, onreset }: Props = $props();
</script>

<div class="zoom-controls">
  <Button
    variant="secondary"
    size="small"
    onclick={() => onzoomout?.()}
    label="-"
    title="ズームアウト"
  />
  <span class="scale-label">{Math.round(scale * 100)}%</span>
  <Button
    variant="secondary"
    size="small"
    onclick={() => onzoomin?.()}
    label="+"
    title="ズームイン"
  />
  <Button
    variant="ghost"
    size="small"
    onclick={() => onreset?.()}
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
