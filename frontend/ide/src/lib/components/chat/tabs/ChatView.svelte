<script lang="ts">
  import { afterUpdate } from "svelte";
  import ChatMessage from "../ChatMessage.svelte";
  import {
    chatMessages,
    isChatLoading,
    chatError,
    type ChatResponse,
  } from "../../../../stores/chat";

  export let conflicts: NonNullable<ChatResponse["conflicts"]> = [];

  let contentEl: HTMLElement;

  // 自動スクロール
  afterUpdate(() => {
    if ($chatMessages.length > 0 && contentEl) {
      // ユーザーがスクロールしていない場合のみスクロールする判定を入れるのが理想だが、
      // 簡易的にメッセージ追加時は常に最下部へ
      setTimeout(() => {
        if (contentEl) contentEl.scrollTop = contentEl.scrollHeight;
      }, 0);
    }
  });
</script>

<div class="chat-view" bind:this={contentEl}>
  {#if $chatError}
    <div class="error-banner" role="alert">
      {$chatError}
    </div>
  {/if}
  {#each $chatMessages as msg (msg.id)}
    <ChatMessage
      role={msg.role}
      content={msg.content}
      timestamp={msg.timestamp}
    />
  {/each}
  {#if $isChatLoading}
    <div class="loading-indicator">
      <span class="dot"></span>
      <span class="dot"></span>
      <span class="dot"></span>
    </div>
  {/if}
  {#if conflicts.length > 0}
    <div class="conflicts-card">
      <div class="conflicts-header">検出されたコンフリクト</div>
      <ul>
        {#each conflicts as conflict}
          <li>
            <div class="conflict-file">{conflict.file}</div>
            <div class="conflict-warning">{conflict.warning}</div>
            {#if conflict.tasks?.length}
              <div class="conflict-tasks">
                関連タスク: {conflict.tasks.join(", ")}
              </div>
            {/if}
          </li>
        {/each}
      </ul>
    </div>
  {/if}
</div>

<style>
  .chat-view {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-height: 0; /* Required for flex item to shrink and enable scrolling */
    overflow-y: auto; /* スクロールはここで制御 */

    /* Scrollbar styling */
    scrollbar-width: thin;
    scrollbar-color: var(--mv-glass-scrollbar) transparent;
  }

  .chat-view::-webkit-scrollbar {
    width: var(--mv-size-scrollbar);
  }

  .chat-view::-webkit-scrollbar-track {
    background: transparent;
  }

  .chat-view::-webkit-scrollbar-thumb {
    background: var(--mv-glass-scrollbar);
    border-radius: var(--mv-border-radius-scrollbar);
  }

  .chat-view::-webkit-scrollbar-thumb:hover {
    background: var(--mv-glass-scrollbar-hover);
  }

  /* ローディングインジケーター */
  .loading-indicator {
    display: flex;
    gap: var(--mv-spacing-xxs);
    padding: var(--mv-spacing-xs);
    justify-content: center;
    align-self: flex-start;
    margin-left: var(--mv-spacing-xs);
  }

  .dot {
    width: var(--mv-size-loading-dot);
    height: var(--mv-size-loading-dot);
    background: var(--mv-primitive-frost-2);
    border-radius: var(--mv-radius-full);
    animation: bounce 1.4s infinite ease-in-out both;
    opacity: 0.8;
  }

  .dot:nth-child(1) {
    animation-delay: -0.32s;
  }
  .dot:nth-child(2) {
    animation-delay: -0.16s;
  }

  @keyframes bounce {
    0%,
    80%,
    100% {
      transform: scale(0);
      opacity: 0.5;
    }
    40% {
      transform: scale(1);
      opacity: 1;
    }
  }

  .error-banner {
    background: var(--mv-color-surface-danger);
    color: var(--mv-color-text-on-danger);
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
    border-radius: var(--mv-radius-sm);
    margin-bottom: var(--mv-spacing-sm);
    font-size: var(--mv-font-size-sm);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-strong);
  }

  .conflicts-card {
    margin-top: var(--mv-spacing-md);
    padding: var(--mv-spacing-md);
    background: var(--mv-color-surface-secondary);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    border-radius: var(--mv-radius-md);
    box-shadow: var(--mv-shadow-card);
  }

  .conflicts-header {
    font-weight: var(--mv-font-weight-semibold);
    margin-bottom: var(--mv-spacing-xs);
  }

  .conflicts-card ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-xs);
  }

  .conflicts-card li {
    padding: var(--mv-spacing-xs);
    border-radius: var(--mv-radius-sm);
    background: var(--mv-color-surface-primary);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
  }

  .conflict-file {
    font-weight: var(--mv-font-weight-semibold);
  }

  .conflict-warning {
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-secondary);
  }

  .conflict-tasks {
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-tertiary);
    margin-top: var(--mv-spacing-xxs);
  }
</style>
