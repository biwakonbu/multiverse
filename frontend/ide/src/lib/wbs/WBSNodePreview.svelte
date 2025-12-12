<script lang="ts">
  import WBSNode from "./WBSNode.svelte";
  import type { WBSNode as WBSNodeType } from "../../stores/wbsStore";
  import type { Task, TaskStatus, PhaseName } from "../../types";

  interface Props {
    // Props
    type?: "phase" | "task";
    label?: string;
    phaseName?: PhaseName;
    status?: TaskStatus;
    level?: number;
    completed?: number;
    total?: number;
    expanded?: boolean;
    hasChildren?: boolean;
    hasImpl?: boolean; // New prop for testing
  }

  let {
    type = "task",
    label = "タスク名",
    phaseName = "実装",
    status = "PENDING",
    level = 0,
    completed = 0,
    total = 1,
    expanded = true,
    hasChildren = false,
    hasImpl = false,
  }: Props = $props();

  // WBSNodeを構築
  function buildNode(): WBSNodeType {
    const task: Task | undefined =
      type === "task"
        ? {
            id: "preview-task",
            title: label,
            status,
            poolId: "default",
            phaseName,
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
            dependencies: [],
          }
        : undefined;

    const childTask: Task = {
      id: "child-task-1",
      title: "子タスク",
      status: "PENDING",
      poolId: "default",
      phaseName,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      dependencies: [],
    };

    return {
      id: type === "phase" ? `phase-${phaseName}` : "task-preview",
      type,
      label,
      task,
      phaseName: type === "phase" ? phaseName : undefined,
      level,
      children: hasChildren
        ? [
            {
              id: "child-1",
              type: "task",
              label: "子タスク",
              task: childTask,
              level: level + 1,
              children: [],
              progress: { completed: 0, total: 0, percentage: 0 },
            },
          ]
        : [],
      progress: {
        completed,
        total,
        percentage: total > 0 ? Math.round((completed / total) * 100) : 0,
      },
    };
  }

  let node = $derived(buildNode());
</script>

<div class="preview-container">
  <WBSNode {node} {expanded} />
</div>

<style>
  .preview-container {
    width: var(--mv-space-400, 400px);
    background: var(--mv-color-surface-primary);
    border-radius: var(--mv-radius-md, 8px);
    padding: var(--mv-space-2, 8px);
  }
</style>
