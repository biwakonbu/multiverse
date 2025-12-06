<script lang="ts">
  import { onMount } from "svelte";
  import WBSNode from "./WBSNode.svelte";
  import {
    wbsTree,
    expandedNodes,
    overallProgress,
  } from "../../stores/wbsStore";
  import type { WBSNode as WBSNodeType } from "../../stores/wbsStore";

  // グラフ描画用の定数
  const NODE_WIDTH = 200;
  const NODE_HEIGHT = 60;
  const HORIZONTAL_GAP = 80;
  const VERTICAL_GAP = 40;
  const PADDING = 40;

  // ノード位置を計算
  interface PositionedNode {
    node: WBSNodeType;
    x: number;
    y: number;
    level: number;
  }

  // ツリーからグラフレイアウトを生成
  function calculateLayout(
    nodes: WBSNodeType[],
    level: number = 0,
    startY: number = PADDING
  ): PositionedNode[] {
    const result: PositionedNode[] = [];
    let currentY = startY;

    for (const node of nodes) {
      const x = PADDING + level * (NODE_WIDTH + HORIZONTAL_GAP);
      const y = currentY;

      result.push({ node, x, y, level });

      // 子ノードを再帰的にレイアウト
      if (node.children.length > 0 && $expandedNodes.has(node.id)) {
        const childNodes = calculateLayout(node.children, level + 1, currentY);
        result.push(...childNodes);
        // 子ノードの高さ分だけ次の位置を調整
        const childHeight = childNodes.length * (NODE_HEIGHT + VERTICAL_GAP);
        currentY += Math.max(NODE_HEIGHT + VERTICAL_GAP, childHeight);
      } else {
        currentY += NODE_HEIGHT + VERTICAL_GAP;
      }
    }

    return result;
  }

  // 接続線のパスを生成
  function getConnectionPath(from: PositionedNode, to: PositionedNode): string {
    const startX = from.x + NODE_WIDTH;
    const startY = from.y + NODE_HEIGHT / 2;
    const endX = to.x;
    const endY = to.y + NODE_HEIGHT / 2;

    // ベジェ曲線で滑らかな接続
    const controlOffset = HORIZONTAL_GAP / 2;
    return `M ${startX} ${startY} C ${startX + controlOffset} ${startY}, ${endX - controlOffset} ${endY}, ${endX} ${endY}`;
  }

  // 親子関係からエッジを生成
  function getEdges(
    nodes: PositionedNode[]
  ): Array<{ from: PositionedNode; to: PositionedNode }> {
    const edges: Array<{ from: PositionedNode; to: PositionedNode }> = [];
    const nodeMap = new Map(nodes.map((n) => [n.node.id, n]));

    for (const positioned of nodes) {
      if (
        positioned.node.children.length > 0 &&
        $expandedNodes.has(positioned.node.id)
      ) {
        for (const child of positioned.node.children) {
          const childPositioned = nodeMap.get(child.id);
          if (childPositioned) {
            edges.push({ from: positioned, to: childPositioned });
          }
        }
      }
    }

    return edges;
  }

  $: positionedNodes = calculateLayout($wbsTree);
  $: edges = getEdges(positionedNodes);
  $: canvasWidth = Math.max(
    800,
    ...positionedNodes.map((n) => n.x + NODE_WIDTH + PADDING)
  );
  $: canvasHeight = Math.max(
    400,
    ...positionedNodes.map((n) => n.y + NODE_HEIGHT + PADDING)
  );

  // ドラッグスクロール
  let container: HTMLDivElement;
  let isDragging = false;
  let startX = 0;
  let startY = 0;
  let scrollLeft = 0;
  let scrollTop = 0;

  function handleMouseDown(e: MouseEvent) {
    if (e.button !== 0) return;
    isDragging = true;
    startX = e.clientX;
    startY = e.clientY;
    scrollLeft = container.scrollLeft;
    scrollTop = container.scrollTop;
  }

  function handleMouseMove(e: MouseEvent) {
    if (!isDragging) return;
    e.preventDefault();
    const deltaX = e.clientX - startX;
    const deltaY = e.clientY - startY;
    container.scrollLeft = scrollLeft - deltaX;
    container.scrollTop = scrollTop - deltaY;
  }

  function handleMouseUp() {
    isDragging = false;
  }

  function getPhaseClass(phaseName: string | undefined): string {
    switch (phaseName) {
      case "概念設計":
        return "phase-concept";
      case "実装設計":
        return "phase-design";
      case "実装":
        return "phase-impl";
      case "検証":
        return "phase-verify";
      default:
        return "";
    }
  }

  function getStatusClass(node: WBSNodeType): string {
    if (node.type === "phase") {
      return "";
    }
    return node.task ? `status-${node.task.status.toLowerCase()}` : "";
  }
</script>

<div class="wbs-graph-view">
  <!-- ヘッダー -->
  <header class="graph-header">
    <div class="header-title">
      <h2>WBS グラフ</h2>
      <span class="task-count">
        {$overallProgress.completed} / {$overallProgress.total} タスク完了
      </span>
    </div>

    <div class="header-progress">
      <div class="progress-bar-large">
        <div
          class="progress-fill"
          style:--progress="{$overallProgress.percentage}%"
        ></div>
      </div>
      <span class="progress-percentage">
        <span class="progress-number">{$overallProgress.percentage}</span><span
          class="progress-symbol">%</span
        >
      </span>
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

  <!-- グラフキャンバス -->
  <div
    class="graph-container"
    class:dragging={isDragging}
    bind:this={container}
    on:mousedown={handleMouseDown}
    on:mousemove={handleMouseMove}
    on:mouseup={handleMouseUp}
    on:mouseleave={handleMouseUp}
    role="application"
    aria-label="WBS グラフ"
    tabindex="0"
  >
    {#if positionedNodes.length === 0}
      <div class="empty-state">
        <p>タスクがありません</p>
        <p class="empty-hint">チャットからタスクを生成してください</p>
      </div>
    {:else}
      <div
        class="canvas"
        style:width="{canvasWidth}px"
        style:height="{canvasHeight}px"
      >
        <!-- 接続線 (SVG) -->
        <svg
          class="connections-layer"
          width={canvasWidth}
          height={canvasHeight}
        >
          <defs>
            <marker
              id="arrowhead"
              markerWidth="10"
              markerHeight="7"
              refX="9"
              refY="3.5"
              orient="auto"
            >
              <polygon
                points="0 0, 10 3.5, 0 7"
                fill="var(--mv-color-border-default)"
              />
            </marker>
          </defs>
          {#each edges as edge}
            <path
              class="connection-path"
              d={getConnectionPath(edge.from, edge.to)}
              marker-end="url(#arrowhead)"
            />
          {/each}
        </svg>

        <!-- ノード -->
        <div class="nodes-layer">
          {#each positionedNodes as { node, x, y }}
            <div
              class="graph-node {getPhaseClass(node.phaseName)} {getStatusClass(
                node
              )}"
              style:left="{x}px"
              style:top="{y}px"
              style:width="{NODE_WIDTH}px"
              style:height="{NODE_HEIGHT}px"
              on:click={() => {
                if (node.children.length > 0) {
                  expandedNodes.toggle(node.id);
                }
              }}
              on:keydown={(e) => {
                if (e.key === "Enter" && node.children.length > 0) {
                  expandedNodes.toggle(node.id);
                }
              }}
              role="button"
              tabindex="0"
            >
              <div class="phase-bar"></div>
              <div class="node-content">
                <div class="node-title">{node.label}</div>
                <div class="node-meta">
                  {#if node.type === "phase"}
                    <span class="phase-badge">{node.label}</span>
                  {:else if node.task}
                    <span class="status-badge">{node.task.status}</span>
                  {/if}
                  {#if node.children.length > 0}
                    <span class="children-count">
                      {$expandedNodes.has(node.id) ? "▼" : "▶"}
                      {node.children.length}
                    </span>
                  {/if}
                </div>
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}
  </div>

  <!-- 操作ヒント -->
  <div class="controls-hint">
    <span>ドラッグでスクロール</span>
    <span>ノードクリックで展開/折りたたみ</span>
  </div>
</div>

<style>
  .wbs-graph-view {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--mv-color-surface-node);
  }

  /* ヘッダー */
  .graph-header {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-sm);
    padding: var(--mv-spacing-md);
    border-bottom: var(--mv-border-width-thin) solid
      var(--mv-color-border-subtle);
    background: var(--mv-color-surface-hover);
    flex-shrink: 0;
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
    background: var(--mv-progress-bar-bg);
    border-radius: var(--mv-radius-sm);
    overflow: hidden;
    box-shadow: var(--mv-shadow-progress-bar);
    border: var(--mv-border-panel);
  }

  .progress-fill {
    height: 100%;
    width: var(--progress, 0%);
    background: var(--mv-progress-bar-fill);
    border-radius: var(--mv-radius-sm);
    transition: width var(--mv-duration-slow);
    box-shadow: var(--mv-shadow-glow-sm);
  }

  .progress-percentage {
    display: flex;
    align-items: baseline;
    font-family: var(--mv-font-mono);
    color: var(--mv-progress-text-color);
    min-width: var(--mv-progress-text-width-md);
    justify-content: flex-end;
    text-shadow: var(--mv-text-shadow-glow);
  }

  .progress-number {
    font-family: var(--mv-font-display);
    font-size: var(--mv-font-size-xl);
    font-weight: var(--mv-font-weight-bold);
    letter-spacing: var(--mv-letter-spacing-tight);
  }

  .progress-symbol {
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-semibold);
    margin-left: var(--mv-border-width-thin);
    opacity: 0.85;
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

  /* グラフコンテナ */
  .graph-container {
    flex: 1;
    overflow: auto;
    position: relative;
    cursor: grab;
    touch-action: none;
  }

  .graph-container.dragging {
    cursor: grabbing;
  }

  /* カスタムスクロールバー */
  .graph-container::-webkit-scrollbar {
    width: var(--mv-scrollbar-width);
    height: var(--mv-scrollbar-width);
  }

  .graph-container::-webkit-scrollbar-track {
    background: var(--mv-color-surface-node);
  }

  .graph-container::-webkit-scrollbar-thumb {
    background: var(--mv-color-border-default);
    border-radius: var(--mv-scrollbar-radius);
  }

  .graph-container::-webkit-scrollbar-thumb:hover {
    background: var(--mv-color-border-strong);
  }

  .graph-container::-webkit-scrollbar-corner {
    background: var(--mv-color-surface-node);
  }

  /* キャンバス */
  .canvas {
    position: relative;
    min-width: var(--mv-size-full);
    min-height: var(--mv-size-full);
  }

  /* 接続線 */
  .connections-layer {
    position: absolute;
    top: 0;
    left: 0;
    pointer-events: none;
  }

  .connection-path {
    fill: none;
    stroke: var(--mv-color-border-default);
    stroke-width: 2;
  }

  /* ノードレイヤー */
  .nodes-layer {
    position: absolute;
    top: 0;
    left: 0;
  }

  /* グラフノード */
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

  .phase-badge,
  .status-badge {
    font-size: var(--mv-font-size-xs);
    padding: 0 var(--mv-spacing-xxs);
    border-radius: var(--mv-radius-sm);
    text-transform: uppercase;
    letter-spacing: var(--mv-letter-spacing-wide);
  }

  .phase-badge {
    background: var(--mv-color-surface-secondary);
    color: var(--mv-color-text-secondary);
  }

  .status-badge {
    font-weight: var(--mv-font-weight-medium);
  }

  /* ステータス別スタイル */
  .status-pending .status-badge {
    background: var(--mv-color-status-pending-bg);
    color: var(--mv-color-status-pending-text);
  }

  .status-running .status-badge {
    background: var(--mv-color-status-running-bg);
    color: var(--mv-color-status-running-text);
  }

  .status-succeeded .status-badge {
    background: var(--mv-color-status-succeeded-bg);
    color: var(--mv-color-status-succeeded-text);
  }

  .status-failed .status-badge {
    background: var(--mv-color-status-failed-bg);
    color: var(--mv-color-status-failed-text);
  }

  .children-count {
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-muted);
    font-family: var(--mv-font-mono);
  }

  /* 空状態 */
  .empty-state {
    position: absolute;
    top: var(--mv-size-half);
    left: var(--mv-size-half);
    transform: translate(-50%, -50%);
    text-align: center;
    color: var(--mv-color-text-muted);
  }

  .empty-state p {
    margin: var(--mv-spacing-xxs) 0;
  }

  .empty-hint {
    font-size: var(--mv-font-size-sm);
  }

  /* 操作ヒント */
  .controls-hint {
    display: flex;
    gap: var(--mv-spacing-sm);
    padding: var(--mv-spacing-xs) var(--mv-spacing-md);
    background: var(--mv-color-surface-secondary);
    border-top: var(--mv-border-width-thin) solid var(--mv-color-border-subtle);
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-muted);
    flex-shrink: 0;
  }

  .controls-hint span {
    background: var(--mv-color-surface-primary);
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    border-radius: var(--mv-radius-sm);
  }
</style>
