<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { fade } from "svelte/transition";

  export let type: "text" | "password" | "search" | "email" = "text";
  export let value = "";
  export let placeholder = "";
  export let label = "";
  export let error = "";
  export let disabled = false;
  export let autofocus = false;
  export let id = "";

  const dispatch = createEventDispatcher();

  function handleInput(event: Event) {
    const target = event.target as HTMLInputElement;
    value = target.value;
    dispatch("input", event);
  }

  function handleChange(event: Event) {
    dispatch("change", event);
  }

  function handleKeydown(event: KeyboardEvent) {
    dispatch("keydown", event);
    if (event.key === "Enter") {
      dispatch("submit");
    }
  }
</script>

<div class="input-wrapper" class:has-error={!!error} class:disabled>
  {#if label}
    <label for={id} class="label">{label}</label>
  {/if}

  <div class="input-container">
    <!-- svelte-ignore a11y-autofocus -->
    <input
      {id}
      {type}
      {value}
      {placeholder}
      {disabled}
      {autofocus}
      class="input"
      on:input={handleInput}
      on:change={handleChange}
      on:keydown={handleKeydown}
      on:focus
      on:blur
    />
  </div>

  {#if error}
    <p class="error-message" transition:fade={{ duration: 150 }}>{error}</p>
  {/if}
</div>

<style>
  .input-wrapper {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-xs);
    width: 100%;
  }

  .label {
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-secondary);
    font-weight: var(--mv-font-weight-medium);
  }

  .input-container {
    position: relative;
    width: 100%;
  }

  .input {
    width: 100%;
    height: var(--mv-input-height-md);
    padding: 0 var(--mv-spacing-md);
    font-family: var(--mv-font-sans);
    font-size: var(--mv-font-size-md);
    color: var(--mv-color-text-primary);

    background: var(--mv-color-surface-secondary);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    border-radius: var(--mv-radius-md);

    transition: var(--mv-transition-base);
  }

  .input:hover:not(:disabled) {
    border-color: var(--mv-color-border-strong);
    background: var(--mv-color-surface-hover);
  }

  .input:focus {
    outline: none;
    border-color: var(--mv-color-interactive-primary);
    background: var(--mv-color-surface-primary);
    box-shadow: var(--mv-shadow-focus);
  }

  /* Disabled State */
  .disabled .input {
    opacity: 0.6;
    cursor: not-allowed;
    background: var(--mv-color-surface-primary);
  }

  /* Error State */
  .has-error .input {
    border-color: var(--mv-color-status-failed-border);
    background: var(--mv-bg-glow-red-lighter);
  }

  .has-error .input:focus {
    border-color: var(--mv-color-status-failed-border);
    box-shadow: var(--mv-shadow-focus-error);
  }

  .has-error .label {
    color: var(--mv-color-status-failed-text);
  }

  .error-message {
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-status-failed-text);
    margin: 0;
    padding-left: var(--mv-spacing-xs);
  }
</style>
