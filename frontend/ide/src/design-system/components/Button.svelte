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
   * ボタンのタイプ
   */
  export let type: 'button' | 'submit' | 'reset' = 'button';

  /**
   * ボタンのラベル（slot の代替、Storybook用）
   */
  export let label = '';

  const dispatch = createEventDispatcher();

  function handleClick(event: MouseEvent) {
    if (!disabled) {
      dispatch('click', event);
    }
  }
</script>

<button
  {type}
  {disabled}
  class="button variant-{variant} size-{size}"
  class:disabled
  on:click={handleClick}
>
  {#if label}
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
    border: 1px solid transparent;
    border-radius: var(--mv-radius-md);
    font-family: var(--mv-font-sans);
    font-weight: var(--mv-font-weight-medium);
    cursor: pointer;
    transition: var(--mv-transition-hover);
    white-space: nowrap;
  }

  /* サイズ */
  .size-small {
    height: 28px;
    padding: 0 var(--mv-spacing-sm);
    font-size: var(--mv-font-size-sm);
  }

  .size-medium {
    height: 36px;
    padding: 0 var(--mv-spacing-md);
    font-size: var(--mv-font-size-md);
  }

  .size-large {
    height: 44px;
    padding: 0 var(--mv-spacing-lg);
    font-size: var(--mv-font-size-lg);
  }

  /* Primary バリアント */
  .variant-primary {
    background: var(--mv-color-status-running-border);
    border-color: var(--mv-color-status-running-border);
    color: var(--mv-color-surface-app);
  }

  .variant-primary:hover:not(.disabled) {
    background: var(--mv-color-status-running-text);
    border-color: var(--mv-color-status-running-text);
  }

  /* Secondary バリアント */
  .variant-secondary {
    background: var(--mv-color-surface-secondary);
    border-color: var(--mv-color-border-default);
    color: var(--mv-color-text-primary);
  }

  .variant-secondary:hover:not(.disabled) {
    background: var(--mv-color-surface-hover);
    border-color: var(--mv-color-border-strong);
  }

  /* Ghost バリアント */
  .variant-ghost {
    background: transparent;
    border-color: transparent;
    color: var(--mv-color-text-secondary);
  }

  .variant-ghost:hover:not(.disabled) {
    background: var(--mv-color-surface-hover);
    color: var(--mv-color-text-primary);
  }

  /* Danger バリアント */
  .variant-danger {
    background: var(--mv-color-status-failed-border);
    border-color: var(--mv-color-status-failed-border);
    color: var(--mv-color-text-primary);
  }

  .variant-danger:hover:not(.disabled) {
    background: var(--mv-color-status-failed-text);
    border-color: var(--mv-color-status-failed-text);
  }

  /* 無効状態 */
  .disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* フォーカス状態 */
  .button:focus-visible {
    outline: 2px solid var(--mv-color-border-focus);
    outline-offset: 2px;
  }
</style>
