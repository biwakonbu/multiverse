<script lang="ts">
  import WBSHeader from "./WBSHeader.svelte";
  import WBSNode from "./WBSNode.svelte";
  import { wbsTree, expandedNodes } from "../../stores/wbsStore";
  import type { WBSNode as WBSNodeType } from "../../stores/wbsStore";

  // 展開状態を取得
  let expandedSet = $derived($expandedNodes);

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
        const childItems = flattenTree(
          node.children,
          parentExpanded && expanded
        );
        result.push(...childItems);
      }
    }

    return result;
  }

  let flatNodes = $derived(flattenTree($wbsTree));
  let visibleNodes = $derived(flatNodes.filter((item) => item.visible));
</script>

<div class="wbs-list-view">
  <!-- ヘッダー：全体進捗 -->
  <WBSHeader />

  <!-- ツリービュー -->
  <div class="wbs-tree" role="tree" aria-label="WBS ツリー">
    {#if visibleNodes.length === 0}
      <div class="empty-state">
        <p>タスクがありません</p>
        <p class="empty-hint">チャットからタスクを生成してください</p>
      </div>
    {:else}
      {#each visibleNodes as { node, expanded }, i (node.id)}
        <WBSNode {node} {expanded} index={i} />
      {/each}
    {/if}
  </div>
</div>

<style>
  .wbs-list-view {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--mv-color-surface-base);
    background-image: radial-gradient(
      var(--mv-color-border-subtle) 1px,
      transparent 1px
    );
    background-size: 20px 20px;
  }

  /* ツリービュー */
  .wbs-tree {
    flex: 1;
    overflow-y: auto;
    padding: var(--mv-spacing-sm);
  }

  /* カスタムスクロールバー */
  .wbs-tree::-webkit-scrollbar {
    width: var(--mv-scrollbar-width);
  }

  .wbs-tree::-webkit-scrollbar-track {
    background: var(--mv-color-surface-node);
  }

  .wbs-tree::-webkit-scrollbar-thumb {
    background: var(--mv-color-border-default);
    border-radius: var(--mv-scrollbar-radius);
  }

  .wbs-tree::-webkit-scrollbar-thumb:hover {
    background: var(--mv-color-border-strong);
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
