<script lang="ts">
  import {
    executionState,
    startExecution,
    pauseExecution,
    resumeExecution,
    stopExecution,
  } from "../../stores/executionStore";
</script>

<div class="execution-controls">
  {#if $executionState === "IDLE"}
    <button
      class="control-btn start"
      on:click={startExecution}
      title="自律実行開始"
      aria-label="Start Execution"
    >
      <span class="icon">▶</span>
      <span class="label">Start</span>
    </button>
  {:else if $executionState === "RUNNING"}
    <button
      class="control-btn pause"
      on:click={pauseExecution}
      title="一時停止"
      aria-label="Pause Execution"
    >
      <span class="icon">⏸</span>
    </button>
    <button
      class="control-btn stop"
      on:click={stopExecution}
      title="停止"
      aria-label="Stop Execution"
    >
      <span class="icon">⏹</span>
    </button>
  {:else if $executionState === "PAUSED"}
    <button
      class="control-btn resume"
      on:click={resumeExecution}
      title="再開"
      aria-label="Resume Execution"
    >
      <span class="icon">▶</span>
    </button>
    <button
      class="control-btn stop"
      on:click={stopExecution}
      title="停止"
      aria-label="Stop Execution"
    >
      <span class="icon">⏹</span>
    </button>
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
    padding: var(--mv-spacing-xxs) var(--mv-spacing-sm);
    background: var(--mv-color-surface-secondary);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    border-radius: var(--mv-radius-sm);
  }

  .control-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: var(--mv-spacing-xs);
    border: none;
    border-radius: var(--mv-radius-sm);
    cursor: pointer;
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-medium);
    transition: all var(--mv-transition-hover);
    padding: var(--mv-spacing-xxs) var(--mv-spacing-sm);
    min-width: var(--mv-btn-min-width-sm);
  }

  .control-btn.start {
    background: var(--mv-color-status-running-bg);
    color: var(--mv-color-status-running-text);
  }

  .control-btn.resume {
    background: var(--mv-color-status-ready-bg);
    color: var(--mv-color-status-ready-text);
  }

  .control-btn.pause {
    background: var(--mv-color-surface-overlay);
    color: var(--mv-color-text-primary);
  }

  .control-btn.stop {
    background: var(--mv-color-status-failed-bg);
    color: var(--mv-color-status-failed-text);
  }

  .control-btn:hover {
    filter: brightness(1.1);
    transform: translateY(-1px);
  }

  .control-btn:active {
    transform: translateY(0);
  }

  .icon {
    font-family: var(--mv-font-mono); /* For generic symbols */
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
    background: var(--mv-color-text-warning);
  }

  .status-text {
    color: var(--mv-color-text-secondary);
  }
</style>
