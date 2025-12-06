<script lang="ts">
  import { formatLocalTime } from "../../utils/time";

  export let role: "user" | "assistant" | "system" = "user";
  export let content: string;
  export let timestamp: string;

  const isUser = role === "user";
  const isSystem = role === "system";

  $: displayTime = formatLocalTime(timestamp);
</script>

<div class="message-container {role}">
  <div class="meta">
    <span class="sender"
      >{isUser ? "You" : isSystem ? "System" : "Antigravity"}</span
    >
    <span class="timestamp">{displayTime}</span>
  </div>
  <div class="content">
    {@html content}
    <!-- Allow basic HTML formatting if safe, or just text -->
  </div>
</div>

<style>
  .message-container {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-xxs);
    padding: var(--mv-spacing-sm) 0;
    border-bottom: var(--mv-border-width-thin) solid
      var(--mv-glass-border-separator);
  }

  .message-container:last-child {
    border-bottom: none;
  }

  /* Metadata Line */
  .meta {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-sm);
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-xs);
    opacity: 0.8;
  }

  .sender {
    font-weight: var(--mv-font-weight-bold);
    text-transform: uppercase;
    letter-spacing: var(--mv-letter-spacing-count);
  }

  .timestamp {
    color: var(--mv-primitive-snow-storm-0); /* Brighter grey */
    font-size: 10px;
    opacity: 0.8; /* Much more visible */
    text-shadow: 0 0 2px rgba(0, 0, 0, 0.5);
  }

  /* Content Block */
  .content {
    font-family: var(--mv-font-sans);
    font-size: var(--mv-font-size-sm);
    line-height: var(--mv-line-height-normal);
    color: var(--mv-color-text-primary);
    font-weight: var(--mv-font-weight-semibold);
    padding-left: var(--mv-spacing-xs);
    white-space: pre-wrap;

    /* Base glow for readability */
    text-shadow: var(--mv-text-shadow-base);
  }

  /* Role Specific Styles */

  /* User */
  .user .sender {
    color: var(--mv-primitive-frost-1);
    text-shadow: var(--mv-text-shadow-cyan);
  }

  .user .content {
    color: var(--mv-color-text-primary);
    font-weight: var(--mv-font-weight-bold);
    text-shadow: var(--mv-text-shadow-cyan-content);
  }

  /* Assistant */
  .assistant .sender {
    color: var(--mv-primitive-aurora-green);
    text-shadow: var(--mv-text-shadow-green);
  }

  .assistant .content {
    color: var(--mv-primitive-snow-storm-2);
    font-weight: var(--mv-font-weight-semibold);
    text-shadow: var(--mv-text-shadow-green-content);
  }

  /* System */
  .system {
    opacity: 1;
  }

  .system .sender {
    color: var(--mv-primitive-aurora-purple);
    text-shadow: var(--mv-text-shadow-purple);
  }

  .system .content {
    color: var(--mv-primitive-snow-storm-0);
    font-style: italic;
    font-weight: var(--mv-font-weight-semibold);
    text-shadow: var(--mv-text-shadow-purple-content);
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-xs);
  }
</style>
