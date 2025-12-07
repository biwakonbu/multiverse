<script lang="ts">
  /**
   * Box Component
   * Basic container primitive with consistent spacing and colors.
   */
  export let as: string = "div";
  export let p: string = "0";
  export let m: string = "0";
  export let bg: string | undefined = undefined;
  export let color: string | undefined = undefined;
  export let border: string | undefined = undefined;
  export let radius: string = "0";
  export let height: string = "auto";
  export let width: string = "auto";
  export let grow: boolean = false;
  export let shrink: boolean = false;

  // Custom class
  let className = "";
  export { className as class };

  // Map simplified props to CSS vars or values
  // This is a simplified implementation. In a full system, we might map 'sm', 'md' to vars.
  // For now, we accept CSS values or var(--mv-...) strings.
</script>

<svelte:element
  this={as}
  class="box {className}"
  class:grow
  class:shrink
  style:--box-p={p}
  style:--box-m={m}
  style:--box-bg={bg}
  style:--box-color={color}
  style:--box-border={border}
  style:--box-radius={radius}
  style:--box-h={height}
  style:--box-w={width}
  role={as === "button" ? "button" : undefined}
  on:click
  on:keydown
>
  <slot />
</svelte:element>

<style>
  .box {
    display: block;
    box-sizing: border-box;
    padding: var(--box-p);
    margin: var(--box-m);
    background: var(--box-bg, transparent);
    color: var(--box-color, inherit);
    border: var(--box-border, none);
    border-radius: var(--box-radius);
    height: var(--box-h);
    width: var(--box-w);
  }

  .grow {
    flex-grow: 1;
  }

  .shrink {
    flex-shrink: 1;
  }
</style>
