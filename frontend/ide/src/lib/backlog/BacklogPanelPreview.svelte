<script module lang="ts">
  // バックログアイテム型（stores/backlogStore.ts からコピー）
  export type BacklogItemType = "FAILURE" | "QUESTION" | "BLOCKER";
  export interface BacklogItem {
    id: string;
    taskId: string;
    type: BacklogItemType;
    title: string;
    description: string;
    priority: number;
    createdAt: string;
    resolvedAt?: string;
    resolution?: string;
    // Wails models 互換用（Storybook プレビューでは未使用）
    convertValues?: (a: unknown) => void;
  }
</script>

<script lang="ts">
  import BacklogItemComponent from "./components/BacklogItem.svelte";
  import ResolveDialog from "./components/ResolveDialog.svelte";
  import EmptyBacklog from "./components/EmptyBacklog.svelte";

  
  interface Props {
    // Props
    items?: BacklogItem[];
    onresolve?: (data: { id: string; resolution: string }) => void;
    ondelete?: (data: { id: string }) => void;
  }

  let { items = [], onresolve, ondelete }: Props = $props();

  // 未解決アイテム数
  let unresolvedCount = $derived(items.filter((item) => !item.resolvedAt).length);

  // 解決ダイアログ
  let resolvingItem: BacklogItem | null = $state(null);

  function openResolveDialog(item: BacklogItem) {
    resolvingItem = item;
  }

  function closeResolveDialog() {
    resolvingItem = null;
  }

  function handleResolve(event: { text: string }) {
    if (!resolvingItem) return;
    onresolve?.({
      id: resolvingItem.id,
      resolution: event.text || "Resolved",
    });
    closeResolveDialog();
  }

  function handleDelete(item: BacklogItem) {
    ondelete?.({ id: item.id });
  }
</script>

<aside class="backlog-panel">
  <header class="panel-header">
    <h3>バックログ ({unresolvedCount})</h3>
  </header>

  <div class="panel-content">
    {#if items.length === 0}
      <EmptyBacklog />
    {:else}
      <ul class="backlog-list">
        {#each items as item (item.id)}
          <BacklogItemComponent
            {item}
            onresolve={() => openResolveDialog(item)}
            ondelete={() => handleDelete(item)}
          />
        {/each}
      </ul>
    {/if}
  </div>
</aside>

<!-- 解決ダイアログ -->
{#if resolvingItem}
  <ResolveDialog
    item={resolvingItem}
    onclose={closeResolveDialog}
    onconfirm={(e) => handleResolve(e)}
  />
{/if}

<style>
  /* === Crystal Glass Panel === */
  .backlog-panel {
    display: flex;
    flex-direction: column;
    height: 100%;

    /* Glassmorphism Background */
    background: var(--mv-glass-bg);
    backdrop-filter: blur(16px);

    /* Subtle glass border */
    border-left: var(--mv-border-width-thin) solid var(--mv-glass-border-strong);

    /* Soft ambient glow */
    box-shadow: var(--mv-shadow-backlog-panel);
  }

  /* === Header with HUD styling === */
  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
    background: var(--mv-glass-bg-strong);
    border-bottom: var(--mv-border-width-thin) solid var(--mv-glass-border);

    /* Inner glow effect */
    box-shadow: var(--mv-shadow-backlog-header);
  }

  .panel-header h3 {
    margin: 0;
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-bold);
    text-transform: uppercase;
    letter-spacing: var(--mv-letter-spacing-header);
    color: var(--mv-color-text-secondary);

    /* Glow text effect */
    text-shadow: var(--mv-text-shadow-frost);
  }

  /* === Scrollable Content === */
  .panel-content {
    flex: 1;
    overflow-y: auto;
    padding: var(--mv-spacing-md);

    /* Smooth scrollbar */
    scrollbar-width: thin;
    scrollbar-color: var(--mv-glass-border) transparent;
  }

  .panel-content::-webkit-scrollbar {
    width: var(--mv-scrollbar-width);
  }

  .panel-content::-webkit-scrollbar-track {
    background: transparent;
  }

  .panel-content::-webkit-scrollbar-thumb {
    background: var(--mv-glass-border);
    border-radius: var(--mv-scrollbar-radius);
  }

  .panel-content::-webkit-scrollbar-thumb:hover {
    background: var(--mv-glass-border-strong);
  }

  /* === Backlog List === */
  .backlog-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-md);
  }
</style>
