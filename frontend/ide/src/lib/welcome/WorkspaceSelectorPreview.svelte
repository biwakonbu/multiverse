<script lang="ts">
  import { WelcomeHeader, RecentWorkspaceList, OpenWorkspaceButton } from './index';
  import type { WorkspaceSummary } from '../../schemas';

  interface Props {
    recentWorkspaces?: WorkspaceSummary[];
    isLoadingRecent?: boolean;
    isLoading?: boolean;
    onOpen?: (id: string) => void;
    onRemove?: (id: string) => void;
    onSelectNew?: () => void;
  }

  let {
    recentWorkspaces = [],
    isLoadingRecent = false,
    isLoading = false,
    onOpen = () => {},
    onRemove = () => {},
    onSelectNew = () => {},
  }: Props = $props();

  function handleOpen(e: CustomEvent<string>) {
    onOpen(e.detail);
  }

  function handleRemove(e: CustomEvent<string>) {
    onRemove(e.detail);
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
      on:open={handleOpen}
      on:remove={handleRemove}
    />

    <!-- アクション -->
    <div class="action-section">
      <OpenWorkspaceButton
        loading={isLoading}
        on:click={onSelectNew}
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
