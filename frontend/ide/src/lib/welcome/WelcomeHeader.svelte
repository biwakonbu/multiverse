<script lang="ts">
  // Welcome Header - Logo and Title
</script>

<header class="welcome-header">
  <svg class="logo" viewBox="0 0 340 72" fill="none" xmlns="http://www.w3.org/2000/svg">
    <!-- Logo Icon - Multiverse Grid -->
    <g class="logo-icon">
      <rect x="8" y="8" width="24" height="24" rx="4" class="grid-cell-primary"/>
      <rect x="36" y="8" width="24" height="24" rx="4" class="grid-cell-secondary"/>
      <rect x="8" y="36" width="24" height="24" rx="4" class="grid-cell-secondary"/>
      <rect x="36" y="36" width="24" height="24" rx="4" class="grid-cell-primary"/>
    </g>

    <!-- Text: MULTIVERSE -->
    <text x="76" y="46" class="title-text">
      <tspan class="title-main">MULTI</tspan><tspan class="title-accent">VERSE</tspan>
    </text>
  </svg>
</header>

<style>
  .welcome-header {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding-bottom: var(--mv-spacing-xl);
    border-bottom: var(--mv-border-width-thin) solid var(--mv-color-border-subtle);
  }

  .logo {
    width: 340px;
    height: 72px;
  }

  .logo-icon {
    color: var(--mv-color-interactive-primary);
  }

  .grid-cell-primary {
    fill: currentColor;
    opacity: 1;
    transition: opacity 0.3s ease;
  }

  .grid-cell-secondary {
    fill: currentColor;
    opacity: 0.5;
    transition: opacity 0.3s ease;
  }

  .title-text {
    font-family: var(--mv-font-family-primary);
    font-size: 32px;
    font-weight: 700;
    letter-spacing: 0.15em;
  }

  .title-main {
    fill: var(--mv-color-text-primary);
  }

  .title-accent {
    fill: var(--mv-color-interactive-primary);
  }

  /* Hover animation */
  .welcome-header:hover .grid-cell-primary {
    opacity: 0.7;
  }

  .welcome-header:hover .grid-cell-secondary {
    opacity: 0.9;
  }
</style>
