<script lang="ts">
  import { overallProgress, expandedNodes } from "../../stores/wbsStore";
  import { getProgressColor } from "./utils";
  import ProgressBar from "./ProgressBar.svelte";

  function splitProgress(percentage: number) {
    const str = Math.round(percentage).toString();
    if (str.length === 1) return { first: str, rest: "" };
    return { first: str[0], rest: str.slice(1) };
  }

  $: percentage = $overallProgress?.percentage ?? 0;
  $: completed = $overallProgress?.completed ?? 0;
  $: total = $overallProgress?.total ?? 0;
  $: progressParts = splitProgress(percentage);
  $: progressColor = getProgressColor(percentage);
</script>

<div class="wbs-header">
  <div class="header-content">
    <div class="header-info">
      <h2 class="title">WBS</h2>

      <div class="progress-section">
        <ProgressBar {percentage} size="sm" />
        <span
          class="progress-percentage"
          style:color={progressColor.fill}
          style:text-shadow={progressColor.textShadowMd}
        >
          <span class="progress-first-digit">{progressParts.first}</span>
          <span class="progress-rest-digits">{progressParts.rest}</span>
          <span class="progress-symbol">%</span>
        </span>
        <span class="task-count">
          {completed}/{total}
        </span>
      </div>
    </div>

    <div class="header-actions">
      <button
        class="action-btn"
        on:click={() => expandedNodes.expandAll()}
        title="Expand All"
      >
        <span class="icon">↕</span>
      </button>
      <button
        class="action-btn"
        on:click={() => expandedNodes.collapseAll()}
        title="Collapse All"
      >
        <span class="icon">⇕</span>
      </button>
    </div>
  </div>
</div>

<style>
  .wbs-header {
    display: inline-flex;
    padding: var(--mv-spacing-sm) var(--mv-spacing-xl);

    /* Phantom Glass: No border, just blur and tint */
    background: var(--mv-glass-bg);
    backdrop-filter: blur(24px);

    /* Micro Frame: Barely visible edge to define volume */
    border: var(--mv-border-width-thin) solid var(--mv-glass-border);
    border-radius: var(--mv-radius-lg);

    /* Soft ambient shadow + Inner highlight for glass edge */
    box-shadow: var(--mv-shadow-glass-panel);

    pointer-events: auto;
    margin: var(--mv-spacing-md);
    min-width: var(--mv-min-width-wbs-header);
    align-items: center;
  }

  .header-content {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xxl); /* Wide breathing room */
    width: 100%;
    justify-content: space-between;
  }

  .header-info {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xl);
  }

  .title {
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-bold);
    color: var(--mv-color-text-secondary);
    margin: 0;
    text-transform: uppercase;
    letter-spacing: var(--mv-letter-spacing-title);
    opacity: 0.6;
  }

  .progress-section {
    display: flex;
    align-items: center; /* Align center for compact look */
    gap: var(--mv-spacing-sm);
  }

  .progress-percentage {
    display: flex;
    align-items: baseline;
    font-family: var(--mv-font-display);
    line-height: 1;
    font-style: italic;
    filter: var(--mv-filter-drop-shadow);
  }

  .progress-first-digit {
    font-size: var(--mv-font-size-xl);
    font-weight: var(--mv-font-weight-bold);
  }

  .progress-rest-digits {
    font-size: var(--mv-font-size-lg);
    font-weight: var(--mv-font-weight-bold);
  }

  .progress-symbol {
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-bold);
    margin-left: var(--mv-margin-progress-symbol);
    opacity: 0.6;
  }

  .task-count {
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-muted);
    font-family: var(--mv-font-mono);
    margin-left: var(--mv-spacing-lg);
    opacity: 0.5;
    letter-spacing: var(--mv-letter-spacing-count);
  }

  .header-actions {
    display: flex;
    gap: var(--mv-spacing-sm);
    padding-left: var(--mv-spacing-xl);
    height: var(--mv-size-header-actions);
    align-items: center;
  }

  .action-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: var(--mv-size-action-btn);
    height: var(--mv-size-action-btn);
    padding: 0;
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-muted);
    background: transparent;
    border: none;
    border-radius: var(--mv-radius-circle);
    cursor: pointer;
    transition: all var(--mv-transition-hover);
  }

  .action-btn:hover {
    color: var(--mv-color-text-primary);
    background: var(--mv-glass-hover);
    box-shadow: var(--mv-shadow-glow-hover);
    transform: scale(1.1);
  }

  .action-btn:active {
    background: var(--mv-glass-active);
    transform: scale(0.95);
  }

  .icon {
    line-height: 1;
    display: block;
  }
</style>
