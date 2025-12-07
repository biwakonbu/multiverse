<script lang="ts">
  import { expandedNodes } from "../../stores/wbsStore";
  import type { WBSNode } from "../../stores/wbsStore";
  import { phaseToCssClass } from "../../schemas";
  import { GRAPH_NODE_WIDTH, GRAPH_NODE_HEIGHT } from "./utils";
  import WBSStatusBadge from "./WBSStatusBadge.svelte"; // Import new component

  export let node: WBSNode;
  export let x: number;
  export let y: number;

  $: expanded = $expandedNodes.has(node.id);
  $: phaseClass = phaseToCssClass(node.phaseName);
  // statusClass removed as it's handled by StatusBadge

  function normalizeStatus(status: string): any {
    return status.toLowerCase();
  }

  function handleGenericClick() {
    if (node.children.length > 0) {
      expandedNodes.toggle(node.id);
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter" && node.children.length > 0) {
      expandedNodes.toggle(node.id);
    }
  }
</script>

```
<!-- svelte-ignore a11y-no-noninteractive-tabindex -->
<div
  class="graph-node {phaseClass}"
  style:left="{x}px"
  style:top="{y}px"
  style:width="{GRAPH_NODE_WIDTH}px"
  style:height="{GRAPH_NODE_HEIGHT}px"
  on:click={handleGenericClick}
  on:keydown={handleKeydown}
  role="button"
  tabindex="0"
>
  <div class="phase-bar"></div>
  <div class="node-content">
    <div class="node-title" title={node.label}>{node.label}</div>
    <div class="node-meta">
      {#if node.type === "phase"}
        <span class="phase-badge">{node.label}</span>
      {:else if node.task}
        <WBSStatusBadge status={normalizeStatus(node.task.status)} />
      {/if}
      {#if node.children.length > 0}
        <span class="children-count">
          {expanded ? "▼" : "▶"}
          {node.children.length}
        </span>
      {/if}
    </div>
  </div>
</div>

<style>
  .graph-node {
    position: absolute;
    background: var(--mv-color-surface-node);
    border: var(--mv-border-width-default) solid var(--mv-color-border-default);
    border-radius: var(--mv-radius-md);
    cursor: pointer;
    display: flex;
    overflow: hidden;
    transition:
      border-color var(--mv-transition-hover),
      box-shadow var(--mv-transition-hover),
      transform var(--mv-transition-hover);
    box-shadow: var(--mv-shadow-node-glow); /* 常時微発光 */
  }

  .graph-node:hover {
    border-color: var(--mv-color-border-focus);
    transform: translateY(-2px);
    box-shadow: var(--mv-shadow-card);
  }

  .graph-node:focus {
    outline: none;
    border-color: var(--mv-color-border-focus);
    box-shadow: var(--mv-shadow-focus);
  }

  /* フェーズバー */
  .phase-bar {
    width: var(--mv-spacing-xxs);
    flex-shrink: 0;
  }

  /* Phase Colors - copied from WBSView/Node */
  .phase-concept .phase-bar {
    background: var(--mv-primitive-frost-3);
  }

  .phase-design .phase-bar {
    background: var(--mv-primitive-aurora-purple);
  }

  .phase-impl .phase-bar {
    background: var(--mv-primitive-aurora-green);
  }

  .phase-verify .phase-bar {
    background: var(--mv-primitive-aurora-yellow);
  }

  /* ノードコンテンツ */
  .node-content {
    flex: 1;
    padding: var(--mv-spacing-xs) var(--mv-spacing-sm);
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: var(--mv-spacing-xxs);
  }

  .node-title {
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .node-meta {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
  }

  .phase-badge {
    font-size: var(--mv-font-size-xs);
    padding: 0 var(--mv-spacing-xxs);
    border-radius: var(--mv-radius-sm);
    text-transform: uppercase;
    letter-spacing: var(--mv-letter-spacing-wide);
    background: var(--mv-color-surface-secondary);
    color: var(--mv-color-text-secondary);
  }

  /* status-badge logic moved to StatusBadge.svelte */

  .children-count {
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-muted);
    font-family: var(--mv-font-mono);
  }
</style>
