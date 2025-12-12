<script lang="ts">
  import { tasks, selectedTaskId } from "../../../stores/taskStore";
  import { statusLabels } from "../../../types";
  import { run } from "svelte/legacy";

  let selectedTask = $derived($tasks.find((t) => t.id === $selectedTaskId));

  function handleClose() {
    selectedTaskId.select(null);
  }

  // Auto-close if task disappears (e.g. filtered out?)
  run(() => {
    if ($selectedTaskId && !selectedTask) {
      // selectedTaskId.select(null); // Optional: keep selection even if not found?
    }
  });

  // Helper for JSON pretty print
  function formatJSON(obj: any): string {
    return JSON.stringify(obj, null, 2);
  }
</script>

{#if selectedTask}
  <div class="prop-panel">
    <div class="header">
      <div class="title-row">
        <h3>{selectedTask.title}</h3>
        <button class="close-btn" onclick={handleClose} aria-label="Close"
          >×</button
        >
      </div>
      <div class="meta-row">
        <span class="status-badge status-{selectedTask.status.toLowerCase()}">
          {statusLabels[selectedTask.status]}
        </span>
        <span class="id">{selectedTask.id}</span>
      </div>
    </div>

    <div class="content">
      <section>
        <h4>Description</h4>
        <p class="description">
          {selectedTask.description || "No description provided."}
        </p>
      </section>

      {#if selectedTask.suggestedImpl}
        <section class="impl-section">
          <h4>Suggested Implementation</h4>
          <div class="field">
            <span class="field-label">Language:</span>
            <span>{selectedTask.suggestedImpl.language || "N/A"}</span>
          </div>

          {#if selectedTask.suggestedImpl.filePaths?.length}
            <div class="field">
              <span class="field-label">Files to Change:</span>
              <ul class="file-list">
                {#each selectedTask.suggestedImpl.filePaths as file}
                  <li>{file}</li>
                {/each}
              </ul>
            </div>
          {/if}

          {#if selectedTask.suggestedImpl.constraints?.length}
            <div class="field">
              <span class="field-label">Constraints:</span>
              <ul class="constraint-list">
                {#each selectedTask.suggestedImpl.constraints as constraint}
                  <li>{constraint}</li>
                {/each}
              </ul>
            </div>
          {/if}
        </section>
      {/if}

      {#if selectedTask.artifacts}
        <section class="artifacts-section">
          <h4>Artifacts</h4>
          {#if selectedTask.artifacts.files?.length}
            <div class="field">
              <span class="field-label">Generated Files:</span>
              <ul class="file-list">
                {#each selectedTask.artifacts.files as file}
                  <li>{file}</li>
                {/each}
              </ul>
            </div>
          {/if}
          {#if selectedTask.artifacts.logs?.length}
            <div class="field">
              <span class="field-label">Logs:</span>
              <ul class="file-list">
                {#each selectedTask.artifacts.logs as log}
                  <li>{log}</li>
                {/each}
              </ul>
            </div>
          {/if}
        </section>
      {/if}

      <section>
        <h4>Details</h4>
        <div class="field">
          <span class="field-label">Phase:</span>
          {selectedTask.phaseName || "-"}
        </div>
        <div class="field">
          <span class="field-label">Dependencies:</span>
          {selectedTask.dependencies?.length || 0}
        </div>
      </section>
    </div>
  </div>
{/if}

<style>
  .prop-panel {
    position: absolute;
    top: var(--mv-spacing-md);
    right: var(--mv-spacing-md);
    width: var(--mv-prop-panel-width, 320px);
    max-height: calc(100% - var(--mv-spacing-md) * 2);
    background: var(--mv-glass-bg-panel);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
    border-radius: var(--mv-radius-lg);
    box-shadow: var(--mv-shadow-glass-panel-full);
    backdrop-filter: blur(var(--mv-blur-md, 12px));
    display: flex;
    flex-direction: column;
    z-index: 2000;
    color: var(--mv-color-text-primary);
    overflow: hidden;
  }

  .header {
    padding: var(--mv-spacing-md);
    border-bottom: var(--mv-border-width-thin) solid
      var(--mv-glass-border-subtle);
    background: var(--mv-glass-bg-header);
  }

  .title-row {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: var(--mv-spacing-sm);
  }

  h3 {
    margin: 0;
    font-size: var(--mv-font-size-md);
    font-weight: var(--mv-font-weight-semibold);
    line-height: var(--mv-line-height-tight);
  }

  .close-btn {
    background: transparent;
    border: none;
    color: var(--mv-color-text-muted);
    font-size: var(--mv-font-size-xl);
    line-height: 1;
    cursor: pointer;
    padding: 0;
  }
  .close-btn:hover {
    color: var(--mv-color-text-primary);
  }

  .meta-row {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-sm);
    margin-top: var(--mv-spacing-sm);
  }

  .id {
    font-size: var(--mv-font-size-xs);
    font-family: var(--mv-font-mono);
    color: var(--mv-color-text-muted);
  }

  .status-badge {
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-bold);
    padding: var(--mv-spacing-xs) var(--mv-spacing-sm);
    border-radius: var(--mv-radius-sm);
    text-transform: uppercase;
  }

  .status-pending {
    color: var(--mv-color-status-pending-text);
    background: var(--mv-glass-bg-dark);
  }

  .status-running {
    color: var(--mv-color-status-running-text);
    background: var(--mv-glass-bg-dark);
  }

  .status-succeeded {
    color: var(--mv-color-status-succeeded-text);
    background: var(--mv-glass-bg-dark);
  }

  .content {
    padding: var(--mv-spacing-md);
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-lg);
  }

  section h4 {
    margin: 0 0 var(--mv-spacing-sm) 0;
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-secondary);
    text-transform: uppercase;
    letter-spacing: var(--mv-letter-spacing-wide);
    border-bottom: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
    padding-bottom: var(--mv-spacing-xs);
  }

  .description {
    font-size: var(--mv-font-size-sm);
    line-height: var(--mv-line-height-normal);
    color: var(--mv-color-text-primary);
  }

  .field {
    margin-bottom: var(--mv-spacing-xs);
    font-size: var(--mv-font-size-sm);
  }

  .field-label {
    color: var(--mv-color-text-muted);
    margin-right: var(--mv-spacing-xs);
    display: inline-block;
  }

  .file-list,
  .constraint-list {
    margin: var(--mv-spacing-xs) 0 0 0;
    padding-left: var(--mv-spacing-md);
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-xs);
    color: var(--mv-primitive-frost-1); /* Cyan-ish */
  }

  .impl-section {
    background: var(--mv-glass-bg-active);
    padding: var(--mv-spacing-sm);
    border-radius: var(--mv-radius-md);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
  }
</style>
