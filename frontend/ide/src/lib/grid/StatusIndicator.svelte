<!--
  StatusIndicator - タスクステータスを示すインジケーター

  7つのステータスに対応し、running 状態ではパルスアニメーションを表示。
  GridNode 内で使用される小さなコンポーネント。
-->
<script lang="ts">
  import type { StatusKey } from "../../design-system/tokens/colors";

  

  

  
  interface Props {
    // ステータス（必須）
    status?: StatusKey;
    // サイズ
    size?: "small" | "medium" | "large";
    // ラベル表示（アクセシビリティ）
    showLabel?: boolean;
  }

  let { status = "pending", size = "medium", showLabel = false }: Props = $props();

  // ステータスラベルのマッピング
  const statusLabels: Record<StatusKey, string> = {
    pending: "待機中",
    ready: "準備完了",
    running: "実行中",
    succeeded: "成功",
    completed: "完了",
    failed: "失敗",
    canceled: "キャンセル",
    blocked: "ブロック",
    retryWait: "リトライ待機",
  };
</script>

<div
  class="status-indicator status-{status === 'retryWait'
    ? 'retry-wait'
    : status} size-{size}"
  role="status"
  aria-label={statusLabels[status]}
>
  <span class="dot">
    {#if status === "running"}
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
    border-radius: var(--mv-radius-full);
    flex-shrink: 0;
  }

  /* サイズバリエーション */
  .size-small .dot {
    width: var(--mv-indicator-size-sm);
    height: var(--mv-indicator-size-sm);
  }

  .size-medium .dot {
    width: var(--mv-indicator-size-md);
    height: var(--mv-indicator-size-md);
  }

  .size-large .dot {
    width: var(--mv-indicator-size-lg);
    height: var(--mv-indicator-size-lg);
  }

  /* ステータス別カラー */
  .status-pending .dot {
    background: var(--mv-color-status-pending-text);
  }

  .status-ready .dot {
    background: var(--mv-color-status-ready-text);
  }

  .status-running .dot {
    background: var(--mv-color-status-running-text);
  }

  .status-succeeded .dot {
    background: var(--mv-color-status-succeeded-text);
  }

  .status-failed .dot {
    background: var(--mv-color-status-failed-text);
  }

  .status-canceled .dot {
    background: var(--mv-color-status-canceled-text);
  }

  .status-blocked .dot {
    background: var(--mv-color-status-blocked-text);
  }

  .status-completed .dot {
    background: var(--mv-color-status-completed-text);
  }

  .status-retry-wait .dot {
    background: var(--mv-color-status-retry-wait-text);
  }

  /* パルスリングアニメーション */
  .pulse-ring {
    position: absolute;
    inset: 0;
    border-radius: var(--mv-radius-full);
    background: var(--mv-color-status-running-text);
    animation: pulse-ring var(--mv-duration-pulse) ease-in-out infinite;
  }

  @keyframes pulse-ring {
    0%,
    100% {
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
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-secondary);
  }

  .size-small .label {
    font-size: var(--mv-font-size-xs);
  }

  .size-large .label {
    font-size: var(--mv-font-size-md);
  }

  /* ステータス別ラベルカラー */
  .status-pending .label {
    color: var(--mv-color-status-pending-text);
  }
  .status-ready .label {
    color: var(--mv-color-status-ready-text);
  }
  .status-running .label {
    color: var(--mv-color-status-running-text);
  }
  .status-succeeded .label {
    color: var(--mv-color-status-succeeded-text);
  }
  .status-completed .label {
    color: var(--mv-color-status-completed-text);
  }
  .status-failed .label {
    color: var(--mv-color-status-failed-text);
  }
  .status-canceled .label {
    color: var(--mv-color-status-canceled-text);
  }
  .status-blocked .label {
    color: var(--mv-color-status-blocked-text);
  }
  .status-retry-wait .label {
    color: var(--mv-color-status-retry-wait-text);
  }
</style>
