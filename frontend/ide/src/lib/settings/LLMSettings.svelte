<script lang="ts">
  import { onMount } from "svelte";
  import Button from "../../design-system/components/Button.svelte";
  import Input from "../../design-system/components/Input.svelte";
  import { toasts } from "../../stores/toastStore";
  import {
    GetLLMConfig,
    SetLLMConfig,
    TestLLMConnection,
  } from "../../../wailsjs/go/main/App";

  interface LLMConfig {
    kind: string;
    model: string;
    baseUrl: string;
    systemPrompt: string;
    hasApiKey: boolean;
  }

  let config: LLMConfig = {
    kind: "codex-cli",
    model: "gpt-4o",
    baseUrl: "",
    systemPrompt: "",
    hasApiKey: false,
  };

  let testing = false;
  let saving = false;
  let loading = true;

  onMount(async () => {
    await loadConfig();
  });

  async function loadConfig() {
    try {
      loading = true;
      config = await GetLLMConfig();
    } catch (e) {
      console.error("Failed to load LLM config:", e);
      toasts.add(`設定の読み込みに失敗: ${e}`, "error");
    } finally {
      loading = false;
    }
  }

  async function saveConfig() {
    try {
      saving = true;
      await SetLLMConfig(config);
      toasts.add("設定を保存しました", "success");
    } catch (e) {
      console.error("Failed to save LLM config:", e);
      toasts.add(`保存に失敗: ${e}`, "error");
    } finally {
      saving = false;
    }
  }

  async function testConnection() {
    try {
      testing = true;
      const result = await TestLLMConnection();
      toasts.add(result, "success");
    } catch (e) {
      console.error("LLM connection test failed:", e);
      toasts.add(`${e}`, "error");
    } finally {
      testing = false;
    }
  }
</script>

<div class="llm-settings glass-panel">
  <h3>LLM 設定</h3>

  {#if loading}
    <p class="loading">読み込み中...</p>
  {:else}
    <div class="form-group">
      <label for="provider">プロバイダ</label>
      <select id="provider" bind:value={config.kind}>
        <option value="mock">モック（開発用）</option>
        <option value="codex-cli">Codex CLI</option>
        <option value="openai-chat">OpenAI（HTTP）</option>
      </select>
    </div>

    {#if config.kind === "codex-cli"}
      <div class="form-group">
        <div class="cli-session-status">
          <span class="label">CLI セッション</span>
          <span class="status info"
            >Codex CLI のセッションが利用されます（~/.codex/auth.json または環境変数）</span
          >
        </div>
      </div>

      <div class="form-group">
        <Input
          label="モデル（オプション）"
          bind:value={config.model}
          placeholder="codex-cli では使用されません"
          id="model"
          disabled={true}
        />
      </div>
    {:else if config.kind === "openai-chat"}
      <div class="form-group">
        <div class="api-key-status">
          <span class="label">API キー</span>
          {#if config.hasApiKey}
            <span class="status success">✓ 設定済み（環境変数）</span>
          {:else}
            <span class="status warning"
              >未設定（OPENAI_API_KEY を設定してください）</span
            >
          {/if}
        </div>
      </div>

      <div class="form-group">
        <Input
          label="モデル"
          bind:value={config.model}
          placeholder="gpt-4o"
          id="model"
        />
      </div>

      <div class="form-group">
        <Input
          label="カスタムエンドポイント（オプション）"
          bind:value={config.baseUrl}
          placeholder="https://api.openai.com/v1"
          id="baseUrl"
        />
      </div>
    {/if}

    <div class="actions">
      <Button variant="primary" on:click={saveConfig} disabled={saving}>
        {saving ? "保存中..." : "保存"}
      </Button>
      <Button variant="secondary" on:click={testConnection} disabled={testing}>
        {testing ? "テスト中..." : "接続テスト"}
      </Button>
    </div>
  {/if}
</div>

<style>
  .llm-settings {
    padding: var(--mv-spacing-lg);
    max-width: var(--mv-content-max-width-sm);
  }

  h3 {
    margin: 0 0 var(--mv-spacing-lg);
    font-size: var(--mv-font-size-lg);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
  }

  .loading {
    color: var(--mv-color-text-muted);
  }

  .form-group {
    margin-bottom: var(--mv-spacing-md);
  }

  label {
    display: block;
    margin-bottom: var(--mv-spacing-xs);
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-secondary);
    font-weight: var(--mv-font-weight-medium);
  }

  select {
    width: 100%;
    height: var(--mv-input-height-md);
    padding: 0 var(--mv-spacing-md);
    font-family: var(--mv-font-sans);
    font-size: var(--mv-font-size-md);
    color: var(--mv-color-text-primary);
    background: var(--mv-color-surface-secondary);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    border-radius: var(--mv-radius-md);
    cursor: pointer;
    transition: var(--mv-transition-base);
  }

  select:hover {
    border-color: var(--mv-color-border-strong);
    background: var(--mv-color-surface-hover);
  }

  select:focus {
    outline: none;
    border-color: var(--mv-color-interactive-primary);
    box-shadow: var(--mv-shadow-focus);
  }

  .api-key-status {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-xs);
  }

  .api-key-status .label {
    margin: 0;
  }

  .status {
    font-size: var(--mv-font-size-sm);
    padding: var(--mv-spacing-xs) var(--mv-spacing-sm);
    border-radius: var(--mv-radius-sm);
  }

  .status.success {
    color: var(--mv-color-status-success-text);
    background: var(--mv-color-status-success-bg);
  }

  .status.warning {
    color: var(--mv-color-status-paused-text);
    background: var(--mv-color-status-paused-bg);
  }

  .status.info {
    color: var(--mv-color-text-secondary);
    background: var(--mv-color-surface-secondary);
  }

  .cli-session-status {
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-xs);
  }

  .cli-session-status .label {
    margin: 0;
  }

  .actions {
    display: flex;
    gap: var(--mv-spacing-sm);
    margin-top: var(--mv-spacing-lg);
  }

  .glass-panel {
    background: var(--mv-glass-bg);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
    border-radius: var(--mv-radius-lg);
    backdrop-filter: var(--mv-glass-blur);
  }
</style>
