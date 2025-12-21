<script lang="ts">
  import Button from "../../design-system/components/Button.svelte";
  import {
    Cpu,
    Zap,
    Code,
    Check,
    CircleAlert,
    RefreshCw,
    Power,
    PowerOff,
    ChevronDown,
    ChevronRight,
    Plus,
    Trash2,
    Copy,
    Pencil,
    X,
  } from "lucide-svelte";

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
    { id: "advanced", label: "Advanced", icon: Code },
  ] as const;
  type TabId = (typeof tabs)[number]["id"];

  // 初期値を計算
  function initializeFromConfig(config: ToolingConfig) {
    const cfg = config ?? {};
    return {
      rawJSON: JSON.stringify(cfg, null, 2),
      activeProfile: cfg.activeProfile ?? "",
      profiles: cfg.profiles ?? [],
      forceEnabled: cfg.force?.enabled ?? false,
      forceTool: cfg.force?.tool ?? "",
      forceModel: cfg.force?.model ?? "",
      selectedProfileId: cfg.activeProfile || cfg.profiles?.[0]?.id || "",
    };
  }

  const init = initializeFromConfig(initialConfig);

  let activeTab = $state<TabId>(initialConfig?.force?.enabled ? "force" : "profiles");
  let rawJSON = $state(init.rawJSON);
  let parseError = $state("");
  let statusMessage = $state("");
  let statusTone = $state<"ok" | "error" | "">("");
  let activeProfile = $state(init.activeProfile);
  let profiles = $state<ToolingProfile[]>(init.profiles);
  let forceEnabled = $state(init.forceEnabled);
  let forceTool = $state(init.forceTool);
  let forceModel = $state(init.forceModel);
  let selectedProfileId = $state(init.selectedProfileId);
  let newCategoryName = $state("");
  let expandedCategories = $state<Set<string>>(new Set());
  let editingProfileName = $state(false);
  let editingProfileId = $state(false);
  let tempProfileName = $state("");
  let tempProfileId = $state("");
  let profileIdError = $state("");

  const selectedProfile = $derived.by(() => profiles.find(p => (p.id ?? "") === selectedProfileId));
  const selectedCategories = $derived.by(() => {
    if (!selectedProfile?.categories) {
      return [] as Array<[string, ToolCategoryConfig]>;
    }
    return Object.entries(selectedProfile.categories).sort(([a], [b]) => a.localeCompare(b));
  });
  const isActiveProfile = $derived.by(() => selectedProfileId === activeProfile);

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

    // 選択プロファイルの維持または初期選択
    const nextSelectedId = profiles.some(p => (p.id ?? "") === selectedProfileId)
      ? selectedProfileId
      : (activeProfile || profiles[0]?.id || "");
    selectedProfileId = nextSelectedId;
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

  function selectProfile(profileId: string) {
    selectedProfileId = profileId;
    editingProfileName = false;
    editingProfileId = false;
  }

  function setActiveProfile(profileId: string) {
    updateConfig((cfg) => {
      cfg.activeProfile = profileId;
    });
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

  function createNewProfile() {
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
      newID = identity.id;
    });
    if (newID) {
      selectedProfileId = newID;
    }
  }

  function duplicateProfile() {
    if (!selectedProfile) return;
    let newID = "";
    updateConfig((cfg) => {
      ensureProfiles(cfg);
      const base = cfg.profiles?.find(profile => profile.id === selectedProfile.id);
      if (!base) return;
      const clone = cloneProfile(base);
      const identity = nextCopyIdentity(base, cfg.profiles ?? []);
      clone.id = identity.id;
      clone.name = identity.name;
      cfg.profiles?.push(clone);
      newID = identity.id;
    });
    if (newID) {
      selectedProfileId = newID;
    }
  }

  function deleteProfile() {
    if (!selectedProfile?.id || profiles.length <= 1) return;
    const deleteID = selectedProfile.id;
    updateConfig((cfg) => {
      ensureProfiles(cfg);
      cfg.profiles = (cfg.profiles ?? []).filter(profile => (profile.id ?? "") !== deleteID);
      if (cfg.activeProfile === deleteID) {
        cfg.activeProfile = cfg.profiles[0]?.id ?? "";
      }
    });
  }

  function startEditProfileName() {
    if (!selectedProfile) return;
    tempProfileName = selectedProfile.name ?? "";
    editingProfileName = true;
  }

  function saveProfileName() {
    if (!selectedProfile?.id) return;
    updateConfig((cfg) => {
      const profile = cfg.profiles?.find(p => (p.id ?? "") === selectedProfile.id);
      if (profile) {
        profile.name = tempProfileName;
      }
    });
    editingProfileName = false;
  }

  function cancelEditProfileName() {
    editingProfileName = false;
  }

  function startEditProfileId() {
    if (!selectedProfile) return;
    tempProfileId = selectedProfile.id ?? "";
    editingProfileId = true;
    profileIdError = "";
  }

  function saveProfileId() {
    if (!selectedProfile?.id) return;
    const currentID = selectedProfile.id;
    const nextID = tempProfileId.trim();

    if (!nextID || nextID === currentID) {
      editingProfileId = false;
      return;
    }

    if (profiles.some(profile => (profile.id ?? "") === nextID)) {
      profileIdError = "ID already exists";
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
    selectedProfileId = nextID;
    editingProfileId = false;
  }

  function cancelEditProfileId() {
    editingProfileId = false;
    profileIdError = "";
  }

  function toggleCategory(categoryName: string) {
    const next = new Set(expandedCategories);
    if (next.has(categoryName)) {
      next.delete(categoryName);
    } else {
      next.add(categoryName);
    }
    expandedCategories = next;
  }

  function addCategory() {
    if (!selectedProfile?.id) return;
    const name = newCategoryName.trim();
    if (!name) return;
    updateConfig((cfg) => {
      const profile = cfg.profiles?.find(p => (p.id ?? "") === selectedProfile.id);
      if (!profile) return;
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
    expandedCategories = new Set([...expandedCategories, name]);
    newCategoryName = "";
  }

  function removeCategory(categoryName: string) {
    if (!selectedProfile?.id) return;
    updateConfig((cfg) => {
      const profile = cfg.profiles?.find(p => (p.id ?? "") === selectedProfile.id);
      if (!profile?.categories) return;
      delete profile.categories[categoryName];
    });
  }

  function updateCategory(categoryName: string, update: (category: ToolCategoryConfig) => void) {
    if (!selectedProfile?.id) return;
    updateConfig((cfg) => {
      const profile = cfg.profiles?.find(p => (p.id ?? "") === selectedProfile.id);
      if (!profile) return;
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

  function updateCandidate(categoryName: string, index: number, field: keyof ToolCandidate, value: string | number) {
    updateCategory(categoryName, (category) => {
      const candidates = category.candidates ?? [];
      if (!candidates[index]) {
        candidates[index] = { tool: "", model: "", weight: 0 };
      }
      if (field === "weight") {
        const numValue = Number(value);
        candidates[index][field] = Number.isFinite(numValue) ? Math.max(0, numValue) : 0;
      } else {
        candidates[index][field] = value as string;
      }
      category.candidates = candidates;
    });
  }

  function modelsForTool(_toolID: string): typeof availableModels {
    return availableModels;
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
      <!-- Profiles タブ - 2カラムレイアウト -->
      <div class="profiles-layout">
        <!-- 左サイドバー: プロファイルリスト -->
        <aside class="profile-sidebar">
          <div class="sidebar-header">
            <h3 class="sidebar-title">Profiles</h3>
            <button class="icon-btn" onclick={createNewProfile} title="Create new profile" type="button">
              <Plus size={16} />
            </button>
          </div>

          <div class="profile-list">
            {#each profiles as profile}
              <button
                class="profile-item"
                class:selected={selectedProfileId === profile.id}
                class:active={activeProfile === profile.id}
                onclick={() => selectProfile(profile.id ?? "")}
                type="button"
              >
                <div class="profile-item-content">
                  <span class="profile-item-name">{profile.name ?? profile.id ?? "Unnamed"}</span>
                  {#if activeProfile === profile.id}
                    <span class="active-badge">Active</span>
                  {/if}
                </div>
                <span class="profile-item-id">{profile.id}</span>
              </button>
            {/each}

            {#if profiles.length === 0}
              <div class="empty-sidebar">
                <p>No profiles yet</p>
                <Button
                  variant="ghost"
                  size="small"
                  label="Create Profile"
                  onclick={createNewProfile}
                />
              </div>
            {/if}
          </div>
        </aside>

        <!-- 右メイン: プロファイル詳細 -->
        <main class="profile-detail">
          {#if selectedProfile}
            <!-- プロファイルヘッダー -->
            <div class="detail-header">
              <div class="profile-info">
                {#if editingProfileName}
                  <div class="inline-edit">
                    <input
                      class="inline-input large"
                      type="text"
                      bind:value={tempProfileName}
                      onkeydown={(e) => e.key === "Enter" && saveProfileName()}
                    />
                    <button class="icon-btn small success" onclick={saveProfileName} type="button">
                      <Check size={14} />
                    </button>
                    <button class="icon-btn small" onclick={cancelEditProfileName} type="button">
                      <X size={14} />
                    </button>
                  </div>
                {:else}
                  <button class="profile-name-btn" type="button" onclick={startEditProfileName}>
                    <span class="profile-name-text">{selectedProfile.name ?? "Unnamed"}</span>
                    <Pencil size={14} class="edit-icon" />
                  </button>
                {/if}

                {#if editingProfileId}
                  <div class="inline-edit">
                    <input
                      class="inline-input mono"
                      type="text"
                      bind:value={tempProfileId}
                      onkeydown={(e) => e.key === "Enter" && saveProfileId()}
                    />
                    <button class="icon-btn small success" onclick={saveProfileId} type="button">
                      <Check size={14} />
                    </button>
                    <button class="icon-btn small" onclick={cancelEditProfileId} type="button">
                      <X size={14} />
                    </button>
                  </div>
                  {#if profileIdError}
                    <span class="error-text">{profileIdError}</span>
                  {/if}
                {:else}
                  <button class="profile-id-btn" type="button" onclick={startEditProfileId}>
                    <span>ID: {selectedProfile.id}</span>
                    <Pencil size={12} class="edit-icon" />
                  </button>
                {/if}
              </div>

              <div class="header-actions">
                {#if !isActiveProfile}
                  <Button
                    variant="crystal"
                    size="small"
                    label="Set Active"
                    onclick={() => setActiveProfile(selectedProfileId)}
                  />
                {/if}
                <button class="icon-btn" onclick={duplicateProfile} title="Duplicate" type="button">
                  <Copy size={16} />
                </button>
                <button
                  class="icon-btn danger"
                  onclick={deleteProfile}
                  title="Delete"
                  disabled={profiles.length <= 1}
                  type="button"
                >
                  <Trash2 size={16} />
                </button>
              </div>
            </div>

            <!-- カテゴリセクション -->
            <div class="categories-section">
              <div class="section-header-row">
                <h3 class="section-title">Categories</h3>
                <div class="add-category-row">
                  <input
                    class="add-category-input"
                    type="text"
                    placeholder="New category name..."
                    bind:value={newCategoryName}
                    onkeydown={(e) => e.key === "Enter" && addCategory()}
                  />
                  <button
                    class="icon-btn"
                    onclick={addCategory}
                    disabled={!newCategoryName.trim()}
                    title="Add category"
                    type="button"
                  >
                    <Plus size={16} />
                  </button>
                </div>
              </div>

              {#if selectedCategories.length === 0}
                <div class="empty-state compact">
                  <Code size={24} class="empty-icon" />
                  <p class="empty-text">No categories. Add one to configure tool priorities.</p>
                </div>
              {:else}
                <div class="category-list">
                  {#each selectedCategories as [categoryName, category]}
                    <div class="category-accordion" class:expanded={expandedCategories.has(categoryName)}>
                      <!-- アコーディオンヘッダー行（ボタン + 削除ボタン） -->
                      <div class="accordion-header-row">
                        <button
                          class="accordion-header"
                          onclick={() => toggleCategory(categoryName)}
                          type="button"
                        >
                          <div class="accordion-left">
                            {#if expandedCategories.has(categoryName)}
                              <ChevronDown size={16} />
                            {:else}
                              <ChevronRight size={16} />
                            {/if}
                            <span class="category-name">{categoryName}</span>
                            <span class="category-meta">
                              {category.strategy ?? "weighted"} · {(category.candidates ?? []).length} candidates
                            </span>
                          </div>
                        </button>
                        <button
                          class="icon-btn small danger accordion-delete"
                          onclick={() => removeCategory(categoryName)}
                          title="Delete category"
                          type="button"
                        >
                          <Trash2 size={14} />
                        </button>
                      </div>

                      <!-- アコーディオンコンテンツ -->
                      {#if expandedCategories.has(categoryName)}
                        <div class="accordion-content">
                          <!-- カテゴリ設定 -->
                          <div class="category-settings">
                            <label class="setting-group">
                              <span class="setting-label">Strategy</span>
                              <select
                                class="setting-select"
                                value={category.strategy ?? "weighted"}
                                onchange={(e) => updateCategory(categoryName, (c) => { c.strategy = (e.currentTarget as HTMLSelectElement).value; })}
                              >
                                <option value="weighted">Weighted</option>
                                <option value="round_robin">Round Robin</option>
                              </select>
                            </label>
                            <label class="setting-group">
                              <span class="setting-label">Rate Limit Fallback</span>
                              <select
                                class="setting-select"
                                value={(category.fallbackOnRateLimit ?? true) ? "true" : "false"}
                                onchange={(e) => updateCategory(categoryName, (c) => { c.fallbackOnRateLimit = (e.currentTarget as HTMLSelectElement).value === "true"; })}
                              >
                                <option value="true">Enabled</option>
                                <option value="false">Disabled</option>
                              </select>
                            </label>
                            <label class="setting-group">
                              <span class="setting-label">Cooldown</span>
                              <div class="setting-input-group">
                                <input
                                  class="setting-input"
                                  type="number"
                                  min="0"
                                  value={category.cooldownSec ?? 120}
                                  oninput={(e) => updateCategory(categoryName, (c) => { c.cooldownSec = Number((e.currentTarget as HTMLInputElement).value) || 0; })}
                                />
                                <span class="setting-suffix">sec</span>
                              </div>
                            </label>
                          </div>

                          <!-- Candidates テーブル -->
                          <div class="candidates-section">
                            <div class="candidates-header">
                              <span class="candidates-title">Candidates</span>
                              <button
                                class="icon-btn small"
                                onclick={() => addCandidate(categoryName)}
                                title="Add candidate"
                                type="button"
                              >
                                <Plus size={14} />
                              </button>
                            </div>

                            {#if (category.candidates ?? []).length === 0}
                              <p class="no-candidates">No candidates. Add tool/model pairs.</p>
                            {:else}
                              <table class="candidates-table">
                                <thead>
                                  <tr>
                                    <th>Tool</th>
                                    <th>Model</th>
                                    <th>Weight</th>
                                    <th></th>
                                  </tr>
                                </thead>
                                <tbody>
                                  {#each category.candidates ?? [] as candidate, index}
                                    <tr>
                                      <td>
                                        <select
                                          class="table-select"
                                          value={candidate.tool ?? ""}
                                          onchange={(e) => updateCandidate(categoryName, index, "tool", (e.currentTarget as HTMLSelectElement).value)}
                                        >
                                          <option value="">Select...</option>
                                          {#each availableTools as tool}
                                            <option value={tool.id}>{tool.name}</option>
                                          {/each}
                                        </select>
                                      </td>
                                      <td>
                                        <select
                                          class="table-select"
                                          value={candidate.model ?? ""}
                                          onchange={(e) => updateCandidate(categoryName, index, "model", (e.currentTarget as HTMLSelectElement).value)}
                                          disabled={!candidate.tool}
                                        >
                                          {#each modelsForTool(candidate.tool ?? "") as model}
                                            <option value={model.id}>{model.name}</option>
                                          {/each}
                                        </select>
                                      </td>
                                      <td>
                                        <input
                                          class="table-input"
                                          type="number"
                                          min="0"
                                          value={candidate.weight ?? 0}
                                          oninput={(e) => updateCandidate(categoryName, index, "weight", Number((e.currentTarget as HTMLInputElement).value))}
                                        />
                                      </td>
                                      <td>
                                        <button
                                          class="icon-btn tiny danger"
                                          onclick={() => removeCandidate(categoryName, index)}
                                          title="Remove"
                                          type="button"
                                        >
                                          <X size={12} />
                                        </button>
                                      </td>
                                    </tr>
                                  {/each}
                                </tbody>
                              </table>
                            {/if}
                          </div>
                        </div>
                      {/if}
                    </div>
                  {/each}
                </div>
              {/if}
            </div>
          {:else}
            <div class="empty-detail">
              <Cpu size={48} class="empty-icon" />
              <h3>No profile selected</h3>
              <p>Select a profile from the list or create a new one.</p>
              <Button
                variant="crystal"
                size="small"
                label="Create Profile"
                onclick={createNewProfile}
              />
            </div>
          {/if}
        </main>
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
              Force mode active: Using <strong>{availableTools.find((t) => t.id === forceTool)?.name ?? forceTool}</strong>
              {#if forceModel}
                with <strong>{availableModels.find((m) => m.id === forceModel)?.name ?? forceModel}</strong>
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
            Edit the raw JSON configuration directly.
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
              <CircleAlert size={14} />
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
          <CircleAlert size={14} />
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
    background: linear-gradient(to bottom, var(--mv-glass-bg-dark), transparent);
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
    overflow: hidden;
  }

  .content-section {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-lg);
    padding: var(--mv-spacing-lg);
    height: 100%;
    overflow-y: auto;
  }

  /* ========================================
     2カラムレイアウト（Profiles タブ）
     ======================================== */
  .profiles-layout {
    display: grid;
    grid-template-columns: 240px 1fr;
    height: 100%;
  }

  /* 左サイドバー */
  .profile-sidebar {
    display: flex;
    flex-direction: column;
    border-right: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
    background: var(--mv-glass-bg-dark);
  }

  .sidebar-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--mv-spacing-md);
    border-bottom: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
  }

  .sidebar-title {
    margin: 0;
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-secondary);
    letter-spacing: var(--mv-letter-spacing-wide);
    text-transform: uppercase;
  }

  .profile-list {
    flex: 1;
    overflow-y: auto;
    padding: var(--mv-spacing-xs);
  }

  .profile-item {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    width: 100%;
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
    margin-bottom: var(--mv-spacing-xxs);
    background: transparent;
    border: var(--mv-border-width-thin) solid transparent;
    border-radius: var(--mv-radius-md);
    cursor: pointer;
    text-align: left;
    transition: all 0.15s ease;
  }

  .profile-item:hover {
    background: var(--mv-glass-bg-darker);
    border-color: var(--mv-glass-border-light);
  }

  .profile-item.selected {
    background: var(--mv-glass-active);
    border-color: var(--mv-primitive-frost-2);
  }

  .profile-item-content {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
    width: 100%;
  }

  .profile-item-name {
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-medium);
    color: var(--mv-color-text-primary);
  }

  .profile-item-id {
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-xxs);
    color: var(--mv-color-text-muted);
  }

  .active-badge {
    padding: 2px 6px;
    background: var(--mv-primitive-frost-2);
    border-radius: var(--mv-radius-full);
    font-size: 10px;
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-primitive-polar-night-0);
    letter-spacing: 0.5px;
    text-transform: uppercase;
  }

  .empty-sidebar {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--mv-spacing-sm);
    padding: var(--mv-spacing-lg);
    text-align: center;
    color: var(--mv-color-text-muted);
    font-size: var(--mv-font-size-sm);
  }

  /* 右メイン詳細 */
  .profile-detail {
    display: flex;
    flex-direction: column;
    padding: var(--mv-spacing-lg);
    overflow-y: auto;
  }

  .detail-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--mv-spacing-md);
    margin-bottom: var(--mv-spacing-lg);
    padding-bottom: var(--mv-spacing-md);
    border-bottom: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
  }

  .profile-info {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-xs);
  }

  .profile-name-btn {
    display: inline-flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
    padding: 0;
    background: none;
    border: none;
    cursor: pointer;
  }

  .profile-name-btn:hover :global(.edit-icon) {
    opacity: 1;
  }

  .profile-name-text {
    font-size: var(--mv-font-size-xl);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
  }

  :global(.edit-icon) {
    opacity: 0.3;
    color: var(--mv-color-text-muted);
    transition: opacity 0.15s ease;
  }

  .profile-id-btn {
    display: inline-flex;
    align-items: center;
    gap: var(--mv-spacing-xxs);
    padding: 0;
    background: none;
    border: none;
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-muted);
    cursor: pointer;
  }

  .profile-id-btn:hover :global(.edit-icon) {
    opacity: 1;
  }

  .inline-edit {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xxs);
  }

  .inline-input {
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    background: var(--mv-glass-bg-darker);
    border: var(--mv-border-width-thin) solid var(--mv-primitive-frost-2);
    border-radius: var(--mv-radius-sm);
    color: var(--mv-color-text-primary);
    font-size: var(--mv-font-size-sm);
  }

  .inline-input.large {
    font-size: var(--mv-font-size-lg);
    font-weight: var(--mv-font-weight-semibold);
  }

  .inline-input.mono {
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-xs);
  }

  .inline-input:focus {
    outline: none;
    box-shadow: 0 0 0 2px var(--mv-primitive-frost-2);
  }

  .error-text {
    font-size: var(--mv-font-size-xs);
    color: var(--mv-primitive-pastel-red);
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
  }

  /* アイコンボタン */
  .icon-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    background: var(--mv-glass-bg-dark);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-light);
    border-radius: var(--mv-radius-md);
    color: var(--mv-color-text-muted);
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .icon-btn:hover:not(:disabled) {
    background: var(--mv-glass-active);
    color: var(--mv-color-text-primary);
    border-color: var(--mv-glass-border-hover);
  }

  .icon-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .icon-btn.small {
    width: 24px;
    height: 24px;
  }

  .icon-btn.tiny {
    width: 20px;
    height: 20px;
    border: none;
    background: transparent;
  }

  .icon-btn.danger:hover:not(:disabled) {
    color: var(--mv-primitive-pastel-red);
    border-color: var(--mv-primitive-aurora-red);
  }

  .icon-btn.success:hover:not(:disabled) {
    color: var(--mv-primitive-aurora-green);
    border-color: var(--mv-primitive-aurora-green);
  }

  /* カテゴリセクション */
  .categories-section {
    flex: 1;
  }

  .section-header-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--mv-spacing-md);
    margin-bottom: var(--mv-spacing-md);
  }

  .section-title {
    margin: 0;
    font-size: var(--mv-font-size-md);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
  }

  .add-category-row {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xxs);
  }

  .add-category-input {
    width: 180px;
    padding: var(--mv-spacing-xs) var(--mv-spacing-sm);
    background: var(--mv-glass-bg-dark);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-light);
    border-radius: var(--mv-radius-md);
    color: var(--mv-color-text-primary);
    font-size: var(--mv-font-size-sm);
  }

  .add-category-input:focus {
    outline: none;
    border-color: var(--mv-primitive-frost-2);
  }

  .add-category-input::placeholder {
    color: var(--mv-color-text-muted);
  }

  /* アコーディオン */
  .category-list {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-xs);
  }

  .category-accordion {
    background: var(--mv-glass-bg-dark);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-light);
    border-radius: var(--mv-radius-md);
    overflow: hidden;
  }

  .category-accordion.expanded {
    border-color: var(--mv-primitive-frost-2);
  }

  .accordion-header-row {
    display: flex;
    align-items: center;
  }

  .accordion-header {
    display: flex;
    align-items: center;
    flex: 1;
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
    background: transparent;
    border: none;
    cursor: pointer;
    transition: background 0.15s ease;
  }

  .accordion-header:hover {
    background: var(--mv-glass-bg-darker);
  }

  .accordion-delete {
    margin-right: var(--mv-spacing-sm);
    flex-shrink: 0;
  }

  .accordion-left {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-sm);
    color: var(--mv-color-text-primary);
  }

  .category-name {
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-semibold);
  }

  .category-meta {
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-muted);
  }

  .accordion-content {
    padding: var(--mv-spacing-md);
    border-top: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
    background: var(--mv-glass-bg-darker);
  }

  /* カテゴリ設定 */
  .category-settings {
    display: flex;
    gap: var(--mv-spacing-lg);
    margin-bottom: var(--mv-spacing-md);
    padding-bottom: var(--mv-spacing-md);
    border-bottom: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
  }

  .setting-group {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-xxs);
    cursor: pointer;
  }

  .setting-label {
    display: block;
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-medium);
    color: var(--mv-color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .setting-select {
    padding: var(--mv-spacing-xxs) var(--mv-spacing-sm);
    background: var(--mv-glass-bg-dark);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-light);
    border-radius: var(--mv-radius-sm);
    color: var(--mv-color-text-primary);
    font-size: var(--mv-font-size-sm);
    cursor: pointer;
  }

  .setting-select:focus {
    outline: none;
    border-color: var(--mv-primitive-frost-2);
  }

  .setting-input-group {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xxs);
  }

  .setting-input {
    width: 80px;
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    background: var(--mv-glass-bg-dark);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-light);
    border-radius: var(--mv-radius-sm);
    color: var(--mv-color-text-primary);
    font-size: var(--mv-font-size-sm);
    text-align: right;
  }

  .setting-input:focus {
    outline: none;
    border-color: var(--mv-primitive-frost-2);
  }

  .setting-suffix {
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-text-muted);
  }

  /* Candidates テーブル */
  .candidates-section {
    margin-top: var(--mv-spacing-sm);
  }

  .candidates-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--mv-spacing-sm);
  }

  .candidates-title {
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .no-candidates {
    margin: 0;
    padding: var(--mv-spacing-sm);
    text-align: center;
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-muted);
  }

  .candidates-table {
    width: 100%;
    border-collapse: collapse;
  }

  .candidates-table th {
    padding: var(--mv-spacing-xs) var(--mv-spacing-sm);
    text-align: left;
    font-size: var(--mv-font-size-xxs);
    font-weight: var(--mv-font-weight-medium);
    color: var(--mv-color-text-muted);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    border-bottom: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
  }

  .candidates-table th:last-child {
    width: 32px;
  }

  .candidates-table td {
    padding: var(--mv-spacing-xs) var(--mv-spacing-sm);
    vertical-align: middle;
  }

  .candidates-table tr:not(:last-child) td {
    border-bottom: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
  }

  .table-select {
    width: 100%;
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    background: var(--mv-glass-bg-dark);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-light);
    border-radius: var(--mv-radius-sm);
    color: var(--mv-color-text-primary);
    font-size: var(--mv-font-size-sm);
    cursor: pointer;
  }

  .table-select:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .table-select:focus {
    outline: none;
    border-color: var(--mv-primitive-frost-2);
  }

  .table-input {
    width: 60px;
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    background: var(--mv-glass-bg-dark);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-light);
    border-radius: var(--mv-radius-sm);
    color: var(--mv-color-text-primary);
    font-size: var(--mv-font-size-sm);
    text-align: center;
  }

  .table-input:focus {
    outline: none;
    border-color: var(--mv-primitive-frost-2);
  }

  /* Empty States */
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

  .empty-state.compact {
    padding: var(--mv-spacing-lg);
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

  .empty-detail {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--mv-spacing-md);
    height: 100%;
    text-align: center;
    color: var(--mv-color-text-muted);
  }

  .empty-detail h3 {
    margin: 0;
    font-size: var(--mv-font-size-lg);
    color: var(--mv-color-text-secondary);
  }

  .empty-detail p {
    margin: 0;
    font-size: var(--mv-font-size-sm);
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
    background: linear-gradient(135deg, var(--mv-glow-red), var(--mv-glow-red-subtle));
    border: var(--mv-border-width-thin) solid var(--mv-glow-red-strong);
    border-radius: var(--mv-radius-md);
    box-shadow: var(--mv-shadow-segment-indicator-off);
    transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .segment-indicator.right {
    left: calc(var(--mv-position-center) + var(--mv-space-px));
    background: linear-gradient(135deg, var(--mv-glow-green), var(--mv-glow-green-subtle));
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
    background: linear-gradient(90deg, var(--mv-glow-green-light), transparent);
    border: var(--mv-border-width-thin) solid var(--mv-glow-green);
    border-radius: var(--mv-radius-md);
    color: var(--mv-primitive-aurora-green);
    font-size: var(--mv-font-size-sm);
  }

  .force-status strong {
    color: var(--mv-color-text-primary);
  }

  /* ========================================
     JSON Editor
     ======================================== */
  .json-editor-wrapper {
    display: flex;
    flex-direction: column;
    flex: 1;
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
    flex: 1;
    width: 100%;
    min-height: 300px;
    padding: var(--mv-spacing-md);
    background: transparent;
    border: none;
    color: var(--mv-primitive-frost-1);
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-xs);
    line-height: var(--mv-line-height-relaxed);
    resize: none;
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
