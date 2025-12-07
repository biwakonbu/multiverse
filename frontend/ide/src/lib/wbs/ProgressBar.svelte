<script lang="ts">
  import { getProgressColor } from "./utils";

  interface Props {
    percentage?: number;
    size?: "sm" | "md" | "mini";
    className?: string;
  }

  let { percentage = 0, size = "sm", className = "" }: Props = $props();

  let progressColor = $derived(getProgressColor(percentage));

  // Calculate dynamic shadow and background for the container
  let containerShadow =
    $derived(size === "md"
      ? `0 0 2px ${progressColor.glow}, ${progressColor.insetShadow}`
      : progressColor.insetShadow);
</script>

<div
  class="progress-bar {size} {className}"
  style:box-shadow={containerShadow}
  style:background-color={progressColor.bg}
>
  <div
    class="progress-fill"
    style:width="{percentage}%"
    style:background-color={progressColor.fill}
    style:box-shadow="var(--mv-shadow-badge-glow-md) {progressColor.glow}"
  ></div>
</div>

<style>
  .progress-bar {
    border-radius: var(--mv-radius-sm);
    overflow: hidden;
  }

  /* Size variants */
  .progress-bar.sm {
    width: var(--mv-progress-bar-width-sm);
    height: var(--mv-progress-bar-height-sm);
  }

  .progress-bar.mini {
    width: var(--mv-progress-bar-width-mini);
    height: var(--mv-progress-bar-height-sm);
  }

  .progress-bar.md {
    width: var(--mv-size-full);
    height: var(--mv-progress-bar-height-md);
    border: var(--mv-border-panel);
  }

  .progress-fill {
    height: var(--mv-size-full);
    width: var(--progress, 0%);
    background: var(--mv-progress-bar-fill);
    border-radius: var(--mv-radius-sm);
    transition: width var(--mv-duration-slow);
    box-shadow: var(--mv-shadow-progress-fill) var(--mv-progress-bar-fill-glow);
  }
</style>
