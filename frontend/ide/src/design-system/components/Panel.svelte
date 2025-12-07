<script lang="ts">
  /**
   * Panel Component
   * "Phantom Glass" container with defined variants.
   */
  import Box from "./Box.svelte";


  interface Props {
    variant?: "default" | "glass" | "glass-strong" | "outlined";
    padding?: "none" | "sm" | "md" | "lg";
    radius?: "none" | "sm" | "md" | "lg" | "full";
    hover?: boolean;
    glow?: boolean;
    class?: string;
    children?: import('svelte').Snippet;
  }

  let {
    variant = "glass",
    padding = "md",
    radius = "md",
    hover = false,
    glow = false,
    class: className = "",
    children
  }: Props = $props();
  

  const paddingMap = {
    none: "0",
    sm: "var(--mv-spacing-sm)",
    md: "var(--mv-spacing-md)",
    lg: "var(--mv-spacing-lg)",
  };

  const radiusMap = {
    none: "0",
    sm: "var(--mv-radius-sm)",
    md: "var(--mv-radius-md)",
    lg: "var(--mv-radius-lg)",
    full: "var(--mv-radius-full)",
  };
</script>

<div
  class="panel variant-{variant} {className}"
  class:hover-effect={hover}
  class:glow-effect={glow}
  style:--panel-p={paddingMap[padding]}
  style:--panel-r={radiusMap[radius]}
>
  {@render children?.()}
</div>

<style>
  .panel {
    box-sizing: border-box;
    padding: var(--panel-p);
    border-radius: var(--panel-r);
    transition: var(--mv-transition-base);
  }

  /* Variants */
  .variant-default {
    background: var(--mv-color-surface-secondary);
    border: var(--mv-border-width-thin) solid transparent;
  }

  .variant-glass {
    background: var(--mv-glass-bg);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
    backdrop-filter: blur(16px);
  }

  .variant-glass-strong {
    background: var(--mv-glass-bg-strong);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border);
    backdrop-filter: blur(24px);
    box-shadow: var(--mv-shadow-lg);
  }

  .variant-outlined {
    background: transparent;
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
  }

  /* Effects */
  .hover-effect:hover {
    background: var(--mv-glass-hover);
    border-color: var(--mv-glass-border-strong);
    transform: translateY(-1px);
    box-shadow: var(--mv-shadow-md);
  }

  .glow-effect {
    box-shadow: var(--mv-shadow-glow-accent);
    border-color: var(--mv-color-interactive-primary);
  }
</style>
