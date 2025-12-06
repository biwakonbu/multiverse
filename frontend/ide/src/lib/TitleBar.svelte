<script lang="ts">
  // Wailsのドラッグ領域として機能させる
</script>

<header class="title-bar" style="--wails-draggable:drag">
  <!-- ドラッグ専用領域 (透明) -->
</header>

<style>
  .title-bar {
    height: 32px;
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent; /* 背景は透明にして下のレイヤー（Appの背景）を見せる */
    user-select: none;
    cursor: default;
    z-index: 9999;
    /* Macの信号機ボタン（左上）を避けるためのパディング */
    padding-left: 80px;
  }
</style>
