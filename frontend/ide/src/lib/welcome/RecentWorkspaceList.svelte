<script lang="ts">
  import WorkspaceCard from "./WorkspaceCard.svelte";
  import type { WorkspaceSummary } from "../../schemas";

  interface Props {
    workspaces?: WorkspaceSummary[];
    loading?: boolean;
    onopen?: (id: string) => void;
    onremove?: (id: string) => void;
  }

  let { workspaces = [], loading = false, onopen, onremove }: Props = $props();

  function handleOpen(id: string) {
    onopen?.(id);
  }

  function handleRemove(id: string) {
    onremove?.(id);
  }
</script>

<section class="recent-workspaces">
  <h2 class="section-title">最近使用したワークスペース</h2>

  {#if loading}
    <div class="loading-state">
      <div class="spinner"></div>
      <p>読み込み中...</p>
    </div>
  {:else if workspaces.length === 0}
    <div class="empty-state">
      <svg
        class="empty-icon"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
      >
        <path
          d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
        />
      </svg>
      <p class="empty-text">まだワークスペースがありません</p>
      <p class="empty-hint">
        「フォルダを開く」からプロジェクトを選択してください
      </p>
    </div>
  {:else}
    <div class="workspace-list">
      {#each workspaces as workspace (workspace.id)}
        <WorkspaceCard
          {workspace}
          onopen={handleOpen}
          onremove={handleRemove}
        />
      {/each}
    </div>
  {/if}
</section>

<style>
  .recent-workspaces {
    width: 100%;
  }

  .section-title {
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-secondary);
    text-transform: uppercase;
    letter-spacing: var(--mv-letter-spacing-wider);
    margin: 0 0 var(--mv-spacing-md) 0;
  }

  .workspace-list {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-sm);
    max-height: var(--mv-container-max-height-list);
    overflow-y: auto;
    padding-right: var(--mv-spacing-xs);
  }

  /* スクロールバーのスタイル */
  .workspace-list::-webkit-scrollbar {
    width: var(--mv-scrollbar-width);
  }

  .workspace-list::-webkit-scrollbar-track {
    background: transparent;
  }

  .workspace-list::-webkit-scrollbar-thumb {
    background: var(--mv-color-border-default);
    border-radius: var(--mv-scrollbar-radius);
  }

  .workspace-list::-webkit-scrollbar-thumb:hover {
    background: var(--mv-color-border-strong);
  }

  /* 空状態 */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--mv-spacing-xxl) var(--mv-spacing-lg);
    text-align: center;
  }

  .empty-icon {
    width: var(--mv-icon-size-xxxl);
    height: var(--mv-icon-size-xxxl);
    color: var(--mv-color-text-disabled);
    margin-bottom: var(--mv-spacing-md);
  }

  .empty-text {
    font-size: var(--mv-font-size-md);
    color: var(--mv-color-text-secondary);
    margin: 0 0 var(--mv-spacing-xs) 0;
  }

  .empty-hint {
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-muted);
    margin: 0;
  }

  /* ローディング状態 */
  .loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--mv-spacing-xxl) var(--mv-spacing-lg);
    gap: var(--mv-spacing-md);
  }

  .loading-state p {
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-muted);
    margin: 0;
  }

  .spinner {
    width: var(--mv-spinner-size-md);
    height: var(--mv-spinner-size-md);
    border: var(--mv-spinner-border-width) solid var(--mv-color-border-subtle);
    border-top-color: var(--mv-color-interactive-primary);
    border-radius: var(--mv-radius-full);
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
