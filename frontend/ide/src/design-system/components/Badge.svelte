<script lang="ts">
  /**
   * バッジコンポーネント
   * ステータス表示やラベル付けに使用
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
    | "retryWait"
    | undefined = undefined;

  export let variant: "default" | "outline" | "glass" = "default";
  export let color:
    | "primary"
    | "secondary"
    | "success"
    | "warning"
    | "danger"
    | "info"
    | "neutral" = "neutral";

  export let size: "small" | "medium" = "medium";
  export let label = "";

  // ステータスからカラーとラベルを自動解決
  const statusConfig: Record<
    NonNullable<typeof status>,
    { color: typeof color; label: string }
  > = {
    pending: { color: "warning", label: "Pending" },
    ready: { color: "info", label: "Ready" },
    running: { color: "success", label: "Running" },
    succeeded: { color: "primary", label: "Succeeded" },
    completed: { color: "primary", label: "Completed" },
    failed: { color: "danger", label: "Failed" },
    canceled: { color: "neutral", label: "Canceled" },
    blocked: { color: "secondary", label: "Blocked" },
    retryWait: { color: "warning", label: "Retry Wait" },
  };

  $: resolvedColor = status ? statusConfig[status].color : color;
  $: resolvedLabel = label || (status ? statusConfig[status].label : "");
</script>

<span
  class="badge variant-{variant} color-{resolvedColor} size-{size}"
  class:pulse={status === "running"}
>
  {#if status === "running"}
    <span class="pulse-dot"></span>
  {/if}
  {resolvedLabel}
  <slot />
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
    line-height: 1;
    white-space: nowrap;
  }

  /* サイズ */
  .size-small {
    padding: var(--mv-spacing-xxxs) var(--mv-spacing-xs);
    font-size: var(--mv-font-size-xxs);
  }

  .size-medium {
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    font-size: var(--mv-font-size-xs);
  }

  /* カラーマッピング (CSS変数) */
  .color-primary {
    --badge-bg: var(--mv-primitive-frost-3);
    --badge-border: var(--mv-primitive-frost-1);
    --badge-text: var(--mv-primitive-frost-1);
    --badge-glow: var(--mv-color-glow-focus);
  }
  .color-secondary {
    --badge-bg: var(--mv-color-surface-secondary);
    --badge-border: var(--mv-color-border-subtle);
    --badge-text: var(--mv-color-text-secondary);
  }
  .color-success {
    --badge-bg: var(--mv-color-status-running-bg);
    --badge-border: var(--mv-primitive-aurora-green);
    --badge-text: var(--mv-primitive-aurora-green);
    --badge-glow: var(--mv-color-glow-running);
  }
  .color-warning {
    --badge-bg: var(--mv-color-status-pending-bg);
    --badge-border: var(--mv-primitive-aurora-yellow);
    --badge-text: var(--mv-primitive-aurora-yellow);
  }
  .color-danger {
    --badge-bg: var(--mv-color-status-failed-bg);
    --badge-border: var(--mv-primitive-aurora-red);
    --badge-text: var(--mv-primitive-aurora-red);
  }
  .color-info {
    --badge-bg: var(--mv-color-status-ready-bg);
    --badge-border: var(--mv-primitive-frost-2);
    --badge-text: var(--mv-primitive-frost-2);
  }
  .color-neutral {
    --badge-bg: var(--mv-color-surface-secondary);
    --badge-border: var(--mv-color-border-subtle);
    --badge-text: var(--mv-color-text-muted);
  }

  /* バリアント */

  /* Default: Flat/Solid-ish */
  .variant-default {
    background: var(--badge-bg);
    border: var(--mv-border-width-thin) solid var(--badge-border);
    color: var(--badge-text);
  }

  /* Outline: Transparent bg */
  .variant-outline {
    background: transparent;
    border: var(--mv-border-width-thin) solid var(--badge-border);
    color: var(--badge-text);
  }

  /* Glass: Rich transparent Look */
  .variant-glass {
    background: var(
      --badge-bg
    ); /* Use same semi-transparent base but maybe add blur? */
    background: color-mix(
      in srgb,
      var(--badge-bg),
      transparent 50%
    ); /* make it subtler */
    border: var(--mv-border-width-thin) solid color-mix(in srgb, var(--badge-border), transparent 30%);
    color: var(--badge-text);
    box-shadow: var(--mv-shadow-badge-glow-lg) var(--badge-bg); /* subtle glow */
    backdrop-filter: blur(4px);
  }

  /* パルスドット */
  .pulse-dot {
    width: var(--mv-status-dot-size);
    height: var(--mv-status-dot-size);
    border-radius: var(--mv-radius-full);
    background-color: var(--badge-text);
    animation: pulse 1.5s infinite;
  }

  @keyframes pulse {
    0% {
      opacity: 1;
      transform: scale(1);
    }
    50% {
      opacity: 0.4;
      transform: scale(0.8);
    }
    100% {
      opacity: 1;
      transform: scale(1);
    }
  }
</style>
