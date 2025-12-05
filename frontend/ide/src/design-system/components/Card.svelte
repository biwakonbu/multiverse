<script lang="ts">
  /**
   * 汎用カードコンポーネント
   * パネルやノードのベースとして使用
   */

  /**
   * カードのバリアント
   * - default: 標準の背景
   * - elevated: 浮き上がった見た目（シャドウ付き）
   * - outlined: ボーダーのみ
   */
  export let variant: 'default' | 'elevated' | 'outlined' = 'default';

  /**
   * パディングサイズ
   */
  export let padding: 'none' | 'small' | 'medium' | 'large' = 'medium';

  /**
   * 選択状態
   */
  export let selected = false;

  /**
   * インタラクティブ（ホバー効果）
   */
  export let interactive = false;
</script>

<!-- svelte-ignore a11y-no-noninteractive-tabindex -->
<div
  class="card variant-{variant} padding-{padding}"
  class:selected
  class:interactive
  role={interactive ? 'button' : undefined}
  tabindex={interactive ? 0 : undefined}
  on:click
  on:keydown
>
  <slot />
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
