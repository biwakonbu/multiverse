<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  /**
   * ボタンのバリアント
   * - primary: メインアクション用（目立つ緑）
   * - secondary: 補助アクション用（控えめ）
   * - ghost: テキストのみ（背景なし）
   * - danger: 破壊的アクション用（赤）
   */
  export let variant: 'primary' | 'secondary' | 'ghost' | 'danger' = 'primary';

  /**
   * ボタンのサイズ
   */
  export let size: 'small' | 'medium' | 'large' = 'medium';

  /**
   * 無効状態
   */
  export let disabled = false;

  /**
   * ローディング状態
   */
  export let loading = false;

  /**
   * ボタンのタイプ
   */
  export let type: 'button' | 'submit' | 'reset' = 'button';

  /**
   * ボタンのラベル（slot の代替、Storybook用）
   */
  export let label = '';

  /**
   * ローディング時のラベル
   */
  export let loadingLabel = '';

  /**
   * ツールチップ用のタイトル属性
   */
  export let title = '';

  const dispatch = createEventDispatcher();

  function handleClick(event: MouseEvent) {
    if (!disabled && !loading) {
      dispatch('click', event);
    }
  }

  $: isDisabled = disabled || loading;
</script>

<button
  {type}
  {title}
  disabled={isDisabled}
  class="button variant-{variant} size-{size}"
  class:disabled={isDisabled}
  class:loading
  on:click={handleClick}
>
  {#if loading}
    <span class="spinner"></span>
    {#if loadingLabel}
      {loadingLabel}
    {:else if label}
      {label}
    {:else}
      <slot />
    {/if}
  {:else if label}
    {label}
  {:else}
    <slot />
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
  transition: var(--mv-transition-hover), transform 120ms ease;
  white-space: nowrap;
  background: var(--btn-bg, var(--mv-color-surface-secondary));
  color: var(--btn-text, var(--mv-color-text-primary));
  box-shadow: var(--btn-shadow, none);
}

.button:hover:not(.disabled) {
  background: var(--btn-bg-hover, var(--btn-bg, var(--mv-color-surface-secondary)));
  border-color: var(--btn-border-hover, var(--btn-border, var(--mv-color-border-default)));
  color: var(--btn-text-hover, var(--btn-text, var(--mv-color-text-primary)));
  box-shadow: var(--btn-shadow-hover, var(--btn-shadow, none));
}

.button:active:not(.disabled) {
  background: var(--btn-bg-active, var(--btn-bg-hover, var(--btn-bg, var(--mv-color-surface-secondary))));
  border-color: var(--btn-border-active, var(--btn-border-hover, var(--btn-border, var(--mv-color-border-default))));
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
    --btn-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.08),
      inset 0 -1px 0 rgba(34, 48, 56, 0.25),
      0 8px 18px -10px rgba(143, 191, 187, 0.35);
    --btn-bg-hover: var(--mv-primitive-frost-1);
    --btn-border-hover: var(--mv-primitive-frost-1);
    --btn-shadow-hover: inset 0 1px 0 rgba(255, 255, 255, 0.12),
      inset 0 -1px 0 rgba(34, 48, 56, 0.28),
      0 10px 22px -12px rgba(136, 192, 208, 0.45);
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
    --btn-border: var(--mv-color-border-subtle);
    --btn-text: var(--mv-color-text-secondary);
    --btn-bg-hover: var(--mv-color-surface-hover);
    --btn-border-hover: var(--mv-color-border-default);
    --btn-text-hover: var(--mv-color-text-primary);
    --btn-bg-active: var(--mv-color-surface-primary);
  }

  /* Danger バリアント */
  .variant-danger {
    --btn-bg: var(--mv-color-status-failed-border);
    --btn-border: var(--mv-color-status-failed-border);
    --btn-text: var(--mv-color-text-primary);
    --btn-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.06),
      inset 0 -1px 0 rgba(58, 28, 28, 0.3),
      0 8px 18px -12px rgba(201, 123, 123, 0.45);
    --btn-bg-hover: var(--mv-color-status-failed-text);
    --btn-border-hover: var(--mv-color-status-failed-text);
    --btn-shadow-hover: inset 0 1px 0 rgba(255, 255, 255, 0.08),
      inset 0 -1px 0 rgba(58, 28, 28, 0.35),
      0 10px 20px -12px rgba(218, 126, 135, 0.5);
    --btn-bg-active: var(--mv-primitive-pastel-red);
    --btn-border-active: var(--mv-primitive-pastel-red);
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

  /* スピナー */
  .spinner {
    width: var(--mv-icon-size-xs);
    height: var(--mv-icon-size-xs);
    border: var(--mv-border-width-default) solid transparent;
    border-top-color: currentColor;
    border-radius: var(--mv-radius-full);
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
