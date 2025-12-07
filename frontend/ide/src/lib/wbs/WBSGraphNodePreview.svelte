<script lang="ts">
  import WBSGraphNode from "./WBSGraphNode.svelte";
  import type { WBSNode } from "../../stores/wbsStore";
  import type { TaskStatus } from "../../schemas";

  
  interface Props {
    // WBSNode properties exposed as controls
    label?: string;
    type?: "phase" | "task";
    phaseName?: string;
    status?: TaskStatus;
    hasChildren?: boolean;
    expanded?: boolean;
  }

  let {
    label = "Sample Node",
    type = "task",
    phaseName = "実装",
    status = "PENDING",
    hasChildren = false,
    expanded = false
  }: Props = $props();

  // Construct node object from flat props
  let node = $derived({
    id: "sample-id",
    type,
    label,
    phaseName,
    level: 1,
    children: hasChildren ? ["child-1"] : [],
    task:
      type === "task"
        ? {
            id: "task-1",
            title: label,
            status,
            phaseName,
            poolId: "default",
            createdAt: "",
            updatedAt: "",
            dependencies: [],
          }
        : undefined,
    progress: { total: 0, completed: 0, percentage: 0 },
  } as unknown as WBSNode); // Cast as WBSNode
</script>

<div
  style="position: relative; width: var(--mv-preview-graph-node-width); height: var(--mv-preview-graph-node-height);"
>
  <WBSGraphNode {node} x={50} y={20} />
</div>
