<script lang="ts">
  import type { WorkspaceSummary } from "../../schemas";

  interface Props {
    workspace: WorkspaceSummary;
    onopen: (id: string) => void;
    onremove: (id: string) => void;
  }

  let { workspace, onopen, onremove }: Props = $props();

  // 日付フォーマット
  function formatDate(dateStr: string): string {
    try {
      const date = new Date(dateStr);
      return date.toLocaleDateString("ja-JP", {
        year: "numeric",
        month: "2-digit",
        day: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
      });
    } catch {
      return dateStr;
    }
  }

  // パスからプロジェクト名を取得
  function getProjectName(path: string): string {
    return path.split("/").filter(Boolean).pop() || path;
  }

  // パスを短縮表示（ホームディレクトリを ~ に置換）
  function shortenPath(path: string): string {
    // 簡易的な処理（実際は環境変数から取得すべき）
    return path.replace(/^\/Users\/[^/]+/, "~");
  }

  function handleClick() {
    onopen(workspace.id);
  }

  function handleRemove(e: MouseEvent) {
    e.stopPropagation();
    onremove(workspace.id);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      onopen(workspace.id);
    }
  }
</script>

<div
  class="workspace-card"
  onclick={handleClick}
  onkeydown={handleKeydown}
  role="button"
  tabindex="0"
>
  <div class="card-icon">
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <path
        d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"
      />
    </svg>
  </div>

  <div class="card-content">
    <h3 class="project-name">
      {workspace.displayName || getProjectName(workspace.projectRoot)}
    </h3>
    <p class="project-path">{shortenPath(workspace.projectRoot)}</p>
    <p class="last-opened">最終使用: {formatDate(workspace.lastOpenedAt)}</p>
  </div>

  <button
    class="remove-btn"
    onclick={handleRemove}
    title="履歴から削除"
    aria-label="履歴から削除"
  >
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <path d="M18 6L6 18M6 6l12 12" />
    </svg>
  </button>
</div>

<style>
  .workspace-card {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-md);
    padding: var(--mv-spacing-md);
    background: var(--mv-color-surface-secondary);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-subtle);
    border-radius: var(--mv-radius-md);
    cursor: pointer;
    transition: all var(--mv-transition-hover);
  }

  .workspace-card:hover {
    background: var(--mv-color-surface-hover);
    border-color: var(--mv-color-border-default);
    transform: var(--mv-transform-hover-lift);
  }

  .workspace-card:focus {
    outline: none;
    box-shadow: var(--mv-shadow-focus);
  }

  .card-icon {
    flex-shrink: 0;
    width: var(--mv-icon-size-xxl);
    height: var(--mv-icon-size-xxl);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--mv-color-interactive-primary);
  }

  .card-icon svg {
    width: var(--mv-icon-size-lg);
    height: var(--mv-icon-size-lg);
  }

  .card-content {
    flex: 1;
    min-width: 0;
  }

  .project-name {
    font-size: var(--mv-font-size-md);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
    margin: 0 0 var(--mv-spacing-xxs) 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .project-path {
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-muted);
    margin: 0 0 var(--mv-spacing-xxs) 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-family: var(--mv-font-mono);
  }

  .last-opened {
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-disabled);
    margin: 0;
  }

  .remove-btn {
    flex-shrink: 0;
    width: var(--mv-icon-size-xl);
    height: var(--mv-icon-size-xl);
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: none;
    border-radius: var(--mv-radius-sm);
    color: var(--mv-color-text-muted);
    cursor: pointer;
    transition: all var(--mv-transition-hover);
    opacity: 0;
  }

  .workspace-card:hover .remove-btn {
    opacity: 1;
  }

  .remove-btn:hover {
    background: var(--mv-color-surface-selected);
    color: var(--mv-color-interactive-danger);
  }

  .remove-btn:focus {
    outline: none;
    opacity: 1;
    box-shadow: var(--mv-shadow-focus);
  }

  .remove-btn svg {
    width: var(--mv-icon-size-sm);
    height: var(--mv-icon-size-sm);
  }
</style>
