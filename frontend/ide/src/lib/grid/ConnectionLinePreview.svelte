<script lang="ts">
  // シンプルなプレビューコンポーネント（ConnectionLine自体は使用しない）
  // ConnectionLineはGridCanvas内のSVGコンテキストで動作するため、
  // ここでは視覚的なデモのみ行う

  
  interface Props {
    // Props
    satisfied?: boolean;
  }

  let { satisfied = true }: Props = $props();

  // SVGマーカー定義用
  let markerId = $derived(satisfied ? 'arrowhead-satisfied' : 'arrowhead-unsatisfied');
</script>

<div class="preview-container">
  <svg class="connection-svg" viewBox="0 0 400 200">
    <!-- マーカー定義 -->
    <defs>
      <marker
        id="arrowhead-satisfied"
        markerWidth="10"
        markerHeight="7"
        refX="9"
        refY="3.5"
        orient="auto"
      >
        <polygon points="0 0, 10 3.5, 0 7" fill="var(--mv-color-status-succeeded-border, #a3be8c)" />
      </marker>
      <marker
        id="arrowhead-unsatisfied"
        markerWidth="10"
        markerHeight="7"
        refX="9"
        refY="3.5"
        orient="auto"
      >
        <polygon points="0 0, 10 3.5, 0 7" fill="var(--mv-color-status-blocked-border, #bf616a)" />
      </marker>
    </defs>

    <!-- ノード表示（プレビュー用） -->
    <rect x="20" y="75" width="100" height="50" rx="8" fill="var(--mv-color-surface-secondary, #3b4252)" stroke={satisfied ? '#a3be8c' : '#ebcb8b'} stroke-width="2" />
    <text x="70" y="105" text-anchor="middle" fill="var(--mv-color-text-primary, #eceff4)" font-size="12">依存元</text>

    <rect x="280" y="75" width="100" height="50" rx="8" fill="var(--mv-color-surface-secondary, #3b4252)" stroke={satisfied ? '#88c0d0' : '#bf616a'} stroke-width="2" />
    <text x="330" y="105" text-anchor="middle" fill="var(--mv-color-text-primary, #eceff4)" font-size="12">依存先</text>

    <!-- 接続線 -->
    <path
      d="M 120 100 C 180 100, 220 100, 280 100"
      fill="none"
      stroke={satisfied ? 'var(--mv-color-status-succeeded-border, #a3be8c)' : 'var(--mv-color-status-blocked-border, #bf616a)'}
      stroke-width="2"
      stroke-dasharray={satisfied ? 'none' : '8 4'}
      marker-end="url(#{markerId})"
    />
  </svg>

  <div class="legend">
    <span class="status">{satisfied ? '満たされた依存（緑・実線）' : '未満の依存（赤・破線）'}</span>
  </div>
</div>

<style>
  .preview-container {
    width: var(--mv-space-420, 420px);
    background: var(--mv-color-surface-primary);
    border-radius: var(--mv-radius-md, 8px);
    padding: var(--mv-space-4, 16px);
  }

  .connection-svg {
    width: 100%;
    height: var(--mv-space-200, 200px);
  }

  .legend {
    text-align: center;
    margin-top: var(--mv-space-2, 8px);
    font-size: var(--mv-font-size-sm, 12px);
    color: var(--mv-color-text-muted);
  }

  .status {
    padding: var(--mv-space-1, 4px) var(--mv-space-2, 8px);
    background: var(--mv-color-surface-secondary);
    border-radius: var(--mv-radius-sm, 4px);
  }
</style>
