<script lang="ts">
  import { onMount } from "svelte";
  import Button from "../../design-system/components/Button.svelte";
  import { getToolingConfigJSON, setToolingConfigJSON } from "../../services/toolSettings";

  type ToolingProfile = {
    id?: string;
    name?: string;
  };

  type ToolingConfig = {
    activeProfile?: string;
    profiles?: ToolingProfile[];
    force?: {
      enabled?: boolean;
      tool?: string;
      model?: string;
    };
  };

  let rawJSON = $state("");
  let parseError = $state("");
  let statusMessage = $state("");
  let statusTone = $state<"ok" | "error" | "">("");

  let activeProfile = $state("");
  let profiles = $state<ToolingProfile[]>([]);
  let forceEnabled = $state(false);
  let forceTool = $state("");
  let forceModel = $state("");

  onMount(() => {
    void loadConfig();
  });

  async function loadConfig() {
    statusMessage = "";
    statusTone = "";
    try {
      rawJSON = await getToolingConfigJSON();
      syncFromRaw(rawJSON);
    } catch (err) {
      parseError = err instanceof Error ? err.message : String(err);
      statusTone = "error";
    }
  }

  function parseConfig(raw: string): ToolingConfig | null {
    try {
      return JSON.parse(raw) as ToolingConfig;
    } catch {
      return null;
    }
  }

  function syncFromRaw(raw: string) {
    const cfg = parseConfig(raw);
    if (!cfg) {
      parseError = "Invalid JSON.";
      return;
    }
    parseError = "";
    activeProfile = cfg.activeProfile ?? "";
    profiles = cfg.profiles ?? [];
    forceEnabled = cfg.force?.enabled ?? false;
    forceTool = cfg.force?.tool ?? "";
    forceModel = cfg.force?.model ?? "";
  }

  function updateConfig(update: (cfg: ToolingConfig) => void) {
    const cfg = parseConfig(rawJSON);
    if (!cfg) {
      parseError = "Invalid JSON.";
      statusTone = "error";
      return;
    }
    update(cfg);
    rawJSON = JSON.stringify(cfg, null, 2);
    syncFromRaw(rawJSON);
  }

  function handleRawInput(event: Event) {
    const target = event.currentTarget as HTMLTextAreaElement;
    rawJSON = target.value;
    syncFromRaw(rawJSON);
  }

  function handleProfileChange(event: Event) {
    const target = event.currentTarget as HTMLSelectElement;
    updateConfig((cfg) => {
      cfg.activeProfile = target.value;
    });
  }

  function handleForceToggle(event: Event) {
    const target = event.currentTarget as HTMLInputElement;
    updateConfig((cfg) => {
      if (!cfg.force) {
        cfg.force = {};
      }
      cfg.force.enabled = target.checked;
    });
  }

  function handleForceTool(event: Event) {
    const target = event.currentTarget as HTMLInputElement;
    updateConfig((cfg) => {
      if (!cfg.force) {
        cfg.force = {};
      }
      cfg.force.tool = target.value;
    });
  }

  function handleForceModel(event: Event) {
    const target = event.currentTarget as HTMLInputElement;
    updateConfig((cfg) => {
      if (!cfg.force) {
        cfg.force = {};
      }
      cfg.force.model = target.value;
    });
  }

  async function applyConfig() {
    statusMessage = "";
    statusTone = "";
    const cfg = parseConfig(rawJSON);
    if (!cfg) {
      parseError = "Invalid JSON.";
      statusTone = "error";
      return;
    }

    try {
      await setToolingConfigJSON(rawJSON);
      statusMessage = "Settings saved.";
      statusTone = "ok";
      parseError = "";
      syncFromRaw(rawJSON);
    } catch (err) {
      statusMessage = err instanceof Error ? err.message : String(err);
      statusTone = "error";
    }
  }
</script>

<div class="panel">
  <div class="section">
    <div class="section-header">
      <span class="section-title">Profiles</span>
      <span class="section-subtitle">Switch weighted mixes by category.</span>
    </div>
    <div class="field-row">
      <label class="label" for="tooling-profile">Active</label>
      <select
        id="tooling-profile"
        class="select"
        value={activeProfile}
        onchange={handleProfileChange}
        disabled={profiles.length === 0}
      >
        {#if profiles.length === 0}
          <option value="">No profiles</option>
        {:else}
          {#each profiles as profile}
            <option value={profile.id ?? ""}>
              {profile.name ?? profile.id ?? "Unnamed"}
            </option>
          {/each}
        {/if}
      </select>
    </div>
  </div>

  <div class="section">
    <div class="section-header">
      <span class="section-title">Force Mode</span>
      <span class="section-subtitle">
        Run all categories with the specified tool/model.
      </span>
    </div>
    <div class="toggle-row">
      <label class="toggle">
        <input
          type="checkbox"
          class="toggle-input"
          checked={forceEnabled}
          onchange={handleForceToggle}
        />
        <span class="toggle-track"></span>
        <span class="toggle-thumb"></span>
        <span class="toggle-label">Enable force mode</span>
      </label>
    </div>
    <div class="field-grid">
      <div class="field">
        <label class="label" for="force-tool">Tool</label>
        <input
          id="force-tool"
          class="text-input"
          type="text"
          value={forceTool}
          placeholder="codex-cli / claude-code / gemini-cli"
          oninput={handleForceTool}
        />
      </div>
      <div class="field">
        <label class="label" for="force-model">Model</label>
        <input
          id="force-model"
          class="text-input"
          type="text"
          value={forceModel}
          placeholder="gpt-5.2 / gemini-3-flash-preview"
          oninput={handleForceModel}
        />
      </div>
    </div>
  </div>

  <div class="section">
    <div class="section-header">
      <span class="section-title">Advanced JSON</span>
      <span class="section-subtitle">
        Edit categories, weights, candidates, and fallbacks.
      </span>
    </div>
    <textarea
      class="json-textarea"
      value={rawJSON}
      oninput={handleRawInput}
      spellcheck={false}
    ></textarea>
    {#if parseError}
      <div class="status error">{parseError}</div>
    {/if}
  </div>

  <div class="actions">
    <Button variant="secondary" label="Reload" onclick={loadConfig} />
    <Button variant="primary" label="Apply" onclick={applyConfig} />
  </div>

  {#if statusMessage}
    <div class="status" class:ok={statusTone === "ok"} class:error={statusTone === "error"}>
      {statusMessage}
    </div>
  {/if}
</div>

<style>
  .panel {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-lg);
    padding: var(--mv-spacing-md);
    color: var(--mv-color-text-primary);
  }

  .section {
    background: var(--mv-color-surface-primary);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    border-radius: var(--mv-radius-lg);
    padding: var(--mv-spacing-md);
    box-shadow: var(--mv-shadow-glass-panel);
  }

  .section-header {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-2xs);
    margin-bottom: var(--mv-spacing-md);
  }

  .section-title {
    font-size: var(--mv-font-size-md);
    font-weight: var(--mv-font-weight-semibold);
    letter-spacing: var(--mv-letter-spacing-wide);
  }

  .section-subtitle {
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-secondary);
  }

  .field-row {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-md);
  }

  .field-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: var(--mv-spacing-md);
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-xs);
  }

  .label {
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-secondary);
    font-weight: var(--mv-font-weight-medium);
  }

  .select {
    flex: 1;
    height: var(--mv-input-height-md);
    padding: 0 var(--mv-spacing-md);
    border-radius: var(--mv-radius-md);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    background: var(--mv-color-surface-secondary);
    color: var(--mv-color-text-primary);
    font-family: var(--mv-font-sans);
    font-size: var(--mv-font-size-sm);
  }

  .select:focus {
    outline: none;
    border-color: var(--mv-color-border-focus);
    box-shadow: var(--mv-shadow-focus);
  }

  .text-input {
    height: var(--mv-input-height-md);
    padding: 0 var(--mv-spacing-md);
    border-radius: var(--mv-radius-md);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    background: var(--mv-color-surface-secondary);
    color: var(--mv-color-text-primary);
    font-family: var(--mv-font-sans);
    font-size: var(--mv-font-size-sm);
  }

  .text-input:focus {
    outline: none;
    border-color: var(--mv-color-border-focus);
    box-shadow: var(--mv-shadow-focus);
  }

  .json-textarea {
    width: 100%;
    min-height: var(--mv-textarea-min-height);
    padding: var(--mv-spacing-md);
    border-radius: var(--mv-radius-md);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    background: var(--mv-color-surface-secondary);
    color: var(--mv-color-text-primary);
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-xs);
    line-height: var(--mv-line-height-relaxed);
    resize: vertical;
  }

  .json-textarea:focus {
    outline: none;
    border-color: var(--mv-color-border-focus);
    box-shadow: var(--mv-shadow-focus);
  }

  .toggle-row {
    margin-bottom: var(--mv-spacing-md);
  }

  .toggle {
    display: inline-flex;
    align-items: center;
    gap: var(--mv-spacing-sm);
    cursor: pointer;
    color: var(--mv-color-text-primary);
    font-size: var(--mv-font-size-sm);
  }

  .toggle-input {
    position: absolute;
    opacity: 0;
    pointer-events: none;
  }

  .toggle-track {
    width: var(--mv-toggle-track-width);
    height: var(--mv-toggle-track-height);
    border-radius: var(--mv-radius-full);
    background: var(--mv-color-surface-secondary);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    position: relative;
    transition: var(--mv-transition-hover);
  }

  .toggle-thumb {
    width: var(--mv-toggle-thumb-size);
    height: var(--mv-toggle-thumb-size);
    border-radius: var(--mv-radius-full);
    background: var(--mv-color-text-secondary);
    position: relative;
    left: var(--mv-toggle-thumb-offset);
    transition: var(--mv-transition-hover);
    box-shadow: var(--mv-shadow-glow-subtle);
  }

  .toggle-input:checked + .toggle-track {
    border-color: var(--mv-color-border-focus);
    box-shadow: var(--mv-shadow-focus);
  }

  .toggle-input:checked + .toggle-track + .toggle-thumb {
    transform: translateX(18px);
    background: var(--mv-primitive-frost-2);
  }

  .toggle-label {
    font-weight: var(--mv-font-weight-medium);
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--mv-spacing-sm);
  }

  .status {
    margin-top: var(--mv-spacing-sm);
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
    border-radius: var(--mv-radius-md);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    background: var(--mv-color-surface-secondary);
    font-size: var(--mv-font-size-sm);
  }

  .status.ok {
    border-color: var(--mv-color-status-success-border);
    color: var(--mv-color-status-success-text);
  }

  .status.error {
    border-color: var(--mv-color-status-failed-border);
    color: var(--mv-color-status-failed-text);
  }
</style>
