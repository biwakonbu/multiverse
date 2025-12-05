<script lang="ts">
  export let role: "user" | "assistant" | "system" = "user";
  export let content: string;
  export let timestamp: string;

  const isUser = role === "user";
  const isSystem = role === "system";
</script>

<div class="log-line {role}">
  <span class="timestamp">[{timestamp}]</span>
  <span class="sender">{isUser ? "You" : "Antigravity"}</span>
  <span class="separator">{@html isSystem ? "" : "&gt;"}</span>
  <span class="content">{content}</span>
</div>

<style>
  .log-line {
    font-family: var(
      --mv-font-mono
    ); /* Monospace font looks more like a system log */
    font-size: var(--mv-font-size-md);
    line-height: 1.4;
    padding: 2px 0;
    text-shadow: 1px 1px 1px rgba(0, 0, 0, 0.8); /* Text shadow for legibility on transparent bg */
    word-break: break-word;
  }

  .timestamp {
    color: var(--mv-color-text-muted);
    font-size: var(--mv-font-size-sm);
    margin-right: 4px;
    opacity: 0.7;
  }

  .sender {
    font-weight: var(--mv-font-weight-bold);
  }

  .separator {
    margin: 0 4px;
    color: var(--mv-color-text-muted);
  }

  /* User: Light Blue / Frost */
  .user .sender {
    color: var(--mv-primitive-frost-1);
  }
  .user .content {
    color: var(--mv-primitive-snow-storm-2);
  }

  /* Assistant: Greenish / Aurora Green (like party members) */
  .assistant .sender {
    color: var(--mv-primitive-aurora-green);
  }
  .assistant .content {
    color: var(--mv-primitive-snow-storm-1);
  }

  /* System: Purple / Aurora Purple or Grey */
  .system {
    color: var(--mv-primitive-aurora-purple);
    font-style: italic;
  }
  .system .sender {
    color: var(--mv-primitive-aurora-purple);
  }
  .system .content {
    color: var(--mv-primitive-snow-storm-0);
  }
</style>
