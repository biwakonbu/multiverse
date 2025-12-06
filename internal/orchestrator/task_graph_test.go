package orchestrator

import (
	"testing"
	"time"
)

func TestTaskGraphManager_BuildGraph(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)
	manager := NewTaskGraphManager(store)

	// タスクを作成
	now := time.Now()
	tasks := []*Task{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusSucceeded, PoolID: "default", CreatedAt: now, Dependencies: []string{}},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"task-1"}},
		{ID: "task-3", Title: "Task 3", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"task-1", "task-2"}},
	}

	for _, task := range tasks {
		if err := store.SaveTask(task); err != nil {
			t.Fatalf("failed to save task: %v", err)
		}
	}

	// グラフを構築
	graph, err := manager.BuildGraph()
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}

	// ノード数を確認
	if len(graph.Nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(graph.Nodes))
	}

	// エッジ数を確認（task-1->task-2, task-1->task-3, task-2->task-3）
	if len(graph.Edges) != 3 {
		t.Errorf("expected 3 edges, got %d", len(graph.Edges))
	}

	// task-1 の Dependents を確認
	node1 := graph.Nodes["task-1"]
	if node1 == nil {
		t.Fatal("task-1 node not found")
	}
	if len(node1.Dependents) != 2 {
		t.Errorf("expected task-1 to have 2 dependents, got %d", len(node1.Dependents))
	}

	// task-3 の InDegree を確認
	node3 := graph.Nodes["task-3"]
	if node3 == nil {
		t.Fatal("task-3 node not found")
	}
	if node3.InDegree != 2 {
		t.Errorf("expected task-3 InDegree=2, got %d", node3.InDegree)
	}
}

func TestTaskGraphManager_GetExecutionOrder(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)
	manager := NewTaskGraphManager(store)

	now := time.Now()
	tasks := []*Task{
		{ID: "a", Title: "A", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{}},
		{ID: "b", Title: "B", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"a"}},
		{ID: "c", Title: "C", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"a"}},
		{ID: "d", Title: "D", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"b", "c"}},
	}

	for _, task := range tasks {
		if err := store.SaveTask(task); err != nil {
			t.Fatalf("failed to save task: %v", err)
		}
	}

	order, err := manager.GetExecutionOrder()
	if err != nil {
		t.Fatalf("GetExecutionOrder failed: %v", err)
	}

	if len(order) != 4 {
		t.Fatalf("expected 4 tasks in order, got %d", len(order))
	}

	// a は最初
	if order[0] != "a" {
		t.Errorf("expected 'a' first, got '%s'", order[0])
	}

	// d は最後
	if order[3] != "d" {
		t.Errorf("expected 'd' last, got '%s'", order[3])
	}

	// b, c は a の後、d の前
	bPos, cPos := -1, -1
	for i, id := range order {
		if id == "b" {
			bPos = i
		}
		if id == "c" {
			cPos = i
		}
	}

	if bPos < 1 || cPos < 1 {
		t.Error("b and c should come after a")
	}
	if bPos > 2 || cPos > 2 {
		t.Error("b and c should come before d")
	}
}

func TestTaskGraphManager_GetExecutionOrder_Cycle(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)
	manager := NewTaskGraphManager(store)

	now := time.Now()
	// サイクル: a -> b -> c -> a
	tasks := []*Task{
		{ID: "a", Title: "A", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"c"}},
		{ID: "b", Title: "B", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"a"}},
		{ID: "c", Title: "C", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"b"}},
	}

	for _, task := range tasks {
		if err := store.SaveTask(task); err != nil {
			t.Fatalf("failed to save task: %v", err)
		}
	}

	_, err := manager.GetExecutionOrder()
	if err == nil {
		t.Error("expected error for cyclic dependencies, got nil")
	}
}

func TestTaskGraphManager_GetBlockedTasks(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)
	manager := NewTaskGraphManager(store)

	now := time.Now()
	tasks := []*Task{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{}},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"task-1"}},
		{ID: "task-3", Title: "Task 3", Status: TaskStatusSucceeded, PoolID: "default", CreatedAt: now, Dependencies: []string{}},
		{ID: "task-4", Title: "Task 4", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"task-3"}},
	}

	for _, task := range tasks {
		if err := store.SaveTask(task); err != nil {
			t.Fatalf("failed to save task: %v", err)
		}
	}

	blocked, err := manager.GetBlockedTasks()
	if err != nil {
		t.Fatalf("GetBlockedTasks failed: %v", err)
	}

	// task-2 はブロック（task-1 が未完了）
	// task-4 はブロックされていない（task-3 が完了済み）
	if len(blocked) != 1 {
		t.Errorf("expected 1 blocked task, got %d: %v", len(blocked), blocked)
	}

	if len(blocked) > 0 && blocked[0] != "task-2" {
		t.Errorf("expected 'task-2' to be blocked, got '%s'", blocked[0])
	}
}

func TestTaskGraphManager_GetReadyTasks(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)
	manager := NewTaskGraphManager(store)

	now := time.Now()
	tasks := []*Task{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{}},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"task-1"}},
		{ID: "task-3", Title: "Task 3", Status: TaskStatusSucceeded, PoolID: "default", CreatedAt: now, Dependencies: []string{}},
		{ID: "task-4", Title: "Task 4", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"task-3"}},
		{ID: "task-5", Title: "Task 5", Status: TaskStatusRunning, PoolID: "default", CreatedAt: now, Dependencies: []string{}},
	}

	for _, task := range tasks {
		if err := store.SaveTask(task); err != nil {
			t.Fatalf("failed to save task: %v", err)
		}
	}

	ready, err := manager.GetReadyTasks()
	if err != nil {
		t.Fatalf("GetReadyTasks failed: %v", err)
	}

	// task-1: PENDING, 依存なし → Ready
	// task-2: PENDING, task-1未完了 → Blocked
	// task-3: SUCCEEDED → Not Ready (すでに完了)
	// task-4: PENDING, task-3完了 → Ready
	// task-5: RUNNING → Not Ready (すでに実行中)
	if len(ready) != 2 {
		t.Errorf("expected 2 ready tasks, got %d: %v", len(ready), ready)
	}

	// ソートされているので task-1, task-4 の順
	expected := []string{"task-1", "task-4"}
	for i, id := range expected {
		if i >= len(ready) || ready[i] != id {
			t.Errorf("expected ready[%d]='%s', got '%v'", i, id, ready)
		}
	}
}

func TestTaskGraphManager_DetectCycle(t *testing.T) {
	t.Run("no cycle", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := NewTaskStore(tmpDir)
		manager := NewTaskGraphManager(store)

		now := time.Now()
		tasks := []*Task{
			{ID: "a", Title: "A", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{}},
			{ID: "b", Title: "B", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"a"}},
		}

		for _, task := range tasks {
			if err := store.SaveTask(task); err != nil {
				t.Fatalf("failed to save task: %v", err)
			}
		}

		hasCycle, cycleNodes, err := manager.DetectCycle()
		if err != nil {
			t.Fatalf("DetectCycle failed: %v", err)
		}

		if hasCycle {
			t.Errorf("expected no cycle, got cycle with nodes: %v", cycleNodes)
		}
	})

	t.Run("with cycle", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := NewTaskStore(tmpDir)
		manager := NewTaskGraphManager(store)

		now := time.Now()
		tasks := []*Task{
			{ID: "a", Title: "A", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"c"}},
			{ID: "b", Title: "B", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"a"}},
			{ID: "c", Title: "C", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"b"}},
		}

		for _, task := range tasks {
			if err := store.SaveTask(task); err != nil {
				t.Fatalf("failed to save task: %v", err)
			}
		}

		hasCycle, cycleNodes, err := manager.DetectCycle()
		if err != nil {
			t.Fatalf("DetectCycle failed: %v", err)
		}

		if !hasCycle {
			t.Error("expected cycle to be detected")
		}

		if len(cycleNodes) != 3 {
			t.Errorf("expected 3 nodes in cycle, got %d: %v", len(cycleNodes), cycleNodes)
		}
	})
}

func TestTaskGraphManager_GetTaskDependencyInfo(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)
	manager := NewTaskGraphManager(store)

	now := time.Now()
	tasks := []*Task{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusSucceeded, PoolID: "default", CreatedAt: now, Dependencies: []string{}},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{}},
		{ID: "task-3", Title: "Task 3", Status: TaskStatusPending, PoolID: "default", CreatedAt: now, Dependencies: []string{"task-1", "task-2"}},
	}

	for _, task := range tasks {
		if err := store.SaveTask(task); err != nil {
			t.Fatalf("failed to save task: %v", err)
		}
	}

	info, err := manager.GetTaskDependencyInfo("task-3")
	if err != nil {
		t.Fatalf("GetTaskDependencyInfo failed: %v", err)
	}

	if info.TaskID != "task-3" {
		t.Errorf("expected TaskID='task-3', got '%s'", info.TaskID)
	}

	if len(info.DependsOn) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(info.DependsOn))
	}

	if len(info.UnsatisfiedDeps) != 1 {
		t.Errorf("expected 1 unsatisfied dependency, got %d: %v", len(info.UnsatisfiedDeps), info.UnsatisfiedDeps)
	}

	if len(info.UnsatisfiedDeps) > 0 && info.UnsatisfiedDeps[0] != "task-2" {
		t.Errorf("expected unsatisfied dep 'task-2', got '%s'", info.UnsatisfiedDeps[0])
	}

	if info.AllDepsAreSatisfied {
		t.Error("expected AllDepsAreSatisfied=false")
	}
}

func TestTaskGraphManager_EmptyGraph(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewTaskStore(tmpDir)
	manager := NewTaskGraphManager(store)

	// 空のグラフ
	graph, err := manager.BuildGraph()
	if err != nil {
		t.Fatalf("BuildGraph failed: %v", err)
	}

	if len(graph.Nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(graph.Nodes))
	}

	if len(graph.Edges) != 0 {
		t.Errorf("expected 0 edges, got %d", len(graph.Edges))
	}

	// 実行順序
	order, err := manager.GetExecutionOrder()
	if err != nil {
		t.Fatalf("GetExecutionOrder failed: %v", err)
	}

	if len(order) != 0 {
		t.Errorf("expected empty order, got %v", order)
	}

	// サイクル検出
	hasCycle, _, err := manager.DetectCycle()
	if err != nil {
		t.Fatalf("DetectCycle failed: %v", err)
	}

	if hasCycle {
		t.Error("expected no cycle in empty graph")
	}
}
