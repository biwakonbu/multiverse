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
          "ツーリング設定パネル。プロファイル選択、Force Mode、高度な設定（JSON編集）を提供します。",
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
        story: "複数のプロファイルが設定されている状態。",
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

// 完全な設定
export const FullConfig: Story = {
  args: {
    initialConfig: {
      activeProfile: "coding",
      profiles: [
        { id: "coding", name: "Coding" },
        { id: "research", name: "Research" },
        { id: "writing", name: "Writing" },
      ],
      force: {
        enabled: true,
        tool: "claude-code",
        model: "claude-sonnet-4-20250514",
      },
    },
  },
  parameters: {
    docs: {
      description: {
        story: "プロファイルとForce Modeの両方が設定されている状態。",
      },
    },
  },
};
