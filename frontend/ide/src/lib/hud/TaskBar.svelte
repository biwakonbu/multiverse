<script lang="ts">
  import { windowStore, type WindowId } from "../../stores/windowStore";
  import { unresolvedCount } from "../../stores/backlogStore";

  // Icons (Simple generic fallback chars for now, can be replaced with better icons later)
  const ICONS: Record<string, string> = {
    chat: "💬",
    process: "⚙️",
    wbs: "📊",
    backlog: "📋",
  };

  function toggle(id: WindowId) {
    windowStore.toggle(id);
  }
</script>

<div class="taskbar-container">
  <div class="taskbar glass-panel">
    <!-- Chat Toggle -->
    <button
      class="taskbar-item"
      class:active={$windowStore.chat.isOpen}
      onclick={() => toggle("chat")}
      title="Chat"
    >
      <span class="icon">{ICONS.chat}</span>
      {#if $windowStore.chat.isOpen}
        <div class="indicator"></div>
      {/if}
    </button>

    <!-- Process Toggle -->
    <button
      class="taskbar-item"
      class:active={$windowStore.process.isOpen}
      onclick={() => toggle("process")}
      title="Process & Resources"
    >
      <span class="icon">{ICONS.process}</span>
      {#if $windowStore.process.isOpen}
        <div class="indicator"></div>
      {/if}
    </button>

    <!-- WBS Toggle (Graph/List) - For now just WBS List Window -->
    <!-- Assuming we want to expose WBS as a window as requested, though existing impl is Graph + Overlay -->
    <!-- Let's keep it consistent: User asked to "mix and match", so toggle WBS window -->
    <!--  
    <button 
        class="taskbar-item" 
        class:active={$windowStore.wbs.isOpen}
        onclick={() => toggle('wbs')}
        title="Work Breakdown Structure"
    >
      <span class="icon">{ICONS.wbs}</span>
      {#if $windowStore.wbs.isOpen}
        <div class="indicator"></div>
      {/if}
    </button>
    -->

    <!-- Backlog Toggle -->
    <button
      class="taskbar-item"
      class:active={$windowStore.backlog.isOpen}
      onclick={() => toggle("backlog")}
      title="Backlog"
    >
      <div class="icon-wrapper">
        <span class="icon">{ICONS.backlog}</span>
        {#if $unresolvedCount > 0}
          <span class="badge">{$unresolvedCount}</span>
        {/if}
      </div>
      {#if $windowStore.backlog.isOpen}
        <div class="indicator"></div>
      {/if}
    </button>
  </div>
</div>

<style>
  .taskbar-container {
    position: fixed;
    bottom: var(--mv-spacing-lg);
    left: var(--mv-position-center);
    transform: translateX(var(--mv-transform-center-x));
    z-index: 2000;
  }

  .taskbar {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-sm);
    padding: var(--mv-spacing-xs) var(--mv-spacing-md);
    border-radius: var(--mv-radius-full);

    background: var(--mv-glass-bg);
    backdrop-filter: var(--mv-glass-blur);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
    box-shadow: var(--mv-shadow-floating-panel);
    transition: all var(--mv-transition-base);
  }

  .taskbar:hover {
    background: var(--mv-glass-hover);
    transform: scale(1.02);
  }

  .taskbar-item {
    position: relative;
    width: var(--mv-size-action-btn);
    height: var(--mv-size-action-btn);
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: transparent;
    border-radius: var(--mv-radius-md);
    cursor: pointer;
    transition: all var(--mv-transition-fast);
    color: var(--mv-color-text-secondary);
  }

  .taskbar-item:hover {
    background: var(--mv-color-surface-hover);
    color: var(--mv-color-text-primary);
    transform: translateY(-2px);
  }

  .taskbar-item.active {
    color: var(--mv-primitive-frost-1); /* Use brand color */
    background: var(--mv-glass-active);
  }

  .icon {
    font-size: var(--mv-font-size-lg);
    line-height: 1;
  }

  .indicator {
    position: absolute;
    bottom: var(--mv-space-1);
    left: var(--mv-position-center);
    transform: translateX(var(--mv-transform-center-x));
    width: var(--mv-space-1);
    height: var(--mv-space-1);
    background: var(--mv-primitive-frost-1);
    border-radius: var(--mv-radius-full);
    box-shadow: 0 0 var(--mv-space-1) var(--mv-primitive-frost-1);
  }

  .icon-wrapper {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .badge {
    position: absolute;
    top: var(--mv-space-2);
    right: var(--mv-space-2);
    background: var(--mv-color-status-failed-bg);
    color: var(--mv-color-status-failed-text);
    font-size: var(--mv-font-size-2xs);
    font-weight: bold;
    padding: var(--mv-space-px) var(--mv-space-1);
    border-radius: var(--mv-space-2-5);
    border: var(--mv-border-width-sm) solid var(--mv-color-status-failed-text);
  }
</style>
