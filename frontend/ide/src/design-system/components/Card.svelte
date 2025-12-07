<script lang="ts">
  import { createBubbler } from 'svelte/legacy';

  const bubble = createBubbler();
  /**
   * 汎用カードコンポーネント
   * パネルやノードのベースとして使用
   */

  

  

  

  
  interface Props {
    /**
   * カードのバリアント
   * - default: 標準の背景
   * - elevated: 浮き上がった見た目（シャドウ付き）
   * - outlined: ボーダーのみ
   */
    variant?: 'default' | 'elevated' | 'outlined';
    /**
   * パディングサイズ
   */
    padding?: 'none' | 'small' | 'medium' | 'large';
    /**
   * 選択状態
   */
    selected?: boolean;
    /**
   * インタラクティブ（ホバー効果）
   */
    interactive?: boolean;
    children?: import('svelte').Snippet;
  }

  let {
    variant = 'default',
    padding = 'medium',
    selected = false,
    interactive = false,
    children
  }: Props = $props();
</script>

<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<div
  class="card variant-{variant} padding-{padding}"
  class:selected
  class:interactive
  role={interactive ? 'button' : undefined}
  tabindex={interactive ? 0 : undefined}
  onclick={bubble('click')}
  onkeydown={bubble('keydown')}
>
  {@render children?.()}
</div>

<style>
  .card {
    border-radius: var(--mv-radius-md);
    transition: var(--mv-transition-hover);
  }

  /* バリアント */
  .variant-default {
    background: var(--mv-color-surface-node);
  }

  .variant-elevated {
    background: var(--mv-color-surface-secondary);
    box-shadow: var(--mv-shadow-card);
  }

  .variant-outlined {
    background: transparent;
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
  }

  /* パディング */
  .padding-none {
    padding: 0;
  }

  .padding-small {
    padding: var(--mv-spacing-xs);
  }

  .padding-medium {
    padding: var(--mv-spacing-md);
  }

  .padding-large {
    padding: var(--mv-spacing-lg);
  }

  /* 選択状態 */
  .selected {
    border: var(--mv-border-width-default) solid var(--mv-color-border-focus);
  }

  /* インタラクティブ */
  .interactive {
    cursor: pointer;
  }

  .interactive:hover {
    background: var(--mv-color-surface-hover);
  }

  .interactive:focus-visible {
    outline: var(--mv-focus-ring-width) solid var(--mv-color-border-focus);
    outline-offset: var(--mv-focus-ring-offset);
  }
</style>
