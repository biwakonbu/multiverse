<script lang="ts">
  import {
    executionState,
    startExecution,
    pauseExecution,
    resumeExecution,
    stopExecution,
  } from "../../stores/executionStore";
  import Button from "../../design-system/components/Button.svelte";
  import { Play, Pause, Square } from "lucide-svelte";
</script>

<div class="execution-controls">
  {#if $executionState === "IDLE"}
    <Button
      variant="primary"
      size="small"
      on:click={startExecution}
      title="Start Execution"
    >
      <Play size="14" />
      <span>Start</span>
    </Button>
  {:else if $executionState === "RUNNING"}
    <Button
      variant="secondary"
      size="small"
      on:click={pauseExecution}
      title="Pause Execution"
    >
      <Pause size="14" />
    </Button>
    <Button
      variant="danger"
      size="small"
      on:click={stopExecution}
      title="Stop Execution"
    >
      <Square size="14" />
    </Button>
  {:else if $executionState === "PAUSED"}
    <Button
      variant="primary"
      size="small"
      on:click={resumeExecution}
      title="Resume Execution"
    >
      <Play size="14" />
    </Button>
    <Button
      variant="danger"
      size="small"
      on:click={stopExecution}
      title="Stop Execution"
    >
      <Square size="14" />
    </Button>
  {/if}

  <div class="state-indicator" class:active={$executionState !== "IDLE"}>
    <span class="status-dot {$executionState.toLowerCase()}"></span>
    <span class="status-text">{$executionState}</span>
  </div>
</div>

<style>
  .execution-controls {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-sm);
  }

  .state-indicator {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xxs);
    margin-left: var(--mv-spacing-xs);
    font-size: var(--mv-font-size-xs);
    font-family: var(--mv-font-mono);
  }

  .status-dot {
    width: var(--mv-status-dot-size);
    height: var(--mv-status-dot-size);
    border-radius: var(--mv-radius-full);
    background: var(--mv-color-text-muted);
  }

  .status-dot.running {
    background: var(--mv-color-status-running-text);
    box-shadow: var(--mv-status-dot-glow-running);
    animation: mv-pulse 1.5s infinite;
  }

  .status-dot.paused {
    background: var(--mv-primitive-aurora-yellow);
  }

  .status-text {
    color: var(--mv-color-text-secondary);
  }
</style>
