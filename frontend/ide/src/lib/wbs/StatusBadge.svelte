<script lang="ts">
  import { statusToCssClass } from "../../schemas";
  import type { TaskStatus } from "../../schemas";

  interface Props {
    status: TaskStatus;
  }

  let { status }: Props = $props();

  let statusClass = $derived(statusToCssClass(status));
</script>

<span class="status-badge status-{statusClass}">
  {status}
</span>

<style>
  .status-badge {
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-bold);
    text-transform: uppercase;
    letter-spacing: var(--mv-letter-spacing-wide);
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    border-radius: var(--mv-radius-sm);
    box-shadow: var(--mv-shadow-badge);
    display: inline-block;
    min-width: var(--mv-badge-min-width);
    text-align: center;
  }

  .status-pending {
    background: var(--mv-color-status-pending-bg);
    color: var(--mv-color-status-pending-text);
    border: var(--mv-border-width-thin) solid
      var(--mv-color-status-pending-border);
  }

  .status-ready {
    background: var(--mv-color-status-ready-bg);
    color: var(--mv-color-status-ready-text);
    border: var(--mv-border-width-thin) solid
      var(--mv-color-status-ready-border);
    box-shadow: var(--mv-shadow-badge-glow-sm) var(--mv-color-status-ready-text);
  }

  .status-running {
    background: var(--mv-color-status-running-bg);
    color: var(--mv-color-status-running-text);
    border: var(--mv-border-width-thin) solid
      var(--mv-color-status-running-border);
    box-shadow: var(--mv-shadow-badge-glow-md)
      var(--mv-color-status-running-glow);
    animation: mv-pulse var(--mv-duration-pulse) infinite;
  }

  .status-succeeded {
    background: var(--mv-color-status-succeeded-bg);
    color: var(--mv-color-status-succeeded-text);
    border: var(--mv-border-width-thin) solid
      var(--mv-color-status-succeeded-border);
    box-shadow: var(--mv-shadow-badge-glow-sm)
      var(--mv-color-status-succeeded-border);
  }

  .status-completed {
    background: var(--mv-color-status-completed-bg);
    color: var(--mv-color-status-completed-text);
    border: var(--mv-border-width-thin) solid
      var(--mv-color-status-completed-border);
    box-shadow: var(--mv-shadow-badge-glow-lg)
      var(--mv-color-status-completed-glow);
  }

  .status-failed {
    background: var(--mv-color-status-failed-bg);
    color: var(--mv-color-status-failed-text);
    border: var(--mv-border-width-thin) solid
      var(--mv-color-status-failed-border);
    box-shadow: var(--mv-shadow-badge-glow-sm)
      var(--mv-color-status-failed-border);
  }

  .status-canceled {
    background: var(--mv-color-status-canceled-bg);
    color: var(--mv-color-status-canceled-text);
    border: var(--mv-border-width-thin) solid
      var(--mv-color-status-canceled-border);
  }

  .status-blocked {
    background: var(--mv-color-status-blocked-bg);
    color: var(--mv-color-status-blocked-text);
    border: var(--mv-border-width-thin) solid
      var(--mv-color-status-blocked-border);
  }

  .status-retry-wait {
    background: var(--mv-color-status-retry-wait-bg);
    color: var(--mv-color-status-retry-wait-text);
    border: var(--mv-border-width-thin) solid
      var(--mv-color-status-retry-wait-border);
    animation: mv-pulse-slow 2s infinite;
  }

  @keyframes mv-pulse-slow {
    0%,
    100% {
      opacity: 1;
    }
    50% {
      opacity: 0.7;
    }
  }
</style>
