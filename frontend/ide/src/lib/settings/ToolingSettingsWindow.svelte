<script lang="ts">
  import DraggableWindow from "../components/ui/window/DraggableWindow.svelte";
  import ToolingSettingsPanel from "./ToolingSettingsPanel.svelte";
  import { windowStore } from "../../stores/windowStore";
  import { Sliders } from "lucide-svelte";

  let isOpen = $derived($windowStore.settings.isOpen);
  let position = $derived($windowStore.settings.position);
  let size = $derived($windowStore.settings.size);
  let zIndex = $derived($windowStore.settings.zIndex);

  function handleClose() {
    windowStore.close("settings");
  }

  function handleDragEnd(data: { x: number; y: number }) {
    windowStore.updatePosition("settings", data.x, data.y);
  }

  function handleResizeEnd(data: { width: number; height: number }) {
    windowStore.updateSize("settings", data.width, data.height);
  }

  function handleClick() {
    windowStore.bringToFront("settings");
  }
</script>

{#if isOpen}
  <DraggableWindow
    id="settings"
    initialPosition={position}
    initialSize={size}
    {zIndex}
    onclose={handleClose}
    ondragend={handleDragEnd}
    onresizeend={handleResizeEnd}
    onclick={handleClick}
  >
    {#snippet header()}
      <div class="window-header">
        <Sliders size={16} class="header-icon" />
        <span class="header-title">Tooling Settings</span>
      </div>
    {/snippet}

    {#snippet children()}
      <div class="settings-content">
        <ToolingSettingsPanel />
      </div>
    {/snippet}
  </DraggableWindow>
{/if}

<style>
  .window-header {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-sm);
    color: var(--mv-color-text-secondary);
  }

  :global(.header-icon) {
    opacity: var(--mv-opacity-70);
  }

  .header-title {
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
    letter-spacing: var(--mv-letter-spacing-widest);
  }

  .settings-content {
    flex: 1;
    overflow: auto;
  }
</style>
