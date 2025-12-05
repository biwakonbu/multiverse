<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  // @ts-ignore - Wails自動生成ファイル
  import { CreateTask } from '../../wailsjs/go/main/App';

  const dispatch = createEventDispatcher<{
    created: void;
  }>();

  let title = '';
  let poolId = 'default';
  let isSubmitting = false;
  let error = '';

  const pools = [
    { id: 'default', name: 'Default' },
    { id: 'codegen', name: 'Codegen' },
    { id: 'test', name: 'Test' },
  ];

  async function handleSubmit() {
    if (!title.trim()) {
      error = 'タイトルを入力してください';
      return;
    }

    error = '';
    isSubmitting = true;

    try {
      await CreateTask(title.trim(), poolId);
      title = '';
      dispatch('created');
    } catch (e) {
      console.error('タスク作成エラー:', e);
      error = 'タスクの作成に失敗しました';
    } finally {
      isSubmitting = false;
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      handleSubmit();
    }
  }
</script>

<form class="task-create-form" on:submit|preventDefault={handleSubmit}>
  <!-- タイトル入力 -->
  <div class="form-group">
    <label for="task-title" class="form-label">タイトル</label>
    <input
      id="task-title"
      type="text"
      class="form-input"
      class:error={!!error}
      bind:value={title}
      on:keydown={handleKeydown}
      placeholder="タスクのタイトルを入力"
      disabled={isSubmitting}
    />
    {#if error}
      <span class="form-error">{error}</span>
    {/if}
  </div>

  <!-- Pool選択 -->
  <div class="form-group">
    <label for="task-pool" class="form-label">Pool</label>
    <select
      id="task-pool"
      class="form-select"
      bind:value={poolId}
      disabled={isSubmitting}
    >
      {#each pools as pool}
        <option value={pool.id}>{pool.name}</option>
      {/each}
    </select>
  </div>

  <!-- 送信ボタン -->
  <div class="form-actions">
    <button
      type="submit"
      class="btn btn-primary"
      disabled={isSubmitting || !title.trim()}
    >
      {#if isSubmitting}
        <span class="spinner"></span>
        作成中...
      {:else}
        タスクを作成
      {/if}
    </button>
  </div>
</form>

<style>
  .task-create-form {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-md);
    padding: var(--mv-spacing-md);
  }

  /* フォームグループ */
  .form-group {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-xxs);
  }

  .form-label {
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-medium);
    color: var(--mv-color-text-secondary);
  }

  /* 入力フィールド */
  .form-input,
  .form-select {
    padding: var(--mv-spacing-sm);
    font-size: var(--mv-font-size-md);
    font-family: var(--mv-font-sans);
    color: var(--mv-color-text-primary);
    background: var(--mv-color-surface-secondary);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    border-radius: var(--mv-radius-sm);
    transition: border-color var(--mv-transition-hover),
                box-shadow var(--mv-transition-hover);
  }

  .form-input:focus,
  .form-select:focus {
    outline: none;
    border-color: var(--mv-color-border-focus);
    box-shadow: var(--mv-shadow-focus);
  }

  .form-input::placeholder {
    color: var(--mv-color-text-muted);
  }

  .form-input.error {
    border-color: var(--mv-color-status-failed-border);
  }

  .form-input:disabled,
  .form-select:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  /* セレクト */
  .form-select {
    cursor: pointer;
    appearance: none;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%23888888' d='M6 8L1 3h10z'/%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right var(--mv-spacing-sm) center;
    padding-right: var(--mv-spacing-xl);
  }

  .form-select option {
    background: var(--mv-color-surface-primary);
    color: var(--mv-color-text-primary);
  }

  /* エラーメッセージ */
  .form-error {
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-status-failed-text);
  }

  /* アクション */
  .form-actions {
    display: flex;
    justify-content: flex-end;
    padding-top: var(--mv-spacing-xs);
  }

  /* ボタン */
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: var(--mv-spacing-xs);
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-medium);
    border-radius: var(--mv-radius-sm);
    cursor: pointer;
    transition: background var(--mv-transition-hover),
                border-color var(--mv-transition-hover);
  }

  .btn-primary {
    background: var(--mv-color-status-running-bg);
    border: var(--mv-border-width-thin) solid var(--mv-color-status-running-border);
    color: var(--mv-color-status-running-text);
  }

  .btn-primary:hover:not(:disabled) {
    background: var(--mv-color-status-running-border);
    color: var(--mv-color-text-primary);
  }

  .btn-primary:focus {
    outline: none;
    box-shadow: var(--mv-shadow-focus);
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* スピナー */
  .spinner {
    width: var(--mv-icon-size-xs);
    height: var(--mv-icon-size-xs);
    border: var(--mv-border-width-default) solid transparent;
    border-top-color: currentColor;
    border-radius: var(--mv-radius-full);
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
