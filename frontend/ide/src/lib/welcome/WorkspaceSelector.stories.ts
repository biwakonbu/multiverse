import type { Meta, StoryObj } from "@storybook/svelte";
import { fn } from "@storybook/test";
import WorkspaceSelectorPreview from "./WorkspaceSelectorPreview.svelte";

const meta = {
  title: "IDE/Welcome/WorkspaceSelector",
  component: WorkspaceSelectorPreview,
  parameters: {
    layout: "fullscreen",
  },
  tags: ["autodocs"],
  argTypes: {
    isLoadingRecent: {
      control: "boolean",
      description: "最近のワークスペース読み込み中",
    },
    isLoading: {
      control: "boolean",
      description: "ワークスペース選択中",
    },
  },
  args: {
    onOpen: fn(),
    onRemove: fn(),
    onSelectNew: fn(),
  },
} as Meta<typeof WorkspaceSelectorPreview>;

export default meta;
type Story = StoryObj<typeof meta>;

/**
 * 初回起動 - ワークスペースなし
 */
export const Empty: Story = {
  args: {
    recentWorkspaces: [],
    isLoadingRecent: false,
    isLoading: false,
  },
};

/**
 * 最近のワークスペースあり
 */
export const WithRecentWorkspaces: Story = {
  args: {
    recentWorkspaces: [
      {
        id: "ws-1",
        displayName: "multiverse",
        projectRoot: "/Users/dev/projects/multiverse",
        lastOpenedAt: new Date(Date.now() - 1000 * 60 * 30).toISOString(), // 30分前
      },
      {
        id: "ws-2",
        displayName: "my-app",
        projectRoot: "/Users/dev/projects/my-app",
        lastOpenedAt: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(), // 2時間前
      },
      {
        id: "ws-3",
        displayName: "backend-api",
        projectRoot: "/Users/dev/projects/backend-api",
        lastOpenedAt: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString(), // 1日前
      },
    ],
    isLoadingRecent: false,
    isLoading: false,
  },
};

/**
 * 多数のワークスペース
 */
export const ManyWorkspaces: Story = {
  args: {
    recentWorkspaces: [
      {
        id: "ws-1",
        displayName: "multiverse",
        projectRoot: "/Users/dev/projects/multiverse",
        lastOpenedAt: new Date(Date.now() - 1000 * 60 * 30).toISOString(),
      },
      {
        id: "ws-2",
        displayName: "my-app",
        projectRoot: "/Users/dev/projects/my-app",
        lastOpenedAt: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(),
      },
      {
        id: "ws-3",
        displayName: "backend-api",
        projectRoot: "/Users/dev/projects/backend-api",
        lastOpenedAt: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString(),
      },
      {
        id: "ws-4",
        displayName: "frontend-web",
        projectRoot: "/Users/dev/projects/frontend-web",
        lastOpenedAt: new Date(Date.now() - 1000 * 60 * 60 * 48).toISOString(),
      },
      {
        id: "ws-5",
        displayName: "mobile-app",
        projectRoot: "/Users/dev/projects/mobile-app",
        lastOpenedAt: new Date(Date.now() - 1000 * 60 * 60 * 72).toISOString(),
      },
    ],
    isLoadingRecent: false,
    isLoading: false,
  },
};

/**
 * 読み込み中
 */
export const Loading: Story = {
  args: {
    recentWorkspaces: [],
    isLoadingRecent: true,
    isLoading: false,
  },
};

/**
 * ワークスペース選択中
 */
export const SelectingWorkspace: Story = {
  args: {
    recentWorkspaces: [
      {
        id: "ws-1",
        displayName: "multiverse",
        projectRoot: "/Users/dev/projects/multiverse",
        lastOpenedAt: new Date(Date.now() - 1000 * 60 * 30).toISOString(),
      },
    ],
    isLoadingRecent: false,
    isLoading: true,
  },
};
