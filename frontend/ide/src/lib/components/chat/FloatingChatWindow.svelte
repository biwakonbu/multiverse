<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import ChatMessage from "./ChatMessage.svelte";
  import ChatInput from "./ChatInput.svelte";

  export let initialPosition = { x: 20, y: 20 };

  let position = { ...initialPosition };
  let isDragging = false;
  let dragOffset = { x: 0, y: 0 };
  let windowEl: HTMLElement;

  // Tabs
  let tabs = ["General", "Log"];
  let activeTab = "General";

  let isMinimized = false;

  function toggleMinimize() {
    isMinimized = !isMinimized;
    // Adjust position if minimized to avoid jumping?
    // Actually, since it's top/left, shrinking height pulls bottom up (visually),
    // but 'top' stays same, so it shrinks downwards from top.
    // If we want it to stick to bottom, we'd need to adjust top.
    // But for now simple height shrink is fine.
  }

  // Use createEventDispatcher for close
  const dispatch = createEventDispatcher();

  function closeWindow() {
    dispatch("close");
  }

  // チャットメッセージ
  export let messages: Array<{
    id: string;
    role: "user" | "assistant" | "system";
    content: string;
    timestamp: string;
  }> = [];

  function startDrag(e: MouseEvent) {
    if (e.button !== 0) return; // Only left click
    // Don't drag if clicking buttons
    if ((e.target as HTMLElement).closest(".window-controls")) return;

    isDragging = true;
    const rect = windowEl.getBoundingClientRect();
    // Calculate offset from the top-left corner of the element
    dragOffset = {
      x: e.clientX - rect.left,
      y: e.clientY - rect.top,
    };

    // Bring to front logic could go here
    window.addEventListener("mouseup", stopDrag);
  }

  function onMouseMove(e: MouseEvent) {
    if (!isDragging) return;

    // Calculate new position relative to viewport (since fixed/absolute)
    // If strict absolute within a container is needed, we need container offset,
    // but clientX/Y are viewport coordinates.
    // For absolute positioning inside a relative container, we need to account for container position,
    // BUT MockMainView sets context.
    // Let's stick to standard drag logic for fixed/absolute elements.

    let newX = e.clientX - dragOffset.x;
    let newY = e.clientY - dragOffset.y;

    // Constraint is handled by CSS/bounds mostly, but we can clamp here if needed
    // For now, let it be free

    position = { x: newX, y: newY };
  }

  function stopDrag() {
    isDragging = false;
    window.removeEventListener("mouseup", stopDrag);
  }

  // チャット機能は将来実装予定
  // 現在は入力を受け付けるが、バックエンド API が未実装のため
  // システムメッセージでその旨を通知する
  function handleSend(e: CustomEvent<string>) {
    const text = e.detail;
    messages = [
      ...messages,
      {
        id: crypto.randomUUID(),
        role: "user",
        content: text,
        timestamp: new Date().toISOString(),
      },
    ];
    setTimeout(() => {
      messages = [
        ...messages,
        {
          id: crypto.randomUUID(),
          role: "assistant",
          content: `受信しました: "${text}"`,
          timestamp: new Date().toISOString(),
        },
      ];
    }, 1000);
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
        <div
          class="tab"
          class:active={activeTab === tab}
          on:click|stopPropagation={() => (activeTab = tab)}
        >
          {tab}
        </div>
      {/each}
    </div>
    <div class="window-controls">
      <!-- Minimize Button -->
      <button
        class="control-btn"
        on:click|stopPropagation={toggleMinimize}
        aria-label="Minimize"
      >
        _
      </button>
      <!-- Close Button -->
      <button
        class="control-btn close"
        on:click|stopPropagation={closeWindow}
        aria-label="Close"
      >
        ×
      </button>
    </div>
  </div>

  {#if !isMinimized}
    <div class="content">
      {#if activeTab === "General"}
        {#each messages as msg (msg.id)}
          <ChatMessage
            role={msg.role}
            content={msg.content}
            timestamp={msg.timestamp}
          />
        {/each}
      {:else if activeTab === "Log"}
        <!-- Filter showing only system/log messages could go here -->
        <div class="log-placeholder">
          <ChatMessage
            role="system"
            content="[System] Log tab selected."
            timestamp={new Date().toISOString()}
          />
          {#each messages.filter((m) => m.role === "system") as msg (msg.id)}
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
      <ChatInput on:send={handleSend} />
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
      var(--mv-color-surface-overlay) 0%,
      var(--mv-color-surface-primary) 100%
    );
    backdrop-filter: blur(var(--mv-scrollbar-radius));
    border-radius: var(--mv-radius-sm);
    box-shadow: 0 var(--mv-spacing-xxs) var(--mv-spacing-sm)
      var(--mv-color-shadow-deep);
    display: flex;
    flex-direction: column;
    z-index: 1000;
    overflow: hidden;
    transition: height 0.2s cubic-bezier(0.16, 1, 0.3, 1); /* Animation for minimize */
  }

  .floating-window.minimized {
    height: var(--mv-icon-size-xl); /* Header height only */
    background: var(--mv-color-surface-overlay); /* Darker when minimized */
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
    align-items: flex-end; /* Tabs sit at the bottom of header */
    justify-content: space-between; /* Space between tabs and controls */
    padding: 0 var(--mv-spacing-xs);
    cursor: grab;
    user-select: none;
    flex-shrink: 0;

    /* Subtle separation */
    background: var(--mv-color-surface-overlay);
  }

  .window-controls {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xxs);
    margin-bottom: var(--mv-spacing-xxs); /* Align with tab vertical center roughly */
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
    color: var(--mv-primitive-flamingo-1); /* Red for close */
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
    box-shadow: var(--mv-shadow-tab-active-inset);
  }

  .content {
    flex: 1;
    overflow-y: auto;
    padding: var(--mv-spacing-xs) var(--mv-spacing-sm);
    display: flex;
    flex-direction: column;

    /* Log messages flow from bottom usually, but standard scroll for now is fine.
       Maybe add logic to keep scroll at bottom. */
    mask-image: linear-gradient(
      to bottom,
      transparent,
      var(--mv-primitive-deep-0) 10px
    ); /* Fade out top slightly */
  }

  /* Custom scrollbar for MMO feel (thin, unobtrusive) */
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
</style>
