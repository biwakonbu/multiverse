<script lang="ts">
  /**
   * タスクステータスを示すバッジ
   * デザインシステムのステータスカラーを使用
   */
  export let status: 'pending' | 'ready' | 'running' | 'succeeded' | 'failed' | 'canceled' | 'blocked' = 'pending';

  /**
   * バッジのサイズ
   */
  export let size: 'small' | 'medium' = 'medium';

  /**
   * ラベル（省略時はステータス名を表示）
   */
  export let label = '';

  const statusLabels: Record<typeof status, string> = {
    pending: 'Pending',
    ready: 'Ready',
    running: 'Running',
    succeeded: 'Succeeded',
    failed: 'Failed',
    canceled: 'Canceled',
    blocked: 'Blocked'
  };

  $: displayLabel = label || statusLabels[status];
</script>

<span class="badge status-{status} size-{size}">
  {#if status === 'running'}
    <span class="pulse-dot"></span>
  {/if}
  {displayLabel}
</span>

<style>
  .badge {
    display: inline-flex;
    align-items: center;
    gap: var(--mv-spacing-xxs);
    border-radius: var(--mv-radius-sm);
    font-family: var(--mv-font-sans);
    font-weight: var(--mv-font-weight-medium);
    text-transform: uppercase;
    letter-spacing: var(--mv-letter-spacing-widest);
  }

  /* サイズ */
  .size-small {
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    font-size: var(--mv-font-size-xs);
  }

  .size-medium {
    padding: var(--mv-spacing-xxs) var(--mv-spacing-sm);
    font-size: var(--mv-font-size-sm);
  }

  /* ステータス別スタイル */
  .status-pending {
    background: var(--mv-color-status-pending-bg);
    border: var(--mv-border-width-thin) solid var(--mv-color-status-pending-border);
    color: var(--mv-color-status-pending-text);
  }

  .status-ready {
    background: var(--mv-color-status-ready-bg);
    border: var(--mv-border-width-thin) solid var(--mv-color-status-ready-border);
    color: var(--mv-color-status-ready-text);
  }

  .status-running {
    background: var(--mv-color-status-running-bg);
    border: var(--mv-border-width-thin) solid var(--mv-color-status-running-border);
    color: var(--mv-color-status-running-text);
  }

  .status-succeeded {
    background: var(--mv-color-status-succeeded-bg);
    border: var(--mv-border-width-thin) solid var(--mv-color-status-succeeded-border);
    color: var(--mv-color-status-succeeded-text);
  }

  .status-failed {
    background: var(--mv-color-status-failed-bg);
    border: var(--mv-border-width-thin) solid var(--mv-color-status-failed-border);
    color: var(--mv-color-status-failed-text);
  }

  .status-canceled {
    background: var(--mv-color-status-canceled-bg);
    border: var(--mv-border-width-thin) solid var(--mv-color-status-canceled-border);
    color: var(--mv-color-status-canceled-text);
  }

  .status-blocked {
    background: var(--mv-color-status-blocked-bg);
    border: var(--mv-border-width-thin) solid var(--mv-color-status-blocked-border);
    color: var(--mv-color-status-blocked-text);
  }

  /* パルスドット（Running 状態用） */
  .pulse-dot {
    width: var(--mv-indicator-size-xs);
    height: var(--mv-indicator-size-xs);
    border-radius: var(--mv-radius-full);
    background: var(--mv-color-status-running-text);
    animation: pulse var(--mv-duration-pulse) var(--mv-easing-default) infinite;
  }

  @keyframes pulse {
    0%, 100% {
      opacity: 1;
      transform: scale(1);
    }
    50% {
      opacity: 0.5;
      transform: scale(0.8);
    }
  }
</style>
