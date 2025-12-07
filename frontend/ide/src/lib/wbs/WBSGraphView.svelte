<script lang="ts">
  import { onMount } from "svelte";
  import WBSGraphNode from "./WBSGraphNode.svelte";
  import WBSHeader from "./WBSHeader.svelte";
  import ZoomControls from "./ZoomControls.svelte";
  import {
    wbsTree,
    expandedNodes,
    overallProgress,
  } from "../../stores/wbsStore";
  import type { WBSNode as WBSNodeType } from "../../stores/wbsStore";
  import {
    GRAPH_NODE_WIDTH as NODE_WIDTH,
    GRAPH_NODE_HEIGHT as NODE_HEIGHT,
    HORIZONTAL_GAP,
    VERTICAL_GAP,
    GRAPH_PADDING as PADDING,
  } from "./utils";

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
    return `M ${startX} ${startY} C ${startX + controlOffset} ${startY}, ${
      endX - controlOffset
    } ${endY}, ${endX} ${endY}`;
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

  let positionedNodes = $derived(calculateLayout($wbsTree));
  let edges = $derived(getEdges(positionedNodes));
  let canvasWidth = $derived(Math.max(
    800,
    ...positionedNodes.map((n) => n.x + NODE_WIDTH + PADDING)
  ));
  let canvasHeight = $derived(Math.max(
    400,
    ...positionedNodes.map((n) => n.y + NODE_HEIGHT + PADDING)
  ));

  // --- Zoom & Pan Logic ---
  let container: HTMLDivElement = $state();
  let isDragging = false;
  let scale = $state(0.5); // Default 50%
  let translateX = $state(0);
  let translateY = $state(0);
  let startX = 0;
  let startY = 0;

  function handleMouseDown(e: MouseEvent) {
    if (e.button !== 0) return; // Only left click for pan (or consider middle click?)
    // If clicking on a node (interactive element), propagation should be stopped there.
    // Assuming container gets event only if background.
    isDragging = true;
    startX = e.clientX - translateX;
    startY = e.clientY - translateY;
    container.style.cursor = "grabbing";
  }

  function handleMouseMove(e: MouseEvent) {
    if (!isDragging) return;
    e.preventDefault();
    translateX = e.clientX - startX;
    translateY = e.clientY - startY;
  }

  function handleMouseUp() {
    isDragging = false;
    if (container) container.style.cursor = "grab";
  }

  function handleWheel(e: WheelEvent) {
    if (e.ctrlKey || e.metaKey || true) {
      // Always zoom on wheel for 'infinite canvas' feel?
      // Or stick to standard scroll unless modifier key.
      // Factorio/Figma: Wheel zooms.
      e.preventDefault();
      const zoomSensitivity = 0.001;
      const delta = -e.deltaY * zoomSensitivity;
      const newScale = Math.min(Math.max(0.1, scale + delta), 3.0);

      // Adjust translation to zoom towards cursor
      // (Optional: simple center zoom for now to keep implementation lighter,
      //  or cursor zoom if calculation is easy)
      // Simple implementation: Zoom centered or just zoom.
      // Better implementation: Zoom towards cursor.
      // Let's do simple zoom for this iteration to avoid complex math bugs initially,
      // or just scale.
      scale = newScale;
    }
  }

  function onZoomIn() {
    scale = Math.min(3.0, scale + 0.1);
  }

  function onZoomOut() {
    scale = Math.max(0.1, scale - 0.1);
  }

  function onReset() {
    scale = 1.0;
    translateX = 0;
    translateY = 0;
  }
</script>

<div class="wbs-graph-view">
  <!-- ヘッダー (Absolute overlay) -->
  <div class="header-overlay">
    <WBSHeader />
  </div>

  <!-- グラフキャンバス -->
  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
  <div
    class="graph-container"
    bind:this={container}
    onmousedown={handleMouseDown}
    onmousemove={handleMouseMove}
    onmouseup={handleMouseUp}
    onmouseleave={handleMouseUp}
    onwheel={handleWheel}
    role="application"
    aria-label="WBS グラフキャンバス"
  >
    {#if positionedNodes.length === 0}
      <div
        class="empty-state"
        style:transform={`translate(${translateX}px, ${translateY}px) scale(${scale})`}
      >
        <p>タスクがありません</p>
        <p class="empty-hint">チャットからタスクを生成してください</p>
      </div>
    {:else}
      <div
        class="canvas-world"
        style:width="{canvasWidth}px"
        style:height="{canvasHeight}px"
      >
        <!-- 
          NOTE: テキストの滲みを防ぐため、コンテナ一括transformではなくレイヤー別に適用する。
          SVGレイヤーは transform: scale で問題ない。
          ノード（DOM）レイヤーは zoom プロパティでリフローさせる。
        -->

        <!-- 接続線 (SVG) -->
        <svg
          class="connections-layer"
          width={canvasWidth}
          height={canvasHeight}
          style:transform={`translate(${translateX}px, ${translateY}px) scale(${scale})`}
          style:transform-origin="0 0"
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
        <div
          class="nodes-layer"
          style:zoom={scale}
          style:transform={`translate(${translateX}px, ${translateY}px)`}
          style:transform-origin="0 0"
        >
          {#each positionedNodes as { node, x, y } (node.id)}
            <WBSGraphNode {node} {x} {y} />
          {/each}
        </div>
      </div>
    {/if}
  </div>

  <ZoomControls
    {scale}
    on:zoomIn={onZoomIn}
    on:zoomOut={onZoomOut}
    on:reset={onReset}
  />

  <!-- 操作ヒント -->
  <div class="controls-hint">
    <span>ドラッグ: 移動</span>
    <span>ホイール: ズーム</span>
    <span>クリック: 選択/展開</span>
  </div>
</div>

<style>
  .wbs-graph-view {
    position: relative;
    width: var(--mv-size-full);
    height: var(--mv-size-full);
    overflow: hidden;
    background: var(--mv-color-surface-base);
    background-image: radial-gradient(
      var(--mv-color-border-subtle) 1px,
      transparent 1px
    );
    background-size: 20px 20px;
  }

  .header-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    z-index: 10;
    pointer-events: none; /* Let clicks pass through to canvas if not on header elements */
  }

  .header-overlay :global(*) {
    pointer-events: auto;
  }

  .graph-container {
    width: var(--mv-size-full);
    height: var(--mv-size-full);
    cursor: grab;
    touch-action: none;
    overflow: hidden;
  }

  .canvas-world {
    position: absolute;
    top: 0;
    left: 0;
    will-change: transform;
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
    transition: stroke var(--mv-transition-hover);
  }

  /* ノードレイヤー */
  .nodes-layer {
    position: absolute;
    top: 0;
    left: 0;
  }

  .empty-state {
    position: absolute;
    top: var(--mv-size-half);
    left: var(--mv-size-half);
    text-align: center;
    color: var(--mv-color-text-muted);
    margin-left: var(--mv-empty-state-margin-left);
    margin-top: var(--mv-empty-state-margin-top);
    width: var(--mv-empty-state-width);
  }

  .empty-state p {
    margin: var(--mv-spacing-xxs) 0;
  }

  .empty-hint {
    font-size: var(--mv-font-size-sm);
  }

  /* 操作ヒント */
  .controls-hint {
    position: absolute;
    bottom: calc(var(--mv-spacing-lg) * 2 + var(--mv-icon-size-xxxl));
    right: var(--mv-spacing-lg);
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-xs);
    padding: var(--mv-spacing-sm);
    background: var(--mv-color-surface-secondary); /* Semi-transparent? */
    border-radius: var(--mv-radius-md);
    box-shadow: var(--mv-shadow-card);
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-muted);
    pointer-events: none;
    opacity: 0.8;
  }

  .controls-hint span {
    background: transparent;
    padding: 0;
  }
</style>
