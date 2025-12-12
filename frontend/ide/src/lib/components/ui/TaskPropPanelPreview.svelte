<script lang="ts">
  import TaskPropPanel from "./TaskPropPanel.svelte";
  import { tasks, selectedTaskId } from "../../../stores/taskStore";
  import type { Task } from "../../../types";

  interface Props {
    task: Task;
  }

  let { task }: Props = $props();

  // Mock store values - update the readable store context
  // Tasks store is readable, so we need to work within the component context
  $effect(() => {
    selectedTaskId.select(task.id);
  });
</script>

<div class="story-wrapper">
  <TaskPropPanel />
</div>

<style>
  .story-wrapper {
    position: relative;
    width: var(--mv-story-wrapper-width, 400px);
    height: var(--mv-story-wrapper-height, 600px);
    background: var(--mv-story-wrapper-bg, var(--mv-glass-bg-dark));
  }
</style>
