/**
 * E2E テスト用 Wails バインディングモック
 *
 * このファイルは VITE_TEST_E2E=true 環境でのみ使用されます。
 * 本番環境（wails dev / wails build）では使用されません。
 *
 * 使用方法:
 * - E2E テスト実行: VITE_TEST_E2E=true npm run dev
 * - 通常開発: wails dev（このファイルは無視される）
 */
console.log('[Mock] Wails E2E test bindings loaded');

// モックワークスペースデータ
const mockWorkspaces = [
    {
        id: "mock-workspace-1",
        displayName: "My Project",
        projectRoot: "/Users/demo/projects/my-project",
        lastOpenedAt: new Date(Date.now() - 1000 * 60 * 30).toISOString() // 30分前
    },
    {
        id: "mock-workspace-2",
        displayName: "Another Project",
        projectRoot: "/Users/demo/projects/another-project",
        lastOpenedAt: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString() // 1日前
    }
];

export function SelectWorkspace() {
    console.log("[Mock] SelectWorkspace called");
    return Promise.resolve("mock-workspace-new");
}

export function ListRecentWorkspaces() {
    console.log("[Mock] ListRecentWorkspaces called");
    const workspaces = JSON.parse(window.localStorage.getItem('mock_workspaces') || JSON.stringify(mockWorkspaces));
    return Promise.resolve(workspaces);
}

export function OpenWorkspaceByID(id) {
    console.log("[Mock] OpenWorkspaceByID called", id);
    return Promise.resolve(id);
}

export function RemoveWorkspace(id) {
    console.log("[Mock] RemoveWorkspace called", id);
    const workspaces = JSON.parse(window.localStorage.getItem('mock_workspaces') || JSON.stringify(mockWorkspaces));
    const filtered = workspaces.filter(w => w.id !== id);
    window.localStorage.setItem('mock_workspaces', JSON.stringify(filtered));
    return Promise.resolve();
}

export function ListTasks() {
    console.log("[Mock] ListTasks called");
    const tasks = JSON.parse(window.localStorage.getItem('mock_tasks') || '[]');
    return Promise.resolve(tasks);
}

export function GetWorkspace(id) {
    console.log("[Mock] GetWorkspace called", id);
    return Promise.resolve({
        version: "1.0",
        projectRoot: "/mock/root",
        displayName: "Mock Project",
        createdAt: new Date().toISOString(),
        lastOpenedAt: new Date().toISOString()
    });
}

export function CreateTask(title, poolId) {
    console.log("[Mock] CreateTask called", title, poolId);
    const tasks = JSON.parse(window.localStorage.getItem('mock_tasks') || '[]');
    const newTask = {
        id: "task-" + Date.now(),
        title: title,
        status: "PENDING",
        poolId: poolId,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
    };
    tasks.push(newTask);
    window.localStorage.setItem('mock_tasks', JSON.stringify(tasks));
    return Promise.resolve(newTask);
}

export function RunTask(taskId) {
    console.log("[Mock] RunTask called", taskId);
    return Promise.resolve();
}

export function ListAttempts(taskId) {
    console.log("[Mock] ListAttempts called", taskId);
    return Promise.resolve([]);
}

export function GetPoolSummaries() {
    console.log("[Mock] GetPoolSummaries called");
    return Promise.resolve([]);
}

export function GetAvailablePools() {
    console.log("[Mock] GetAvailablePools called");
    return Promise.resolve([
        { id: "default", name: "Default", description: "汎用タスク実行用" },
        { id: "codegen", name: "Codegen", description: "コード生成タスク用" },
        { id: "test", name: "Test", description: "テスト実行タスク用" }
    ]);
}

export function GetTaskGraph() {
    console.log("[Mock] GetTaskGraph called");
    const tasks = JSON.parse(window.localStorage.getItem('mock_tasks') || '[]');
    // 単純なグラフ構造を返す（すべてのタスクをノードとして、依存関係をエッジとして構築）
    const nodes = tasks.map(t => ({
        task: t,
        col: 0,
        row: 0
    }));
    const edges = [];
    tasks.forEach(t => {
        if (t.dependencies) {
            t.dependencies.forEach(depId => {
                edges.push({ from: depId, to: t.id, satisfied: false });
            });
        }
    });
    
    return Promise.resolve({
        nodes: nodes,
        edges: edges,
        blockedTasks: [],
        readyTasks: []
    });
}

export function SendChatMessage(sessionId, message) {
    console.log("[Mock] SendChatMessage called", sessionId, message);
    
    // Golden Test: TODO App
    if (message.includes("TODO アプリを作成して")) {
         const newTask = {
            id: "task-golden-todo",
            title: "TODO アプリを作成して",
            status: "PENDING",
            poolId: "default",
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString()
        };
        // Update local storage so it appears in lists too
        const tasks = JSON.parse(window.localStorage.getItem('mock_tasks') || '[]');
        tasks.push(newTask);
        window.localStorage.setItem('mock_tasks', JSON.stringify(tasks));

        return Promise.resolve({
            message: {
                id: "msg-" + Date.now(),
                role: "assistant",
                content: "承知しました。TODOアプリの作成タスクを生成しました。",
                timestamp: new Date().toISOString()
            },
            generatedTasks: [newTask],
            understanding: "TODOアプリの作成要件を理解しました。",
            conflicts: []
        });
    }

    return Promise.resolve({
        message: {
             id: "msg-" + Date.now(),
             role: "assistant",
             content: "Mock response",
             timestamp: new Date().toISOString()
        },
        generatedTasks: [],
        understanding: "Mock understanding",
        conflicts: []
    });
}

export function GetChatHistory(sessionId) {
    console.log("[Mock] GetChatHistory called", sessionId);
    return Promise.resolve([]);
}

export function CreateChatSession() {
    console.log("[Mock] CreateChatSession called");
    return Promise.resolve("mock-session-" + Date.now());
}

// 実行状態
let executionState = 'IDLE';

// Execution control
export function StartExecution() {
    console.log("[Mock] StartExecution called");
    executionState = 'RUNNING';
    return Promise.resolve();
}

export function PauseExecution() {
    console.log("[Mock] PauseExecution called");
    executionState = 'PAUSED';
    return Promise.resolve();
}

export function ResumeExecution() {
    console.log("[Mock] ResumeExecution called");
    executionState = 'RUNNING';
    return Promise.resolve();
}

export function StopExecution() {
    console.log("[Mock] StopExecution called");
    executionState = 'IDLE';
    return Promise.resolve();
}

export function GetExecutionState() {
    console.log("[Mock] GetExecutionState called");
    return Promise.resolve(executionState);
}

// Backlog
export function GetBacklogItems() {
    console.log("[Mock] GetBacklogItems called");
    const items = JSON.parse(window.localStorage.getItem('mock_backlog') || '[]');
    return Promise.resolve(items);
}

export function ResolveBacklogItem(id, resolution) {
    console.log("[Mock] ResolveBacklogItem called", id, resolution);
    const items = JSON.parse(window.localStorage.getItem('mock_backlog') || '[]');
    const index = items.findIndex(i => i.id === id);
    if (index >= 0) {
        items[index].resolvedAt = new Date().toISOString();
        items[index].resolution = resolution;
        window.localStorage.setItem('mock_backlog', JSON.stringify(items));
    }
    return Promise.resolve();
}

export function DeleteBacklogItem(id) {
    console.log("[Mock] DeleteBacklogItem called", id);
    const items = JSON.parse(window.localStorage.getItem('mock_backlog') || '[]');
    const filtered = items.filter(i => i.id !== id);
    window.localStorage.setItem('mock_backlog', JSON.stringify(filtered));
    return Promise.resolve();
}
