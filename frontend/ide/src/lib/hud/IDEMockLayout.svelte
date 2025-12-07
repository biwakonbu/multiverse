<script lang="ts">
  import ProcessHUD from "./ProcessHUD.svelte";
  import type { ResourceNode } from "./types";

  interface Props {
    executionState?: "IDLE" | "RUNNING" | "PAUSED";
    activeTaskTitle?: string | undefined;
  }

  let {
    executionState = "RUNNING",
    activeTaskTitle = "Designing new UI components",
  }: Props = $props();

  // モック用リソースデータ
  const resources: ResourceNode[] = [
    {
      id: "sys",
      name: "Multiverse Orchestrator",
      type: "ORCHESTRATOR",
      status: "RUNNING",
      expanded: true,
      children: [
        {
          id: "meta-1",
          name: "Meta-Agent",
          type: "META",
          status: "THINKING",
          detail: "Analyzing code structure...",
          children: [],
        },
      ],
    },
  ];

  // モック用タスクノード
  const mockNodes = [
    { id: "1", title: "UI設計", status: "SUCCEEDED", x: 100, y: 80 },
    { id: "2", title: "コンポーネント実装", status: "RUNNING", x: 350, y: 80 },
    { id: "3", title: "テスト作成", status: "PENDING", x: 600, y: 80 },
    { id: "4", title: "ドキュメント", status: "PENDING", x: 350, y: 200 },
  ];
</script>

<div class="ide-mock-container">
  <!-- Toolbar モック -->
  <div class="toolbar-mock">
    <div class="toolbar-left">
      <span class="brand">MULTIVERSE</span>
    </div>
    <div class="toolbar-center">
      <span class="workspace-name">Demo Workspace</span>
    </div>
    <div class="toolbar-right">
      <button class="view-btn active">Graph</button>
      <button class="view-btn">List</button>
    </div>
  </div>

  <!-- Main Content (WBSGraphView風) -->
  <div class="main-content">
    <!-- グリッド背景 -->
    <svg class="grid-background">
      <defs>
        <pattern id="grid" width="40" height="40" patternUnits="userSpaceOnUse">
          <path d="M 40 0 L 0 0 0 40" fill="none" stroke="var(--mv-glass-border-subtle)" stroke-width="0.5" />
        </pattern>
      </defs>
      <rect width="100%" height="100%" fill="url(#grid)" />
    </svg>

    <!-- モックノード -->
    <div class="nodes-container">
      {#each mockNodes as node}
        <div
          class="mock-node"
          class:succeeded={node.status === "SUCCEEDED"}
          class:running={node.status === "RUNNING"}
          class:pending={node.status === "PENDING"}
          style="left: {node.x}px; top: {node.y}px;"
        >
          <div class="node-status"></div>
          <div class="node-title">{node.title}</div>
        </div>
      {/each}

      <!-- 接続線 -->
      <svg class="connections">
        <line x1="220" y1="110" x2="350" y2="110" />
        <line x1="470" y1="110" x2="600" y2="110" />
        <line x1="410" y1="130" x2="410" y2="200" />
      </svg>
    </div>
  </div>

  <!-- ProcessHUD -->
  <ProcessHUD {executionState} {resources} {activeTaskTitle} />
</div>

<style>
  .ide-mock-container {
    display: flex;
    flex-direction: column;
    width: 100vw;
    height: 100vh;
    background: var(--mv-color-bg-primary);
    color: var(--mv-color-text-primary);
    font-family: var(--mv-font-sans);
    overflow: hidden;
    position: relative;
  }

  /* Toolbar モック */
  .toolbar-mock {
    height: var(--mv-layout-toolbar-height);
    background: var(--mv-glass-bg-dark);
    border-bottom: var(--mv-border-width-sm) solid var(--mv-glass-border-subtle);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 var(--mv-spacing-md);
    flex-shrink: 0;
  }

  .toolbar-left {
    display: flex;
    align-items: center;
  }

  .brand {
    font-size: var(--mv-font-size-lg);
    font-weight: var(--mv-font-weight-bold);
    letter-spacing: var(--mv-letter-spacing-heading);
    color: var(--mv-color-brand-primary);
  }

  .toolbar-center {
    display: flex;
    align-items: center;
  }

  .workspace-name {
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-secondary);
  }

  .toolbar-right {
    display: flex;
    gap: var(--mv-spacing-xs);
  }

  .view-btn {
    padding: var(--mv-spacing-xs) var(--mv-spacing-sm);
    background: transparent;
    border: var(--mv-border-width-sm) solid var(--mv-glass-border-subtle);
    border-radius: var(--mv-radius-sm);
    color: var(--mv-color-text-secondary);
    font-size: var(--mv-font-size-sm);
    cursor: pointer;
    transition: var(--mv-transition-base);
  }

  .view-btn:hover {
    background: var(--mv-glass-hover);
  }

  .view-btn.active {
    background: var(--mv-color-brand-primary);
    border-color: var(--mv-color-brand-primary);
    color: var(--mv-color-text-primary);
  }

  /* Main Content */
  .main-content {
    flex: 1;
    position: relative;
    overflow: hidden;
  }

  .grid-background {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
  }

  .nodes-container {
    position: absolute;
    inset: 0;
  }

  /* モックノード */
  .mock-node {
    position: absolute;
    width: var(--mv-mock-node-width);
    padding: var(--mv-spacing-sm);
    background: var(--mv-glass-bg);
    border: var(--mv-border-width-sm) solid var(--mv-glass-border-subtle);
    border-radius: var(--mv-radius-md);
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
  }

  .node-status {
    width: var(--mv-indicator-size-sm);
    height: var(--mv-indicator-size-sm);
    border-radius: var(--mv-radius-full);
    flex-shrink: 0;
  }

  .mock-node.succeeded .node-status {
    background: var(--mv-color-status-succeeded);
  }

  .mock-node.running .node-status {
    background: var(--mv-color-status-running);
    animation: pulse 1.5s ease-in-out infinite;
  }

  .mock-node.pending .node-status {
    background: var(--mv-color-status-pending);
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }

  .node-title {
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* 接続線 */
  .connections {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
  }

  .connections line {
    stroke: var(--mv-glass-border-subtle);
    stroke-width: 2;
  }
</style>
