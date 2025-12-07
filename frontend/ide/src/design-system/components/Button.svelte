<script lang="ts" module>
  export interface Props {
    /** ボタンのバリアント */
    variant?: "primary" | "secondary" | "ghost" | "danger" | "crystal";
    /** ボタンのサイズ */
    size?: "small" | "medium" | "large";
    /** 無効状態 */
    disabled?: boolean;
    /** ローディング状態 */
    loading?: boolean;
    /** ボタンのタイプ */
    type?: "button" | "submit" | "reset";
    /** ボタンのラベル（Storybook用） */
    label?: string;
    /** ローディング時のラベル */
    loadingLabel?: string;
    /** ツールチップ用のタイトル属性 */
    title?: string;
    /** クリックイベントのコールバック */
    onclick?: ((event: MouseEvent) => void) | undefined;
    children?: import('svelte').Snippet;
  }
</script>

<script lang="ts">
  import Spinner from "./Spinner.svelte";

  let {
    variant = "primary",
    size = "medium",
    disabled = false,
    loading = false,
    type = "button",
    label = "",
    loadingLabel = "",
    title = "",
    onclick = undefined,
    children
  }: Props = $props();

  function handleClick(event: MouseEvent) {
    if (!disabled && !loading) {
      onclick?.(event);
    }
  }

  let isDisabled = $derived(disabled || loading);
</script>

<button
  {type}
  {title}
  disabled={isDisabled}
  class="button variant-{variant} size-{size}"
  class:disabled={isDisabled}
  class:loading
  class:crystal-glow={variant === "crystal"}
  onclick={handleClick}
>
  {#if loading}
    <Spinner size={size === "small" ? "xs" : "sm"} />
    {#if loadingLabel}
      {loadingLabel}
    {:else if label}
      {label}
    {:else}
      {@render children?.()}
    {/if}
  {:else if label}
    {label}
  {:else}
    {@render children?.()}
  {/if}
</button>

<style>
  .button {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: var(--mv-spacing-xs);
    border: var(--mv-border-width-thin) solid var(--btn-border, transparent);
    border-radius: var(--mv-radius-md);
    font-family: var(--mv-font-sans);
    font-weight: var(--mv-font-weight-medium);
    cursor: pointer;
    transition:
      var(--mv-transition-hover),
      transform 120ms ease,
      box-shadow 200ms ease;
    white-space: nowrap;
    background: var(--btn-bg, var(--mv-color-surface-secondary));
    color: var(--btn-text, var(--mv-color-text-primary));
    box-shadow: var(--btn-shadow, none);
  }

  .button:hover:not(.disabled) {
    background: var(
      --btn-bg-hover,
      var(--btn-bg, var(--mv-color-surface-secondary))
    );
    border-color: var(
      --btn-border-hover,
      var(--btn-border, var(--mv-color-border-default))
    );
    color: var(--btn-text-hover, var(--btn-text, var(--mv-color-text-primary)));
    box-shadow: var(--btn-shadow-hover, var(--btn-shadow, none));
  }

  .button:active:not(.disabled) {
    background: var(
      --btn-bg-active,
      var(--btn-bg-hover, var(--btn-bg, var(--mv-color-surface-secondary)))
    );
    border-color: var(
      --btn-border-active,
      var(--btn-border-hover, var(--btn-border, var(--mv-color-border-default)))
    );
    transform: translateY(1px);
  }

  /* サイズ */
  .size-small {
    height: var(--mv-input-height-sm);
    padding: 0 var(--mv-spacing-sm);
    font-size: var(--mv-font-size-sm);
  }

  .size-medium {
    height: var(--mv-input-height-md);
    padding: 0 var(--mv-spacing-md);
    font-size: var(--mv-font-size-md);
  }

  .size-large {
    height: var(--mv-input-height-lg);
    padding: 0 var(--mv-spacing-lg);
    font-size: var(--mv-font-size-lg);
  }

  /* Primary バリアント */
  .variant-primary {
    --btn-bg: var(--mv-primitive-frost-2);
    --btn-border: var(--mv-primitive-frost-3);
    --btn-text: var(--mv-color-text-primary);
    --btn-shadow: var(--mv-btn-shadow-primary);
    --btn-bg-hover: var(--mv-primitive-frost-1);
    --btn-border-hover: var(--mv-primitive-frost-1);
    --btn-shadow-hover: var(--mv-btn-shadow-primary-hover);
    --btn-bg-active: var(--mv-primitive-frost-3);
    --btn-border-active: var(--mv-primitive-frost-3);
  }

  /* Secondary バリアント */
  .variant-secondary {
    --btn-bg: var(--mv-color-surface-secondary);
    --btn-border: var(--mv-color-border-default);
    --btn-text: var(--mv-color-text-primary);
    --btn-shadow: 0 0 0 1px var(--mv-color-border-default);
    --btn-bg-hover: var(--mv-color-surface-hover);
    --btn-border-hover: var(--mv-color-border-strong);
    --btn-shadow-hover: 0 0 0 1px var(--mv-color-border-strong);
    --btn-bg-active: var(--mv-color-surface-primary);
  }

  /* Ghost バリアント */
  .variant-ghost {
    --btn-bg: transparent;
    --btn-border: transparent;
    --btn-text: var(--mv-color-text-secondary);
    --btn-bg-hover: var(--mv-glass-hover);
    --btn-border-hover: transparent;
    --btn-text-hover: var(--mv-color-text-primary);
    --btn-bg-active: var(--mv-glass-active);
  }

  /* Danger バリアント */
  .variant-danger {
    --btn-bg: var(--mv-color-status-failed-border);
    --btn-border: var(--mv-color-status-failed-border);
    --btn-text: var(--mv-color-text-primary);
    --btn-shadow: var(--mv-btn-shadow-danger);
    --btn-bg-hover: var(--mv-color-status-failed-text);
    --btn-border-hover: var(--mv-color-status-failed-text);
    --btn-shadow-hover: var(--mv-btn-shadow-danger-hover);
    --btn-bg-active: var(--mv-primitive-pastel-red);
    --btn-border-active: var(--mv-primitive-pastel-red);
  }

  /* Crystal Variant (New) */
  .variant-crystal {
    --btn-bg: var(--mv-btn-crystal-bg);
    --btn-border: var(--mv-btn-crystal-border);
    --btn-text: var(--mv-primitive-frost-nord9);
    --btn-shadow: var(--mv-btn-crystal-shadow);

    --btn-bg-hover: var(--mv-btn-crystal-bg-hover);
    --btn-border-hover: var(--mv-primitive-frost-nord9);
    --btn-text-hover: var(--mv-primitive-frost-nord8);
    --btn-shadow-hover: var(--mv-btn-crystal-shadow-hover);

    --btn-bg-active: var(--mv-btn-crystal-bg-active);
    background: var(--btn-bg);
    backdrop-filter: blur(10px);
  }

  .variant-crystal:hover {
    text-shadow: var(--mv-btn-crystal-text-shadow);
  }

  /* 無効状態 */
  .disabled {
    opacity: 0.6;
    cursor: not-allowed;
    background: var(--mv-color-surface-secondary);
    border-color: var(--mv-color-border-subtle);
    color: var(--mv-color-text-disabled);
    box-shadow: none;
  }

  /* フォーカス状態 */
  .button:focus-visible {
    outline: var(--mv-focus-ring-width) solid var(--mv-color-border-focus);
    outline-offset: var(--mv-focus-ring-offset);
  }
</style>
