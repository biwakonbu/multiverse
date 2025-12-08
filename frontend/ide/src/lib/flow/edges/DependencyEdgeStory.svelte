<script lang="ts">
  import { SvelteFlow, MarkerType } from "@xyflow/svelte";
  import DependencyEdge from "./DependencyEdge.svelte";
  import "@xyflow/svelte/dist/style.css";
  import TaskNode from "../nodes/TaskNode.svelte"; // Use TaskNode for realism
  import type { Task } from "../../../types";

  interface Props {
    satisfied?: boolean;
    animated?: boolean;
  }

  let { satisfied = false, animated = false }: Props = $props();

  const nodeTypes = { custom: TaskNode };
  const edgeTypes = { dependency: DependencyEdge };

  // Create two nodes to connect
  const sourceTask: Task = {
    id: "source",
    poolId: "S-1",
    title: "Source Task",
    status: "COMPLETED",
    phaseName: "概念設計",
    dependencies: [],
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  };

  const targetTask: Task = {
    id: "target",
    poolId: "T-1",
    title: "Target Task",
    status: "PENDING",
    phaseName: "実装設計",
    dependencies: ["source"],
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  };

  let nodes = [
    {
      id: "source",
      type: "custom",
      position: { x: 50, y: 150 },
      data: { task: sourceTask },
    },
    {
      id: "target",
      type: "custom",
      position: { x: 450, y: 150 },
      data: { task: targetTask },
    },
  ];

  let edges = $derived([
    {
      id: "e1-2",
      source: "source",
      target: "target",
      type: "dependency",
      animated: animated,
      data: {
        satisfied: satisfied, // Assuming DependencyEdge uses this data prop
      },
    },
  ]);
</script>

<div class="edge-story-container">
  <!-- Include Grid Background for context -->
  <div class="grid-background" style="position: absolute; inset: 0;">
    <svg class="grid-pattern" width="100%" height="100%">
      <defs>
        <pattern
          id="grid-cross-edge-story"
          width="200"
          height="140"
          patternUnits="userSpaceOnUse"
        >
          <path
            d="M96 70H104M100 66V74"
            stroke="var(--mv-primitive-aurora-yellow)"
            stroke-width="1"
            opacity="0.15"
          />
        </pattern>
      </defs>
      <rect width="100%" height="100%" fill="url(#grid-cross-edge-story)" />
    </svg>
  </div>

  <!-- Define Markers (Copied from UnifiedFlowCanvas) -->
  <svg class="markers-defs">
    <defs>
      <marker
        id="marker-source"
        markerWidth="8"
        markerHeight="8"
        refX="4"
        refY="4"
        orient="auto"
        markerUnits="userSpaceOnUse"
      >
        <circle
          cx="4"
          cy="4"
          r="2.5"
          fill="var(--mv-color-surface-app)"
          stroke="var(--mv-color-text-muted)"
          stroke-width="1"
        />
      </marker>
      <marker
        id="marker-terminal-satisfied"
        markerWidth="10"
        markerHeight="10"
        refX="5"
        refY="5"
        orient="auto"
        markerUnits="userSpaceOnUse"
      >
        <rect
          x="2"
          y="2"
          width="6"
          height="6"
          fill="var(--mv-color-status-succeeded-border)"
        />
      </marker>
      <marker
        id="marker-terminal-unsatisfied"
        markerWidth="12"
        markerHeight="12"
        refX="6"
        refY="6"
        orient="auto"
        markerUnits="userSpaceOnUse"
      >
        <path
          d="M6 1 L11 6 L6 11 L1 6 Z"
          fill="var(--mv-color-status-blocked-border)"
        />
      </marker>
    </defs>
  </svg>

  <SvelteFlow {nodes} {edges} {nodeTypes} {edgeTypes} fitView />
</div>

<style>
  .edge-story-container {
    width: 100%;
    height: var(--mv-preview-chat-height);
    background: var(--mv-color-surface-app);
    position: relative;
  }

  .markers-defs {
    position: absolute;
    width: var(--mv-space-0);
    height: var(--mv-space-0);
  }

  /* stylelint-disable selector-class-pattern -- Svelte Flow library classes */
  :global(.svelte-flow) {
    background: transparent !important;
  }

  :global(.svelte-flow__pane) {
    background: transparent !important;
  }
  /* stylelint-enable selector-class-pattern */
</style>
