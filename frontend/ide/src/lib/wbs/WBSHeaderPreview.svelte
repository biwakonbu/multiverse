<script lang="ts">
  import { run } from 'svelte/legacy';

  import WBSHeader from "./WBSHeader.svelte";
  import { tasks } from "../../stores/taskStore";
  import type { Task, TaskStatus } from "../../types";

  interface Props {
    percentage?: number; // Not used directly, derived from completed/total
    completed?: number;
    total?: number;
  }

  let { percentage = 0, completed = 0, total = 10 }: Props = $props();

  // Function to create dummy task
  const createDummyTask = (id: string, status: TaskStatus): Task => ({
    id,
    title: `Task ${id}`,
    status,
    phaseName: "実装",
    poolId: "default",
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    dependencies: [],
  });

  run(() => {
    const dummyTasks: Task[] = [];
    // Add completed tasks
    for (let i = 0; i < completed; i++) {
      dummyTasks.push(createDummyTask(`c-${i}`, "SUCCEEDED"));
    }
    // Add remaining tasks (pending)
    const remaining = Math.max(0, total - completed);
    for (let i = 0; i < remaining; i++) {
      dummyTasks.push(createDummyTask(`p-${i}`, "PENDING"));
    }

    // Update store
    tasks.setTasks(dummyTasks);
  });
</script>

<div
  style="width: var(--mv-preview-header-width); background: var(--mv-color-surface-app); padding: var(--mv-preview-header-padding);"
>
  <WBSHeader />
</div>
