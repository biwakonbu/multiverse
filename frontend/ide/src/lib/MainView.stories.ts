import type { Meta, StoryObj } from "@storybook/svelte";
import MainViewPreview from "./MainViewPreview.svelte";
import type { Task, TaskStatus } from "../types";

const meta = {
  title: "IDE/MainView",
  component: MainViewPreview,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
    docs: {
      description: {
        component:
          "IDE本体のメインビュー。Toolbar、WBS/Graphビュー、チャットウィンドウを含むワークスペース画面。",
      },
    },
  },
  argTypes: {
    viewMode: {
      control: { type: "select" },
      options: ["graph", "wbs"],
      description: "表示モード（Graph / WBS）",
    },

  },
} as Meta<typeof MainViewPreview>;

export default meta;
type Story = StoryObj<typeof meta>;

// VRT用に固定タイムスタンプを使用（動的な値は視覚回帰テストを不安定にする）
const FIXED_DATE = new Date('2024-01-15T10:00:00Z').toISOString();

// ヘルパー: タスクを生成
function createTask(
  id: string,
  title: string,
  status: TaskStatus,
  opts: Partial<Task> = {}
): Task {
  return {
    id,
    title,
    status,
    poolId: opts.poolId || "codegen",
    createdAt: opts.createdAt || FIXED_DATE,
    updatedAt: opts.updatedAt || FIXED_DATE,
    startedAt: opts.startedAt,
    doneAt: opts.doneAt,
    description: opts.description,
    dependencies: opts.dependencies || [],
    parentId: opts.parentId,
    wbsLevel: opts.wbsLevel || 3,
    phaseName: opts.phaseName || "実装",
    acceptanceCriteria: opts.acceptanceCriteria || [],
  };
}

// ヘルパー: ステータス別カウントを計算
function countByStatus(tasks: Task[]): Record<TaskStatus, number> {
  const counts: Record<TaskStatus, number> = {
    PENDING: 0,
    READY: 0,
    RUNNING: 0,
    SUCCEEDED: 0,
    COMPLETED: 0,
    FAILED: 0,
    CANCELED: 0,
    BLOCKED: 0,
    RETRY_WAIT: 0,
  };
  for (const task of tasks) {
    counts[task.status]++;
  }
  return counts;
}

// ヘルパー: 進捗を計算
function calcProgress(tasks: Task[]) {
  const total = tasks.length;
  const completed = tasks.filter(
    (t) => t.status === "SUCCEEDED" || t.status === "COMPLETED"
  ).length;
  return {
    total,
    completed,
    percentage: total > 0 ? Math.round((completed / total) * 100) : 0,
  };
}

// === Story: プロジェクト開始直後 ===
const projectStartTasks: Task[] = [
  createTask("task-1", "要件定義書作成", "SUCCEEDED", { phaseName: "概念設計", wbsLevel: 2 }),
  createTask("task-2", "アーキテクチャ設計", "RUNNING", { phaseName: "実装設計", wbsLevel: 2, dependencies: ["task-1"] }),
  createTask("task-3", "データベース設計", "PENDING", { phaseName: "実装設計", wbsLevel: 3, dependencies: ["task-2"] }),
  createTask("task-4", "API 設計", "PENDING", { phaseName: "実装設計", wbsLevel: 3, dependencies: ["task-2"] }),
];

export const ProjectStart: Story = {
  args: {
    viewMode: "wbs",
    taskList: projectStartTasks,
    poolSummaries: [
      { poolId: "codegen", running: 1, queued: 2, failed: 0, total: 4, counts: {} },
    ],
    overallProgress: calcProgress(projectStartTasks),
    taskCountsByStatus: countByStatus(projectStartTasks),
    selectedTask: null,
    showChat: true,
    chatPosition: { x: 600, y: 300 },
  },
  parameters: {
    docs: {
      description: {
        story:
          "プロジェクト開始直後の状態。要件定義が完了し、アーキテクチャ設計が実行中。後続タスクは待機中。",
      },
    },
  },
};

// === Story: 開発進行中（50%） ===
const developmentInProgressTasks: Task[] = [
  createTask("task-1", "要件定義書作成", "SUCCEEDED", { phaseName: "概念設計", wbsLevel: 2 }),
  createTask("task-2", "アーキテクチャ設計", "SUCCEEDED", { phaseName: "実装設計", wbsLevel: 2 }),
  createTask("task-3", "データベース設計", "SUCCEEDED", { phaseName: "実装設計", wbsLevel: 3 }),
  createTask("task-4", "API 設計", "SUCCEEDED", { phaseName: "実装設計", wbsLevel: 3 }),
  createTask("task-5", "ユーザー認証機能", "RUNNING", { phaseName: "実装", wbsLevel: 3, poolId: "codegen" }),
  createTask("task-6", "ダッシュボード画面", "RUNNING", { phaseName: "実装", wbsLevel: 3, poolId: "codegen" }),
  createTask("task-7", "設定画面", "PENDING", { phaseName: "実装", wbsLevel: 3 }),
  createTask("task-8", "通知機能", "PENDING", { phaseName: "実装", wbsLevel: 3 }),
  createTask("task-9", "単体テスト", "BLOCKED", { phaseName: "検証", wbsLevel: 3, dependencies: ["task-5", "task-6"] }),
  createTask("task-10", "E2Eテスト", "BLOCKED", { phaseName: "検証", wbsLevel: 3, dependencies: ["task-9"] }),
];

export const DevelopmentInProgress: Story = {
  args: {
    viewMode: "wbs",
    taskList: developmentInProgressTasks,
    poolSummaries: [
      { poolId: "codegen", running: 2, queued: 2, failed: 0, total: 6, counts: {} },
      { poolId: "test", running: 0, queued: 0, failed: 0, total: 2, counts: {} },
    ],
    overallProgress: calcProgress(developmentInProgressTasks),
    taskCountsByStatus: countByStatus(developmentInProgressTasks),
    selectedTask: null,
    showChat: true,
    chatPosition: { x: 600, y: 300 },
  },
  parameters: {
    docs: {
      description: {
        story:
          "開発進行中の状態（進捗約40%）。複数のタスクが並行して実行中。テストタスクはブロック状態。",
      },
    },
  },
};

// === Story: タスク選択状態 ===
const selectedTaskData = createTask("task-5", "ユーザー認証機能", "RUNNING", {
  phaseName: "実装",
  wbsLevel: 3,
  description:
    "JWT ベースのユーザー認証機能を実装する。ログイン、ログアウト、トークンリフレッシュ、パスワードリセットを含む。",
  acceptanceCriteria: [
    "ログインフォームからメール/パスワードでログインできる",
    "JWT トークンが発行され、ローカルストレージに保存される",
    "トークンの有効期限切れ時に自動リフレッシュされる",
    "パスワードリセットメールが送信できる",
  ],
  dependencies: ["task-3", "task-4"],
});

export const TaskSelected: Story = {
  args: {
    viewMode: "wbs",
    taskList: developmentInProgressTasks,
    poolSummaries: [
      { poolId: "codegen", running: 2, queued: 2, failed: 0, total: 6, counts: {} },
    ],
    overallProgress: calcProgress(developmentInProgressTasks),
    taskCountsByStatus: countByStatus(developmentInProgressTasks),
    selectedTask: selectedTaskData,
    showChat: true,
    chatPosition: { x: 600, y: 300 },
  },
  parameters: {
    docs: {
      description: {
        story:
          "タスクを選択した状態。右側の詳細パネルに説明、受け入れ条件、実行履歴が表示される。",
      },
    },
  },
};

// === Story: エラー発生時 ===
const errorStateTasks: Task[] = [
  createTask("task-1", "要件定義書作成", "SUCCEEDED", { phaseName: "概念設計" }),
  createTask("task-2", "アーキテクチャ設計", "SUCCEEDED", { phaseName: "実装設計" }),
  createTask("task-3", "データベース設計", "SUCCEEDED", { phaseName: "実装設計" }),
  createTask("task-4", "API 設計", "FAILED", { phaseName: "実装設計" }),
  createTask("task-5", "ユーザー認証機能", "FAILED", { phaseName: "実装", poolId: "codegen" }),
  createTask("task-6", "ダッシュボード画面", "RUNNING", { phaseName: "実装", poolId: "codegen" }),
  createTask("task-7", "設定画面", "BLOCKED", { phaseName: "実装", dependencies: ["task-5"] }),
  createTask("task-8", "通知機能", "CANCELED", { phaseName: "実装" }),
];

const failedTaskData = createTask("task-5", "ユーザー認証機能", "FAILED", {
  phaseName: "実装",
  wbsLevel: 3,
  description: "JWT ベースのユーザー認証機能を実装する。",
});

export const ErrorState: Story = {
  args: {
    viewMode: "wbs",
    taskList: errorStateTasks,
    poolSummaries: [
      { poolId: "codegen", running: 1, queued: 0, failed: 2, total: 4, counts: {} },
    ],
    overallProgress: calcProgress(errorStateTasks),
    taskCountsByStatus: countByStatus(errorStateTasks),
    selectedTask: failedTaskData,
    showChat: true,
    chatPosition: { x: 600, y: 300 },
  },
  parameters: {
    docs: {
      description: {
        story:
          "エラーが発生した状態。複数のタスクが失敗し、依存タスクがブロックされている。詳細パネルにエラー履歴を表示。",
      },
    },
  },
};

// === Story: プロジェクト完了間近（90%） ===
const nearCompletionTasks: Task[] = [
  createTask("task-1", "要件定義書作成", "SUCCEEDED", { phaseName: "概念設計" }),
  createTask("task-2", "アーキテクチャ設計", "SUCCEEDED", { phaseName: "実装設計" }),
  createTask("task-3", "データベース設計", "SUCCEEDED", { phaseName: "実装設計" }),
  createTask("task-4", "API 設計", "SUCCEEDED", { phaseName: "実装設計" }),
  createTask("task-5", "ユーザー認証機能", "SUCCEEDED", { phaseName: "実装" }),
  createTask("task-6", "ダッシュボード画面", "SUCCEEDED", { phaseName: "実装" }),
  createTask("task-7", "設定画面", "SUCCEEDED", { phaseName: "実装" }),
  createTask("task-8", "通知機能", "SUCCEEDED", { phaseName: "実装" }),
  createTask("task-9", "単体テスト", "SUCCEEDED", { phaseName: "検証" }),
  createTask("task-10", "E2Eテスト", "RUNNING", { phaseName: "検証", poolId: "test" }),
  createTask("task-11", "パフォーマンステスト", "PENDING", { phaseName: "検証", dependencies: ["task-10"] }),
  createTask("task-12", "ドキュメント作成", "PENDING", { phaseName: "検証" }),
];

export const NearCompletion: Story = {
  args: {
    viewMode: "wbs",
    taskList: nearCompletionTasks,
    poolSummaries: [
      { poolId: "codegen", running: 0, queued: 0, failed: 0, total: 8, counts: {} },
      { poolId: "test", running: 1, queued: 2, failed: 0, total: 4, counts: {} },
    ],
    overallProgress: calcProgress(nearCompletionTasks),
    taskCountsByStatus: countByStatus(nearCompletionTasks),
    selectedTask: null,
    showChat: true,
    chatPosition: { x: 600, y: 300 },
  },
  parameters: {
    docs: {
      description: {
        story:
          "プロジェクト完了間近の状態（進捗75%）。ほとんどの実装タスクが完了し、テストフェーズが進行中。",
      },
    },
  },
};

// === Story: プロジェクト完了 ===
const completedTasks: Task[] = [
  createTask("task-1", "要件定義書作成", "SUCCEEDED", { phaseName: "概念設計" }),
  createTask("task-2", "アーキテクチャ設計", "SUCCEEDED", { phaseName: "実装設計" }),
  createTask("task-3", "データベース設計", "SUCCEEDED", { phaseName: "実装設計" }),
  createTask("task-4", "API 設計", "SUCCEEDED", { phaseName: "実装設計" }),
  createTask("task-5", "ユーザー認証機能", "SUCCEEDED", { phaseName: "実装" }),
  createTask("task-6", "ダッシュボード画面", "SUCCEEDED", { phaseName: "実装" }),
  createTask("task-7", "設定画面", "SUCCEEDED", { phaseName: "実装" }),
  createTask("task-8", "通知機能", "SUCCEEDED", { phaseName: "実装" }),
  createTask("task-9", "単体テスト", "SUCCEEDED", { phaseName: "検証" }),
  createTask("task-10", "E2Eテスト", "SUCCEEDED", { phaseName: "検証" }),
  createTask("task-11", "パフォーマンステスト", "SUCCEEDED", { phaseName: "検証" }),
  createTask("task-12", "ドキュメント作成", "SUCCEEDED", { phaseName: "検証" }),
];

export const ProjectCompleted: Story = {
  args: {
    viewMode: "wbs",
    taskList: completedTasks,
    poolSummaries: [
      { poolId: "codegen", running: 0, queued: 0, failed: 0, total: 8, counts: {} },
      { poolId: "test", running: 0, queued: 0, failed: 0, total: 4, counts: {} },
    ],
    overallProgress: { total: 12, completed: 12, percentage: 100 },
    taskCountsByStatus: countByStatus(completedTasks),
    selectedTask: null,
    showChat: false,
    chatPosition: { x: 600, y: 300 },
  },
  parameters: {
    docs: {
      description: {
        story:
          "プロジェクト完了状態。全タスクが成功し、進捗100%。チャットウィンドウは閉じている。",
      },
    },
  },
};

// === Story: Graph ビュー ===
export const GraphView: Story = {
  args: {
    viewMode: "graph",
    taskList: developmentInProgressTasks,
    poolSummaries: [
      { poolId: "codegen", running: 2, queued: 2, failed: 0, total: 6, counts: {} },
    ],
    overallProgress: calcProgress(developmentInProgressTasks),
    taskCountsByStatus: countByStatus(developmentInProgressTasks),
    selectedTask: null,
    showChat: true,
    chatPosition: { x: 600, y: 300 },
  },
  parameters: {
    docs: {
      description: {
        story:
          "Graph ビューモード。タスクを2D俯瞰で依存関係グラフとして表示。",
      },
    },
  },
};

// === Story: チャット非表示 ===
export const ChatHidden: Story = {
  args: {
    viewMode: "wbs",
    taskList: developmentInProgressTasks,
    poolSummaries: [
      { poolId: "codegen", running: 2, queued: 2, failed: 0, total: 6, counts: {} },
    ],
    overallProgress: calcProgress(developmentInProgressTasks),
    taskCountsByStatus: countByStatus(developmentInProgressTasks),
    selectedTask: null,
    showChat: false,
    chatPosition: { x: 600, y: 300 },
  },
  parameters: {
    docs: {
      description: {
        story:
          "チャットウィンドウを閉じた状態。右下にチャット再表示ボタン（FAB）が表示される。",
      },
    },
  },
};

// === Story: 空のプロジェクト ===
export const EmptyProject: Story = {
  args: {
    viewMode: "wbs",
    taskList: [],
    poolSummaries: [],
    overallProgress: { total: 0, completed: 0, percentage: 0 },
    taskCountsByStatus: {
      PENDING: 0,
      READY: 0,
      RUNNING: 0,
      SUCCEEDED: 0,
      COMPLETED: 0,
      FAILED: 0,
      CANCELED: 0,
      BLOCKED: 0,
      RETRY_WAIT: 0,
    },
    selectedTask: null,
    showChat: true,
    chatPosition: { x: 600, y: 300 },
  },
  parameters: {
    docs: {
      description: {
        story:
          "タスクがない空のプロジェクト。新規ワークスペース作成直後の状態。",
      },
    },
  },
};

// === Story: 大規模プロジェクト ===
const largeTasks: Task[] = [];
const phases = ["概念設計", "実装設計", "実装", "検証"] as const;
const statuses: TaskStatus[] = ["SUCCEEDED", "SUCCEEDED", "RUNNING", "PENDING", "BLOCKED"];

for (let i = 0; i < 50; i++) {
  const phaseIndex = Math.floor(i / 15);
  const phase = phases[Math.min(phaseIndex, phases.length - 1)];
  const statusIndex = Math.floor((i % 15) / 3);
  const status = i < 30 ? "SUCCEEDED" : statuses[Math.min(statusIndex, statuses.length - 1)];

  largeTasks.push(
    createTask(`task-${i + 1}`, `タスク ${i + 1}: ${phase}フェーズの作業`, status, {
      phaseName: phase,
      wbsLevel: 3,
      poolId: phase === "検証" ? "test" : "codegen",
      dependencies: i > 0 && i % 5 === 0 ? [`task-${i}`] : [],
    })
  );
}

export const LargeProject: Story = {
  args: {
    viewMode: "wbs",
    taskList: largeTasks,
    poolSummaries: [
      { poolId: "codegen", running: 3, queued: 10, failed: 0, total: 38, counts: {} },
      { poolId: "test", running: 1, queued: 5, failed: 0, total: 12, counts: {} },
    ],
    overallProgress: calcProgress(largeTasks),
    taskCountsByStatus: countByStatus(largeTasks),
    selectedTask: null,
    showChat: true,
    chatPosition: { x: 600, y: 300 },
  },
  parameters: {
    docs: {
      description: {
        story:
          "50タスクの大規模プロジェクト。スクロールとパフォーマンスの確認用。",
      },
    },
  },
};

// === Story: バックログパネル表示 ===
export const WithBacklogOpen: Story = {
  args: {
    viewMode: "wbs",
    taskList: errorStateTasks,
    poolSummaries: [
      { poolId: "codegen", running: 1, queued: 0, failed: 2, total: 4, counts: {} },
    ],
    overallProgress: calcProgress(errorStateTasks),
    taskCountsByStatus: countByStatus(errorStateTasks),
    selectedTask: null,
    showChat: true,
    chatPosition: { x: 600, y: 300 },
    showBacklog: true,
    unresolvedCount: 3,
  },
  parameters: {
    docs: {
      description: {
        story:
          "バックログパネルが展開された状態。エラー発生時にバックログで問題を管理。",
      },
    },
  },
};



// === Story: 全オーバーレイ表示 ===
export const WithAllOverlays: Story = {
  args: {
    viewMode: "wbs",
    taskList: errorStateTasks,
    poolSummaries: [
      { poolId: "codegen", running: 1, queued: 0, failed: 2, total: 4, counts: {} },
    ],
    overallProgress: calcProgress(errorStateTasks),
    taskCountsByStatus: countByStatus(errorStateTasks),
    selectedTask: null,
    showChat: true,
    chatPosition: { x: 600, y: 300 },
    showBacklog: true,
    unresolvedCount: 3,
  },
  parameters: {
    docs: {
      description: {
        story:
          "チャット、バックログが同時に表示された状態。複合UIのレイアウト確認用。",
      },
    },
  },
};

// === Story: Graph ビュー + バックログ ===
export const GraphViewWithBacklog: Story = {
  args: {
    viewMode: "graph",
    taskList: errorStateTasks,
    poolSummaries: [
      { poolId: "codegen", running: 1, queued: 0, failed: 2, total: 4, counts: {} },
    ],
    overallProgress: calcProgress(errorStateTasks),
    taskCountsByStatus: countByStatus(errorStateTasks),
    selectedTask: null,
    showChat: false,
    chatPosition: { x: 600, y: 300 },
    showBacklog: true,
    unresolvedCount: 2,
  },
  parameters: {
    docs: {
      description: {
        story:
          "Graphビューモードでバックログパネルを表示。チャットは非表示。",
      },
    },
  },
};
