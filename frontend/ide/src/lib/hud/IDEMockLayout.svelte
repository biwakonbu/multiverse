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

  // Mock Resources
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
</script>

<div class="ide-mock-container">
  <!-- Sidebar -->
  <div class="sidebar">
    <div class="sidebar-item active">Explorer</div>
    <div class="sidebar-item">Search</div>
    <div class="sidebar-item">Source Control</div>
  </div>

  <!-- Main Content -->
  <div class="main-content">
    <!-- Editor Tabs -->
    <div class="tabs">
      <div class="tab active">App.svelte</div>
      <div class="tab">processStore.ts</div>
    </div>
    <!-- Editor Area -->
    <div class="editor">
      <div class="line">1 import {"{"} onMount {"}"} from 'svelte';</div>
      <div class="line">2</div>
      <div class="line">3 // This is a mock editor view</div>
      <div class="line">4 // To demonstrate HUD overlay</div>
      <div class="line">5</div>
      <div class="line">6 export let name;</div>
    </div>
  </div>

  <!-- Chat Panel (Bottom Right normally, let's say Right Side) -->
  <div class="chat-panel">
    <div class="chat-header">Multiverse Chat</div>
    <div class="chat-body">
      <div class="msg user">Analyze this file</div>
      <div class="msg agent">Processing...</div>
    </div>
  </div>

  <!-- The HUD Overlay -->
  <ProcessHUD {executionState} {resources} {activeTaskTitle} />
</div>

<style>
  .ide-mock-container {
    display: flex;
    width: 100vw;
    height: 100vh;
    background: var(--mv-color-bg-primary);
    color: var(--mv-color-text-secondary);
    font-family: var(--mv-font-sans);
    overflow: hidden;
    position: relative;
  }

  .sidebar {
    width: var(--mv-space-14);
    background: var(--mv-glass-bg-dark);
    border-right: var(--mv-border-width-sm) solid var(--mv-glass-border-subtle);
    display: flex;
    flex-direction: column;
  }

  .sidebar-item {
    padding: var(--mv-space-2) var(--mv-space-4);
    cursor: pointer;
  }
  .sidebar-item.active {
    background: var(--mv-glass-hover);
    color: var(--mv-color-text-primary);
  }

  .main-content {
    flex: 1;
    display: flex;
    flex-direction: column;
  }

  .tabs {
    height: var(--mv-space-8);
    background: var(--mv-glass-bg);
    display: flex;
    align-items: center;
  }

  .tab {
    padding: var(--mv-space-2) var(--mv-space-3);
    background: var(--mv-glass-bg);
    border-right: var(--mv-border-width-sm) solid var(--mv-glass-border-subtle);
    font-size: var(--mv-font-size-sm);
    cursor: pointer;
  }
  .tab.active {
    background: var(--mv-color-bg-primary);
    color: var(--mv-color-text-primary);
  }

  .editor {
    flex: 1;
    padding: var(--mv-space-4);
    font-family: var(--mv-font-mono);
    line-height: var(--mv-line-height-relaxed);
  }

  .chat-panel {
    width: var(--mv-space-16);
    background: var(--mv-glass-bg-dark);
    border-left: var(--mv-border-width-sm) solid var(--mv-glass-border-subtle);
    display: flex;
    flex-direction: column;
  }

  .chat-header {
    padding: var(--mv-space-2);
    background: var(--mv-glass-border);
    font-weight: var(--mv-font-weight-bold);
  }

  .chat-body {
    flex: 1;
    padding: var(--mv-space-2);
  }

  .msg {
    margin-bottom: var(--mv-space-2);
    padding: var(--mv-space-2);
    border-radius: var(--mv-radius-sm);
    font-size: var(--mv-font-size-sm);
  }

  .msg.user {
    background: var(--mv-glass-hover);
    align-self: flex-end;
  }

  .msg.agent {
    background: var(--mv-color-brand-primary);
  }
</style>
