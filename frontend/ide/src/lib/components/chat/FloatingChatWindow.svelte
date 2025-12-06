<script lang="ts">
  import { createEventDispatcher, onMount } from "svelte";
  import ChatMessage from "./ChatMessage.svelte";
  import ChatInput from "./ChatInput.svelte";
  import { chatStore, chatMessages, isChatLoading } from "../../../stores/chat";

  export let initialPosition = { x: 20, y: 20 };

  let position = { ...initialPosition };
  let isDragging = false;
  let windowEl: HTMLElement | undefined;
  let contentEl: HTMLElement | undefined;

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
    if (!text.trim()) return;

    const response = await chatStore.sendMessage(text);

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
      {:else if activeTab === "Log"}
        <div class="log-placeholder">
          <ChatMessage
            role="system"
            content="[System] Log tab selected."
            timestamp={new Date().toISOString()}
          />
          {#each $chatMessages.filter((m) => m.role === "system") as msg (msg.id)}
            <ChatMessage
              role={msg.role}
              content={msg.content}
              timestamp={msg.timestamp}
            />
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
  }
</style>
