<script lang="ts">
  import { toasts } from "../../stores/toastStore";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";

  function handleClick(id: string) {
    toasts.remove(id);
  }
</script>

<div class="toast-container">
  {#each $toasts as toast (toast.id)}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
    <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
    <div
      class="toast {toast.type}"
      animate:flip
      transition:fly={{ y: 20, duration: 300 }}
      role="alert"
      onclick={() => handleClick(toast.id)}
      tabindex="0"
    >
      <div class="message">{toast.message}</div>
    </div>
  {/each}
</div>

<style>
  .toast-container {
    position: fixed;
    bottom: var(--mv-spacing-xl);
    left: var(--mv-toast-left);
    transform: translateX(-50%);
    z-index: var(--mv-z-toast);
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-sm);
    pointer-events: none;
    width: max-content;
    max-width: var(--mv-toast-max-width);
  }

  .toast {
    pointer-events: auto;
    background: var(--mv-glass-bg);
    color: var(--mv-color-text-primary);
    padding: var(--mv-spacing-sm) var(--mv-spacing-lg);
    border-radius: var(--mv-radius-md);
    box-shadow: var(--mv-shadow-card);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border);
    text-align: center;
    cursor: pointer;
    backdrop-filter: blur(8px);
    font-family: var(--mv-font-sans);
    font-size: var(--mv-font-size-sm);
  }

  .toast.info {
    border-color: var(--mv-primitive-frost-2);
  }

  .toast.success {
    border-color: var(--mv-primitive-aurora-green);
  }

  .toast.warning {
    border-color: var(--mv-primitive-aurora-yellow);
  }

  .toast.error {
    border-color: var(--mv-primitive-aurora-red);
    background: var(--mv-bg-glow-red-light);
  }
</style>
