<script lang="ts">
  import { run } from 'svelte/legacy';

  import { formatLocalTime } from "../../utils/time";
  import { marked } from "marked";
  import DOMPurify from "dompurify";

  interface Props {
    role?: "user" | "assistant" | "system";
    content: string;
    timestamp: string;
  }

  let { role = "user", content, timestamp }: Props = $props();

  const isUser = role === "user";
  const isSystem = role === "system";

  let displayTime = $derived(formatLocalTime(timestamp));

  let htmlContent = $state("");

  run(() => {
    // marked.parse returns string | Promise<string>.
    // Without async extensions, it is synchronous.
    // We cast to string to satisfy TS if we are sure we aren't using async features.
    // Or we handle the promise if needed. For now assuming sync.
    const raw = marked.parse(content, { async: false }) as string;
    htmlContent = DOMPurify.sanitize(raw);
  });
</script>

<div class="message-container {role}">
  <div class="meta">
    <span class="sender"
      >{isUser ? "You" : isSystem ? "System" : "Antigravity"}</span
    >
    <span class="timestamp">{displayTime}</span>
  </div>
  <div class="content markdown-body">
    {@html htmlContent}
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
    color: var(--mv-primitive-snow-storm-0);
    font-size: var(--mv-font-size-timestamp);
    opacity: 0.8;
    text-shadow: var(--mv-text-shadow-timestamp);
  }

  /* Content Block */
  .content {
    font-family: var(--mv-font-sans);
    font-size: var(--mv-font-size-sm);
    line-height: var(--mv-line-height-normal);
    color: var(--mv-color-text-primary);
    font-weight: var(--mv-font-weight-semibold);
    padding-left: var(--mv-spacing-xs);

    /* Removed white-space: pre-wrap because marked handles formatting */

    /* Base glow for readability */
    text-shadow: var(--mv-text-shadow-base);
  }

  /* Markdown Styles Scope */
  .content :global(p) {
    margin: var(--mv-spacing-sm) 0;
  }
  .content :global(p:first-child) {
    margin-top: 0;
  }
  .content :global(p:last-child) {
    margin-bottom: 0;
  }

  .content :global(pre) {
    background: var(--mv-glass-bg-base);
    padding: var(--mv-spacing-sm);
    border-radius: var(--mv-radius-md);
    overflow-x: auto;
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
    margin: var(--mv-spacing-sm) 0;
  }

  .content :global(code) {
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-sm);
    background: var(--mv-glass-hover);
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    border-radius: var(--mv-radius-sm);
  }

  .content :global(pre code) {
    background: none;
    padding: 0;
    color: var(--mv-color-text-primary);
  }

  .content :global(ul),
  .content :global(ol) {
    margin: var(--mv-spacing-sm) 0;
    padding-left: var(--mv-spacing-lg);
  }

  .content :global(li) {
    margin: var(--mv-spacing-xxs) 0;
  }

  .content :global(a) {
    color: var(--mv-primitive-frost-2);
    text-decoration: none;
    border-bottom: var(--mv-border-width-thin) dashed
      var(--mv-primitive-frost-2);
  }

  .content :global(a:hover) {
    color: var(--mv-primitive-frost-1);
    border-bottom-style: solid;
  }

  .content :global(blockquote) {
    border-left: var(--mv-border-width-thick) solid var(--mv-primitive-frost-3);
    margin: var(--mv-spacing-sm) 0;
    padding-left: var(--mv-spacing-sm);
    color: var(--mv-color-text-secondary);
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
    font-size: var(--mv-font-size-sm);
  }
</style>
