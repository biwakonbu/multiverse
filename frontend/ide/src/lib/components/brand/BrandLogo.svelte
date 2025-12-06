<script lang="ts">
  import logo from "../../../assets/logo.svg";

  export let size: "sm" | "md" | "lg" | "xl" | "custom" = "md";
  export let customSize: string = "32px";

  const sizes = {
    sm: "24px",
    md: "48px",
    lg: "96px",
    xl: "192px",
    custom: customSize,
  };

  $: computedSize = size === "custom" ? customSize : sizes[size];
</script>

<div class="brand-logo" style:--logo-size={computedSize}>
  <img src={logo} alt="Multiverse Logo" />
</div>

<style>
  .brand-logo {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: var(--logo-size);
    height: var(--logo-size);
  }

  img {
    width: 100%;
    height: 100%;
    object-fit: contain;
    filter: var(--mv-filter-drop-shadow-brand);
    transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  }

  .brand-logo:hover img {
    transform: translateY(-4px) scale(1.05) rotate(-3deg);
    filter: var(--mv-filter-drop-shadow-brand-hover);
  }
</style>
