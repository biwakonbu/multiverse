<script lang="ts">
  import { windowStore, type WindowId } from "../../stores/windowStore";
  import { unresolvedCount } from "../../stores/backlogStore";
  import { MessageSquare, Cpu, ListTodo, ClipboardList, Sliders } from "lucide-svelte";

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
      aria-label="Toggle Chat"
    >
      <div class="icon-wrapper">
        <MessageSquare size={20} absoluteStrokeWidth class="icon" />
      </div>
      {#if $windowStore.chat.isOpen}
        <div class="active-glow"></div>
      {/if}
    </button>

    <!-- Process Toggle -->
    <button
      class="taskbar-item"
      class:active={$windowStore.process.isOpen}
      onclick={() => toggle("process")}
      title="Process & Resources"
      aria-label="Toggle Process View"
    >
      <div class="icon-wrapper">
        <Cpu size={20} absoluteStrokeWidth class="icon" />
      </div>
      {#if $windowStore.process.isOpen}
        <div class="active-glow"></div>
      {/if}
    </button>

    <!-- Backlog Toggle -->
    <button
      class="taskbar-item"
      class:active={$windowStore.backlog.isOpen}
      onclick={() => toggle("backlog")}
      title="Backlog"
      aria-label="Toggle Backlog"
    >
      <div class="icon-wrapper">
        <ClipboardList size={20} absoluteStrokeWidth class="icon" />
        {#if $unresolvedCount > 0}
          <span class="badge">{$unresolvedCount}</span>
        {/if}
      </div>
      {#if $windowStore.backlog.isOpen}
        <div class="active-glow"></div>
      {/if}
    </button>

    <!-- Settings Toggle -->
    <button
      class="taskbar-item"
      class:active={$windowStore.settings.isOpen}
      onclick={() => toggle("settings")}
      title="Tooling Settings"
      aria-label="Toggle Tooling Settings"
    >
      <div class="icon-wrapper">
        <Sliders size={20} absoluteStrokeWidth class="icon" />
      </div>
      {#if $windowStore.settings.isOpen}
        <div class="active-glow"></div>
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
    z-index: var(--mv-z-toast);
  }

  .taskbar {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-sm);
    padding: var(--mv-spacing-xs) var(--mv-spacing-md);

    /* Sophisticated Glassmorphism */
    background: var(--mv-taskbar-bg);
    backdrop-filter: blur(20px) saturate(180%);

    border: var(--mv-border-width-thin) solid var(--mv-taskbar-border);
    border-top: var(--mv-border-width-thin) solid var(--mv-taskbar-border-top);
    border-radius: var(--mv-space-hidden);

    box-shadow: var(--mv-taskbar-shadow);

    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .taskbar:hover {
    background: var(--mv-taskbar-bg-hover);
    box-shadow: var(--mv-taskbar-shadow-hover);
    transform: translateY(calc(-1 * var(--mv-space-0-5)));
  }

  .taskbar-item {
    position: relative;
    width: var(--mv-input-height-lg);
    height: var(--mv-input-height-lg);
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: transparent;
    border-radius: var(--mv-radius-full);
    cursor: pointer;
    transition: all 0.2s ease;

    color: var(--mv-color-text-secondary);
  }

  .taskbar-item:hover {
    background: var(--mv-btn-hover-bg);
    color: var(--mv-color-text-primary);
  }

  .taskbar-item:active {
    transform: scale(0.95);
  }

  .taskbar-item.active {
    color: var(--mv-primitive-frost-2);
    background: var(--mv-taskbar-item-active-bg);
    box-shadow: var(--mv-taskbar-item-active-shadow);
  }

  .active-glow {
    position: absolute;
    bottom: calc(-1 * var(--mv-indicator-size-xs));
    left: var(--mv-position-center);
    transform: translateX(var(--mv-transform-center-x));
    width: var(--mv-space-1);
    height: var(--mv-space-1);
    background-color: var(--mv-primitive-frost-2);
    border-radius: var(--mv-radius-full);
    box-shadow: var(--mv-taskbar-glow);
  }

  .icon-wrapper {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  /* Icon styling handled by Lucide component classes via global CSS or simpler: */
  :global(.icon) {
    stroke-width: var(--mv-border-width-md);
    opacity: var(--mv-opacity-80);
    transition: opacity 0.2s;
  }

  .taskbar-item:hover :global(.icon),
  .taskbar-item.active :global(.icon) {
    opacity: var(--mv-opacity-100);
  }

  /* Badge styling */
  .badge {
    position: absolute;
    top: calc(-1 * var(--mv-indicator-size-xs));
    right: calc(-1 * var(--mv-spacing-xs));

    background: var(--mv-primitive-aurora-red);
    color: var(--mv-color-text-on-accent);

    font-size: var(--mv-font-size-2xs);
    font-weight: var(--mv-font-weight-bold);
    line-height: var(--mv-line-height-tight);

    padding: var(--mv-space-0-75) var(--mv-space-1-5);
    border-radius: var(--mv-radius-md);
    border: var(--mv-border-width-md) solid var(--mv-badge-border);

    box-shadow: var(--mv-badge-shadow);
    min-width: var(--mv-icon-size-sm);
    text-align: center;
  }
</style>
