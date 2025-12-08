<script lang="ts">
  interface Props {
    /**
     * ワークスペースを開くボタン
     * VSCode/Cursor 風のクリーンでミニマルなデザイン
     */
    loading?: boolean;
    disabled?: boolean;
    onclick?: () => void;
  }

  let { loading = false, disabled = false, onclick }: Props = $props();

  function handleClick() {
    if (!loading && !disabled) {
      onclick?.();
    }
  }

  let isDisabled = $derived(disabled || loading);
</script>

<button
  type="button"
  class="open-workspace-button"
  class:loading
  class:disabled={isDisabled}
  disabled={isDisabled}
  onclick={handleClick}
>
  {#if loading}
    <svg class="spinner" viewBox="0 0 24 24" fill="none">
      <circle
        cx="12"
        cy="12"
        r="10"
        stroke="currentColor"
        stroke-width="2"
        opacity="0.25"
      />
      <path
        d="M12 2a10 10 0 0 1 10 10"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
      />
    </svg>
  {:else}
    <svg class="icon" viewBox="0 0 24 24" fill="none">
      <path
        d="M3 7V17C3 18.1046 3.89543 19 5 19H19C20.1046 19 21 18.1046 21 17V9C21 7.89543 20.1046 7 19 7H12L10 5H5C3.89543 5 3 5.89543 3 7Z"
        stroke="currentColor"
        stroke-width="1.5"
        stroke-linejoin="round"
      />
    </svg>
  {/if}
  <span class="label">
    {#if loading}
      Opening...
    {:else}
      Workspaceを開く
    {/if}
  </span>
</button>

<style>
  .open-workspace-button {
    display: inline-flex;
    align-items: center;
    gap: var(--mv-spacing-sm);
    padding: var(--mv-spacing-sm) var(--mv-spacing-lg);
    background: transparent;
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    border-radius: var(--mv-radius-md);
    color: var(--mv-color-text-secondary);
    font-family: var(--mv-font-sans);
    font-size: var(--mv-font-size-md);
    cursor: pointer;
    transition: all var(--mv-duration-fast) var(--mv-easing-out);
  }

  .open-workspace-button:hover:not(.disabled) {
    background: var(--mv-color-surface-hover);
    border-color: var(--mv-color-border-strong);
    color: var(--mv-color-text-primary);
  }

  .open-workspace-button:active:not(.disabled) {
    background: var(--mv-color-surface-selected);
  }

  .open-workspace-button:focus-visible {
    outline: var(--mv-focus-ring-width) solid var(--mv-color-border-focus);
    outline-offset: var(--mv-focus-ring-offset);
  }

  .icon {
    width: var(--mv-icon-size-md);
    height: var(--mv-icon-size-md);
    flex-shrink: 0;
  }

  .spinner {
    width: var(--mv-icon-size-md);
    height: var(--mv-icon-size-md);
    flex-shrink: 0;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .label {
    font-weight: var(--mv-font-weight-normal);
  }

  .disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
