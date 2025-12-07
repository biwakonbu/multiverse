<script lang="ts">
  import { createEventDispatcher } from "svelte";

  // Wails models (type: string) と Storybook プレビュー互換の型定義
  type BacklogItemType = "FAILURE" | "QUESTION" | "BLOCKER";
  interface BacklogItemProps {
    id: string;
    taskId: string;
    type: BacklogItemType | string; // Wails models は string を生成するため
    title: string;
    description: string;
    priority: number;
    createdAt: string | Date; // Wails Go time 型対応
    resolvedAt?: string | Date;
    resolution?: string;
  }

  export let item: BacklogItemProps;

  let resolutionText = "";

  const dispatch = createEventDispatcher<{
    close: void;
    confirm: { text: string };
  }>();

  function handleResolve() {
    dispatch("confirm", { text: resolutionText || "Resolved" });
  }

  function handleClose() {
    dispatch("close");
  }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- svelte-ignore a11y-no-static-element-interactions -->
<div class="dialog-overlay" on:click={handleClose}>
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div class="dialog" on:click|stopPropagation>
    <h4>バックログを解決</h4>
    <p class="dialog-item-title">{item.title}</p>
    <label>
      解決方法:
      <textarea
        bind:value={resolutionText}
        placeholder="どのように解決したかを入力..."
        rows="3"
      />
    </label>
    <div class="dialog-actions">
      <button class="btn-cancel" on:click={handleClose}> キャンセル </button>
      <button class="btn-confirm" on:click={handleResolve}> 解決 </button>
    </div>
  </div>
</div>

<style>
  /* === Dialog Overlay === */
  .dialog-overlay {
    position: fixed;
    inset: 0;
    background: var(--mv-glow-polar);
    backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  /* === Dialog Box === */
  .dialog {
    background: var(--mv-glass-bg);
    backdrop-filter: blur(24px);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-strong);
    border-radius: var(--mv-radius-lg);
    padding: var(--mv-spacing-xl);
    min-width: var(--mv-dialog-min-width-sm);
    max-width: var(--mv-dialog-max-width-sm);

    box-shadow: var(--mv-shadow-dialog);
  }

  .dialog h4 {
    margin: 0 0 var(--mv-spacing-md);
    font-size: var(--mv-font-size-md);
    font-weight: var(--mv-font-weight-bold);
    color: var(--mv-color-text-primary);
    text-shadow: var(--mv-text-shadow-frost-lg);
  }

  .dialog-item-title {
    margin: 0 0 var(--mv-spacing-lg);
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
    background: var(--mv-glass-bg-strong);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border);
    border-radius: var(--mv-radius-sm);
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-secondary);
  }

  .dialog label {
    display: block;
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-muted);
    text-transform: uppercase;
    letter-spacing: var(--mv-letter-spacing-badge);
    margin-bottom: var(--mv-spacing-xs);
  }

  .dialog textarea {
    width: 100%;
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
    background: var(--mv-glass-bg-dark);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border);
    border-radius: var(--mv-radius-sm);
    color: var(--mv-color-text-primary);
    font-family: var(--mv-font-sans);
    font-size: var(--mv-font-size-sm);
    resize: vertical;
    transition: all 0.2s ease;
  }

  .dialog textarea::placeholder {
    color: var(--mv-color-text-disabled);
    font-style: italic;
  }

  .dialog textarea:focus {
    outline: none;
    border-color: var(--mv-primitive-frost-2);
    box-shadow: var(--mv-shadow-glow-frost-2-md);
  }

  .dialog-actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--mv-spacing-sm);
    margin-top: var(--mv-spacing-lg);
    padding-top: var(--mv-spacing-md);
    border-top: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
  }

  .btn-cancel,
  .btn-confirm {
    padding: var(--mv-spacing-xs) var(--mv-spacing-lg);
    border-radius: var(--mv-radius-sm);
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-semibold);
    letter-spacing: var(--mv-letter-spacing-count);
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.2, 0.8, 0.2, 1);
  }

  .btn-cancel {
    background: transparent;
    color: var(--mv-color-text-muted);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border);
  }

  .btn-cancel:hover {
    background: var(--mv-glass-hover);
    color: var(--mv-color-text-primary);
    border-color: var(--mv-glass-border-strong);
  }

  .btn-confirm {
    background: var(--mv-bg-glow-green-mid);
    color: var(--mv-primitive-aurora-green);
    border: var(--mv-border-width-thin) solid var(--mv-glow-green-strong);
  }

  .btn-confirm:hover {
    background: var(--mv-bg-glow-green-hover);
    border-color: var(--mv-primitive-aurora-green);
    box-shadow: var(--mv-shadow-glow-green-lg);
    transform: translateY(-1px);
  }
</style>
