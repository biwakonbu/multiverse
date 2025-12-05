<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  // @ts-ignore - Wails自動生成ファイル
  import { SelectWorkspace } from '../../wailsjs/go/main/App';

  const dispatch = createEventDispatcher<{
    selected: string;
  }>();

  let isLoading = false;

  async function select() {
    if (isLoading) return;

    isLoading = true;
    try {
      const id = await SelectWorkspace();
      if (id) {
        dispatch('selected', id);
      }
    } catch (e) {
      console.error('Workspace選択エラー:', e);
    } finally {
      isLoading = false;
    }
  }
</script>

<div class="welcome-screen">
  <div class="welcome-content">
    <!-- ロゴ/タイトル -->
    <div class="logo-section">
      <h1 class="app-title">multiverse IDE</h1>
      <p class="app-subtitle">AI開発タスク管理プラットフォーム</p>
    </div>

    <!-- アクション -->
    <div class="action-section">
      <p class="instruction">プロジェクトを選択して開始</p>

      <button class="btn btn-primary" on:click={select} disabled={isLoading}>
        {#if isLoading}
          <span class="spinner"></span>
          読み込み中...
        {:else}
          <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
          </svg>
          Workspaceを開く
        {/if}
      </button>
    </div>

    <!-- ヒント -->
    <div class="hints">
      <p class="hint">Workspaceはプロジェクトのルートディレクトリを選択してください</p>
    </div>
  </div>
</div>

<style>
  .welcome-screen {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100vh;
    background: var(--mv-color-surface-app);
    padding: var(--mv-spacing-xl);
  }

  .welcome-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--mv-spacing-xxl);
    text-align: center;
    max-width: var(--mv-container-max-width-sm);
  }

  /* ロゴセクション */
  .logo-section {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--mv-spacing-sm);
  }

  .app-title {
    font-size: var(--mv-font-size-xxl);
    font-weight: var(--mv-font-weight-bold);
    color: var(--mv-color-text-primary);
    margin: 0;
    letter-spacing: var(--mv-letter-spacing-tight);
  }

  .app-subtitle {
    font-size: var(--mv-font-size-md);
    color: var(--mv-color-text-secondary);
    margin: 0;
  }

  /* アクションセクション */
  .action-section {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--mv-spacing-md);
  }

  .instruction {
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-muted);
    margin: 0;
  }

  /* ボタン */
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: var(--mv-spacing-xs);
    padding: var(--mv-spacing-sm) var(--mv-spacing-lg);
    font-size: var(--mv-font-size-md);
    font-weight: var(--mv-font-weight-medium);
    border-radius: var(--mv-radius-md);
    cursor: pointer;
    transition: background var(--mv-transition-hover),
                border-color var(--mv-transition-hover),
                transform var(--mv-transition-hover);
  }

  .btn:active:not(:disabled) {
    transform: var(--mv-transform-press);
  }

  .btn-primary {
    background: var(--mv-color-status-ready-bg);
    border: var(--mv-border-width-default) solid var(--mv-color-status-ready-border);
    color: var(--mv-color-status-ready-text);
  }

  .btn-primary:hover:not(:disabled) {
    background: var(--mv-color-status-ready-border);
    color: var(--mv-color-text-primary);
  }

  .btn-primary:focus {
    outline: none;
    box-shadow: var(--mv-shadow-selected);
  }

  .btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .icon {
    width: var(--mv-icon-size-md);
    height: var(--mv-icon-size-md);
  }

  /* スピナー */
  .spinner {
    width: var(--mv-icon-size-sm);
    height: var(--mv-icon-size-sm);
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

  /* ヒント */
  .hints {
    padding-top: var(--mv-spacing-lg);
    border-top: var(--mv-border-width-thin) solid var(--mv-color-border-subtle);
  }

  .hint {
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-disabled);
    margin: 0;
  }
</style>
