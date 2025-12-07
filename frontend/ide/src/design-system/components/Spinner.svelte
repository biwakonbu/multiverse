<script lang="ts">
  

  interface Props {
    /**
   * Spinner Component
   * Loading indicator.
   */
    size?: "xs" | "sm" | "md" | "lg";
    color?: "current" | "primary" | "secondary";
    class?: string;
  }

  let { size = "sm", color = "current", class: className = "" }: Props = $props();
  

  const sizeMap = {
    xs: "var(--mv-icon-size-xs)",
    sm: "var(--mv-icon-size-sm)",
    md: "var(--mv-icon-size-md)",
    lg: "var(--mv-icon-size-lg)",
  };
</script>

<span
  class="spinner color-{color} {className}"
  style:--spinner-size={sizeMap[size]}
  aria-label="Loading"
></span>

<style>
  .spinner {
    display: inline-block;
    width: var(--spinner-size);
    height: var(--spinner-size);
    border: var(--mv-border-width-default) solid transparent;
    border-top-color: currentColor;
    border-radius: var(--mv-radius-full);
    animation: spin 0.8s linear infinite;
  }

  .color-current {
    color: currentColor;
  }

  .color-primary {
    color: var(--mv-primitive-frost-nord8);
  }

  .color-secondary {
    color: var(--mv-color-text-secondary);
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
