<script lang="ts">
  import DraggableWindow from "../components/ui/window/DraggableWindow.svelte";
  import ResourceList from "./ResourceList.svelte";
  import type { ResourceNode } from "./types";
  import { windowStore } from "../../stores/windowStore";

  interface Props {
    resources?: ResourceNode[];
  }

  let { resources = [] }: Props = $props();

  let isOpen = $derived($windowStore.process.isOpen);
  let position = $derived($windowStore.process.position);
  let size = $derived($windowStore.process.size);
  let zIndex = $derived($windowStore.process.zIndex);

  function handleClose() {
    windowStore.close('process');
  }

  function handleMinimize(data: { minimized: boolean }) {
    windowStore.minimize('process', data.minimized);
  }

  function handleDragEnd(data: { x: number; y: number }) {
    windowStore.updatePosition('process', data.x, data.y);
  }

  function handleClick() {
    windowStore.bringToFront('process');
  }
</script>

{#if isOpen}
  <DraggableWindow
    id="process"
    title="Process & Resources"
    initialPosition={position}
    initialSize={size}
    {zIndex}
    onclose={handleClose}
    onminimize={handleMinimize}
    ondragend={handleDragEnd}
    onclick={handleClick}
  >
    {#snippet children()}
        <div class="resource-window-content">
            <ResourceList {resources} />
        </div>
    {/snippet}
  </DraggableWindow>
{/if}

<style>
    .resource-window-content {
        flex: 1;
        overflow-y: auto;

        /* Matches ResourceList usage in ProcessHUD but adapted for window */
    }
</style>
