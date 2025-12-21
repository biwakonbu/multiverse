<script lang="ts">
  import { onMount } from "svelte";
  import Button from "../../design-system/components/Button.svelte";
  import {
    getToolingConfigJSON,
    setToolingConfigJSON,
    getAvailableTools,
    getAvailableModels,
    getModelsForTool,
    type ToolOption,
    type ModelOption,
  } from "../../services/toolSettings";
  import { Cpu, Zap, Code2, Check, AlertCircle, RefreshCw, Power, PowerOff } from "lucide-svelte";

  type ToolCandidate = {
    tool?: string;
    model?: string;
    weight?: number;
  };

  type ToolCategoryConfig = {
    strategy?: string;
    candidates?: ToolCandidate[];
    fallbackOnRateLimit?: boolean;
    cooldownSec?: number;
  };

  type ToolingProfile = {
    id?: string;
    name?: string;
    categories?: Record<string, ToolCategoryConfig>;
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

  // Go側から取得する利用可能なツールとモデル
  let availableTools = $state<ToolOption[]>([]);
  let availableModels = $state<ModelOption[]>([]);
  // 選択中のツールでサポートされるモデル（フィルタリング済み）
  let filteredModels = $state<ModelOption[]>([]);

  // タブ管理
  const tabs = [
    { id: "profiles", label: "Profiles", icon: Cpu },
    { id: "force", label: "Force Mode", icon: Zap },
    { id: "advanced", label: "Advanced", icon: Code2 },
  ] as const;
  type TabId = (typeof tabs)[number]["id"];
  let activeTab = $state<TabId>("profiles");

  let rawJSON = $state("");
  let parseError = $state("");
  let statusMessage = $state("");
  let statusTone = $state<"ok" | "error" | "">("");

  let activeProfile = $state("");
  let profiles = $state<ToolingProfile[]>([]);
  let forceEnabled = $state(false);
  let forceTool = $state("");
  let forceModel = $state("");
  let editProfileId = $state("");
  let newCategoryName = $state("");
  let profileIdError = $state("");
  let modelsByTool = $state<Record<string, ModelOption[]>>({});

  const editingProfile = $derived.by(() => profiles.find(p => (p.id ?? "") === editProfileId));
  const editingCategories = $derived.by(() => {
    if (!editingProfile?.categories) {
      return [] as Array<[string, ToolCategoryConfig]>;
    }
    return Object.entries(editingProfile.categories).sort(([a], [b]) => a.localeCompare(b));
  });

  onMount(() => {
    void loadAll();
  });

  async function loadAll() {
    // ツールとモデルの一覧をGoから取得
    try {
      const [tools, models] = await Promise.all([
        getAvailableTools(),
        getAvailableModels(),
      ]);
      availableTools = tools;
      availableModels = models;
    } catch (err) {
      console.error("Failed to load tools/models:", err);
    }
    // 設定を読み込み
    await loadConfig();
    await primeModelsFromConfig();
    // 初期状態でモデルをフィルタリング
    await updateFilteredModels(forceTool);
  }

  // 選択中のツールに応じてモデル一覧をフィルタリング
  async function updateFilteredModels(toolID: string) {
    if (!toolID) {
      // ツール未選択時は全モデルを表示
      filteredModels = availableModels;
      return;
    }
    try {
      const models = await getModelsForTool(toolID);
      filteredModels = models.length > 0 ? models : availableModels;
    } catch (err) {
      console.error("Failed to get models for tool:", err);
      filteredModels = availableModels;
    }
  }

  function modelsForTool(toolID: string): ModelOption[] {
    if (!toolID) {
      return availableModels;
    }
    return modelsByTool[toolID] ?? availableModels;
  }

  async function ensureModelsForTool(toolID: string) {
    if (!toolID || modelsByTool[toolID]) {
      return;
    }
    try {
      const models = await getModelsForTool(toolID);
      modelsByTool = {
        ...modelsByTool,
        [toolID]: models.length > 0 ? models : availableModels,
      };
    } catch (err) {
      console.error("Failed to get models for tool:", err);
      modelsByTool = {
        ...modelsByTool,
        [toolID]: availableModels,
      };
    }
  }

  async function primeModelsFromConfig() {
    const cfg = parseConfig(rawJSON);
    if (!cfg) {
      return;
    }
    const toolIDs = new Set<string>();
    if (cfg.force?.tool) {
      toolIDs.add(cfg.force.tool);
    }
    for (const profile of cfg.profiles ?? []) {
      for (const category of Object.values(profile.categories ?? {})) {
        for (const candidate of category.candidates ?? []) {
          if (candidate.tool) {
            toolIDs.add(candidate.tool);
          }
        }
      }
    }
    for (const toolID of toolIDs) {
      await ensureModelsForTool(toolID);
    }
  }

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
    const nextEditId = profiles.some(p => (p.id ?? "") === editProfileId)
      ? editProfileId
      : (activeProfile || profiles[0]?.id || "");
    editProfileId = nextEditId;
    profileIdError = "";
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

  function handleEditProfileChange(event: Event) {
    const target = event.currentTarget as HTMLSelectElement;
    editProfileId = target.value;
    profileIdError = "";
  }

  function ensureProfiles(cfg: ToolingConfig) {
    if (!cfg.profiles) {
      cfg.profiles = [];
    }
  }

  function cloneProfile(profile?: ToolingProfile): ToolingProfile {
    if (!profile) {
      return { id: "", name: "", categories: {} };
    }
    return JSON.parse(JSON.stringify(profile)) as ToolingProfile;
  }

  function nextCustomIdentity(existing: ToolingProfile[]) {
    const existingIDs = new Set(existing.map(profile => profile.id ?? ""));
    let index = 1;
    let id = `custom-${index}`;
    while (existingIDs.has(id)) {
      index += 1;
      id = `custom-${index}`;
    }
    return { id, name: `Custom ${index}` };
  }

  function nextCopyIdentity(base: ToolingProfile, existing: ToolingProfile[]) {
    const existingIDs = new Set(existing.map(profile => profile.id ?? ""));
    const existingNames = new Set(existing.map(profile => profile.name ?? ""));
    const baseID = (base.id ?? "profile").trim() || "profile";
    const baseName = (base.name ?? base.id ?? "Profile").trim() || "Profile";
    let id = `${baseID}-copy`;
    let index = 2;
    while (existingIDs.has(id)) {
      id = `${baseID}-copy-${index}`;
      index += 1;
    }
    let name = `${baseName} Copy`;
    let nameIndex = 2;
    while (existingNames.has(name)) {
      name = `${baseName} Copy ${nameIndex}`;
      nameIndex += 1;
    }
    return { id, name };
  }

  function createProfileFromDefault() {
    let newID = "";
    updateConfig((cfg) => {
      ensureProfiles(cfg);
      const base = cfg.profiles?.find(profile => profile.id === "balanced") ?? cfg.profiles?.[0];
      const clone = cloneProfile(base);
      const identity = nextCustomIdentity(cfg.profiles ?? []);
      clone.id = identity.id;
      clone.name = identity.name;
      if (!clone.categories) {
        clone.categories = {};
      }
      cfg.profiles?.push(clone);
      cfg.activeProfile = identity.id;
      newID = identity.id;
    });
    if (newID) {
      editProfileId = newID;
    }
  }

  function duplicateProfile() {
    if (!editingProfile) {
      return;
    }
    let newID = "";
    updateConfig((cfg) => {
      ensureProfiles(cfg);
      const base = cfg.profiles?.find(profile => profile.id === editingProfile.id);
      if (!base) {
        return;
      }
      const clone = cloneProfile(base);
      const identity = nextCopyIdentity(base, cfg.profiles ?? []);
      clone.id = identity.id;
      clone.name = identity.name;
      cfg.profiles?.push(clone);
      newID = identity.id;
    });
    if (newID) {
      editProfileId = newID;
    }
  }

  function deleteProfile() {
    if (!editingProfile?.id) {
      return;
    }
    if (!window.confirm(`Delete profile "${editingProfile.name ?? editingProfile.id}"?`)) {
      return;
    }
    const deleteID = editingProfile.id;
    updateConfig((cfg) => {
      ensureProfiles(cfg);
      cfg.profiles = (cfg.profiles ?? []).filter(profile => (profile.id ?? "") !== deleteID);
      if (cfg.activeProfile === deleteID) {
        cfg.activeProfile = cfg.profiles[0]?.id ?? "";
      }
    });
  }

  function handleProfileNameInput(event: Event) {
    if (!editingProfile?.id) {
      return;
    }
    const target = event.currentTarget as HTMLInputElement;
    updateConfig((cfg) => {
      const profile = cfg.profiles?.find(p => (p.id ?? "") === editingProfile.id);
      if (profile) {
        profile.name = target.value;
      }
    });
  }

  function changeProfileId() {
    if (!editingProfile?.id) {
      return;
    }
    const currentID = editingProfile.id;
    const nextID = window.prompt("New profile ID", currentID)?.trim() ?? "";
    if (!nextID || nextID === currentID) {
      return;
    }
    if (profiles.some(profile => (profile.id ?? "") === nextID)) {
      profileIdError = "Profile ID already exists.";
      return;
    }
    profileIdError = "";
    updateConfig((cfg) => {
      const profile = cfg.profiles?.find(p => (p.id ?? "") === currentID);
      if (profile) {
        profile.id = nextID;
      }
      if (cfg.activeProfile === currentID) {
        cfg.activeProfile = nextID;
      }
    });
    editProfileId = nextID;
  }

  function addCategory() {
    if (!editingProfile?.id) {
      return;
    }
    const name = newCategoryName.trim();
    if (!name) {
      return;
    }
    updateConfig((cfg) => {
      const profile = cfg.profiles?.find(p => (p.id ?? "") === editingProfile.id);
      if (!profile) {
        return;
      }
      if (!profile.categories) {
        profile.categories = {};
      }
      if (!profile.categories[name]) {
        profile.categories[name] = {
          strategy: "weighted",
          fallbackOnRateLimit: true,
          cooldownSec: 120,
          candidates: [],
        };
      }
    });
    newCategoryName = "";
  }

  function removeCategory(categoryName: string) {
    if (!editingProfile?.id) {
      return;
    }
    if (!window.confirm(`Delete category "${categoryName}"?`)) {
      return;
    }
    updateConfig((cfg) => {
      const profile = cfg.profiles?.find(p => (p.id ?? "") === editingProfile.id);
      if (!profile?.categories) {
        return;
      }
      delete profile.categories[categoryName];
    });
  }

  function updateCategory(categoryName: string, update: (category: ToolCategoryConfig) => void) {
    if (!editingProfile?.id) {
      return;
    }
    updateConfig((cfg) => {
      const profile = cfg.profiles?.find(p => (p.id ?? "") === editingProfile.id);
      if (!profile) {
        return;
      }
      if (!profile.categories) {
        profile.categories = {};
      }
      const category = profile.categories[categoryName] ?? {
        strategy: "weighted",
        fallbackOnRateLimit: true,
        cooldownSec: 120,
        candidates: [],
      };
      update(category);
      profile.categories[categoryName] = category;
    });
  }

  function updateCandidate(categoryName: string, index: number, update: (candidate: ToolCandidate) => void) {
    updateCategory(categoryName, (category) => {
      const candidates = category.candidates ?? [];
      if (!candidates[index]) {
        candidates[index] = { tool: "", model: "", weight: 0 };
      }
      update(candidates[index]);
      category.candidates = candidates;
    });
  }

  function addCandidate(categoryName: string) {
    updateCategory(categoryName, (category) => {
      const candidates = category.candidates ?? [];
      candidates.push({
        tool: availableTools[0]?.id ?? "",
        model: "",
        weight: 10,
      });
      category.candidates = candidates;
    });
  }

  function removeCandidate(categoryName: string, index: number) {
    updateCategory(categoryName, (category) => {
      const candidates = category.candidates ?? [];
      category.candidates = candidates.filter((_, idx) => idx !== index);
    });
  }

  async function handleCandidateToolChange(categoryName: string, index: number, event: Event) {
    const target = event.currentTarget as HTMLSelectElement;
    const nextTool = target.value;
    await ensureModelsForTool(nextTool);
    updateCandidate(categoryName, index, (candidate) => {
      candidate.tool = nextTool;
      const supportedModels = modelsForTool(nextTool);
      if (candidate.model && !supportedModels.some(model => model.id === candidate.model)) {
        candidate.model = "";
      }
    });
  }

  function handleCandidateModelChange(categoryName: string, index: number, event: Event) {
    const target = event.currentTarget as HTMLSelectElement;
    updateCandidate(categoryName, index, (candidate) => {
      candidate.model = target.value;
    });
  }

  function handleCandidateWeightChange(categoryName: string, index: number, event: Event) {
    const target = event.currentTarget as HTMLInputElement;
    const value = Number(target.value);
    updateCandidate(categoryName, index, (candidate) => {
      candidate.weight = Number.isFinite(value) ? Math.max(0, value) : 0;
    });
  }

  function handleCategoryStrategyChange(categoryName: string, event: Event) {
    const target = event.currentTarget as HTMLSelectElement;
    updateCategory(categoryName, (category) => {
      category.strategy = target.value;
    });
  }

  function handleCategoryFallbackChange(categoryName: string, event: Event) {
    const target = event.currentTarget as HTMLSelectElement;
    updateCategory(categoryName, (category) => {
      category.fallbackOnRateLimit = target.value === "true";
    });
  }

  function handleCategoryCooldownChange(categoryName: string, event: Event) {
    const target = event.currentTarget as HTMLInputElement;
    const value = Number(target.value);
    updateCategory(categoryName, (category) => {
      category.cooldownSec = Number.isFinite(value) ? Math.max(0, value) : 0;
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

  async function handleForceTool(event: Event) {
    const target = event.currentTarget as HTMLSelectElement;
    const newTool = target.value;

    // モデル一覧をフィルタリング
    await updateFilteredModels(newTool);
    await ensureModelsForTool(newTool);

    // 現在選択中のモデルがサポートされていない場合はリセット
    const isModelSupported = filteredModels.some(m => m.id === forceModel);

    updateConfig((cfg) => {
      if (!cfg.force) {
        cfg.force = {};
      }
      cfg.force.tool = newTool;
      // サポートされていないモデルはデフォルト（空）にリセット
      if (!isModelSupported) {
        cfg.force.model = "";
      }
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
      statusMessage = "Settings saved successfully.";
      statusTone = "ok";
      parseError = "";
      syncFromRaw(rawJSON);
    } catch (err) {
      statusMessage = err instanceof Error ? err.message : String(err);
      statusTone = "error";
    }
  }

  // フィルタリングされたモデルをグループ化（動的データに対応）
  const modelGroups = $derived.by(() => {
    const groups = new Map<string, ModelOption[]>();
    // filteredModels が空の場合は availableModels を使用
    const modelsToGroup = filteredModels.length > 0 ? filteredModels : availableModels;
    for (const model of modelsToGroup) {
      const existing = groups.get(model.group) ?? [];
      existing.push(model);
      groups.set(model.group, existing);
    }
    return groups;
  });
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
          <h3 class="section-title">Profiles & Categories</h3>
          <p class="section-subtitle">
            Create custom profiles and edit category-level tool weights.
          </p>
        </div>

        <div class="crystal-field">
          <label class="field-label" for="tooling-profile">
            <span class="label-text">Active Profile</span>
            <span class="label-hint">Select the configuration profile to use</span>
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
              No profiles configured yet. Create a profile from the default.
            </p>
            <Button
              variant="crystal"
              size="small"
              label="Create from Default"
              onclick={createProfileFromDefault}
            />
          </div>
        {:else}
          <div class="profile-editor">
            <div class="profile-editor-header">
              <div>
                <h4 class="profile-editor-title">Profile Builder</h4>
                <p class="profile-editor-subtitle">
                  Manage profiles, categories, and candidate weights without JSON.
                </p>
              </div>
              <div class="profile-editor-actions">
                <Button
                  variant="crystal"
                  size="small"
                  label="New from Default"
                  onclick={createProfileFromDefault}
                />
                <Button
                  variant="ghost"
                  size="small"
                  label="Duplicate"
                  onclick={duplicateProfile}
                  disabled={!editingProfile}
                />
                <Button
                  variant="ghost"
                  size="small"
                  label="Delete"
                  onclick={deleteProfile}
                  disabled={profiles.length <= 1}
                />
              </div>
            </div>

            <div class="profile-meta-grid">
              <div class="crystal-field">
                <label class="field-label" for="edit-profile">
                  <span class="label-text">Editing Profile</span>
                  <span class="label-hint">Select a profile to edit</span>
                </label>
                <div class="select-wrapper">
                  <select
                    id="edit-profile"
                    class="crystal-select"
                    value={editProfileId}
                    onchange={handleEditProfileChange}
                  >
                    {#each profiles as profile}
                      <option value={profile.id ?? ""}>
                        {profile.name ?? profile.id ?? "Unnamed"}
                      </option>
                    {/each}
                  </select>
                  <div class="select-arrow"></div>
                </div>
              </div>

              <div class="crystal-field">
                <label class="field-label" for="profile-name">
                  <span class="label-text">Profile Name</span>
                  <span class="label-hint">Displayed in the profile selector</span>
                </label>
                <input
                  id="profile-name"
                  class="crystal-input"
                  type="text"
                  value={editingProfile?.name ?? ""}
                  placeholder="Custom profile"
                  oninput={handleProfileNameInput}
                  disabled={!editingProfile}
                />
              </div>

              <div class="crystal-field">
                <label class="field-label">
                  <span class="label-text">Profile ID</span>
                  <span class="label-hint">Unique identifier for this profile</span>
                </label>
                <div class="profile-id-row">
                  <span class="profile-id">{editingProfile?.id ?? "-"}</span>
                  <Button
                    variant="ghost"
                    size="small"
                    label="Change ID"
                    onclick={changeProfileId}
                    disabled={!editingProfile}
                  />
                </div>
                {#if profileIdError}
                  <div class="field-error">{profileIdError}</div>
                {/if}
              </div>
            </div>

            <div class="category-editor">
              <div class="category-editor-header">
                <div>
                  <h4 class="category-editor-title">Categories</h4>
                  <p class="category-editor-subtitle">
                    Tune strategy, cooldown, and candidate weights per category.
                  </p>
                </div>
                <div class="category-actions">
                  <input
                    class="crystal-input category-name-input"
                    type="text"
                    placeholder="category name"
                    value={newCategoryName}
                    oninput={(event) => (newCategoryName = (event.currentTarget as HTMLInputElement).value)}
                    disabled={!editingProfile}
                  />
                  <Button
                    variant="ghost"
                    size="small"
                    label="Add Category"
                    onclick={addCategory}
                    disabled={!newCategoryName.trim() || !editingProfile}
                  />
                </div>
              </div>

              {#if editingCategories.length === 0}
                <div class="empty-state compact">
                  <Code2 size={28} class="empty-icon" />
                  <p class="empty-text">No categories yet. Add your first category.</p>
                </div>
              {:else}
                {#each editingCategories as [categoryName, category]}
                  <div class="category-card">
                    <div class="category-card-header">
                      <div class="category-title">
                        <span class="category-name">{categoryName}</span>
                        <span class="category-badge">{category.strategy ?? "weighted"}</span>
                      </div>
                      <Button
                        variant="ghost"
                        size="small"
                        label="Delete"
                        onclick={() => removeCategory(categoryName)}
                      />
                    </div>

                    <div class="category-fields">
                      <div class="crystal-field">
                        <label class="field-label">
                          <span class="label-text">Strategy</span>
                        </label>
                        <div class="select-wrapper">
                          <select
                            class="crystal-select"
                            value={category.strategy ?? "weighted"}
                            onchange={(event) => handleCategoryStrategyChange(categoryName, event)}
                          >
                            <option value="weighted">weighted</option>
                            <option value="round_robin">round_robin</option>
                          </select>
                          <div class="select-arrow"></div>
                        </div>
                      </div>

                      <div class="crystal-field">
                        <label class="field-label">
                          <span class="label-text">Fallback on Rate Limit</span>
                        </label>
                        <div class="select-wrapper">
                          <select
                            class="crystal-select"
                            value={(category.fallbackOnRateLimit ?? true) ? "true" : "false"}
                            onchange={(event) => handleCategoryFallbackChange(categoryName, event)}
                          >
                            <option value="true">Enabled</option>
                            <option value="false">Disabled</option>
                          </select>
                          <div class="select-arrow"></div>
                        </div>
                      </div>

                      <div class="crystal-field">
                        <label class="field-label">
                          <span class="label-text">Cooldown (sec)</span>
                        </label>
                        <input
                          class="crystal-input"
                          type="number"
                          min="0"
                          value={category.cooldownSec ?? 0}
                          oninput={(event) => handleCategoryCooldownChange(categoryName, event)}
                        />
                      </div>
                    </div>

                    <div class="candidate-block">
                      <div class="candidate-header">
                        <span class="candidate-title">Candidates</span>
                        <Button
                          variant="ghost"
                          size="small"
                          label="Add Candidate"
                          onclick={() => addCandidate(categoryName)}
                        />
                      </div>

                      {#if (category.candidates ?? []).length === 0}
                        <div class="empty-state compact">
                          <p class="empty-text">No candidates yet. Add tool/model pairs.</p>
                        </div>
                      {:else}
                        {#each category.candidates ?? [] as candidate, index}
                          <div class="candidate-row">
                            <div class="crystal-field">
                              <label class="field-label">
                                <span class="label-text">Tool</span>
                              </label>
                              <div class="select-wrapper">
                                <select
                                  class="crystal-select"
                                  value={candidate.tool ?? ""}
                                  onchange={(event) => handleCandidateToolChange(categoryName, index, event)}
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
                              <label class="field-label">
                                <span class="label-text">Model</span>
                              </label>
                              <div class="select-wrapper">
                                <select
                                  class="crystal-select"
                                  value={candidate.model ?? ""}
                                  onchange={(event) => handleCandidateModelChange(categoryName, index, event)}
                                  disabled={!candidate.tool}
                                >
                                  {#each modelsForTool(candidate.tool ?? "") as model}
                                    <option value={model.id}>{model.name}</option>
                                  {/each}
                                </select>
                                <div class="select-arrow"></div>
                              </div>
                            </div>

                            <div class="crystal-field">
                              <label class="field-label">
                                <span class="label-text">Weight</span>
                              </label>
                              <input
                                class="crystal-input"
                                type="number"
                                min="0"
                                value={candidate.weight ?? 0}
                                oninput={(event) => handleCandidateWeightChange(categoryName, index, event)}
                              />
                            </div>

                            <div class="candidate-actions">
                              <Button
                                variant="ghost"
                                size="small"
                                label="Remove"
                                onclick={() => removeCandidate(categoryName, index)}
                              />
                            </div>
                          </div>
                        {/each}
                      {/if}
                    </div>
                  </div>
                {/each}
              {/if}
            </div>
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
              Force mode active: Using <strong>{availableTools.find(t => t.id === forceTool)?.name ?? forceTool}</strong>
              {#if forceModel}
                with <strong>{availableModels.find(m => m.id === forceModel)?.name ?? forceModel}</strong>
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
            Edit categories, weights, candidates, and fallback settings directly.
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
      <div class="status-message" class:success={statusTone === "ok"} class:error={statusTone === "error"}>
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
      <Button variant="crystal" size="small" label="Apply" onclick={applyConfig} />
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
    background: linear-gradient(
      to bottom,
      var(--mv-glass-bg-dark),
      transparent
    );
    border-bottom: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
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
    border-bottom: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
  }

  .header-glow {
    position: absolute;
    top: 0;
    left: 0;
    width: var(--mv-header-glow-width);
    height: var(--mv-space-0-75);
    background: linear-gradient(90deg, var(--mv-primitive-frost-2), transparent);
    border-radius: var(--mv-radius-full);
    box-shadow: var(--mv-shadow-glow-frost-2);
  }

  .header-glow.force {
    background: linear-gradient(90deg, var(--mv-primitive-aurora-yellow), transparent);
    box-shadow: var(--mv-shadow-glow-yellow);
  }

  .header-glow.advanced {
    background: linear-gradient(90deg, var(--mv-primitive-aurora-purple), transparent);
    box-shadow: var(--mv-shadow-glow-purple);
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
    top: var(--mv-position-center);
    transform: translateY(var(--mv-transform-center-x));
    width: var(--mv-space-0);
    height: var(--mv-space-0);
    border-left: var(--mv-arrow-size) solid transparent;
    border-right: var(--mv-arrow-size) solid transparent;
    border-top: var(--mv-arrow-size) solid var(--mv-color-text-muted);
    pointer-events: none;
  }

  /* ========================================
     Segmented Control (ON/OFF Toggle)
     ======================================== */
  .segmented-control {
    position: relative;
    display: inline-grid;
    grid-template-columns: 1fr 1fr;
    padding: var(--mv-space-0-75);
    background: var(--mv-glass-bg-darker);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-light);
    border-radius: var(--mv-radius-lg);
    gap: var(--mv-spacing-xxxs);
    min-width: var(--mv-segmented-min-width);
    max-width: var(--mv-segmented-max-width);
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
    top: var(--mv-space-0-75);
    left: var(--mv-space-0-75);
    width: calc(var(--mv-position-center) - var(--mv-space-3-5));
    height: calc(var(--mv-size-full) - var(--mv-space-1-5));
    background: linear-gradient(
      135deg,
      var(--mv-glow-red),
      var(--mv-glow-red-subtle)
    );
    border: var(--mv-border-width-thin) solid var(--mv-glow-red-strong);
    border-radius: var(--mv-radius-md);
    box-shadow: var(--mv-shadow-segment-indicator-off);
    transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .segment-indicator.right {
    left: calc(var(--mv-position-center) + var(--mv-space-px));
    background: linear-gradient(
      135deg,
      var(--mv-glow-green),
      var(--mv-glow-green-subtle)
    );
    border-color: var(--mv-glow-green-border);
    box-shadow: var(--mv-shadow-segment-indicator-on);
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
    background: linear-gradient(
      90deg,
      var(--mv-glow-green-light),
      transparent
    );
    border: var(--mv-border-width-thin) solid var(--mv-glow-green);
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

  .empty-state.compact {
    padding: var(--mv-spacing-md);
    gap: var(--mv-spacing-xs);
  }

  /* ========================================
     Profile Builder
     ======================================== */
  .profile-editor {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-lg);
    padding: var(--mv-spacing-md);
    background: var(--mv-glass-bg-dark);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
    border-radius: var(--mv-radius-md);
  }

  .profile-editor-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--mv-spacing-md);
    flex-wrap: wrap;
  }

  .profile-editor-title {
    margin: 0 0 var(--mv-spacing-xxs);
    font-size: var(--mv-font-size-md);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
    letter-spacing: var(--mv-letter-spacing-wide);
  }

  .profile-editor-subtitle {
    margin: 0;
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-muted);
  }

  .profile-editor-actions {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
    flex-wrap: wrap;
  }

  .profile-meta-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: var(--mv-spacing-md);
  }

  .crystal-input {
    width: 100%;
    height: var(--mv-input-height-md);
    padding: 0 var(--mv-spacing-md);
    background: var(--mv-glass-bg-dark);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-light);
    border-radius: var(--mv-radius-md);
    color: var(--mv-color-text-primary);
    font-family: var(--mv-font-sans);
    font-size: var(--mv-font-size-sm);
    transition: all 0.2s ease;
  }

  .crystal-input:hover:not(:disabled) {
    border-color: var(--mv-glass-border-hover);
    background: var(--mv-glass-bg-darker);
  }

  .crystal-input:focus {
    outline: none;
    border-color: var(--mv-primitive-frost-2);
    box-shadow: var(--mv-shadow-glow-subtle);
  }

  .crystal-input:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .profile-id-row {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-sm);
    flex-wrap: wrap;
  }

  .profile-id {
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    background: var(--mv-glass-bg-darker);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
    border-radius: var(--mv-radius-sm);
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-muted);
  }

  .field-error {
    margin-top: var(--mv-spacing-xxs);
    font-size: var(--mv-font-size-xs);
    color: var(--mv-primitive-pastel-red);
  }

  .category-editor {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-md);
  }

  .category-editor-header {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: var(--mv-spacing-md);
    flex-wrap: wrap;
  }

  .category-editor-title {
    margin: 0 0 var(--mv-spacing-xxs);
    font-size: var(--mv-font-size-md);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
  }

  .category-editor-subtitle {
    margin: 0;
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-muted);
  }

  .category-actions {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
    flex-wrap: wrap;
  }

  .category-name-input {
    min-width: var(--mv-empty-state-width);
  }

  .category-card {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-md);
    padding: var(--mv-spacing-md);
    background: var(--mv-glass-bg-darker);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-light);
    border-radius: var(--mv-radius-md);
  }

  .category-card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--mv-spacing-sm);
    flex-wrap: wrap;
  }

  .category-title {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
  }

  .category-name {
    font-size: var(--mv-font-size-md);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
  }

  .category-badge {
    padding: var(--mv-spacing-xxxs) var(--mv-spacing-xs);
    border-radius: var(--mv-radius-full);
    background: var(--mv-glass-bg-dark);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
    font-size: var(--mv-font-size-xxs);
    color: var(--mv-color-text-muted);
    letter-spacing: var(--mv-letter-spacing-wide);
  }

  .category-fields {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: var(--mv-spacing-md);
  }

  .candidate-block {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-sm);
  }

  .candidate-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--mv-spacing-md);
    flex-wrap: wrap;
  }

  .candidate-title {
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-secondary);
    letter-spacing: var(--mv-letter-spacing-wide);
  }

  .candidate-row {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
    gap: var(--mv-spacing-sm);
    align-items: end;
    padding: var(--mv-spacing-sm);
    background: var(--mv-glass-bg-dark);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
    border-radius: var(--mv-radius-md);
  }

  .candidate-actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    padding-bottom: var(--mv-spacing-xxs);
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
    border-bottom: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
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
    width: var(--mv-icon-size-lg);
    height: var(--mv-icon-size-lg);
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
    width: var(--mv-size-full);
    min-height: var(--mv-textarea-json-min-height);
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
    background: var(--mv-glow-red-light);
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
