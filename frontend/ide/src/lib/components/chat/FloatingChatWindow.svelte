<script lang="ts">
  import { createEventDispatcher, onMount } from "svelte";
  import ChatMessage from "./ChatMessage.svelte";
  import ChatInput from "./ChatInput.svelte";
  import { chatStore, chatMessages, isChatLoading } from "../../../stores/chat";

  export let initialPosition = { x: 20, y: 20 };

  let position = { ...initialPosition };
  let isDragging = false;
  let dragOffset = { x: 0, y: 0 };
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
    const rect = windowEl.getBoundingClientRect();
    dragOffset = {
      x: e.clientX - rect.left,
      y: e.clientY - rect.top,
    };

    window.addEventListener("mouseup", stopDrag);
  }

  function onMouseMove(e: MouseEvent) {
    if (!isDragging) return;

    const newX = e.clientX - dragOffset.x;
    const newY = e.clientY - dragOffset.y;

    position = { x: newX, y: newY };
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
    background: linear-gradient(
      180deg,
      var(--mv-color-surface-secondary) 0%,
      var(--mv-color-surface-primary) 100%
    );
    backdrop-filter: blur(var(--mv-scrollbar-radius));
    border-radius: var(--mv-radius-sm);
    border: var(--mv-border-panel);
    box-shadow: var(--mv-shadow-floating-window);
    display: flex;
    flex-direction: column;
    z-index: 1000;
    overflow: hidden;
    transition: height 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .floating-window.minimized {
    height: var(--mv-icon-size-xl);
    background: var(--mv-color-surface-overlay);
  }

  .floating-window:focus-within {
    background: linear-gradient(
      180deg,
      var(--mv-color-surface-secondary) 0%,
      var(--mv-color-surface-hover) 100%
    );
  }

  .header {
    height: var(--mv-icon-size-xl);
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    padding: 0 var(--mv-spacing-xs);
    cursor: grab;
    user-select: none;
    flex-shrink: 0;
    background: var(--mv-color-surface-overlay);
  }

  .window-controls {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xxs);
    margin-bottom: var(--mv-spacing-xxs);
  }

  .control-btn {
    background: transparent;
    border: none;
    color: var(--mv-color-text-muted);
    cursor: pointer;
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-md);
    padding: 0 var(--mv-spacing-xxs);
    line-height: 1;
    opacity: 0.7;
    transition: opacity var(--mv-duration-fast);
  }

  .control-btn:hover {
    opacity: 1;
    color: var(--mv-color-text-primary);
  }

  .control-btn.close:hover {
    color: var(--mv-primitive-flamingo-1);
  }

  .header:active {
    cursor: grabbing;
  }

  .tabs {
    display: flex;
    gap: calc(var(--mv-spacing-xxs) / 2);
  }

  .tab {
    padding: var(--mv-spacing-xxs) var(--mv-spacing-sm);
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-muted);
    background: var(--mv-color-surface-primary);
    border: none;
    border-top-left-radius: var(--mv-radius-sm);
    border-top-right-radius: var(--mv-radius-sm);
    cursor: pointer;
    text-shadow: var(--mv-border-width-thin) var(--mv-border-width-thin)
      var(--mv-border-width-default) var(--mv-primitive-deep-0);
    transition: all var(--mv-duration-fast);
  }

  .tab:hover {
    background: var(--mv-color-surface-hover);
    color: var(--mv-color-text-secondary);
  }

  .tab.active {
    background: var(--mv-color-surface-selected);
    color: var(--mv-primitive-aurora-yellow);
    font-weight: bold;
    box-shadow: var(--mv-shadow-button-glow);
  }

  .content {
    flex: 1;
    overflow-y: auto;
    padding: var(--mv-spacing-xs) var(--mv-spacing-sm);
    display: flex;
    flex-direction: column;
    mask-image: linear-gradient(
      to bottom,
      transparent,
      var(--mv-primitive-deep-0) 10px
    );
  }

  .content::-webkit-scrollbar {
    width: var(--mv-scrollbar-width);
  }
  .content::-webkit-scrollbar-track {
    background: var(--mv-color-surface-overlay);
  }
  .content::-webkit-scrollbar-thumb {
    background: var(--mv-color-surface-hover);
    border-radius: var(--mv-scrollbar-radius);
  }

  .footer {
    flex-shrink: 0;
  }

  /* ローディングインジケーター */
  .loading-indicator {
    display: flex;
    gap: var(--mv-spacing-xxs);
    padding: var(--mv-spacing-xs);
    justify-content: center;
  }

  .dot {
    width: var(--mv-indicator-size-sm);
    height: var(--mv-indicator-size-sm);
    background: var(--mv-color-text-muted);
    border-radius: var(--mv-radius-full);
    animation: bounce 1.4s infinite ease-in-out both;
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
    }
    40% {
      transform: scale(1);
    }
  }

  .log-placeholder {
    display: flex;
    flex-direction: column;
  }
</style>
