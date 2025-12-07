<script lang="ts">
  /**
   * Grid Component
   * CSS Grid container helper.
   */
  import Box from "./Box.svelte";

  export let columns: string = "1fr"; // e.g. "repeat(3, 1fr)"
  export let rows: string = "auto";
  export let gap: string = "var(--mv-spacing-md)";
  export let align: "start" | "center" | "end" | "stretch" = "stretch";
  export let justify: "start" | "center" | "end" | "stretch" = "stretch";

  // Box props
  export let p: string = "0";
  export let m: string = "0";
  export let bg: string | undefined = undefined;
  export let height: string = "auto";
  export let width: string = "auto";

  let className = "";
  export { className as class };
</script>

<div
  class="grid {className}"
  style:--grid-cols={columns}
  style:--grid-rows={rows}
  style:--grid-gap={gap}
  style:--grid-align={align}
  style:--grid-justify={justify}
  style:--box-p={p}
  style:--box-m={m}
  style:--box-bg={bg}
  style:--box-h={height}
  style:--box-w={width}
>
  <slot />
</div>

<style>
  .grid {
    display: grid;
    box-sizing: border-box;
    grid-template-columns: var(--grid-cols);
    grid-template-rows: var(--grid-rows);
    gap: var(--grid-gap);
    place-items: var(--grid-align) var(--grid-justify);

    padding: var(--box-p);
    margin: var(--box-m);
    background: var(--box-bg, transparent);
    height: var(--box-h);
    width: var(--box-w);
  }
</style>
