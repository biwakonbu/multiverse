<!--
  StatusIndicator - タスクステータスを示すインジケーター

  7つのステータスに対応し、running 状態ではパルスアニメーションを表示。
  GridNode 内で使用される小さなコンポーネント。
-->
<script lang="ts">
  import type { StatusKey } from '../../design-system/tokens/colors';

  // ステータス（必須）
  export let status: StatusKey = 'pending';

  // サイズ
  export let size: 'small' | 'medium' | 'large' = 'medium';

  // ラベル表示（アクセシビリティ）
  export let showLabel = false;

  // ステータスラベルのマッピング
  const statusLabels: Record<StatusKey, string> = {
    pending: '待機中',
    ready: '準備完了',
    running: '実行中',
    succeeded: '成功',
    failed: '失敗',
    canceled: 'キャンセル',
    blocked: 'ブロック',
  };
</script>

<div
  class="status-indicator status-{status} size-{size}"
  role="status"
  aria-label={statusLabels[status]}
>
  <span class="dot">
    {#if status === 'running'}
      <span class="pulse-ring"></span>
    {/if}
  </span>
  {#if showLabel}
    <span class="label">{statusLabels[status]}</span>
  {/if}
</div>

<style>
  .status-indicator {
    display: inline-flex;
    align-items: center;
    gap: var(--mv-spacing-xs, 8px);
  }

  .dot {
    position: relative;
    border-radius: 50%;
    flex-shrink: 0;
  }

  /* サイズバリエーション */
  .size-small .dot {
    width: 8px;
    height: 8px;
  }

  .size-medium .dot {
    width: 12px;
    height: 12px;
  }

  .size-large .dot {
    width: 16px;
    height: 16px;
  }

  /* ステータス別カラー */
  .status-pending .dot {
    background: var(--mv-color-status-pending-text, #ff9800);
  }

  .status-ready .dot {
    background: var(--mv-color-status-ready-text, #5588ff);
  }

  .status-running .dot {
    background: var(--mv-color-status-running-text, #4caf50);
  }

  .status-succeeded .dot {
    background: var(--mv-color-status-succeeded-text, #228822);
  }

  .status-failed .dot {
    background: var(--mv-color-status-failed-text, #f44336);
  }

  .status-canceled .dot {
    background: var(--mv-color-status-canceled-text, #888888);
  }

  .status-blocked .dot {
    background: var(--mv-color-status-blocked-text, #ccaa22);
  }

  /* パルスリングアニメーション */
  .pulse-ring {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    background: var(--mv-color-status-running-text, #4caf50);
    animation: pulse-ring var(--mv-animation-pulse-duration, 2000ms) ease-in-out infinite;
  }

  @keyframes pulse-ring {
    0%, 100% {
      transform: scale(1);
      opacity: 1;
    }
    50% {
      transform: scale(1.8);
      opacity: 0;
    }
  }

  /* ラベル */
  .label {
    font-size: var(--mv-font-size-sm, 12px);
    color: var(--mv-color-text-secondary, #aaaaaa);
  }

  .size-small .label {
    font-size: var(--mv-font-size-xs, 10px);
  }

  .size-large .label {
    font-size: var(--mv-font-size-md, 14px);
  }

  /* ステータス別ラベルカラー */
  .status-pending .label { color: var(--mv-color-status-pending-text, #ff9800); }
  .status-ready .label { color: var(--mv-color-status-ready-text, #5588ff); }
  .status-running .label { color: var(--mv-color-status-running-text, #4caf50); }
  .status-succeeded .label { color: var(--mv-color-status-succeeded-text, #228822); }
  .status-failed .label { color: var(--mv-color-status-failed-text, #f44336); }
  .status-canceled .label { color: var(--mv-color-status-canceled-text, #888888); }
  .status-blocked .label { color: var(--mv-color-status-blocked-text, #ccaa22); }
</style>
