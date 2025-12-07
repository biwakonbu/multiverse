<script lang="ts">
  import { onMount } from 'svelte';
  import { windowStore } from '../../../stores/windowStore';

  interface Props {
      initialState?: any;
      children?: import('svelte').Snippet;
  }

  let { initialState = {}, children }: Props = $props();

  onMount(() => {
    // Reset store to initial state for the story
    if (initialState.chat) {
        windowStore.update(s => ({ ...s, chat: { ...s.chat, ...initialState.chat } }));
    }
  });
</script>

<div class="store-wrapper">
    {#if children}
        {@render children()}
    {/if}
</div>

<style>
    .store-wrapper {
        width: 100%;
        height: 100vh;
        position: relative;
    }
</style>
