<script lang="ts">
  import BrandLogo from "./BrandLogo.svelte";
  import BrandText from "./BrandText.svelte";

  interface Props {
    size?: "sm" | "md" | "lg" | "xl";
    layout?: "horizontal" | "vertical";
  }

  let { size = "md", layout = "horizontal" }: Props = $props();

  // Adjust logo size to match text size visually
  const sizeMap: Record<
    string,
    { logo: "sm" | "md" | "lg" | "xl"; text: "sm" | "md" | "lg" | "xl" }
  > = {
    sm: { logo: "sm", text: "sm" },
    md: { logo: "md", text: "md" }, // approx 48px logo, 24px text
    lg: { logo: "lg", text: "lg" }, // approx 96px logo, 48px text
    xl: { logo: "xl", text: "xl" }, // huge
  };
</script>

<div class="brand-full {layout}">
  <div class="logo-wrapper">
    <BrandLogo size={sizeMap[size].logo} />
  </div>
  <div class="text-wrapper">
    <BrandText size={sizeMap[size].text} />
  </div>
</div>

<style>
  .brand-full {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: var(--mv-brand-gap-horizontal);
  }

  .brand-full.vertical {
    flex-direction: column;
    gap: var(--mv-brand-gap-vertical);
  }

  .brand-full.horizontal {
    flex-direction: row;
  }

  /* Adjust gap based on size */
  :global(.brand-full .logo-wrapper) {
    display: flex;
    align-items: center;
    justify-content: center;
  }
</style>
