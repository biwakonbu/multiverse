<script lang="ts">
  /**
   * Text Component
   * Standard typography component.
   */
  export let variant:
    | "primary"
    | "secondary"
    | "muted"
    | "disabled"
    | "success"
    | "warning"
    | "error"
    | "info" = "primary";

  export let size: "xs" | "sm" | "md" | "lg" | "xl" = "md";
  export let weight: "normal" | "medium" | "semibold" | "bold" = "normal";
  export let mono: boolean = false;
  export let glow: boolean = false;

  export let as: string = "p";

  let className = "";
  export { className as class };

  const variantMap = {
    primary: "var(--mv-color-text-primary)",
    secondary: "var(--mv-color-text-secondary)",
    muted: "var(--mv-color-text-muted)",
    disabled: "var(--mv-color-text-disabled)",
    success: "var(--mv-color-status-running-text)",
    warning: "var(--mv-color-status-blocked-text)",
    error: "var(--mv-color-status-failed-text)",
    info: "var(--mv-color-status-ready-text)",
  };

  const sizeMap = {
    xs: "var(--mv-font-size-xs)",
    sm: "var(--mv-font-size-sm)",
    md: "var(--mv-font-size-md)",
    lg: "var(--mv-font-size-lg)",
    xl: "var(--mv-font-size-xl)",
  };

  const weightMap = {
    normal: "var(--mv-font-weight-normal)",
    medium: "var(--mv-font-weight-medium)",
    semibold: "var(--mv-font-weight-semibold)",
    bold: "var(--mv-font-weight-bold)",
  };
</script>

<svelte:element
  this={as}
  class="text {className}"
  class:mono
  class:glow
  style:--text-color={variantMap[variant]}
  style:--text-size={sizeMap[size]}
  style:--text-weight={weightMap[weight]}
  style:--text-glow-color={variantMap[variant]}
>
  <slot />
</svelte:element>

<style>
  .text {
    margin: 0;
    color: var(--text-color);
    font-size: var(--text-size);
    font-weight: var(--text-weight);
    font-family: var(--mv-font-sans);
    line-height: var(--mv-line-height-base);
  }

  .mono {
    font-family: var(--mv-font-mono);
  }

  .glow {
    text-shadow: var(--mv-shadow-text-glow-md) var(--text-glow-color);
  }
</style>
