<script lang="ts">
  import { createEventDispatcher, onMount } from "svelte";
  import ChatMessage from "./ChatMessage.svelte";
  import ChatInput from "./ChatInput.svelte";
  import {
    chatStore,
    chatMessages,
    isChatLoading,
    chatLog,
    currentSessionId,
    chatError,
  } from "../../../stores/chat";
  import { get } from "svelte/store";
  import type { ChatResponse } from "../../../stores/chat";

  export let initialPosition = { x: 20, y: 20 };

  let position = { ...initialPosition };
  let isDragging = false;
  let windowEl: HTMLElement | null = null;
  let contentEl: HTMLElement | null = null;
  let conflicts: NonNullable<ChatResponse["conflicts"]> = [];

  // Tabs
  const tabs = ["General", "Log"];
  let activeTab = "General";

  let isMinimized = false;

  // セッション初期化
  onMount(async () => {
    await chatStore.createSession();
  });

  // メッセージ追加時に自動スクロール
  $: if ($chatMessages.length > 0 && contentEl) {
    const el = contentEl;
    setTimeout(() => {
      el.scrollTop = el.scrollHeight;
    }, 100);
  }

  function toggleMinimize() {
    isMinimized = !isMinimized;
  }

  const dispatch = createEventDispatcher();

  function closeWindow() {
    dispatch("close");
  }

  function startDrag(e: MouseEvent) {
    if (e.button !== 0) return;
    if ((e.target as HTMLElement).closest(".window-controls")) return;
    if (!windowEl) return;

    isDragging = true;
    window.addEventListener("mouseup", stopDrag);
  }

  function onMouseMove(e: MouseEvent) {
    if (!isDragging) return;

    position = {
      x: position.x + e.movementX,
      y: position.y + e.movementY,
    };
  }

  function stopDrag() {
    isDragging = false;
    window.removeEventListener("mouseup", stopDrag);
  }

  // チャット送信処理
  async function handleSend(e: CustomEvent<string>) {
    const text = e.detail;
    if (!text.trim()) {
      console.warn("FloatingChatWindow: empty text");
      return;
    }

    const currentId = get(currentSessionId);

    // セッションIDがない場合は再作成を試みる
    if (!currentId) {
      console.warn("Session ID is missing, attempting to recreate...");
      await chatStore.createSession();
      if (!get(currentSessionId)) {
        console.error("Failed to create session.");
        alert(
          "Chat session initialization failed. Please try reloading the workspace."
        );
        return;
      }
    }

    const response = await chatStore.sendMessage(text);
    conflicts = response?.conflicts ?? [];

    // タスクが生成された場合はイベントを発行
    if (response?.generatedTasks && response.generatedTasks.length > 0) {
      dispatch("tasksGenerated", {
        tasks: response.generatedTasks,
        understanding: response.understanding,
      });
    }
  }
</script>

<svelte:window on:mousemove={onMouseMove} />

<!-- svelte-ignore a11y-no-static-element-interactions -->
<div
  class="floating-window"
  class:minimized={isMinimized}
  style="top: {position.y}px; left: {position.x}px;"
  bind:this={windowEl}
>
  <div class="header" on:mousedown={startDrag}>
    <div class="tabs">
      {#each tabs as tab}
        <!-- svelte-ignore a11y-click-events-have-key-events -->
        <button
          class="tab"
          class:active={activeTab === tab}
          on:click|stopPropagation={() => (activeTab = tab)}
          type="button"
        >
          {tab}
        </button>
      {/each}
    </div>
    <div class="window-controls">
      <button
        class="control-btn"
        on:click|stopPropagation={toggleMinimize}
        aria-label="Minimize"
        type="button"
      >
        _
      </button>
      <button
        class="control-btn close"
        on:click|stopPropagation={closeWindow}
        aria-label="Close"
        type="button"
      >
        ×
      </button>
    </div>
  </div>

  {#if !isMinimized}
    <div class="content" bind:this={contentEl}>
      {#if activeTab === "General"}
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
                    <div class="conflict-tasks">関連タスク: {conflict.tasks.join(", ")}</div>
                  {/if}
                </li>
              {/each}
            </ul>
          </div>
        {/if}
      {:else if activeTab === "Log"}
        <div class="log-placeholder">
          <div class="log-entry system">
            <span class="timestamp">{new Date().toLocaleTimeString()}</span>
            <span class="step">[System]</span>
            <span class="message">Log tab selected.</span>
          </div>
          {#each $chatLog as log}
            <div class="log-entry">
              <span class="timestamp"
                >{new Date(log.timestamp).toLocaleTimeString()}</span
              >
              <span class="step">[{log.step}]</span>
              <span class="message">{log.message}</span>
            </div>
          {/each}
        </div>
      {/if}
    </div>

    <div class="footer">
      <ChatInput on:send={handleSend} disabled={$isChatLoading} />
    </div>
  {/if}
</div>

<style>
  .floating-window {
    position: fixed;
    width: var(--mv-floating-window-width);
    height: var(--mv-floating-window-height);

    /* Crystal HUD: Slightly more assertive than Header */
    background: var(--mv-glass-bg-chat);
    backdrop-filter: blur(24px);

    /* Assertive Border */
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-strong);
    border-top: var(--mv-border-width-thin) solid var(--mv-glass-border-top);
    border-bottom: var(--mv-border-width-thin) solid
      var(--mv-glass-border-bottom);

    border-radius: var(--mv-radius-lg);

    /* Deep Shadow */
    box-shadow: var(--mv-shadow-floating-panel);

    display: flex;
    flex-direction: column;
    z-index: 1000;
    overflow: hidden;
    transition: height 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .floating-window.minimized {
    height: var(--mv-size-floating-header);
    background: var(--mv-glass-bg-minimized);
  }

  /* Header Area */
  .header {
    height: var(--mv-size-floating-header);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 var(--mv-spacing-md);
    cursor: grab;
    user-select: none;
    flex-shrink: 0;
    background: transparent;
    border-bottom: var(--mv-border-width-thin) solid
      var(--mv-glass-border-subtle);
  }

  .header:active {
    cursor: grabbing;
  }

  .tabs {
    display: flex;
    gap: var(--mv-spacing-sm);
  }

  /* Glass Tabs */
  .tab {
    padding: var(--mv-spacing-xxs) var(--mv-spacing-sm);
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-bold);
    color: var(--mv-color-text-muted);
    background: transparent;
    border: var(--mv-border-width-thin) solid transparent;
    border-radius: var(--mv-radius-sm);
    cursor: pointer;
    text-transform: uppercase;
    letter-spacing: var(--mv-letter-spacing-tab);
    transition: all var(--mv-duration-fast);
  }

  .tab:hover {
    color: var(--mv-color-text-primary);
    background: var(--mv-glass-active);
  }

  .tab.active {
    color: var(--mv-primitive-frost-1);
    background: var(--mv-glass-tab-bg);
    border-color: var(--mv-glass-border-tab);
    box-shadow: var(--mv-shadow-tab-glow);
  }

  .window-controls {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
  }

  .control-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: var(--mv-size-control-btn);
    height: var(--mv-size-control-btn);
    background: transparent;
    border: var(--mv-border-width-thin) solid transparent;
    border-radius: var(--mv-radius-sm);
    color: var(--mv-color-text-muted);
    cursor: pointer;
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-sm);
    padding: 0;
    transition: all var(--mv-duration-fast);
  }

  .control-btn:hover {
    background: var(--mv-glass-hover);
    color: var(--mv-color-text-primary);
  }

  .control-btn.close:hover {
    background: var(--mv-glass-close-bg);
    color: var(--mv-primitive-aurora-red);
    border-color: var(--mv-glass-close-border);
  }

  .content {
    flex: 1;
    overflow-y: auto;
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
    display: flex;
    flex-direction: column;

    /* Clean scroll */
    scrollbar-width: thin;
    scrollbar-color: var(--mv-glass-scrollbar) transparent;
  }

  .content::-webkit-scrollbar {
    width: var(--mv-size-scrollbar);
  }

  .content::-webkit-scrollbar-track {
    background: transparent;
  }

  .content::-webkit-scrollbar-thumb {
    background: var(--mv-glass-scrollbar);
    border-radius: var(--mv-border-radius-scrollbar);
  }

  .content::-webkit-scrollbar-thumb:hover {
    background: var(--mv-glass-scrollbar-hover);
  }

  .footer {
    flex-shrink: 0;
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
    border-top: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
    background: var(--mv-glass-footer);
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

  .log-placeholder {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-xs);
  }

  .log-entry {
    display: flex;
    gap: var(--mv-spacing-xs);
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-muted);
    padding: var(--mv-spacing-xxs);
    border-bottom: 1px solid var(--mv-glass-border-subtle);
  }

  .log-entry .timestamp {
    color: var(--mv-color-text-secondary);
    min-width: 60px;
  }

  .log-entry .step {
    color: var(--mv-primitive-frost-2);
    font-weight: bold;
    min-width: 80px;
  }

  .log-entry .message {
    color: var(--mv-color-text-primary);
    word-break: break-all;
  }

  .error-banner {
    background: var(--mv-color-surface-danger);
    color: var(--mv-color-text-on-danger);
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
    border-radius: var(--mv-radius-sm);
    margin-bottom: var(--mv-spacing-sm);
    font-size: var(--mv-font-size-sm);
    border: 1px solid var(--mv-color-border-strong);
  }

  .conflicts-card {
    margin-top: var(--mv-spacing-md);
    padding: var(--mv-spacing-md);
    background: var(--mv-color-surface-secondary);
    border: 1px solid var(--mv-color-border-default);
    border-radius: var(--mv-radius-md);
    box-shadow: var(--mv-shadow-card);
  }

  .conflicts-header {
    font-weight: 600;
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
    border: 1px solid var(--mv-color-border-default);
  }

  .conflict-file {
    font-weight: 600;
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
