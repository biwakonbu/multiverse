<script lang="ts">
  /**
   * バッジのバリアント (status)
   */
  export let status:
    | "pending"
    | "ready"
    | "running"
    | "succeeded"
    | "completed"
    | "failed"
    | "canceled"
    | "blocked"
    | "retry_wait" = "pending";

  /**
   * サイズ - StatusBadge は固定サイズだが、汎用性のために残す（ただしデフォルトスタイルを優先）
   */
  export let size: "small" | "medium" = "medium";

  /**
   * ラベル（指定がなければ slot or statusの大文字）
   */
  export let label = "";

  // ステータス文字列の正規化（RETRY_WAIT -> retry-wait）
  $: normalizedStatus = status.toLowerCase().replace("_", "-");
</script>

<span class="badge status-{normalizedStatus} size-{size}">
  {#if label}
    {label}
  {:else}
    <slot>{status.replace("_", " ").toUpperCase()}</slot>
  {/if}
</span>

<style>
  .badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;

    font-family: var(--mv-font-mono);
    font-weight: var(--mv-font-weight-bold);
    letter-spacing: var(--mv-letter-spacing-wide);
    text-transform: uppercase;

    border-radius: var(--mv-radius-sm);
    box-shadow: var(--mv-shadow-badge);

    white-space: nowrap;
    transition: var(--mv-transition-base);
  }

  /* Sizes */
  .size-small {
    height: var(--mv-badge-height-sm, 18px);
    padding: 0 var(--mv-spacing-xs);
    font-size: var(--mv-font-size-xxs, 9px);
  }

  .size-medium {
    /* StatusBadge default equivalent */
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    font-size: var(--mv-font-size-xs);
    min-width: var(--mv-badge-min-width);
  }

  /* 
    Status Colors
    Maps to `src/design-system/variables/status.css`
    Uses -bg, -border, -text suffixes
  */

  .status-pending {
    background: var(--mv-color-status-pending-bg);
    border: var(--mv-border-width-thin) solid
      var(--mv-color-status-pending-border);
    color: var(--mv-color-status-pending-text);
  }

  .status-ready {
    background: var(--mv-color-status-ready-bg);
    border: var(--mv-border-width-thin) solid
      var(--mv-color-status-ready-border);
    color: var(--mv-color-status-ready-text);
    box-shadow: var(--mv-shadow-badge-glow-sm) var(--mv-color-status-ready-text);
  }

  .status-running {
    background: var(--mv-color-status-running-bg);
    border: var(--mv-border-width-thin) solid
      var(--mv-color-status-running-border);
    color: var(--mv-color-status-running-text);
    box-shadow: var(--mv-shadow-badge-glow-md)
      var(--mv-color-status-running-glow);
    animation: pulse var(--mv-duration-pulse) infinite;
  }

  .status-succeeded {
    background: var(--mv-color-status-succeeded-bg);
    border: var(--mv-border-width-thin) solid
      var(--mv-color-status-succeeded-border);
    color: var(--mv-color-status-succeeded-text);
    box-shadow: var(--mv-shadow-badge-glow-sm)
      var(--mv-color-status-succeeded-border);
  }

  .status-completed {
    background: var(--mv-color-status-completed-bg);
    border: var(--mv-border-width-thin) solid
      var(--mv-color-status-completed-border);
    color: var(--mv-color-status-completed-text);
    box-shadow: var(--mv-shadow-badge-glow-lg)
      var(--mv-color-status-completed-glow);
  }

  .status-failed {
    background: var(--mv-color-status-failed-bg);
    border: var(--mv-border-width-thin) solid
      var(--mv-color-status-failed-border);
    color: var(--mv-color-status-failed-text);
    box-shadow: var(--mv-shadow-badge-glow-sm)
      var(--mv-color-status-failed-border);
  }

  .status-canceled {
    background: var(--mv-color-status-canceled-bg);
    border: var(--mv-border-width-thin) solid
      var(--mv-color-status-canceled-border);
    color: var(--mv-color-status-canceled-text);
  }

  .status-blocked {
    background: var(--mv-color-status-blocked-bg);
    border: var(--mv-border-width-thin) solid
      var(--mv-color-status-blocked-border);
    color: var(--mv-color-status-blocked-text);
  }

  .status-retry-wait {
    background: var(--mv-color-status-retry-wait-bg);
    border: var(--mv-border-width-thin) solid
      var(--mv-color-status-retry-wait-border);
    color: var(--mv-color-status-retry-wait-text);
    animation: pulse-slow 2s infinite;
  }

  @keyframes pulse {
    0% {
      opacity: 1;
    }
    50% {
      opacity: 0.8;
    }
    100% {
      opacity: 1;
    }
  }

  @keyframes pulse-slow {
    0%,
    100% {
      opacity: 1;
    }
    50% {
      opacity: 0.7;
    }
  }
</style>
