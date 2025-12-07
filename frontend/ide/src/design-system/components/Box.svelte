<script lang="ts">
  import { createBubbler } from 'svelte/legacy';

  const bubble = createBubbler();
  

  // Custom class
  interface Props {
    /**
   * Box Component
   * Basic container primitive with consistent spacing and colors.
   */
    as?: string;
    p?: string;
    m?: string;
    bg?: string | undefined;
    color?: string | undefined;
    border?: string | undefined;
    radius?: string;
    height?: string;
    width?: string;
    grow?: boolean;
    shrink?: boolean;
    class?: string;
    children?: import('svelte').Snippet;
  }

  let {
    as = "div",
    p = "0",
    m = "0",
    bg = undefined,
    color = undefined,
    border = undefined,
    radius = "0",
    height = "auto",
    width = "auto",
    grow = false,
    shrink = false,
    class: className = "",
    children
  }: Props = $props();
  

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
  onclick={bubble('click')}
  onkeydown={bubble('keydown')}
>
  {@render children?.()}
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
