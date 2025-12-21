import type { Meta, StoryObj } from "@storybook/svelte";
import ToolingSettingsPanelPreview from "./ToolingSettingsPanelPreview.svelte";

const meta = {
  title: "Settings/ToolingSettingsPanel",
  component: ToolingSettingsPanelPreview,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
    docs: {
      description: {
        component:
          "ツーリング設定パネル。プロファイル選択、Force Mode、高度な設定（JSON編集）を提供します。2カラムレイアウトでプロファイル管理が直感的に行えます。",
      },
    },
  },
  argTypes: {
    initialConfig: {
      control: "object",
      description: "初期設定",
    },
  },
} satisfies Meta<typeof ToolingSettingsPanelPreview>;

export default meta;
type Story = StoryObj<typeof meta>;

// デフォルト状態（空の設定）
export const Default: Story = {
  args: {
    initialConfig: {},
  },
};

// プロファイルが設定されている状態
export const WithProfiles: Story = {
  args: {
    initialConfig: {
      activeProfile: "coding",
      profiles: [
        { id: "coding", name: "Coding" },
        { id: "research", name: "Research" },
        { id: "writing", name: "Writing" },
      ],
    },
  },
  parameters: {
    docs: {
      description: {
        story: "複数のプロファイルが設定されている状態。左サイドバーでプロファイルを選択し、右側で編集できます。",
      },
    },
  },
};

// カテゴリを含む完全な設定
export const WithCategories: Story = {
  args: {
    initialConfig: {
      activeProfile: "balanced",
      profiles: [
        {
          id: "balanced",
          name: "Balanced",
          categories: {
            coding: {
              strategy: "weighted",
              fallbackOnRateLimit: true,
              cooldownSec: 120,
              candidates: [
                { tool: "claude-code", model: "claude-sonnet-4-20250514", weight: 50 },
                { tool: "codex-cli", model: "gpt-4.1", weight: 30 },
                { tool: "gemini-cli", model: "gemini-2.5-pro", weight: 20 },
              ],
            },
            research: {
              strategy: "round_robin",
              fallbackOnRateLimit: true,
              cooldownSec: 60,
              candidates: [
                { tool: "claude-code", model: "claude-opus-4-20250514", weight: 10 },
                { tool: "gemini-cli", model: "gemini-2.5-flash", weight: 10 },
              ],
            },
          },
        },
        {
          id: "fast",
          name: "Fast & Cheap",
          categories: {
            coding: {
              strategy: "weighted",
              fallbackOnRateLimit: true,
              cooldownSec: 30,
              candidates: [
                { tool: "codex-cli", model: "gpt-4.1-mini", weight: 70 },
                { tool: "gemini-cli", model: "gemini-2.5-flash", weight: 30 },
              ],
            },
          },
        },
        {
          id: "quality",
          name: "High Quality",
          categories: {
            coding: {
              strategy: "weighted",
              fallbackOnRateLimit: false,
              cooldownSec: 300,
              candidates: [
                { tool: "claude-code", model: "claude-opus-4-20250514", weight: 100 },
              ],
            },
          },
        },
      ],
    },
  },
  parameters: {
    docs: {
      description: {
        story: "カテゴリと候補を含む完全な設定。アコーディオンを展開してカテゴリ設定を確認・編集できます。",
      },
    },
  },
};

// Force Mode が有効な状態
export const ForceModeEnabled: Story = {
  args: {
    initialConfig: {
      force: {
        enabled: true,
        tool: "claude-code",
        model: "claude-opus-4-20250514",
      },
    },
  },
  parameters: {
    docs: {
      description: {
        story: "Force Mode が有効で、特定のツールとモデルが選択されている状態。",
      },
    },
  },
};

// Force Mode が無効な状態
export const ForceModeDisabled: Story = {
  args: {
    initialConfig: {
      force: {
        enabled: false,
        tool: "codex-cli",
        model: "gpt-4.1",
      },
    },
  },
  parameters: {
    docs: {
      description: {
        story: "Force Mode が無効な状態。フィールドは薄く表示される。",
      },
    },
  },
};

// 完全な設定（プロファイル + カテゴリ + Force Mode）
export const FullConfig: Story = {
  args: {
    initialConfig: {
      activeProfile: "balanced",
      profiles: [
        {
          id: "balanced",
          name: "Balanced",
          categories: {
            coding: {
              strategy: "weighted",
              fallbackOnRateLimit: true,
              cooldownSec: 120,
              candidates: [
                { tool: "claude-code", model: "claude-sonnet-4-20250514", weight: 50 },
                { tool: "codex-cli", model: "gpt-4.1", weight: 30 },
                { tool: "gemini-cli", model: "gemini-2.5-pro", weight: 20 },
              ],
            },
            research: {
              strategy: "round_robin",
              fallbackOnRateLimit: true,
              cooldownSec: 60,
              candidates: [
                { tool: "claude-code", model: "claude-opus-4-20250514", weight: 10 },
              ],
            },
            documentation: {
              strategy: "weighted",
              fallbackOnRateLimit: true,
              cooldownSec: 180,
              candidates: [
                { tool: "gemini-cli", model: "gemini-2.5-pro", weight: 60 },
                { tool: "claude-code", model: "claude-sonnet-4-20250514", weight: 40 },
              ],
            },
          },
        },
        {
          id: "fast",
          name: "Fast & Cheap",
          categories: {
            coding: {
              strategy: "weighted",
              fallbackOnRateLimit: true,
              cooldownSec: 30,
              candidates: [
                { tool: "codex-cli", model: "gpt-4.1-mini", weight: 70 },
                { tool: "gemini-cli", model: "gemini-2.5-flash", weight: 30 },
              ],
            },
          },
        },
      ],
      force: {
        enabled: false,
        tool: "claude-code",
        model: "claude-sonnet-4-20250514",
      },
    },
  },
  parameters: {
    docs: {
      description: {
        story: "プロファイル、カテゴリ、Force Modeの全ての設定が含まれている完全な状態。",
      },
    },
  },
};
