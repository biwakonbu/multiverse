import type { Meta, StoryObj } from "@storybook/svelte";
import ToolbarPreview from "./ToolbarPreview.svelte";
import type { PoolSummary } from "../../types";

const meta = {
  title: "IDE/Toolbar",
  component: ToolbarPreview,
  tags: ["autodocs"],
  argTypes: {
    viewMode: {
      control: { type: "select" },
      options: ["graph", "wbs"],
      description: "表示モード",
    },
    overallProgress: {
      control: "object",
      description: "全体進捗",
    },
    poolSummaries: {
      control: "object",
      description: "Pool別サマリ",
    },
    taskCountsByStatus: {
      control: "object",
      description: "ステータス別タスク数",
    },
  },
  parameters: {
    layout: "fullscreen",
    docs: {
      description: {
        component:
          "IDEのメインツールバー。ブランド表示、ステータスサマリ、進捗バー、ビュー切替を表示します。",
      },
    },
  },
} satisfies Meta<ToolbarPreview>;

export default meta;
type Story = StoryObj<typeof meta>;

// デフォルト（Pool情報なし、グラフモード）
export const Default: Story = {
  args: {
    viewMode: "graph",
    overallProgress: { percentage: 0, completed: 0, total: 0 },
    poolSummaries: [],
    taskCountsByStatus: {
      PENDING: 0,
      READY: 0,
      RUNNING: 0,
      SUCCEEDED: 0,
      COMPLETED: 0,
      FAILED: 0,
      CANCELED: 0,
      BLOCKED: 0,
    },
  },
};

// Pool別サマリ表示
export const WithPoolSummaries: Story = {
  args: {
    viewMode: "graph",
    overallProgress: { percentage: 45, completed: 9, total: 20 },
    poolSummaries: [
      { poolId: "codegen", running: 2, queued: 5, failed: 1, total: 15 },
      { poolId: "test", running: 1, queued: 2, failed: 0, total: 5 },
    ] as PoolSummary[],
    taskCountsByStatus: {
      PENDING: 7,
      READY: 0,
      RUNNING: 3,
      SUCCEEDED: 9,
      COMPLETED: 0,
      FAILED: 1,
      CANCELED: 0,
      BLOCKED: 0,
    },
  },
  parameters: {
    docs: {
      description: {
        story: "Pool別のステータスサマリを表示します。",
      },
    },
  },
};

// ステータス別サマリ表示（Pool情報なし）
export const WithStatusSummary: Story = {
  args: {
    viewMode: "graph",
    overallProgress: { percentage: 30, completed: 3, total: 10 },
    poolSummaries: [],
    taskCountsByStatus: {
      PENDING: 4,
      READY: 0,
      RUNNING: 2,
      SUCCEEDED: 3,
      COMPLETED: 0,
      FAILED: 1,
      CANCELED: 0,
      BLOCKED: 0,
    },
  },
  parameters: {
    docs: {
      description: {
        story: "Pool情報がない場合はステータス別のサマリを表示します。",
      },
    },
  },
};

// 高進捗率
export const HighProgress: Story = {
  args: {
    viewMode: "graph",
    overallProgress: { percentage: 85, completed: 17, total: 20 },
    poolSummaries: [
      { poolId: "codegen", running: 1, queued: 2, failed: 0, total: 15 },
    ] as PoolSummary[],
    taskCountsByStatus: {
      PENDING: 2,
      READY: 0,
      RUNNING: 1,
      SUCCEEDED: 17,
      COMPLETED: 0,
      FAILED: 0,
      CANCELED: 0,
      BLOCKED: 0,
    },
  },
  parameters: {
    docs: {
      description: {
        story: "進捗率が高い状態（85%）。",
      },
    },
  },
};

// 完了状態
export const AllCompleted: Story = {
  args: {
    viewMode: "graph",
    overallProgress: { percentage: 100, completed: 20, total: 20 },
    poolSummaries: [
      { poolId: "codegen", running: 0, queued: 0, failed: 0, total: 15 },
      { poolId: "test", running: 0, queued: 0, failed: 0, total: 5 },
    ] as PoolSummary[],
    taskCountsByStatus: {
      PENDING: 0,
      READY: 0,
      RUNNING: 0,
      SUCCEEDED: 20,
      COMPLETED: 0,
      FAILED: 0,
      CANCELED: 0,
      BLOCKED: 0,
    },
  },
  parameters: {
    docs: {
      description: {
        story: "全タスク完了（進捗100%）。",
      },
    },
  },
};

// WBSモード
export const WBSMode: Story = {
  args: {
    viewMode: "wbs",
    overallProgress: { percentage: 50, completed: 10, total: 20 },
    poolSummaries: [
      { poolId: "codegen", running: 1, queued: 4, failed: 0, total: 15 },
    ] as PoolSummary[],
    taskCountsByStatus: {
      PENDING: 5,
      READY: 0,
      RUNNING: 1,
      SUCCEEDED: 10,
      COMPLETED: 0,
      FAILED: 0,
      CANCELED: 0,
      BLOCKED: 4,
    },
  },
  parameters: {
    docs: {
      description: {
        story: "WBSモード。ズームコントロールは非表示。",
      },
    },
  },
};

// エラーが多い状態
export const ManyFailures: Story = {
  args: {
    viewMode: "graph",
    overallProgress: { percentage: 40, completed: 8, total: 20 },
    poolSummaries: [
      { poolId: "codegen", running: 0, queued: 2, failed: 5, total: 15 },
      { poolId: "test", running: 0, queued: 0, failed: 3, total: 5 },
    ] as PoolSummary[],
    taskCountsByStatus: {
      PENDING: 2,
      READY: 0,
      RUNNING: 0,
      SUCCEEDED: 8,
      COMPLETED: 0,
      FAILED: 8,
      CANCELED: 2,
      BLOCKED: 0,
    },
  },
  parameters: {
    docs: {
      description: {
        story: "失敗タスクが多い状態。",
      },
    },
  },
};
