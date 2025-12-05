<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import Button from "../../../design-system/components/Button.svelte";
  import Input from "../../../design-system/components/Input.svelte";

  const dispatch = createEventDispatcher<{ send: string }>();

  let value = "";

  function handleSend() {
    if (value.trim()) {
      dispatch("send", value);
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
  <span class="prompt-icon">&gt;</span>
  <div class="input-wrapper">
    <textarea
      bind:value
      placeholder="何か話す..."
      class="transparent-input"
      on:keydown={handleKeydown}
      rows="2"
    ></textarea>
  </div>
</div>

<style>
  .chat-input-container {
    display: flex;
    align-items: flex-start; /* Align to top for multi-line */
    gap: 8px;
    padding: 8px; /* Slightly more padding */
    background: rgba(0, 0, 0, 0.4);
    border-top: 1px solid rgba(255, 255, 255, 0.1);
  }

  .prompt-icon {
    color: var(--mv-primitive-frost-1); /* User color */
    font-weight: bold;
    font-family: var(--mv-font-mono);
    margin-top: 4px; /* Align with first line of text */
  }

  .input-wrapper {
    flex: 1;
  }

  .transparent-input {
    width: 100%;
    background: transparent;
    border: none;
    color: var(--mv-color-text-primary);
    font-family: var(--mv-font-sans);
    font-size: var(--mv-font-size-md);
    outline: none;
    text-shadow: 1px 1px 1px black;
    resize: none; /* User can't resize manually, fixed to rows */
    display: block;
    line-height: 1.4;
  }

  .transparent-input::placeholder {
    color: rgba(255, 255, 255, 0.3);
  }
</style>
