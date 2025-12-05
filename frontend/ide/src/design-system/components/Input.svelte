<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  /**
   * 入力フィールドのタイプ
   */
  export let type: 'text' | 'password' | 'email' | 'number' | 'search' = 'text';

  /**
   * 入力値
   */
  export let value = '';

  /**
   * プレースホルダー
   */
  export let placeholder = '';

  /**
   * ラベル
   */
  export let label = '';

  /**
   * 無効状態
   */
  export let disabled = false;

  /**
   * エラー状態
   */
  export let error = false;

  /**
   * エラーメッセージ
   */
  export let errorMessage = '';

  /**
   * サイズ
   */
  export let size: 'small' | 'medium' | 'large' = 'medium';

  /**
   * 入力フィールドのID（label関連付け用）
   */
  export let id: string = `input-${Math.random().toString(36).slice(2, 9)}`;

  const dispatch = createEventDispatcher();

  function handleInput(event: Event) {
    const target = event.target as HTMLInputElement;
    value = target.value;
    dispatch('input', value);
  }

  function handleChange(event: Event) {
    dispatch('change', value);
  }
</script>

<div class="input-wrapper size-{size}">
  {#if label}
    <label class="label" class:disabled for={id}>
      {label}
    </label>
  {/if}
  <input
    {id}
    {type}
    {value}
    {placeholder}
    {disabled}
    class="input"
    class:error
    class:disabled
    on:input={handleInput}
    on:change={handleChange}
    on:focus
    on:blur
  />
  {#if error && errorMessage}
    <span class="error-message">{errorMessage}</span>
  {/if}
</div>

<style>
  .input-wrapper {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-xxs);
  }

  .label {
    font-family: var(--mv-font-sans);
    font-weight: var(--mv-font-weight-medium);
    color: var(--mv-color-text-secondary);
  }

  .label.disabled {
    color: var(--mv-color-text-disabled);
  }

  .input {
    font-family: var(--mv-font-sans);
    background: var(--mv-color-surface-primary);
    border: 1px solid var(--mv-color-border-default);
    border-radius: var(--mv-radius-md);
    color: var(--mv-color-text-primary);
    transition: var(--mv-transition-hover);
  }

  .input::placeholder {
    color: var(--mv-color-text-muted);
  }

  /* サイズ */
  .size-small .label {
    font-size: var(--mv-font-size-xs);
  }

  .size-small .input {
    height: 28px;
    padding: 0 var(--mv-spacing-xs);
    font-size: var(--mv-font-size-sm);
  }

  .size-medium .label {
    font-size: var(--mv-font-size-sm);
  }

  .size-medium .input {
    height: 36px;
    padding: 0 var(--mv-spacing-sm);
    font-size: var(--mv-font-size-md);
  }

  .size-large .label {
    font-size: var(--mv-font-size-md);
  }

  .size-large .input {
    height: 44px;
    padding: 0 var(--mv-spacing-md);
    font-size: var(--mv-font-size-lg);
  }

  /* 状態 */
  .input:hover:not(.disabled) {
    border-color: var(--mv-color-border-strong);
  }

  .input:focus {
    outline: none;
    border-color: var(--mv-color-border-focus);
    box-shadow: 0 0 0 2px rgba(76, 175, 80, 0.2);
  }

  .input.error {
    border-color: var(--mv-color-status-failed-border);
  }

  .input.error:focus {
    box-shadow: 0 0 0 2px rgba(244, 67, 54, 0.2);
  }

  .input.disabled {
    opacity: 0.5;
    cursor: not-allowed;
    background: var(--mv-color-surface-secondary);
  }

  .error-message {
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-status-failed-text);
  }
</style>
