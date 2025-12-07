<script lang="ts">
  /**
   * Flex Component
   * Flexbox container helper.
   */
  import Box from "./Box.svelte";

  export let direction: "row" | "column" | "row-reverse" | "column-reverse" =
    "row";
  export let align: "start" | "center" | "end" | "stretch" | "baseline" =
    "stretch";
  export let justify:
    | "start"
    | "center"
    | "end"
    | "between"
    | "around"
    | "evenly" = "start";
  export let wrap: "nowrap" | "wrap" | "wrap-reverse" = "nowrap";
  export let gap: string = "0";

  // Box props
  export let p: string = "0";
  export let m: string = "0";
  export let bg: string | undefined = undefined;
  export let height: string = "auto";
  export let width: string = "auto";
  export let grow: boolean = false;

  let className = "";
  export { className as class };

  const justifyMap = {
    start: "flex-start",
    center: "center",
    end: "flex-end",
    between: "space-between",
    around: "space-around",
    evenly: "space-evenly",
  };

  const alignMap = {
    start: "flex-start",
    center: "center",
    end: "flex-end",
    stretch: "stretch",
    baseline: "baseline",
  };
</script>

<div
  class="flex {className}"
  class:grow
  style:--flex-dir={direction}
  style:--flex-align={alignMap[align]}
  style:--flex-justify={justifyMap[justify]}
  style:--flex-wrap={wrap}
  style:--flex-gap={gap}
  style:--box-p={p}
  style:--box-m={m}
  style:--box-bg={bg}
  style:--box-h={height}
  style:--box-w={width}
  on:click
  on:keydown
  role="group"
>
  <slot />
</div>

<style>
  .flex {
    display: flex;
    box-sizing: border-box;
    flex-flow: var(--flex-dir) var(--flex-wrap);
    align-items: var(--flex-align);
    justify-content: var(--flex-justify);
    gap: var(--flex-gap);

    padding: var(--box-p);
    margin: var(--box-m);
    background: var(--box-bg, transparent);
    height: var(--box-h);
    width: var(--box-w);
  }

  .grow {
    flex-grow: 1;
  }
</style>
