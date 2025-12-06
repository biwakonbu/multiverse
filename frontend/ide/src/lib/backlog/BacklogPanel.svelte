<script lang="ts">
  import {
    backlogItems,
    unresolvedCount,
    resolveItem,
    deleteItem,
    type BacklogItem,
  } from "../../stores/backlogStore";

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

  async function handleResolve() {
    if (!resolvingItem) return;
    try {
      await resolveItem(resolvingItem.id, resolutionText || "Resolved");
      closeResolveDialog();
    } catch {
      // エラーは store でログ出力済み
    }
  }

  async function handleDelete(item: BacklogItem) {
    if (confirm(`「${item.title}」を削除しますか？`)) {
      try {
        await deleteItem(item.id);
      } catch {
        // エラーは store でログ出力済み
      }
    }
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
    <h3>バックログ ({$unresolvedCount})</h3>
  </header>

  <div class="panel-content">
    {#if $backlogItems.length === 0}
      <div class="empty-state">
        <span class="empty-icon">&#10003;</span>
        <p>バックログは空です</p>
      </div>
    {:else}
      <ul class="backlog-list">
        {#each $backlogItems as item (item.id)}
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
            {#if item.metadata?.error}
              <pre class="error-detail">{item.metadata.error}</pre>
            {/if}
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
  .backlog-panel {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--mv-color-surface-secondary);
    border-left: var(--mv-border-width-thin) solid var(--mv-color-border-default);
  }

  .panel-header {
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
    border-bottom: var(--mv-border-width-thin) solid
      var(--mv-color-border-default);
  }

  .panel-header h3 {
    margin: 0;
    font-size: var(--mv-font-size-md);
    font-weight: var(--mv-font-weight-semibold);
  }

  .panel-content {
    flex: 1;
    overflow-y: auto;
    padding: var(--mv-spacing-sm);
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--mv-color-text-muted);
    text-align: center;
  }

  .empty-icon {
    font-size: var(--mv-icon-size-xxl);
    margin-bottom: var(--mv-spacing-sm);
    color: var(--mv-color-status-success-text);
  }

  .backlog-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-sm);
  }

  .backlog-item {
    background: var(--mv-color-surface-primary);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    border-radius: var(--mv-radius-md);
    padding: var(--mv-spacing-sm);
  }

  .backlog-item.failure {
    border-left: 3px solid var(--mv-color-status-failed-text);
  }

  .item-header {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
    margin-bottom: var(--mv-spacing-xs);
    font-size: var(--mv-font-size-xs);
  }

  .type-badge {
    padding: 2px 6px;
    border-radius: var(--mv-radius-sm);
    font-weight: var(--mv-font-weight-medium);
    text-transform: uppercase;
    font-size: 10px;
  }

  .type-badge.failure {
    background: var(--mv-color-status-failed-bg);
    color: var(--mv-color-status-failed-text);
  }

  .type-badge.question {
    background: var(--mv-color-status-pending-bg);
    color: var(--mv-color-status-pending-text);
  }

  .type-badge.blocker {
    background: var(--mv-color-status-blocked-bg);
    color: var(--mv-color-status-blocked-text);
  }

  .priority {
    color: var(--mv-color-text-secondary);
  }

  .date {
    margin-left: auto;
    color: var(--mv-color-text-muted);
  }

  .item-title {
    margin: 0 0 var(--mv-spacing-xs);
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-medium);
  }

  .item-description {
    margin: 0 0 var(--mv-spacing-xs);
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-secondary);
    line-height: 1.4;
  }

  .error-detail {
    margin: var(--mv-spacing-xs) 0;
    padding: var(--mv-spacing-xs);
    background: var(--mv-color-surface-overlay);
    border-radius: var(--mv-radius-sm);
    font-size: var(--mv-font-size-xs);
    font-family: var(--mv-font-mono);
    color: var(--mv-color-status-failed-text);
    overflow-x: auto;
    white-space: pre-wrap;
    word-break: break-all;
  }

  .item-actions {
    display: flex;
    gap: var(--mv-spacing-xs);
    margin-top: var(--mv-spacing-sm);
  }

  .btn-resolve,
  .btn-delete {
    padding: var(--mv-spacing-xxs) var(--mv-spacing-sm);
    border: none;
    border-radius: var(--mv-radius-sm);
    font-size: var(--mv-font-size-xs);
    cursor: pointer;
    transition: all var(--mv-transition-hover);
  }

  .btn-resolve {
    background: var(--mv-color-status-success-bg);
    color: var(--mv-color-status-success-text);
  }

  .btn-resolve:hover {
    filter: brightness(1.1);
  }

  .btn-delete {
    background: transparent;
    color: var(--mv-color-text-muted);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
  }

  .btn-delete:hover {
    background: var(--mv-color-status-failed-bg);
    color: var(--mv-color-status-failed-text);
    border-color: var(--mv-color-status-failed-text);
  }

  /* ダイアログ */
  .dialog-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .dialog {
    background: var(--mv-color-surface-primary);
    border-radius: var(--mv-radius-lg);
    padding: var(--mv-spacing-lg);
    min-width: 400px;
    max-width: 90vw;
    box-shadow: var(--mv-shadow-modal);
  }

  .dialog h4 {
    margin: 0 0 var(--mv-spacing-sm);
    font-size: var(--mv-font-size-lg);
  }

  .dialog-item-title {
    margin: 0 0 var(--mv-spacing-md);
    padding: var(--mv-spacing-sm);
    background: var(--mv-color-surface-secondary);
    border-radius: var(--mv-radius-sm);
    font-size: var(--mv-font-size-sm);
  }

  .dialog label {
    display: block;
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-secondary);
  }

  .dialog textarea {
    width: 100%;
    margin-top: var(--mv-spacing-xs);
    padding: var(--mv-spacing-sm);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    border-radius: var(--mv-radius-sm);
    background: var(--mv-color-surface-secondary);
    color: var(--mv-color-text-primary);
    font-family: inherit;
    font-size: var(--mv-font-size-sm);
    resize: vertical;
  }

  .dialog textarea:focus {
    outline: none;
    border-color: var(--mv-color-border-focus);
  }

  .dialog-actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--mv-spacing-sm);
    margin-top: var(--mv-spacing-md);
  }

  .btn-cancel,
  .btn-confirm {
    padding: var(--mv-spacing-xs) var(--mv-spacing-md);
    border: none;
    border-radius: var(--mv-radius-sm);
    font-size: var(--mv-font-size-sm);
    cursor: pointer;
    transition: all var(--mv-transition-hover);
  }

  .btn-cancel {
    background: var(--mv-color-surface-secondary);
    color: var(--mv-color-text-secondary);
  }

  .btn-cancel:hover {
    background: var(--mv-color-surface-hover);
  }

  .btn-confirm {
    background: var(--mv-color-status-success-bg);
    color: var(--mv-color-status-success-text);
  }

  .btn-confirm:hover {
    filter: brightness(1.1);
  }
</style>
