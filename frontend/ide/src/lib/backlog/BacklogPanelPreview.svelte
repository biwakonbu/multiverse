<script context="module" lang="ts">
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
  }
</script>

<script lang="ts">
  import { createEventDispatcher } from "svelte";

  // Props
  export let items: BacklogItem[] = [];

  const dispatch = createEventDispatcher<{
    resolve: { id: string; resolution: string };
    delete: { id: string };
  }>();

  // 未解決アイテム数
  $: unresolvedCount = items.filter((item) => !item.resolvedAt).length;

  // 解決ダイアログ
  let resolvingItem: BacklogItem | null = null;
  let resolutionText = "";

  function openResolveDialog(item: BacklogItem) {
    resolvingItem = item;
    resolutionText = "";
  }

  function closeResolveDialog() {
    resolvingItem = null;
    resolutionText = "";
  }

  function handleResolve() {
    if (!resolvingItem) return;
    dispatch("resolve", {
      id: resolvingItem.id,
      resolution: resolutionText || "Resolved",
    });
    closeResolveDialog();
  }

  function handleDelete(item: BacklogItem) {
    dispatch("delete", { id: item.id });
  }

  function getTypeLabel(type: BacklogItem["type"]): string {
    switch (type) {
      case "FAILURE":
        return "失敗";
      case "QUESTION":
        return "質問";
      case "BLOCKER":
        return "ブロッカー";
      default:
        return type;
    }
  }

  function getPriorityLabel(priority: number): string {
    if (priority >= 5) return "最高";
    if (priority >= 4) return "高";
    if (priority >= 3) return "中";
    if (priority >= 2) return "低";
    return "最低";
  }

  function formatDate(dateStr: string): string {
    const date = new Date(dateStr);
    return date.toLocaleString("ja-JP", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }
</script>

<aside class="backlog-panel">
  <header class="panel-header">
    <h3>バックログ ({unresolvedCount})</h3>
  </header>

  <div class="panel-content">
    {#if items.length === 0}
      <div class="empty-state">
        <span class="empty-icon">&#10003;</span>
        <p>バックログは空です</p>
      </div>
    {:else}
      <ul class="backlog-list">
        {#each items as item (item.id)}
          <li class="backlog-item" class:failure={item.type === "FAILURE"}>
            <div class="item-header">
              <span class="type-badge {item.type.toLowerCase()}"
                >{getTypeLabel(item.type)}</span
              >
              <span class="priority">{getPriorityLabel(item.priority)}</span>
              <span class="date">{formatDate(item.createdAt)}</span>
            </div>
            <h4 class="item-title">{item.title}</h4>
            <p class="item-description">{item.description}</p>
            <div class="item-actions">
              <button
                class="btn-resolve"
                on:click={() => openResolveDialog(item)}
              >
                解決
              </button>
              <button class="btn-delete" on:click={() => handleDelete(item)}>
                削除
              </button>
            </div>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</aside>

<!-- 解決ダイアログ -->
{#if resolvingItem}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div class="dialog-overlay" on:click={closeResolveDialog}>
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div class="dialog" on:click|stopPropagation>
      <h4>バックログを解決</h4>
      <p class="dialog-item-title">{resolvingItem.title}</p>
      <label>
        解決方法:
        <textarea
          bind:value={resolutionText}
          placeholder="どのように解決したかを入力..."
          rows="3"
        />
      </label>
      <div class="dialog-actions">
        <button class="btn-cancel" on:click={closeResolveDialog}>
          キャンセル
        </button>
        <button class="btn-confirm" on:click={handleResolve}> 解決 </button>
      </div>
    </div>
  </div>
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

  /* === Empty State with Glow === */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--mv-color-text-muted);
    text-align: center;
    padding: var(--mv-spacing-xl);
  }

  .empty-icon {
    font-size: var(--mv-font-size-xxxl);
    margin-bottom: var(--mv-spacing-md);
    color: var(--mv-primitive-aurora-green);

    /* Success glow */
    filter: var(--mv-filter-success-glow);
    animation: gentle-pulse 3s ease-in-out infinite;
  }

  @keyframes gentle-pulse {
    0%,
    100% {
      opacity: 0.8;
      transform: scale(1);
    }
    50% {
      opacity: 1;
      transform: scale(1.05);
    }
  }

  .empty-state p {
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-muted);
    letter-spacing: var(--mv-letter-spacing-count);
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

  /* === Backlog Item Card === */
  .backlog-item {
    position: relative;

    /* Glass Card */
    background: var(--mv-glass-bg-strong);
    backdrop-filter: blur(8px);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border);
    border-radius: var(--mv-radius-md);
    padding: var(--mv-spacing-md);

    /* Card shadow */
    box-shadow: var(--mv-shadow-card);

    /* Animation */
    transition: all 0.25s cubic-bezier(0.2, 0.8, 0.2, 1);
    overflow: hidden;
  }

  .backlog-item::before {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    width: var(--mv-border-width-thick);
    height: 100%;
    background: var(--mv-glass-border);
    opacity: 0.5;
    transition: all 0.25s ease;
  }

  .backlog-item:hover {
    background: var(--mv-glass-hover);
    border-color: var(--mv-glass-border-strong);
    transform: translateX(4px);

    box-shadow: var(--mv-shadow-card-hover);
  }

  .backlog-item:hover::before {
    opacity: 1;
    background: var(--mv-primitive-frost-2);
    box-shadow: var(--mv-shadow-glow-frost-2);
  }

  /* === Failure Type - Glowing Red Edge === */
  .backlog-item.failure::before {
    background: var(--mv-primitive-aurora-red);
    opacity: 1;
    box-shadow: var(--mv-shadow-glow-red);
  }

  .backlog-item.failure:hover::before {
    box-shadow: var(--mv-shadow-glow-red-lg);
  }

  /* === Item Header === */
  .item-header {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-sm);
    margin-bottom: var(--mv-spacing-sm);
    font-size: var(--mv-font-size-xs);
  }

  /* === Type Badge with Glow === */
  .type-badge {
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    border-radius: var(--mv-radius-sm);
    font-weight: var(--mv-font-weight-bold);
    text-transform: uppercase;
    font-size: var(--mv-font-size-xxs);
    letter-spacing: var(--mv-letter-spacing-badge);

    /* Glass effect */
    backdrop-filter: blur(4px);
    border: var(--mv-border-width-thin) solid transparent;
  }

  .type-badge.failure {
    background: var(--mv-bg-glow-red-light);
    color: var(--mv-primitive-aurora-red);
    border-color: var(--mv-glow-failed);
    box-shadow: 0 0 8px var(--mv-glow-red);
  }

  .type-badge.question {
    background: var(--mv-bg-glow-yellow-mid);
    color: var(--mv-primitive-aurora-yellow);
    border-color: var(--mv-border-glow-yellow);
    box-shadow: var(--mv-shadow-glow-yellow);
  }

  .type-badge.blocker {
    background: var(--mv-glow-frost-2-mid);
    color: var(--mv-primitive-frost-2);
    border-color: var(--mv-glow-frost-2-border);
    box-shadow: var(--mv-shadow-glow-frost-2);
  }

  /* === Priority Badge === */
  .priority {
    font-size: var(--mv-font-size-xxs);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-muted);
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    background: var(--mv-glass-active);
    border-radius: var(--mv-radius-sm);
    letter-spacing: var(--mv-letter-spacing-count);
  }

  /* === Date === */
  .date {
    margin-left: auto;
    font-size: var(--mv-font-size-xxs);
    font-family: var(--mv-font-mono);
    color: var(--mv-color-text-disabled);
    letter-spacing: var(--mv-letter-spacing-widest);
  }

  /* === Item Title === */
  .item-title {
    margin: 0 0 var(--mv-spacing-xs);
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
    line-height: var(--mv-line-height-snug);

    /* Subtle glow on text */
    text-shadow: var(--mv-text-shadow-snow-subtle);
  }

  /* === Item Description === */
  .item-description {
    margin: 0 0 var(--mv-spacing-sm);
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-secondary);
    line-height: var(--mv-line-height-relaxed);
    opacity: 0.9;
  }

  /* === Action Buttons === */
  .item-actions {
    display: flex;
    gap: var(--mv-spacing-sm);
    margin-top: var(--mv-spacing-md);
    padding-top: var(--mv-spacing-sm);
    border-top: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
  }

  .btn-resolve,
  .btn-delete {
    padding: var(--mv-spacing-xs) var(--mv-spacing-md);
    border-radius: var(--mv-radius-sm);
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-semibold);
    letter-spacing: var(--mv-letter-spacing-count);
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.2, 0.8, 0.2, 1);
  }

  /* === Resolve Button === */
  .btn-resolve {
    background: var(--mv-bg-glow-green-mid);
    color: var(--mv-primitive-aurora-green);
    border: var(--mv-border-width-thin) solid var(--mv-border-glow-green);
  }

  .btn-resolve:hover {
    background: var(--mv-bg-glow-green-hover);
    border-color: var(--mv-primitive-aurora-green);
    box-shadow: var(--mv-shadow-glow-green-md);
    transform: translateY(-1px);
  }

  /* === Delete Button === */
  .btn-delete {
    background: transparent;
    color: var(--mv-color-text-muted);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border);
  }

  .btn-delete:hover {
    background: var(--mv-bg-glow-red-light);
    color: var(--mv-primitive-aurora-red);
    border-color: var(--mv-glow-red-border);
    box-shadow: var(--mv-shadow-glow-red-md);
  }

  /* === Dialog Overlay === */
  .dialog-overlay {
    position: fixed;
    inset: 0;
    background: var(--mv-glow-polar);
    backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  /* === Dialog Box === */
  .dialog {
    background: var(--mv-glass-bg);
    backdrop-filter: blur(24px);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-strong);
    border-radius: var(--mv-radius-lg);
    padding: var(--mv-spacing-xl);
    min-width: var(--mv-dialog-min-width);
    max-width: var(--mv-dialog-max-width);

    box-shadow: var(--mv-shadow-dialog);
  }

  .dialog h4 {
    margin: 0 0 var(--mv-spacing-md);
    font-size: var(--mv-font-size-md);
    font-weight: var(--mv-font-weight-bold);
    color: var(--mv-color-text-primary);
    text-shadow: 0 0 20px var(--mv-glow-frost-2);
  }

  .dialog-item-title {
    margin: 0 0 var(--mv-spacing-lg);
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
    background: var(--mv-glass-bg-strong);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border);
    border-radius: var(--mv-radius-sm);
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-secondary);
  }

  .dialog label {
    display: block;
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.08em;
    margin-bottom: var(--mv-spacing-xs);
  }

  .dialog textarea {
    width: 100%;
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
    background: var(--mv-glass-bg-dark);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border);
    border-radius: var(--mv-radius-sm);
    color: var(--mv-color-text-primary);
    font-family: var(--mv-font-sans);
    font-size: var(--mv-font-size-sm);
    resize: vertical;
    transition: all 0.2s ease;
  }

  .dialog textarea::placeholder {
    color: var(--mv-color-text-disabled);
    font-style: italic;
  }

  .dialog textarea:focus {
    outline: none;
    border-color: var(--mv-primitive-frost-2);
    box-shadow: 0 0 12px var(--mv-glow-frost-2-mid);
  }

  .dialog-actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--mv-spacing-sm);
    margin-top: var(--mv-spacing-lg);
    padding-top: var(--mv-spacing-md);
    border-top: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
  }

  .btn-cancel,
  .btn-confirm {
    padding: var(--mv-spacing-xs) var(--mv-spacing-lg);
    border-radius: var(--mv-radius-sm);
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-semibold);
    letter-spacing: var(--mv-letter-spacing-count);
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.2, 0.8, 0.2, 1);
  }

  .btn-cancel {
    background: transparent;
    color: var(--mv-color-text-muted);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border);
  }

  .btn-cancel:hover {
    background: var(--mv-glass-hover);
    color: var(--mv-color-text-primary);
    border-color: var(--mv-glass-border-strong);
  }

  .btn-confirm {
    background: var(--mv-bg-glow-green-mid);
    color: var(--mv-primitive-aurora-green);
    border: var(--mv-border-width-thin) solid var(--mv-glow-green-strong);
  }

  .btn-confirm:hover {
    background: var(--mv-bg-glow-green-hover);
    border-color: var(--mv-primitive-aurora-green);
    box-shadow: var(--mv-shadow-glow-green-lg);
    transform: translateY(-1px);
  }
</style>
