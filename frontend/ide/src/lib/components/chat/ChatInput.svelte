<script lang="ts">
  import BrandLogo from "../brand/BrandLogo.svelte";

  interface Props {
    disabled?: boolean;
    onsend?: (text: string) => void;
  }

  let { disabled = false, onsend }: Props = $props();

  let value = $state("");

  function handleSend() {
    if (value.trim() && !disabled) {
      onsend?.(value);
      value = "";
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  }
</script>

<div class="chat-input-container">
  <div class="input-wrapper">
    <textarea
      bind:value
      placeholder="Ask Multiverse..."
      class="transparent-input"
      class:disabled
      onkeydown={handleKeydown}
      rows="1"
      {disabled}
    ></textarea>
  </div>
  <button
    class="send-btn"
    onclick={handleSend}
    disabled={disabled || !value.trim()}
    aria-label="Send"
  >
    <div class="logo-wrapper">
      <BrandLogo size="sm" />
    </div>
  </button>
</div>

<style>
  .chat-input-container {
    display: flex;
    align-items: flex-end;
    gap: var(--mv-spacing-sm);
    padding: var(--mv-spacing-sm);
    background: transparent;
  }

  .input-wrapper {
    flex: 1;
    display: flex;
    align-items: center;
    background: var(--mv-glass-bg-dark);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
    border-radius: var(--mv-radius-md);
    padding: var(--mv-spacing-xs) var(--mv-spacing-md);
    transition: all var(--mv-duration-fast);
  }

  .input-wrapper:focus-within {
    background: var(--mv-glass-bg-darker);
    border-color: var(--mv-glass-border-hover);
    box-shadow: var(--mv-shadow-ambient-sm);
  }

  .transparent-input {
    width: 100%;
    background: transparent;
    border: none;
    color: var(--mv-color-text-primary);
    font-family: var(--mv-font-sans);
    font-size: var(--mv-font-size-sm);
    outline: none;
    resize: none;
    display: block;
    line-height: var(--mv-line-height-relaxed);
    min-height: var(--mv-min-height-input);
    max-height: var(--mv-max-height-input);
    padding: 0;
  }

  .transparent-input::placeholder {
    color: var(--mv-color-text-disabled);
    opacity: 0.5;
    font-style: italic;
  }

  .transparent-input.disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .send-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: var(--mv-size-send-btn);
    height: var(--mv-size-send-btn);
    padding: 0;
    background: transparent;
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
    border-radius: var(--mv-radius-circle);
    cursor: pointer;
    transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
    flex-shrink: 0;
  }

  .send-btn:not(:disabled):hover {
    background: var(--mv-glass-hover);
    transform: scale(1.1);
    box-shadow: var(--mv-shadow-glow-accent);
    border-color: var(--mv-shadow-glow-accent-border);
  }

  .send-btn:not(:disabled):hover .logo-wrapper {
    filter: var(--mv-shadow-glow-accent-strong);
  }

  .send-btn:active {
    transform: scale(0.95);
    background: var(--mv-glass-active);
  }

  .send-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
    background: transparent;
    border-color: transparent;
  }

  .logo-wrapper {
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.3s ease;
    opacity: 0.8;
  }
</style>
