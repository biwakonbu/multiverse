<script lang="ts">
  import { onMount } from "svelte";
  import ChatMessage from "../ChatMessage.svelte";
  import {
    chatMessages,
    isChatLoading,
    chatError,
    type ChatResponse,
  } from "../../../../stores/chat";

  export let conflicts: NonNullable<ChatResponse["conflicts"]> = [];

  const MAX_DISPLAY_MESSAGES = 10000;

  let contentEl: HTMLElement;
  let showScrollToBottom = false;
  let hasNewMessages = false;
  let lastSeenMessageCount = 0;

  // スクロール位置を監視して「最新に移動」ボタンの表示を制御
  function handleScroll() {
    if (!contentEl) return;
    const scrollBottom =
      contentEl.scrollHeight - contentEl.scrollTop - contentEl.clientHeight;
    // 最下部から100px以上離れている場合にボタンを表示
    showScrollToBottom = scrollBottom > 100;
    // 最下部にいる場合は新着フラグをリセット
    if (!showScrollToBottom) {
      hasNewMessages = false;
      lastSeenMessageCount = $chatMessages.length;
    }
  }

  // 新しいメッセージが追加されたかチェック
  $: if ($chatMessages.length > lastSeenMessageCount && showScrollToBottom) {
    hasNewMessages = true;
  }

  // 表示するメッセージ（最新10000件まで）
  $: displayMessages = $chatMessages.slice(-MAX_DISPLAY_MESSAGES);

  // 最新に移動
  function scrollToBottom() {
    if (contentEl) {
      contentEl.scrollTop = contentEl.scrollHeight;
      showScrollToBottom = false;
      hasNewMessages = false;
      lastSeenMessageCount = $chatMessages.length;
    }
  }

  onMount(() => {
    lastSeenMessageCount = $chatMessages.length;
  });
</script>

<div class="chat-view-container">
  <div class="chat-view" bind:this={contentEl} on:scroll={handleScroll}>
    {#if $chatError}
      <div class="error-banner" role="alert">
        {$chatError}
      </div>
    {/if}
    {#each displayMessages as msg (msg.id)}
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

  <!-- スクロールボタン: フェードイン/スライドアップアニメーション -->
  <button
    class="scroll-fab"
    class:visible={showScrollToBottom}
    class:has-new={hasNewMessages}
    on:click={scrollToBottom}
    type="button"
    aria-label="最新のメッセージに移動"
    aria-hidden={!showScrollToBottom}
  >
    <span class="fab-glow"></span>
    <span class="fab-inner">
      <svg
        class="fab-icon"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <polyline points="6 9 12 15 18 9"></polyline>
      </svg>
      {#if hasNewMessages}
        <span class="fab-badge">
          <span class="badge-dot"></span>
        </span>
      {/if}
    </span>
  </button>
</div>

<style>
  .chat-view-container {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-height: 0;
    position: relative;
  }

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

  /* ========================================
     Floating Action Button - Crystal HUD スタイル
     ======================================== */
  .scroll-fab {
    --scroll-fab-size: var(--mv-scroll-fab-size, var(--mv-size-action-btn));
    --scroll-fab-offset: var(--mv-scroll-fab-offset, 16px);
    --scroll-fab-glow-inset: var(--mv-scroll-fab-glow-inset, -4px);
    --scroll-fab-badge-offset: var(--mv-scroll-fab-badge-offset, var(--mv-spacing-xxs));
    --scroll-fab-icon-size: var(--mv-scroll-fab-icon-size, var(--mv-icon-size-md));
    --scroll-fab-dot-size: var(--mv-scroll-fab-dot-size, 10px);

    position: absolute;
    right: var(--mv-spacing-md);
    bottom: var(--mv-spacing-lg);
    width: var(--scroll-fab-size);
    height: var(--scroll-fab-size);
    padding: 0;
    border: none;
    background: transparent;
    cursor: pointer;
    z-index: 10;

    /* 初期状態: 非表示 */
    opacity: 0;
    transform: translateY(var(--scroll-fab-offset)) scale(0.8);
    pointer-events: none;

    /* スムーズなトランジション */
    transition:
      opacity var(--mv-duration-normal) var(--mv-easing-out),
      transform var(--mv-duration-normal) var(--mv-easing-spring);
  }

  .scroll-fab.visible {
    opacity: 1;
    transform: translateY(0) scale(1);
    pointer-events: auto;
  }

  /* グロー背景レイヤー */
  .fab-glow {
    position: absolute;
    inset: var(--scroll-fab-glow-inset);
    border-radius: var(--mv-radius-full);
    background: radial-gradient(
      circle,
      var(--mv-glow-frost-2-light) 0%,
      transparent 70%
    );
    opacity: 0;
    transition: opacity var(--mv-duration-fast) var(--mv-easing-out);
  }

  .scroll-fab:hover .fab-glow {
    opacity: 1;
  }

  .scroll-fab.has-new .fab-glow {
    background: radial-gradient(
      circle,
      var(--mv-glow-green-mid) 0%,
      transparent 70%
    );
    opacity: 1;
    animation: pulse-glow 2s var(--mv-easing-default) infinite;
  }

  /* メインボタン内側 */
  .fab-inner {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    border-radius: var(--mv-radius-full);

    /* Crystal Button スタイル */
    background: var(--mv-btn-crystal-bg);
    border: var(--mv-border-width-thin) solid var(--mv-btn-crystal-border);
    box-shadow: var(--mv-btn-crystal-shadow);
    backdrop-filter: blur(12px);

    /* トランジション */
    transition:
      background var(--mv-duration-fast) var(--mv-easing-out),
      border-color var(--mv-duration-fast) var(--mv-easing-out),
      box-shadow var(--mv-duration-fast) var(--mv-easing-out),
      transform var(--mv-duration-instant) var(--mv-easing-out);
  }

  .scroll-fab:hover .fab-inner {
    background: var(--mv-btn-crystal-bg-hover);
    border-color: var(--mv-glow-frost-2-border);
    box-shadow: var(--mv-btn-crystal-shadow-hover);
    transform: var(--mv-transform-hover-lift);
  }

  .scroll-fab:active .fab-inner {
    background: var(--mv-btn-crystal-bg-active);
    transform: var(--mv-transform-press);
  }

  .scroll-fab.has-new .fab-inner {
    border-color: var(--mv-border-glow-green);
    box-shadow: var(--mv-shadow-glow-green-md);
  }

  /* アイコン */
  .fab-icon {
    width: var(--scroll-fab-icon-size);
    height: var(--scroll-fab-icon-size);
    color: var(--mv-primitive-frost-2);
    filter: drop-shadow(0 0 4px var(--mv-glow-frost-2));
    transition:
      color var(--mv-duration-fast),
      filter var(--mv-duration-fast),
      transform var(--mv-duration-fast) var(--mv-easing-spring);
  }

  .scroll-fab:hover .fab-icon {
    color: var(--mv-primitive-snow-0);
    filter: drop-shadow(0 0 8px var(--mv-glow-frost-2-strong));
    transform: translateY(2px);
  }

  .scroll-fab.has-new .fab-icon {
    color: var(--mv-primitive-aurora-green);
    filter: drop-shadow(0 0 6px var(--mv-glow-green));
  }

  /* 新着バッジ */
  .fab-badge {
    position: absolute;
    top: var(--scroll-fab-badge-offset);
    right: var(--scroll-fab-badge-offset);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .badge-dot {
    width: var(--scroll-fab-dot-size);
    height: var(--scroll-fab-dot-size);
    border-radius: var(--mv-radius-full);
    background: var(--mv-primitive-aurora-green);
    box-shadow: var(
      --scroll-fab-badge-shadow,
      0 0 6px var(--mv-glow-green-strong), 0 0 12px var(--mv-glow-green)
    );
    animation: badge-pulse 1.5s var(--mv-easing-default) infinite;
  }

  /* アニメーション */
  @keyframes pulse-glow {
    0%,
    100% {
      opacity: 0.6;
      transform: scale(1);
    }
    50% {
      opacity: 1;
      transform: scale(1.1);
    }
  }

  @keyframes badge-pulse {
    0%,
    100% {
      transform: scale(1);
      opacity: 1;
    }
    50% {
      transform: scale(1.2);
      opacity: 0.8;
    }
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
