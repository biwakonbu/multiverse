<script lang="ts">
  import type { LogEntry } from "../../stores/logStore";
  import { logs as logStore } from "../../stores/logStore";

  interface Props {
    logs?: LogEntry[];
    height?: string;
  }

  let { logs = [], height = "200px" }: Props = $props();

  let container: HTMLDivElement | undefined = $state();
  let autoScroll = $state(true);

  // Scroll to bottom when logs update, if autoScroll is enabled
  $effect(() => {
    logs; // dependency for reactivity
    if (autoScroll && container) {
      container.scrollTop = container.scrollHeight;
    }
  });

  function onScroll() {
    if (!container) return;
    const { scrollTop, scrollHeight, clientHeight } = container;
    // Tolerance of 5px to consider "at bottom"
    const atBottom = scrollHeight - scrollTop - clientHeight <= 20;

    // If not at bottom, disable auto-scroll
    if (!atBottom) {
      autoScroll = false;
    } else {
      // If user scrolls back to bottom, re-enable auto-scroll
      autoScroll = true;
    }
  }

  function clearLogs() {
    // Since logs prop might be immutable or controlled by parent,
    // ideally we emit an event or call store. But for global log view, calling store directly is pragmatic for MVP.
    logStore.clear();
  }

  function getLogColor(stream: LogEntry["stream"]): string {
    return stream === "stderr"
      ? "var(--mv-color-status-failed-text)"
      : "var(--mv-color-text-secondary)";
  }
</script>

<div class="log-stream" style:height>
  <div class="log-header">
    <div class="header-left">
      <span class="title">LIVE LOGS</span>
      <!-- Debug/Status info -->
      <span class="count">{logs.length} lines</span>
    </div>
    <div class="header-right">
      {#if !autoScroll}
        <button
          class="action-btn resume"
          onclick={() => {
            autoScroll = true;
          }}
        >
          RESUME AUTO-SCROLL
        </button>
      {/if}
      <button class="action-btn clear" onclick={clearLogs}>CLEAR</button>
      <span
        class="status-badge"
        class:active={autoScroll}
        title={autoScroll ? "Auto-scroll enabled" : "Auto-scroll paused"}
      >
        {autoScroll ? "AUTO" : "PAUSED"}
      </span>
    </div>
  </div>

  <div class="log-content" bind:this={container} onscroll={onScroll}>
    {#each logs as log (log.id)}
      <div class="log-line">
        <span class="timestamp"
          >{log.timestamp.split("T")[1]?.split(".")[0] || ""}</span
        >
        <span class="stream" class:stderr={log.stream === "stderr"}
          >{log.stream === "stdout" ? "OUT" : "ERR"}</span
        >
        <span class="message" style:color={getLogColor(log.stream)}
          >{log.line}</span
        >
      </div>
    {/each}
    {#if logs.length === 0}
      <div class="empty-state">Waiting for output...</div>
    {/if}
  </div>
</div>

<style>
  .log-stream {
    display: flex;
    flex-direction: column;
    background: var(--mv-glass-bg-darker);
    border-radius: var(--mv-radius-md);
    border: var(--mv-border-width-sm) solid var(--mv-glass-border-subtle);
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-xs);
    overflow: hidden;
    box-shadow: var(--mv-shadow-inset-dark);
  }

  .log-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--mv-space-1-5) var(--mv-space-3);
    background: var(--mv-glass-border);
    border-bottom: var(--mv-border-width-sm) solid var(--mv-glass-border-subtle);
    backdrop-filter: var(--mv-glass-blur-light);
  }

  .header-left,
  .header-right {
    display: flex;
    align-items: center;
    gap: var(--mv-space-3);
  }

  .title {
    color: var(--mv-color-text-secondary);
    font-weight: var(--mv-font-weight-bold);
    letter-spacing: var(--mv-letter-spacing-wide);
    font-size: var(--mv-font-size-2xs);
    text-transform: uppercase;
    text-shadow: var(--mv-text-shadow-subtle);
  }

  .count {
    font-size: var(--mv-font-size-3xs);
    color: var(--mv-color-text-muted);
    opacity: var(--mv-opacity-60);
  }

  .action-btn {
    background: transparent;
    border: none;
    color: var(--mv-color-text-muted);
    font-size: var(--mv-font-size-2xs);
    font-weight: var(--mv-font-weight-bold);
    cursor: pointer;
    padding: var(--mv-space-0-5) var(--mv-space-1-5);
    border-radius: var(--mv-radius-sm);
    transition: all var(--mv-transition-base);
  }
  .action-btn:hover {
    background: var(--mv-glass-hover);
    color: var(--mv-color-text-primary);
  }
  .action-btn.resume {
    color: var(--mv-color-status-info-text);
  }
  .action-btn.clear:hover {
    background: var(--mv-color-status-failed-bg);
    color: var(--mv-color-status-failed-text);
  }

  .status-badge {
    font-size: var(--mv-font-size-3xs);
    padding: var(--mv-space-0-5) var(--mv-space-1-5);
    border-radius: var(--mv-radius-sm);
    background: var(--mv-glass-border-light);
    color: var(--mv-color-text-muted);
    border: var(--mv-border-width-sm) solid transparent;
    transition: all var(--mv-transition-base);
    user-select: none;
  }

  .status-badge.active {
    background: var(--mv-color-status-success-bg);
    color: var(--mv-color-status-success-text);
    border-color: var(--mv-color-status-success-border);
    box-shadow: var(--mv-shadow-text-glow-xs);
  }

  .log-content {
    flex: 1;
    overflow-y: auto;
    padding: var(--mv-space-2) var(--mv-space-3);
    display: flex;
    flex-direction: column;
    gap: var(--mv-space-px);
  }

  .log-line {
    display: flex;
    gap: var(--mv-space-3);
    line-height: var(--mv-line-height-relaxed);
    word-break: break-all;
    padding: var(--mv-space-px) var(--mv-space-1);
    border-radius: var(--mv-radius-xs);
    transition: background var(--mv-transition-hover);
  }

  .log-line:hover {
    background: var(--mv-glass-hover);
  }

  .timestamp {
    color: var(--mv-color-text-muted);
    opacity: var(--mv-opacity-40);
    flex-shrink: 0;
    width: var(--mv-space-16);
    font-size: var(--mv-font-size-2xs);
    text-shadow: var(--mv-text-shadow-timestamp);
  }

  .stream {
    flex-shrink: 0;
    width: var(--mv-space-8);
    text-align: center;
    font-weight: var(--mv-font-weight-bold);
    font-size: var(--mv-font-size-3xs);
    padding: var(--mv-space-px) 0;
    border-radius: var(--mv-radius-xs);
    background: var(--mv-glass-border-light);
    color: var(--mv-color-text-secondary);
    letter-spacing: var(--mv-letter-spacing-wide);
  }

  .stream.stderr {
    background: var(--mv-color-status-failed-bg);
    color: var(--mv-color-status-failed-text);
    border: var(--mv-border-width-sm) solid var(--mv-color-status-failed-border);
    box-shadow: var(--mv-shadow-glow-error);
  }

  .message {
    white-space: pre-wrap;
    color: var(--mv-color-text-primary);
    opacity: var(--mv-opacity-90);
  }

  .empty-state {
    padding: var(--mv-space-8);
    text-align: center;
    color: var(--mv-color-text-muted);
    font-style: italic;
    opacity: var(--mv-opacity-60);
  }

  /* Custom Scrollbar */
  .log-content::-webkit-scrollbar {
    width: var(--mv-space-1-5);
  }
  .log-content::-webkit-scrollbar-track {
    background: transparent;
  }
  .log-content::-webkit-scrollbar-thumb {
    background: var(--mv-glass-border-light);
    border-radius: var(--mv-radius-xs);
    border: var(--mv-border-width-sm) solid transparent;
    background-clip: content-box;
  }
  .log-content::-webkit-scrollbar-thumb:hover {
    background: var(--mv-glass-border-hover);
    background-clip: content-box;
  }
</style>
