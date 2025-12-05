console.log('[Mock] Wails bindings loaded');

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
