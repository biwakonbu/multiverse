<script lang="ts">
  import WBSNode from './WBSNode.svelte';
  import {
    wbsTree,
    expandedNodes,
    overallProgress,
  } from '../../stores/wbsStore';
  import type { WBSNode as WBSNodeType } from '../../stores/wbsStore';

  // 展開状態を取得
  $: expandedSet = $expandedNodes;

  // ノードが展開されているか判定
  function isExpanded(nodeId: string): boolean {
    return expandedSet.has(nodeId);
  }

  // ツリーをフラット化（再帰的に展開）
  function flattenTree(
    nodes: WBSNodeType[],
    parentExpanded: boolean = true
  ): Array<{ node: WBSNodeType; expanded: boolean; visible: boolean }> {
    const result: Array<{
      node: WBSNodeType;
      expanded: boolean;
      visible: boolean;
    }> = [];

    for (const node of nodes) {
      const expanded = isExpanded(node.id);
      result.push({ node, expanded, visible: parentExpanded });

      if (node.children.length > 0) {
        const childItems = flattenTree(node.children, parentExpanded && expanded);
        result.push(...childItems);
      }
    }

    return result;
  }

  $: flatNodes = flattenTree($wbsTree);
  $: visibleNodes = flatNodes.filter((item) => item.visible);
</script>

<div class="wbs-view">
  <!-- ヘッダー：全体進捗 -->
  <header class="wbs-header">
    <div class="header-title">
      <h2>作業分解構造</h2>
      <span class="task-count">
        {$overallProgress.completed} / {$overallProgress.total} タスク完了
      </span>
    </div>

    <div class="header-progress">
      <div class="progress-bar-large">
        <div class="progress-fill" style:--progress="{$overallProgress.percentage}%"></div>
      </div>
      <span class="progress-percentage">{$overallProgress.percentage}%</span>
    </div>

    <div class="header-actions">
      <button
        class="action-btn"
        on:click={() => expandedNodes.expandAll()}
        title="すべて展開"
      >
        ↕ 全展開
      </button>
      <button
        class="action-btn"
        on:click={() => expandedNodes.collapseAll()}
        title="すべて折りたたむ"
      >
        ⇕ 全折
      </button>
    </div>
  </header>

  <!-- ツリービュー -->
  <div class="wbs-tree" role="tree" aria-label="WBS ツリー">
    {#if visibleNodes.length === 0}
      <div class="empty-state">
        <p>タスクがありません</p>
        <p class="empty-hint">チャットからタスクを生成してください</p>
      </div>
    {:else}
      {#each visibleNodes as { node, expanded } (node.id)}
        <WBSNode {node} {expanded} />
      {/each}
    {/if}
  </div>
</div>

<style>
  .wbs-view {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--mv-color-surface-primary);
  }

  /* ヘッダー */
  .wbs-header {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-sm);
    padding: var(--mv-spacing-md);
    border-bottom: var(--mv-border-width-thin) solid var(--mv-color-border-subtle);
    background: var(--mv-color-surface-secondary);
  }

  .header-title {
    display: flex;
    align-items: baseline;
    gap: var(--mv-spacing-sm);
  }

  .header-title h2 {
    font-size: var(--mv-font-size-lg);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
    margin: 0;
  }

  .task-count {
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-muted);
  }

  .header-progress {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-sm);
  }

  .progress-bar-large {
    flex: 1;
    height: var(--mv-progress-bar-height-md);
    background: var(--mv-color-surface-primary);
    border-radius: var(--mv-radius-sm);
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    width: var(--progress, 0%);
    background: var(--mv-color-status-succeeded-border);
    border-radius: var(--mv-radius-sm);
    transition: width var(--mv-duration-slow);
  }

  .progress-percentage {
    font-size: var(--mv-font-size-lg);
    font-weight: var(--mv-font-weight-bold);
    font-family: var(--mv-font-mono);
    color: var(--mv-color-status-succeeded-text);
    min-width: var(--mv-progress-text-width-md);
    text-align: right;
  }

  .header-actions {
    display: flex;
    gap: var(--mv-spacing-xs);
  }

  .action-btn {
    padding: var(--mv-spacing-xxs) var(--mv-spacing-sm);
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-medium);
    color: var(--mv-color-text-secondary);
    background: var(--mv-color-surface-primary);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    border-radius: var(--mv-radius-sm);
    cursor: pointer;
    transition:
      background-color var(--mv-transition-hover),
      color var(--mv-transition-hover);
  }

  .action-btn:hover {
    background: var(--mv-color-surface-hover);
    color: var(--mv-color-text-primary);
  }

  /* ツリービュー */
  .wbs-tree {
    flex: 1;
    overflow-y: auto;
    padding: var(--mv-spacing-sm);
  }

  /* 空状態 */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--mv-color-text-muted);
  }

  .empty-state p {
    margin: var(--mv-spacing-xxs) 0;
  }

  .empty-hint {
    font-size: var(--mv-font-size-sm);
  }
</style>
