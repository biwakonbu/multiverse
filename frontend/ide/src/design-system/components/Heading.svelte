<script lang="ts">
  /**
   * Heading Component
   * Brand typography for titles and headers.
   */
  export let level: 1 | 2 | 3 | 4 | 5 | 6 = 1;
  export let variant: "default" | "gradient" = "default";

  let className = "";
  export { className as class };

  // Font family determined by brand
  // Level affects size

  const sizeMap = {
    1: "var(--mv-font-size-3xl)",
    2: "var(--mv-font-size-2xl)",
    3: "var(--mv-font-size-xl)",
    4: "var(--mv-font-size-lg)",
    5: "var(--mv-font-size-md)",
    6: "var(--mv-font-size-sm)",
  };
</script>

<svelte:element
  this={`h${level}`}
  class="heading level-{level} variant-{variant} {className}"
  style:--heading-size={sizeMap[level]}
>
  <slot />
</svelte:element>

<style>
  .heading {
    margin: 0;
    font-family: var(--mv-font-brand);
    font-weight: var(--mv-font-weight-bold);
    font-size: var(--heading-size);
    line-height: var(--mv-line-height-tight);
    color: var(--mv-color-text-primary);
    letter-spacing: var(--mv-letter-spacing-count);
    text-transform: uppercase;
  }

  /* Variants */
  .variant-gradient {
    background: linear-gradient(
      135deg,
      var(--mv-primitive-snow-storm-nord6) 0%,
      var(--mv-primitive-frost-nord9) 100%
    );
    background-clip: text;
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    filter: var(--mv-filter-heading-glow);
  }

  /* Levels specific tweaks if needed */
  .level-1 {
    letter-spacing: var(--mv-letter-spacing-heading);
  }
</style>
