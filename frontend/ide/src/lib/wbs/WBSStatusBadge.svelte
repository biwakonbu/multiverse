<script lang="ts">
  import Badge from "../../design-system/components/Badge.svelte";
  import type { ComponentProps } from "svelte";

  // Badgeと同じpropsを受け取る
  type BadgeProps = ComponentProps<Badge>;
  interface Props {
    status: BadgeProps["status"];
  }

  let { status }: Props = $props();

  // WBS用の固定幅バッジ
  // 表示がガタガタになるのを防ぐため、十分な固定幅を確保する
  const FIXED_WIDTH = "100px";

  // RETRY_WAIT などが一番長いのでそれに合わせる
</script>

<div class="wbs-badge-wrapper" style:width={FIXED_WIDTH}>
  <Badge {status} size="medium">
    <!-- デフォルトのスロットコンテンツ（ステータス名）を使用 -->
  </Badge>
</div>

<style>
  .wbs-badge-wrapper {
    display: inline-flex;
    justify-content: center;
  }

  /* Badgeコンポーネントのスタイルを上書きして幅を広げる必要がある場合 */
  :global(.wbs-badge-wrapper .badge) {
    width: 100%;
  }
</style>
