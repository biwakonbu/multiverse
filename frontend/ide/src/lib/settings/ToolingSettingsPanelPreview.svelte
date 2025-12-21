<script lang="ts">
  import Button from "../../design-system/components/Button.svelte";
  import {
    Cpu,
    Zap,
    Code2,
    Check,
    AlertCircle,
    RefreshCw,
    Power,
    PowerOff,
  } from "lucide-svelte";

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

  // Props
  interface Props {
    initialConfig?: ToolingConfig;
  }

  let { initialConfig = {} }: Props = $props();

  // 利用可能なツール
  const availableTools = [
    { id: "codex-cli", name: "Codex CLI", description: "OpenAI Codex" },
    { id: "claude-code", name: "Claude Code", description: "Anthropic Claude" },
    { id: "gemini-cli", name: "Gemini CLI", description: "Google Gemini" },
  ];

  // 利用可能なモデル（ツールごとにグループ化）
  const availableModels = [
    { id: "", name: "Default", group: "Auto" },
    { id: "gpt-4.1", name: "GPT-4.1", group: "OpenAI" },
    { id: "gpt-4.1-mini", name: "GPT-4.1 Mini", group: "OpenAI" },
    { id: "gpt-4.5-preview", name: "GPT-4.5 Preview", group: "OpenAI" },
    { id: "o3", name: "o3", group: "OpenAI" },
    { id: "o4-mini", name: "o4-mini", group: "OpenAI" },
    {
      id: "claude-sonnet-4-20250514",
      name: "Claude Sonnet 4",
      group: "Anthropic",
    },
    {
      id: "claude-opus-4-20250514",
      name: "Claude Opus 4",
      group: "Anthropic",
    },
    { id: "gemini-2.5-pro", name: "Gemini 2.5 Pro", group: "Google" },
    { id: "gemini-2.5-flash", name: "Gemini 2.5 Flash", group: "Google" },
  ];

  // タブ管理
  const tabs = [
    { id: "profiles", label: "Profiles", icon: Cpu },
    { id: "force", label: "Force Mode", icon: Zap },
    { id: "advanced", label: "Advanced", icon: Code2 },
  ] as const;
  type TabId = (typeof tabs)[number]["id"];

  // $derived で props からの初期値を反映
  let activeTab = $state<TabId>(initialConfig.force?.enabled ? "force" : "profiles");
  let rawJSON = $state("");
  let parseError = $state("");
  let statusMessage = $state("");
  let statusTone = $state<"ok" | "error" | "">("");
  let activeProfile = $state("");
  let profiles = $state<ToolingProfile[]>([]);
  let forceEnabled = $state(false);
  let forceTool = $state("");
  let forceModel = $state("");

  // initialConfig が変更されたら状態を更新
  $effect(() => {
    rawJSON = JSON.stringify(initialConfig, null, 2);
    activeProfile = initialConfig.activeProfile ?? "";
    profiles = initialConfig.profiles ?? [];
    forceEnabled = initialConfig.force?.enabled ?? false;
    forceTool = initialConfig.force?.tool ?? "";
    forceModel = initialConfig.force?.model ?? "";
  });

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

  function setForceEnabled(enabled: boolean) {
    updateConfig((cfg) => {
      if (!cfg.force) {
        cfg.force = {};
      }
      cfg.force.enabled = enabled;
    });
  }

  function handleForceTool(event: Event) {
    const target = event.currentTarget as HTMLSelectElement;
    updateConfig((cfg) => {
      if (!cfg.force) {
        cfg.force = {};
      }
      cfg.force.tool = target.value;
    });
  }

  function handleForceModel(event: Event) {
    const target = event.currentTarget as HTMLSelectElement;
    updateConfig((cfg) => {
      if (!cfg.force) {
        cfg.force = {};
      }
      cfg.force.model = target.value;
    });
  }

  function loadConfig() {
    statusMessage = "Configuration reloaded.";
    statusTone = "ok";
  }

  function applyConfig() {
    const cfg = parseConfig(rawJSON);
    if (!cfg) {
      parseError = "Invalid JSON.";
      statusTone = "error";
      return;
    }
    statusMessage = "Settings saved successfully.";
    statusTone = "ok";
    parseError = "";
  }

  // モデルをグループ化
  function getModelGroups() {
    const groups = new Map<string, typeof availableModels>();
    for (const model of availableModels) {
      const existing = groups.get(model.group) ?? [];
      existing.push(model);
      groups.set(model.group, existing);
    }
    return groups;
  }
  const modelGroups = getModelGroups();
</script>

<div class="settings-panel">
  <!-- タブヘッダー -->
  <div class="tab-bar">
    {#each tabs as tab}
      <button
        class="tab-item"
        class:active={activeTab === tab.id}
        onclick={() => (activeTab = tab.id)}
        type="button"
      >
        <tab.icon size={14} />
        <span class="tab-label">{tab.label}</span>
        {#if activeTab === tab.id}
          <div class="tab-indicator"></div>
        {/if}
      </button>
    {/each}
  </div>

  <!-- タブコンテンツ -->
  <div class="tab-content">
    {#if activeTab === "profiles"}
      <!-- Profiles タブ -->
      <div class="content-section">
        <div class="section-header">
          <div class="header-glow"></div>
          <h3 class="section-title">Profile Selection</h3>
          <p class="section-subtitle">
            Switch between weighted tool configurations by category.
          </p>
        </div>

        <div class="crystal-field">
          <label class="field-label" for="tooling-profile">
            <span class="label-text">Active Profile</span>
            <span class="label-hint"
              >Select the configuration profile to use</span
            >
          </label>
          <div class="select-wrapper">
            <select
              id="tooling-profile"
              class="crystal-select"
              value={activeProfile}
              onchange={handleProfileChange}
              disabled={profiles.length === 0}
            >
              {#if profiles.length === 0}
                <option value="">No profiles configured</option>
              {:else}
                {#each profiles as profile}
                  <option value={profile.id ?? ""}>
                    {profile.name ?? profile.id ?? "Unnamed"}
                  </option>
                {/each}
              {/if}
            </select>
            <div class="select-arrow"></div>
          </div>
        </div>

        {#if profiles.length === 0}
          <div class="empty-state">
            <Code2 size={32} class="empty-icon" />
            <p class="empty-text">
              No profiles configured yet. Add profiles in the Advanced tab.
            </p>
          </div>
        {/if}
      </div>
    {:else if activeTab === "force"}
      <!-- Force Mode タブ -->
      <div class="content-section">
        <div class="section-header">
          <div class="header-glow force"></div>
          <h3 class="section-title">Force Mode</h3>
          <p class="section-subtitle">
            Override all categories with a specific tool and model.
          </p>
        </div>

        <!-- Segmented Control Toggle -->
        <div class="segmented-control">
          <button
            class="segment"
            class:active={!forceEnabled}
            onclick={() => setForceEnabled(false)}
            type="button"
          >
            <PowerOff size={14} />
            <span>OFF</span>
          </button>
          <button
            class="segment"
            class:active={forceEnabled}
            onclick={() => setForceEnabled(true)}
            type="button"
          >
            <Power size={14} />
            <span>ON</span>
          </button>
          <div class="segment-indicator" class:right={forceEnabled}></div>
        </div>

        <!-- Force Mode Fields -->
        <div class="force-fields" class:disabled={!forceEnabled}>
          <div class="crystal-field">
            <label class="field-label" for="force-tool">
              <span class="label-text">Tool</span>
              <span class="label-hint">CLI tool to use for all operations</span>
            </label>
            <div class="select-wrapper">
              <select
                id="force-tool"
                class="crystal-select"
                value={forceTool}
                onchange={handleForceTool}
                disabled={!forceEnabled}
              >
                <option value="">Select a tool...</option>
                {#each availableTools as tool}
                  <option value={tool.id}>{tool.name}</option>
                {/each}
              </select>
              <div class="select-arrow"></div>
            </div>
          </div>

          <div class="crystal-field">
            <label class="field-label" for="force-model">
              <span class="label-text">Model</span>
              <span class="label-hint">Override the default model</span>
            </label>
            <div class="select-wrapper">
              <select
                id="force-model"
                class="crystal-select"
                value={forceModel}
                onchange={handleForceModel}
                disabled={!forceEnabled}
              >
                {#each [...modelGroups.entries()] as [group, models]}
                  <optgroup label={group}>
                    {#each models as model}
                      <option value={model.id}>{model.name}</option>
                    {/each}
                  </optgroup>
                {/each}
              </select>
              <div class="select-arrow"></div>
            </div>
          </div>
        </div>

        {#if forceEnabled && forceTool}
          <div class="force-status">
            <Zap size={14} />
            <span>
              Force mode active: Using <strong
                >{availableTools.find((t) => t.id === forceTool)?.name ??
                  forceTool}</strong
              >
              {#if forceModel}
                with <strong
                  >{availableModels.find((m) => m.id === forceModel)?.name ??
                    forceModel}</strong
                >
              {/if}
            </span>
          </div>
        {/if}
      </div>
    {:else if activeTab === "advanced"}
      <!-- Advanced タブ -->
      <div class="content-section">
        <div class="section-header">
          <div class="header-glow advanced"></div>
          <h3 class="section-title">Advanced Configuration</h3>
          <p class="section-subtitle">
            Edit categories, weights, candidates, and fallback settings
            directly.
          </p>
        </div>

        <div class="json-editor-wrapper">
          <div class="editor-header">
            <span class="editor-label">tooling.json</span>
            <div class="editor-actions">
              <button
                class="editor-btn"
                onclick={loadConfig}
                title="Reload configuration"
                type="button"
              >
                <RefreshCw size={14} />
              </button>
            </div>
          </div>
          <textarea
            class="json-editor"
            value={rawJSON}
            oninput={handleRawInput}
            spellcheck={false}
          ></textarea>
          {#if parseError}
            <div class="editor-error">
              <AlertCircle size={14} />
              <span>{parseError}</span>
            </div>
          {/if}
        </div>
      </div>
    {/if}
  </div>

  <!-- フッターアクション -->
  <div class="panel-footer">
    {#if statusMessage}
      <div
        class="status-message"
        class:success={statusTone === "ok"}
        class:error={statusTone === "error"}
      >
        {#if statusTone === "ok"}
          <Check size={14} />
        {:else if statusTone === "error"}
          <AlertCircle size={14} />
        {/if}
        <span>{statusMessage}</span>
      </div>
    {/if}

    <div class="footer-actions">
      <Button variant="ghost" size="small" label="Reload" onclick={loadConfig} />
      <Button
        variant="crystal"
        size="small"
        label="Apply"
        onclick={applyConfig}
      />
    </div>
  </div>
</div>

<style>
  .settings-panel {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--mv-window-content-bg);
  }

  /* ========================================
     タブバー
     ======================================== */
  .tab-bar {
    display: flex;
    gap: var(--mv-spacing-xxs);
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
    background: linear-gradient(to bottom, var(--mv-glass-bg-dark), transparent);
    border-bottom: var(--mv-border-width-thin) solid
      var(--mv-glass-border-subtle);
  }

  .tab-item {
    position: relative;
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
    padding: var(--mv-spacing-xs) var(--mv-spacing-md);
    background: transparent;
    border: none;
    border-radius: var(--mv-radius-sm);
    color: var(--mv-color-text-muted);
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-medium);
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .tab-item:hover {
    color: var(--mv-color-text-primary);
    background: var(--mv-glass-active);
  }

  .tab-item.active {
    color: var(--mv-color-text-primary);
    font-weight: var(--mv-font-weight-semibold);
  }

  .tab-label {
    letter-spacing: var(--mv-letter-spacing-wide);
  }

  .tab-indicator {
    position: absolute;
    bottom: calc(-1 * var(--mv-spacing-sm) - var(--mv-border-width-thin));
    left: 0;
    right: 0;
    height: var(--mv-border-width-md);
    background: var(--mv-primitive-frost-2);
    box-shadow: var(--mv-shadow-tab-glow);
    border-top-left-radius: var(--mv-border-width-md);
    border-top-right-radius: var(--mv-border-width-md);
  }

  /* ========================================
     タブコンテンツ
     ======================================== */
  .tab-content {
    flex: 1;
    overflow-y: auto;
    padding: var(--mv-spacing-lg);
  }

  .content-section {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-lg);
  }

  /* ========================================
     セクションヘッダー（Holographic）
     ======================================== */
  .section-header {
    position: relative;
    padding-bottom: var(--mv-spacing-md);
    border-bottom: var(--mv-border-width-thin) solid
      var(--mv-glass-border-subtle);
  }

  .header-glow {
    position: absolute;
    top: 0;
    left: 0;
    width: 60px;
    height: 3px;
    background: linear-gradient(90deg, var(--mv-primitive-frost-2), transparent);
    border-radius: var(--mv-radius-full);
    box-shadow: var(--mv-shadow-glow-frost-2);
  }

  .header-glow.force {
    background: linear-gradient(
      90deg,
      var(--mv-primitive-aurora-yellow),
      transparent
    );
    box-shadow: var(--mv-shadow-glow-yellow);
  }

  .header-glow.advanced {
    background: linear-gradient(
      90deg,
      var(--mv-primitive-aurora-purple),
      transparent
    );
    box-shadow: 0 0 8px var(--mv-primitive-aurora-purple);
  }

  .section-title {
    margin: var(--mv-spacing-sm) 0 var(--mv-spacing-xxs);
    font-size: var(--mv-font-size-lg);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
    letter-spacing: var(--mv-letter-spacing-wide);
    text-shadow: var(--mv-text-shadow-frost);
  }

  .section-subtitle {
    margin: 0;
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-muted);
    line-height: var(--mv-line-height-relaxed);
  }

  /* ========================================
     Crystal Field
     ======================================== */
  .crystal-field {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-xs);
  }

  .field-label {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-xxxs);
  }

  .label-text {
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-secondary);
    letter-spacing: var(--mv-letter-spacing-wide);
  }

  .label-hint {
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-muted);
  }

  /* ========================================
     Crystal Select
     ======================================== */
  .select-wrapper {
    position: relative;
  }

  .crystal-select {
    width: 100%;
    height: var(--mv-input-height-md);
    padding: 0 var(--mv-spacing-xl) 0 var(--mv-spacing-md);
    background: var(--mv-glass-bg-dark);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-light);
    border-radius: var(--mv-radius-md);
    color: var(--mv-color-text-primary);
    font-family: var(--mv-font-sans);
    font-size: var(--mv-font-size-sm);
    cursor: pointer;
    appearance: none;
    transition: all 0.2s ease;
  }

  .crystal-select:hover:not(:disabled) {
    border-color: var(--mv-glass-border-hover);
    background: var(--mv-glass-bg-darker);
  }

  .crystal-select:focus {
    outline: none;
    border-color: var(--mv-primitive-frost-2);
    box-shadow: var(--mv-shadow-glow-subtle);
  }

  .crystal-select:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .select-arrow {
    position: absolute;
    right: var(--mv-spacing-md);
    top: 50%;
    transform: translateY(-50%);
    width: 0;
    height: 0;
    border-left: 5px solid transparent;
    border-right: 5px solid transparent;
    border-top: 5px solid var(--mv-color-text-muted);
    pointer-events: none;
  }

  /* ========================================
     Segmented Control (ON/OFF Toggle)
     ======================================== */
  .segmented-control {
    position: relative;
    display: inline-grid;
    grid-template-columns: 1fr 1fr;
    padding: 3px;
    background: var(--mv-glass-bg-darker);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-light);
    border-radius: var(--mv-radius-lg);
    gap: 2px;
    min-width: 180px;
    max-width: 220px;
  }

  .segment {
    position: relative;
    z-index: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: var(--mv-spacing-xs);
    padding: var(--mv-spacing-sm) var(--mv-spacing-lg);
    background: transparent;
    border: none;
    border-radius: var(--mv-radius-md);
    color: var(--mv-color-text-muted);
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-semibold);
    letter-spacing: var(--mv-letter-spacing-wide);
    cursor: pointer;
    transition: color 0.2s ease;
  }

  .segment:hover:not(.active) {
    color: var(--mv-color-text-secondary);
  }

  .segment.active {
    color: var(--mv-color-text-primary);
  }

  .segment-indicator {
    position: absolute;
    top: 3px;
    left: 3px;
    width: calc(50% - 3.5px);
    height: calc(100% - 6px);
    background: linear-gradient(
      135deg,
      rgba(191, 97, 106, 0.3),
      rgba(191, 97, 106, 0.15)
    );
    border: var(--mv-border-width-thin) solid rgba(191, 97, 106, 0.4);
    border-radius: var(--mv-radius-md);
    box-shadow:
      0 0 12px rgba(191, 97, 106, 0.3),
      inset 0 1px 0 rgba(255, 255, 255, 0.1);
    transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .segment-indicator.right {
    left: calc(50% + 1px);
    background: linear-gradient(
      135deg,
      rgba(163, 190, 140, 0.3),
      rgba(163, 190, 140, 0.15)
    );
    border-color: rgba(163, 190, 140, 0.5);
    box-shadow:
      0 0 16px rgba(163, 190, 140, 0.4),
      inset 0 1px 0 rgba(255, 255, 255, 0.15);
  }

  /* ========================================
     Force Fields Container
     ======================================== */
  .force-fields {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-md);
    padding: var(--mv-spacing-md);
    background: var(--mv-glass-bg-dark);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
    border-radius: var(--mv-radius-md);
    transition: opacity 0.2s ease;
  }

  .force-fields.disabled {
    opacity: 0.4;
    pointer-events: none;
  }

  /* ========================================
     Force Status
     ======================================== */
  .force-status {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-sm);
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
    background: linear-gradient(90deg, rgba(163, 190, 140, 0.1), transparent);
    border: var(--mv-border-width-thin) solid rgba(163, 190, 140, 0.3);
    border-radius: var(--mv-radius-md);
    color: var(--mv-primitive-aurora-green);
    font-size: var(--mv-font-size-sm);
  }

  .force-status strong {
    color: var(--mv-color-text-primary);
  }

  /* ========================================
     Empty State
     ======================================== */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--mv-spacing-sm);
    padding: var(--mv-spacing-xl);
    background: var(--mv-glass-bg-dark);
    border: var(--mv-border-width-thin) dashed var(--mv-glass-border-light);
    border-radius: var(--mv-radius-md);
    text-align: center;
  }

  :global(.empty-icon) {
    color: var(--mv-color-text-muted);
    opacity: 0.5;
  }

  .empty-text {
    margin: 0;
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-muted);
  }

  /* ========================================
     JSON Editor
     ======================================== */
  .json-editor-wrapper {
    display: flex;
    flex-direction: column;
    background: var(--mv-glass-bg-darker);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-light);
    border-radius: var(--mv-radius-md);
    overflow: hidden;
  }

  .editor-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--mv-spacing-xs) var(--mv-spacing-md);
    background: var(--mv-glass-bg-dark);
    border-bottom: var(--mv-border-width-thin) solid
      var(--mv-glass-border-subtle);
  }

  .editor-label {
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-muted);
    letter-spacing: var(--mv-letter-spacing-wide);
  }

  .editor-actions {
    display: flex;
    gap: var(--mv-spacing-xxs);
  }

  .editor-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    background: transparent;
    border: none;
    border-radius: var(--mv-radius-sm);
    color: var(--mv-color-text-muted);
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .editor-btn:hover {
    color: var(--mv-color-text-primary);
    background: var(--mv-glass-active);
  }

  .json-editor {
    width: 100%;
    min-height: 300px;
    padding: var(--mv-spacing-md);
    background: transparent;
    border: none;
    color: var(--mv-primitive-frost-1);
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-xs);
    line-height: var(--mv-line-height-relaxed);
    resize: vertical;
  }

  .json-editor:focus {
    outline: none;
  }

  .editor-error {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
    background: rgba(191, 97, 106, 0.1);
    border-top: var(--mv-border-width-thin) solid var(--mv-primitive-aurora-red);
    color: var(--mv-primitive-pastel-red);
    font-size: var(--mv-font-size-sm);
  }

  /* ========================================
     Panel Footer
     ======================================== */
  .panel-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--mv-spacing-md);
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
    background: var(--mv-window-footer-bg);
    border-top: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
  }

  .status-message {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-muted);
  }

  .status-message.success {
    color: var(--mv-primitive-aurora-green);
  }

  .status-message.error {
    color: var(--mv-primitive-pastel-red);
  }

  .footer-actions {
    display: flex;
    gap: var(--mv-spacing-sm);
    margin-left: auto;
  }
</style>
