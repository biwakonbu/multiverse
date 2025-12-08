<script lang="ts">
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

  interface Props {
    item: BacklogItemProps;
    onresolve?: () => void;
    ondelete?: () => void;
  }

  let { item, onresolve, ondelete }: Props = $props();

  function getTypeLabel(type: BacklogItemType | string): string {
    switch (type) {
      case "FAILURE":
        return "失敗";
      case "QUESTION":
        return "質問";
      case "BLOCKER":
        return "ブロッカー";
      default:
        return type;
    }
  }

  function getPriorityLabel(priority: number): string {
    if (priority >= 5) return "最高";
    if (priority >= 4) return "高";
    if (priority >= 3) return "中";
    if (priority >= 2) return "低";
    return "最低";
  }

  function formatDate(dateValue: string | Date): string {
    const date =
      typeof dateValue === "string" ? new Date(dateValue) : dateValue;
    return date.toLocaleString("ja-JP", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }
</script>

<li class="backlog-item" class:failure={item.type === "FAILURE"}>
  <div class="item-header">
    <span class="type-badge {item.type.toLowerCase()}"
      >{getTypeLabel(item.type)}</span
    >
    <span class="priority">{getPriorityLabel(item.priority)}</span>
    <span class="date">{formatDate(item.createdAt)}</span>
  </div>
  <h4 class="item-title">{item.title}</h4>
  <p class="item-description">{item.description}</p>
  <div class="item-actions">
    <button class="btn-resolve" onclick={() => onresolve?.()}> 解決 </button>
    <button class="btn-delete" onclick={() => ondelete?.()}> 削除 </button>
  </div>
</li>

<style>
  /* === Backlog Item Card === */
  .backlog-item {
    position: relative;

    /* Glass Card */
    background: var(--mv-glass-bg-strong);
    backdrop-filter: blur(8px);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border);
    border-radius: var(--mv-radius-md);
    padding: var(--mv-spacing-md);

    /* Card shadow */
    box-shadow: var(--mv-shadow-card);

    /* Animation */
    transition: all 0.25s cubic-bezier(0.2, 0.8, 0.2, 1);
    overflow: hidden;
  }

  .backlog-item::before {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    width: var(--mv-border-width-thick);
    height: 100%;
    background: var(--mv-glass-border);
    opacity: 0.5;
    transition: all 0.25s ease;
  }

  .backlog-item:hover {
    background: var(--mv-glass-hover);
    border-color: var(--mv-glass-border-strong);
    transform: translateX(4px);

    box-shadow: var(--mv-shadow-card-hover);
  }

  .backlog-item:hover::before {
    opacity: 1;
    background: var(--mv-primitive-frost-2);
    box-shadow: var(--mv-shadow-glow-frost-2);
  }

  /* === Failure Type - Glowing Red Edge === */
  .backlog-item.failure::before {
    background: var(--mv-primitive-aurora-red);
    opacity: 1;
    box-shadow: var(--mv-shadow-glow-red);
  }

  .backlog-item.failure:hover::before {
    box-shadow: var(--mv-shadow-glow-red-lg);
  }

  /* === Item Header === */
  .item-header {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-sm);
    margin-bottom: var(--mv-spacing-sm);
    font-size: var(--mv-font-size-xs);
  }

  /* === Type Badge with Glow === */
  .type-badge {
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    border-radius: var(--mv-radius-sm);
    font-weight: var(--mv-font-weight-bold);
    text-transform: uppercase;
    font-size: var(--mv-font-size-xxs);
    letter-spacing: var(--mv-letter-spacing-badge);

    /* Glass effect */
    backdrop-filter: blur(4px);
    border: var(--mv-border-width-thin) solid transparent;
  }

  .type-badge.failure {
    background: var(--mv-bg-glow-red-light);
    color: var(--mv-primitive-aurora-red);
    border-color: var(--mv-glow-failed);
    box-shadow: var(--mv-shadow-backlog-item);
  }

  .type-badge.question {
    background: var(--mv-bg-glow-yellow-mid);
    color: var(--mv-primitive-aurora-yellow);
    border-color: var(--mv-border-glow-yellow);
    box-shadow: var(--mv-shadow-glow-yellow);
  }

  .type-badge.blocker {
    background: var(--mv-glow-frost-2-mid);
    color: var(--mv-primitive-frost-2);
    border-color: var(--mv-glow-frost-2-border);
    box-shadow: var(--mv-shadow-glow-frost-2);
  }

  /* === Priority Badge === */
  .priority {
    font-size: var(--mv-font-size-xxs);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-muted);
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    background: var(--mv-glass-active);
    border-radius: var(--mv-radius-sm);
    letter-spacing: var(--mv-letter-spacing-count);
  }

  /* === Date === */
  .date {
    margin-left: auto;
    font-size: var(--mv-font-size-xxs);
    font-family: var(--mv-font-mono);
    color: var(--mv-color-text-disabled);
    letter-spacing: var(--mv-letter-spacing-widest);
  }

  /* === Item Title === */
  .item-title {
    margin: 0 0 var(--mv-spacing-xs);
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
    line-height: var(--mv-line-height-snug);

    /* Subtle glow on text */
    text-shadow: var(--mv-text-shadow-snow-subtle);
  }

  /* === Item Description === */
  .item-description {
    margin: 0 0 var(--mv-spacing-sm);
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-secondary);
    line-height: var(--mv-line-height-relaxed);
    opacity: 0.9;
  }

  /* === Action Buttons === */
  .item-actions {
    display: flex;
    gap: var(--mv-spacing-sm);
    margin-top: var(--mv-spacing-md);
    padding-top: var(--mv-spacing-sm);
    border-top: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
  }

  .btn-resolve,
  .btn-delete {
    padding: var(--mv-spacing-xs) var(--mv-spacing-md);
    border-radius: var(--mv-radius-sm);
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-semibold);
    letter-spacing: var(--mv-letter-spacing-count);
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.2, 0.8, 0.2, 1);
  }

  /* === Resolve Button === */
  .btn-resolve {
    background: var(--mv-bg-glow-green-mid);
    color: var(--mv-primitive-aurora-green);
    border: var(--mv-border-width-thin) solid var(--mv-border-glow-green);
  }

  .btn-resolve:hover {
    background: var(--mv-bg-glow-green-hover);
    border-color: var(--mv-primitive-aurora-green);
    box-shadow: var(--mv-shadow-glow-green-md);
    transform: translateY(-1px);
  }

  /* === Delete Button === */
  .btn-delete {
    background: transparent;
    color: var(--mv-color-text-muted);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border);
  }

  .btn-delete:hover {
    background: var(--mv-bg-glow-red-light);
    color: var(--mv-primitive-aurora-red);
    border-color: var(--mv-glow-red-border);
    box-shadow: var(--mv-shadow-glow-red-md);
  }
</style>
