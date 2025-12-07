import type { Meta, StoryObj } from "@storybook/svelte";
import { fn } from "@storybook/test";
import LLMSettingsPreview from "./LLMSettingsPreview.svelte";

const meta = {
  title: "IDE/Settings/LLMSettings",
  component: LLMSettingsPreview,
  parameters: {
    layout: "centered",
  },
  tags: ["autodocs"],
  argTypes: {
    loading: {
      control: "boolean",
      description: "ローディング状態",
    },
    saving: {
      control: "boolean",
      description: "保存中状態",
    },
    testing: {
      control: "boolean",
      description: "接続テスト中状態",
    },
  },
  args: {
    onSave: fn(),
    onTest: fn(),
  },
} as Meta<typeof LLMSettingsPreview>;

export default meta;
type Story = StoryObj<typeof meta>;

/**
 * デフォルト状態 - Codex CLI プロバイダ
 */
export const Default: Story = {
  args: {
    initialConfig: {
      kind: "codex-cli",
      model: "gpt-4o",
      baseUrl: "",
      systemPrompt: "",
      hasApiKey: false,
    },
  },
};

/**
 * OpenAI プロバイダ - API キー設定済み
 */
export const OpenAIWithApiKey: Story = {
  args: {
    initialConfig: {
      kind: "openai-chat",
      model: "gpt-4o",
      baseUrl: "",
      systemPrompt: "",
      hasApiKey: true,
    },
  },
};

/**
 * OpenAI プロバイダ - API キー未設定
 */
export const OpenAIWithoutApiKey: Story = {
  args: {
    initialConfig: {
      kind: "openai-chat",
      model: "gpt-4o",
      baseUrl: "",
      systemPrompt: "",
      hasApiKey: false,
    },
  },
};

/**
 * Gemini CLI プロバイダ
 */
export const GeminiCLI: Story = {
  args: {
    initialConfig: {
      kind: "gemini-cli",
      model: "",
      baseUrl: "",
      systemPrompt: "",
      hasApiKey: false,
    },
  },
};

/**
 * モックプロバイダ（開発用）
 */
export const MockProvider: Story = {
  args: {
    initialConfig: {
      kind: "mock",
      model: "",
      baseUrl: "",
      systemPrompt: "",
      hasApiKey: false,
    },
  },
};

/**
 * ローディング状態
 */
export const Loading: Story = {
  args: {
    loading: true,
  },
};

/**
 * 保存中状態
 */
export const Saving: Story = {
  args: {
    initialConfig: {
      kind: "codex-cli",
      model: "gpt-4o",
      baseUrl: "",
      systemPrompt: "",
      hasApiKey: false,
    },
    saving: true,
  },
};

/**
 * 接続テスト中状態
 */
export const Testing: Story = {
  args: {
    initialConfig: {
      kind: "openai-chat",
      model: "gpt-4o",
      baseUrl: "",
      systemPrompt: "",
      hasApiKey: true,
    },
    testing: true,
  },
};

/**
 * カスタムエンドポイント設定済み
 */
export const WithCustomEndpoint: Story = {
  args: {
    initialConfig: {
      kind: "openai-chat",
      model: "gpt-4o-mini",
      baseUrl: "https://custom-api.example.com/v1",
      systemPrompt: "",
      hasApiKey: true,
    },
  },
};
