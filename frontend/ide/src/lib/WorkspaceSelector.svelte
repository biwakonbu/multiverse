<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { WelcomeHeader, RecentWorkspaceList, OpenWorkspaceButton } from './welcome';
  import type { WorkspaceSummary } from '../schemas';
  // @ts-ignore - Wails自動生成ファイル
  import { SelectWorkspace, ListRecentWorkspaces, OpenWorkspaceByID, RemoveWorkspace } from '../../wailsjs/go/main/App';

  const dispatch = createEventDispatcher<{
    selected: string;
  }>();

  let recentWorkspaces: WorkspaceSummary[] = $state([]);
  let isLoading = $state(false);
  let isLoadingRecent = $state(true);

  onMount(async () => {
    await loadRecentWorkspaces();
  });

  // 最近使ったワークスペース一覧を読み込み
  async function loadRecentWorkspaces() {
    isLoadingRecent = true;
    try {
      const workspaces = await ListRecentWorkspaces();
      recentWorkspaces = workspaces || [];
    } catch (e) {
      console.error('最近のワークスペース読み込みエラー:', e);
      recentWorkspaces = [];
    } finally {
      isLoadingRecent = false;
    }
  }

  // 最近使ったワークスペースを開く
  async function handleOpenRecent(e: CustomEvent<string>) {
    const id = e.detail;
    try {
      const resultId = await OpenWorkspaceByID(id);
      if (resultId) {
        dispatch('selected', resultId);
      }
    } catch (e) {
      console.error('ワークスペースを開くエラー:', e);
    }
  }

  // ワークスペースを履歴から削除
  async function handleRemoveWorkspace(e: CustomEvent<string>) {
    const id = e.detail;
    try {
      await RemoveWorkspace(id);
      await loadRecentWorkspaces();
    } catch (e) {
      console.error('ワークスペース削除エラー:', e);
    }
  }

  // 新しいワークスペースを選択
  async function selectNew() {
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
    <!-- ヘッダー: ロゴ・タイトル -->
    <WelcomeHeader />

    <!-- 最近使ったワークスペース一覧 -->
    <RecentWorkspaceList
      workspaces={recentWorkspaces}
      loading={isLoadingRecent}
      on:open={handleOpenRecent}
      on:remove={handleRemoveWorkspace}
    />

    <!-- アクション -->
    <div class="action-section">
      <OpenWorkspaceButton
        loading={isLoading}
        on:click={selectNew}
      />
    </div>

    <!-- ヒント -->
    <div class="hints">
      <p class="hint">プロジェクトのルートディレクトリを選択してください</p>
    </div>
  </div>

  <!-- バージョン表示 -->
  <footer class="version-footer">
    <span class="version">v0.1.0</span>
  </footer>
</div>

<style>
  .welcome-screen {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    background: var(--mv-color-surface-app);
    padding: var(--mv-spacing-xl);
  }

  .welcome-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--mv-spacing-xl);
    width: 100%;
    max-width: var(--mv-container-max-width-sm);
  }

  /* アクションセクション */
  .action-section {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--mv-spacing-md);
    width: 100%;
    padding-top: var(--mv-spacing-md);
  }

  /* ヒント */
  .hints {
    text-align: center;
  }

  .hint {
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-disabled);
    margin: 0;
  }

  /* バージョンフッター */
  .version-footer {
    position: fixed;
    bottom: var(--mv-spacing-md);
    right: var(--mv-spacing-md);
  }

  .version {
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-disabled);
    font-family: var(--mv-font-mono);
  }
</style>
