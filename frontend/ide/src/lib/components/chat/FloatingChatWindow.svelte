<script lang="ts">
  import ChatMessage from "./ChatMessage.svelte";
  import ChatInput from "./ChatInput.svelte";

  export let initialPosition = { x: 20, y: 20 };

  let position = { ...initialPosition };
  let isDragging = false;
  let dragOffset = { x: 0, y: 0 };
  let windowEl: HTMLElement;

  // Mock messages for display if empty
  export let messages: Array<{
    id: string;
    role: "user" | "assistant" | "system";
    content: string;
    timestamp: string;
  }> = [];

  function startDrag(e: MouseEvent) {
    isDragging = true;
    const rect = windowEl.getBoundingClientRect();
    dragOffset = {
      x: e.clientX - rect.left,
      y: e.clientY - rect.top,
    };

    // Bring to front logic could go here
  }

  function onMouseMove(e: MouseEvent) {
    if (!isDragging) return;

    let newX = e.clientX - dragOffset.x;
    let newY = e.clientY - dragOffset.y;

    // Constraint to viewport
    newX = Math.max(
      0,
      Math.min(window.innerWidth - windowEl.offsetWidth, newX)
    );
    newY = Math.max(
      0,
      Math.min(window.innerHeight - windowEl.offsetHeight, newY)
    );

    position = { x: newX, y: newY };
  }

  function stopDrag() {
    isDragging = false;
  }

  function handleSend(e: CustomEvent<string>) {
    const text = e.detail;
    // Add user message
    messages = [
      ...messages,
      {
        id: crypto.randomUUID(),
        role: "user",
        content: text,
        timestamp: new Date().toLocaleTimeString(),
      },
    ];

    // Mock response after delay
    setTimeout(() => {
      messages = [
        ...messages,
        {
          id: crypto.randomUUID(),
          role: "assistant",
          content: `受信しました: "${text}"`,
          timestamp: new Date().toLocaleTimeString(),
        },
      ];
    }, 1000);
  }
</script>

<svelte:window on:mousemove={onMouseMove} on:mouseup={stopDrag} />

<!-- svelte-ignore a11y-no-static-element-interactions -->
<div
  class="floating-window"
  style="transform: translate({position.x}px, {position.y}px);"
  bind:this={windowEl}
>
  <div class="header" on:mousedown={startDrag}>
    <div class="tabs">
      <div class="tab active">General</div>
      <div class="tab">Log</div>
    </div>
    <!-- <div class="controls"> -->
    <!-- Plus button or settings cog could go here -->
    <!-- </div> -->
  </div>

  <div class="content">
    {#each messages as msg (msg.id)}
      <ChatMessage
        role={msg.role}
        content={msg.content}
        timestamp={msg.timestamp}
      />
    {/each}
  </div>

  <div class="footer">
    <ChatInput on:send={handleSend} />
  </div>
</div>

<style>
  .floating-window {
    position: fixed;
    top: 0;
    left: 0;
    width: 600px; /* Wider based on feedback */
    height: 350px; /* Slightly taller to match width */
    background: linear-gradient(
      180deg,
      rgba(0, 0, 0, 0.4) 0%,
      rgba(20, 20, 30, 0.5) 100%
    ); /* More transparent */
    backdrop-filter: blur(
      3px
    ); /* Slightly less blur to emphasize transparency */
    border-radius: 4px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
    display: flex;
    flex-direction: column;
    z-index: 1000;
    overflow: hidden;
    /* No border property to feel more integrated/semi-transparent */
  }

  .floating-window:focus-within {
    background: linear-gradient(
      180deg,
      rgba(0, 0, 0, 0.7) 0%,
      rgba(30, 30, 45, 0.8) 100%
    );
  }

  .header {
    height: 32px;
    display: flex;
    align-items: flex-end; /* Tabs sit at the bottom of header */
    padding: 0 var(--mv-spacing-xs);
    cursor: grab;
    user-select: none;
    flex-shrink: 0;
    /* Subtle separation */
    background: rgba(0, 0, 0, 0.3);
  }

  .header:active {
    cursor: grabbing;
  }

  .tabs {
    display: flex;
    gap: 2px;
  }

  .tab {
    padding: 4px 12px;
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-muted);
    background: rgba(255, 255, 255, 0.05);
    border-top-left-radius: 4px;
    border-top-right-radius: 4px;
    cursor: pointer;
    text-shadow: 1px 1px 2px black;
    transition: all 0.2s;
  }

  .tab:hover {
    background: rgba(255, 255, 255, 0.1);
    color: var(--mv-color-text-secondary);
  }

  .tab.active {
    background: rgba(255, 255, 255, 0.15);
    color: var(--mv-primitive-aurora-yellow); /* Active tab highlight */
    font-weight: bold;
    box-shadow: 0 -2px 0 0 var(--mv-primitive-aurora-yellow) inset; /* Top highlight line */
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
      black 10px
    ); /* Fade out top slightly */
  }

  /* Custom scrollbar for MMO feel (thin, unobtrusive) */
  .content::-webkit-scrollbar {
    width: 6px;
  }
  .content::-webkit-scrollbar-track {
    background: rgba(0, 0, 0, 0.2);
  }
  .content::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.2);
    border-radius: 3px;
  }

  .footer {
    flex-shrink: 0;
  }
</style>
