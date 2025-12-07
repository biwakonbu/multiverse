<script lang="ts">
  import { slide } from "svelte/transition";
  import type { ResourceNode, ResourceType, ResourceStatus } from "./types";

  interface Props {
    resources?: ResourceNode[];
  }

  let { resources = $bindable([]) }: Props = $props();

  // Flatten the tree for rendering
  function flatten(
    nodes: ResourceNode[],
    depth = 0
  ): Array<ResourceNode & { depth: number }> {
    let result: Array<ResourceNode & { depth: number }> = [];
    for (const node of nodes) {
      result.push({ ...node, depth });
      if (node.children && node.expanded !== false) {
        // Default to expanded if undefined
        result = result.concat(flatten(node.children, depth + 1));
      }
    }
    return result;
  }

  let flatNodes = $derived(flatten(resources));

  function toggleExpand(nodeId: string) {
    resources = toggleNode(resources, nodeId);
  }

  function toggleNode(nodes: ResourceNode[], id: string): ResourceNode[] {
    return nodes.map((node) => {
      if (node.id === id) {
        return { ...node, expanded: node.expanded === false ? true : false };
      }
      if (node.children) {
        return { ...node, children: toggleNode(node.children, id) };
      }
      return node;
    });
  }

  function getStatusColor(status: ResourceStatus): string {
    switch (status) {
      case "RUNNING":
        return "var(--mv-color-status-success-text)"; // Aurora Green
      case "THINKING":
        return "var(--mv-color-status-info-text)"; // Frost Blue
      case "ERROR":
        return "var(--mv-color-status-failed-text)"; // Aurora Red
      case "PAUSED":
        return "var(--mv-color-status-warning-text)"; // Aurora Yellow
      case "DONE":
        return "var(--mv-color-text-muted)";
      default:
        return "var(--mv-color-text-muted)";
    }
  }

  function getTypeBadgeStyle(type: ResourceType): string {
    switch (type) {
      case "META":
        return "background: rgba(136, 192, 208, 0.2); color: #88c0d0; border: 1px solid rgba(136, 192, 208, 0.3);";
      case "WORKER":
        return "background: rgba(163, 190, 140, 0.2); color: #a3be8c; border: 1px solid rgba(163, 190, 140, 0.3);";
      case "CONTAINER":
        return "background: rgba(180, 142, 173, 0.2); color: #b48ead; border: 1px solid rgba(180, 142, 173, 0.3);";
      default:
        return "background: rgba(255,255,255,0.1); color: #ccc;";
    }
  }
</script>

<div class="resource-list">
  <div class="header-row">
    <div class="col-name">Resource</div>
    <div class="col-type">Type</div>
    <div class="col-status">Status</div>
    <div class="col-activity">Activity / Detail</div>
  </div>

  <div class="list-body">
    {#each flatNodes as node (node.id)}
      <div
        class="resource-row"
        onclick={() => toggleExpand(node.id)}
        role="button"
        tabindex="0"
        onkeydown={(e) => e.key === "Enter" && toggleExpand(node.id)}
      >
        <!-- Name Column with Indent -->
        <div class="col-name" style:--depth={node.depth}>
          {#if node.children && node.children.length > 0}
            <span class="disclosure-icon"
              >{node.expanded !== false ? "▼" : "▶"}</span
            >
          {:else}
            <span class="disclosure-placeholder"></span>
          {/if}
          <span class="node-name">{node.name}</span>
        </div>

        <!-- Type Column -->
        <div class="col-type">
          <span class="type-badge" style={getTypeBadgeStyle(node.type)}
            >{node.type}</span
          >
        </div>

        <!-- Status Column -->
        <div class="col-status">
          <div class="status-indicator-wrapper">
            <div
              class="status-dot"
              style:background={getStatusColor(node.status)}
            >
              {#if node.status === "RUNNING" || node.status === "THINKING"}
                <div
                  class="status-pulse"
                  style:border-color={getStatusColor(node.status)}
                ></div>
              {/if}
            </div>
            <span class="status-label" style:color={getStatusColor(node.status)}
              >{node.status}</span
            >
          </div>
        </div>

        <!-- Activity Column -->
        <div class="col-activity">
          <span class="detail-text">{node.detail || "-"}</span>
        </div>
      </div>
    {/each}
  </div>
</div>

<style>
  .resource-list {
    display: flex;
    flex-direction: column;
    width: 100%;
    background: var(--mv-glass-bg);
    backdrop-filter: var(--mv-glass-blur);
    border: var(--mv-border-width-sm) solid var(--mv-glass-border-subtle);
    border-radius: var(--mv-radius-md);
    font-family: var(--mv-font-sans);
    overflow: hidden;
  }

  .header-row {
    display: grid;
    grid-template-columns: 2fr var(--mv-space-24) var(--mv-space-28) 3fr;
    padding: var(--mv-space-2) var(--mv-space-3);
    background: var(--mv-glass-border);
    border-bottom: var(--mv-border-width-sm) solid var(--mv-glass-border-subtle);
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-bold);
    color: var(--mv-color-text-muted);
    text-transform: uppercase;
    letter-spacing: var(--mv-letter-spacing-wide);
  }

  .list-body {
    max-height: var(--mv-space-96);
    overflow-y: auto;
  }

  .resource-row {
    display: grid;
    grid-template-columns: 2fr var(--mv-space-24) var(--mv-space-28) 3fr;
    padding: var(--mv-space-2) 0;
    border-bottom: var(--mv-border-width-sm) solid var(--mv-glass-border-active);
    align-items: center;
    cursor: pointer;
    transition: background var(--mv-transition-hover);
    font-size: var(--mv-font-size-sm);
  }

  .resource-row:hover {
    background: var(--mv-glass-hover);
  }

  .resource-row:last-child {
    border-bottom: none;
  }

  .col-name {
    display: flex;
    align-items: center;
    gap: var(--mv-space-1-5);
    color: var(--mv-color-text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    padding-left: calc(var(--depth, 0) * var(--mv-space-5) + var(--mv-space-3));
  }

  .disclosure-icon,
  .disclosure-placeholder {
    width: var(--mv-space-3);
    font-size: var(--mv-font-size-2xs);
    color: var(--mv-color-text-muted);
    text-align: center;
  }

  .node-name {
    font-weight: var(--mv-font-weight-medium);
  }

  .col-type {
    display: flex;
    align-items: center;
  }

  .type-badge {
    font-size: var(--mv-font-size-2xs);
    padding: var(--mv-space-0-5) var(--mv-space-1-5);
    border-radius: var(--mv-radius-sm);
    font-weight: var(--mv-font-weight-bold);
    letter-spacing: var(--mv-letter-spacing-wide);
  }

  .col-status {
    display: flex;
    align-items: center;
  }

  .status-indicator-wrapper {
    display: flex;
    align-items: center;
    gap: var(--mv-space-2);
  }

  .status-dot {
    width: var(--mv-space-1-5);
    height: var(--mv-space-1-5);
    border-radius: var(--mv-radius-full);
    position: relative;
  }

  .status-pulse {
    position: absolute;
    top: calc(-1 * var(--mv-space-0-75));
    left: calc(-1 * var(--mv-space-0-75));
    width: var(--mv-space-3);
    height: var(--mv-space-3);
    border-radius: var(--mv-radius-full);
    border: var(--mv-border-width-sm) solid;
    opacity: 0;
    animation: pulse var(--mv-duration-slow) infinite;
  }

  .status-label {
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-semibold);
  }

  .col-activity {
    color: var(--mv-color-text-secondary);
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-xs);
    opacity: var(--mv-opacity-80);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    padding-right: var(--mv-space-3);
  }

  @keyframes pulse {
    0% {
      transform: scale(var(--mv-scale-half));
      opacity: 0;
    }
    50% {
      opacity: var(--mv-opacity-60);
    }
    100% {
      transform: scale(var(--mv-scale-150));
      opacity: 0;
    }
  }
</style>
