<script lang="ts">
  import FloatingChatWindow from "./FloatingChatWindow.svelte";

  export let chatMessages: Array<{
    id: string;
    role: "user" | "assistant" | "system";
    content: string;
    timestamp: string;
  }> = [];
</script>

<div class="main-view">
  <FloatingChatWindow messages={chatMessages} />
</div>

<style>
  .main-view {
    width: 100vw;
    height: 100vh;
    background-color: var(--mv-color-surface-app);
    position: relative;
    overflow: hidden;
    color: var(--mv-color-text-primary);
  }
</style>
