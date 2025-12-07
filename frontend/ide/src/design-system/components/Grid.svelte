<script lang="ts">
  /**
   * Grid Component
   * CSS Grid container helper.
   */
  import Box from "./Box.svelte";


  

  interface Props {
    columns?: string; // e.g. "repeat(3, 1fr)"
    rows?: string;
    gap?: string;
    align?: "start" | "center" | "end" | "stretch";
    justify?: "start" | "center" | "end" | "stretch";
    // Box props
    p?: string;
    m?: string;
    bg?: string | undefined;
    height?: string;
    width?: string;
    class?: string;
    children?: import('svelte').Snippet;
  }

  let {
    columns = "1fr",
    rows = "auto",
    gap = "var(--mv-spacing-md)",
    align = "stretch",
    justify = "stretch",
    p = "0",
    m = "0",
    bg = undefined,
    height = "auto",
    width = "auto",
    class: className = "",
    children
  }: Props = $props();
  
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
  {@render children?.()}
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
