<script lang="ts">
  import type { ExecutionState } from "../../stores/executionStore";
  import type { ResourceNode } from "./types";
  import ResourceList from "./ResourceList.svelte";
  import { slide } from "svelte/transition";

  interface Props {
    executionState?: ExecutionState;
    resources?: ResourceNode[]; // Add resources prop
    activeTaskTitle?: string | undefined;
  }

  let {
    executionState = "IDLE",
    resources = [],
    activeTaskTitle = undefined,
  }: Props = $props();

  let isExpanded = $state(false);

  function toggle() {
    isExpanded = !isExpanded;
  }

  function getStateColor(state: string) {
    switch (state) {
      case "RUNNING":
        return "var(--mv-color-status-success-text)";
      case "PAUSED":
        return "var(--mv-color-status-warning-text)";
      default:
        return "var(--mv-color-text-muted)";
    }
  }
  let stateColor = $derived(getStateColor(executionState));
</script>

<div class="process-hud-container" class:expanded={isExpanded}>
  <!-- Header / Collapsed View -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="hud-header" onclick={toggle} role="button" tabindex="0">
    <div class="left-group">
      <div class="indicator" style:background={stateColor}>
        {#if executionState === "RUNNING"}
          <div class="pulse-ring" style:border-color={stateColor}></div>
        {/if}
      </div>
      <span class="status-text">{executionState}</span>
      {#if executionState === "RUNNING"}
        <span class="separator">|</span>
        <span class="active-task-hint" title={activeTaskTitle}>
          {activeTaskTitle || "Worker Active"}
        </span>
      {/if}
    </div>

    <div class="right-group">
      <!-- Show active resource count or something relevant -->
      <span class="log-counter"
        >{resources.length > 0 ? resources[0].children?.length || 0 : 0} agents</span
      >
      <span class="toggle-icon">{isExpanded ? "▼" : "▲"}</span>
    </div>
  </div>

  <!-- Expanded Content -->
  {#if isExpanded}
    <div class="hud-content" transition:slide={{ duration: 200 }}>
      <div class="content-wrapper">
        <ResourceList {resources} />
      </div>
    </div>
  {/if}
</div>

<style>
  .process-hud-container {
    position: fixed;
    bottom: 0;
    left: var(--mv-position-center); /* 中央寄せ */
    transform: translateX(var(--mv-transform-center-x));
    width: var(--mv-space-96);
    max-width: var(--mv-max-width-viewport-90);
    background: var(--mv-glass-bg); /* Phantom Glass */
    backdrop-filter: var(--mv-glass-blur);
    border: var(--mv-border-width-sm) solid var(--mv-glass-border-subtle);
    border-bottom: none;
    border-top-left-radius: var(--mv-radius-lg);
    border-top-right-radius: var(--mv-radius-lg);
    box-shadow: var(--mv-shadow-floating-panel); /* より深いシャドウに */
    z-index: var(--mv-z-index-overlay);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    transition: all var(--mv-transition-state);
  }

  .hud-header {
    height: var(--mv-space-10);
    padding: 0 var(--mv-space-4);
    display: flex;
    align-items: center;
    justify-content: space-between;
    cursor: pointer;
    background: var(--mv-glass-border);
    border-bottom: var(--mv-border-width-sm) solid transparent;
    transition: background var(--mv-transition-hover);
  }

  .process-hud-container.expanded .hud-header {
    border-bottom-color: var(--mv-glass-border-subtle);
    background: var(--mv-glass-bg-dark);
  }

  .hud-header:hover {
    background: var(--mv-glass-hover);
  }

  .left-group,
  .right-group {
    display: flex;
    align-items: center;
    gap: var(--mv-space-3);
  }

  .status-text {
    font-family: var(--mv-font-display);
    font-weight: var(--mv-font-weight-bold);
    font-size: var(--mv-font-size-sm);
    letter-spacing: var(--mv-letter-spacing-wider);
    color: var(--mv-color-text-primary);
    text-shadow: var(--mv-text-shadow-subtle);
  }

  .separator {
    color: var(--mv-glass-border-light);
    height: var(--mv-space-3);
    width: var(--mv-space-px);
    background: var(--mv-color-current);
    border: none;
    overflow: hidden;
    text-indent: var(--mv-space-hidden);
  }

  .active-task-hint {
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-secondary);
    font-style: italic;
    opacity: var(--mv-opacity-80);
    max-width: var(--mv-space-72);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .log-counter {
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-2xs);
    color: var(--mv-color-text-muted);
    background: var(--mv-glass-border);
    padding: var(--mv-space-0-5) var(--mv-space-1-5);
    border-radius: var(--mv-radius-full);
  }

  .toggle-icon {
    font-size: var(--mv-font-size-3xs);
    color: var(--mv-color-text-muted);
    opacity: var(--mv-opacity-70);
    transition: transform var(--mv-transition-transform);
  }

  .process-hud-container.expanded .toggle-icon {
    transform: rotate(180deg);
  }

  .hud-content {
    background: var(--mv-glass-bg-dark);
  }

  .content-wrapper {
    padding: var(--mv-space-3);
  }

  /* Status Indicator Styles */
  .indicator {
    width: var(--mv-space-2);
    height: var(--mv-space-2);
    border-radius: var(--mv-radius-full);
    position: relative;
    box-shadow: var(--mv-shadow-glow-subtle);
  }

  .pulse-ring {
    position: absolute;
    inset: calc(-1 * var(--mv-space-1));
    border-radius: var(--mv-radius-full);
    border: var(--mv-border-width-sm) solid;
    opacity: 0;
    animation: pulse var(--mv-duration-slow) infinite ease-out;
    box-shadow: var(--mv-shadow-glow-current);
  }

  @keyframes pulse {
    0% {
      transform: scale(0.5);
      opacity: 0;
    }
    50% {
      opacity: 0.6;
    }
    100% {
      transform: scale(1.5); /* より大きく */
      opacity: 0;
    }
  }
</style>
