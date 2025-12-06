package orchestrator

import (
	"fmt"
	"sort"
)

// GraphNode は依存グラフ内のノードを表す
type GraphNode struct {
	Task       *Task    // タスク本体
	InDegree   int      // 入次数（このタスクに依存しているタスク数）
	OutDegree  int      // 出次数（このタスクが依存しているタスク数）
	Dependents []string // このタスクに依存しているタスクID
}

// TaskEdge はタスク間の依存関係エッジを表す
type TaskEdge struct {
	From      string // 依存元タスクID（先に実行すべきタスク）
	To        string // 依存先タスクID（後に実行すべきタスク）
	Satisfied bool   // 依存が満たされているか（Fromタスクが完了済みか）
}

// TaskGraph はタスク依存関係グラフを表す
type TaskGraph struct {
	Nodes map[string]*GraphNode // タスクID -> ノード
	Edges []TaskEdge            // 全エッジのリスト
}

// TaskGraphManager はタスク依存グラフを管理する
type TaskGraphManager struct {
	TaskStore *TaskStore
}

// NewTaskGraphManager は新しい TaskGraphManager を作成する
func NewTaskGraphManager(taskStore *TaskStore) *TaskGraphManager {
	return &TaskGraphManager{TaskStore: taskStore}
}

// BuildGraph は全タスクから依存グラフを構築する
func (m *TaskGraphManager) BuildGraph() (*TaskGraph, error) {
	tasks, err := m.TaskStore.ListAllTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	graph := &TaskGraph{
		Nodes: make(map[string]*GraphNode),
		Edges: []TaskEdge{},
	}

	// ノードを作成
	for i := range tasks {
		task := &tasks[i]
		graph.Nodes[task.ID] = &GraphNode{
			Task:       task,
			InDegree:   0,
			OutDegree:  len(task.Dependencies),
			Dependents: []string{},
		}
	}

	// エッジを構築し、次数を計算
	completedStatuses := map[TaskStatus]bool{
		TaskStatusSucceeded: true,
		TaskStatusCanceled:  true,
	}

	for _, task := range tasks {
		if len(task.Dependencies) == 0 {
			continue
		}

		for _, depID := range task.Dependencies {
			// 依存先ノードが存在するか確認
			depNode, exists := graph.Nodes[depID]
			if !exists {
				// 依存先が存在しない場合はエッジを作成しない
				continue
			}

			// 依存が満たされているか判定
			satisfied := completedStatuses[depNode.Task.Status]

			edge := TaskEdge{
				From:      depID,
				To:        task.ID,
				Satisfied: satisfied,
			}
			graph.Edges = append(graph.Edges, edge)

			// 入次数を更新
			if node, ok := graph.Nodes[task.ID]; ok {
				node.InDegree++
			}

			// Dependents を更新
			depNode.Dependents = append(depNode.Dependents, task.ID)
		}
	}

	return graph, nil
}

// GetExecutionOrder はトポロジカルソートによる実行順序を返す
// サイクルがある場合はエラーを返す
func (m *TaskGraphManager) GetExecutionOrder() ([]string, error) {
	graph, err := m.BuildGraph()
	if err != nil {
		return nil, err
	}

	// Kahn's algorithm によるトポロジカルソート
	// 入次数が0のノードからスタート
	inDegree := make(map[string]int)
	for id, node := range graph.Nodes {
		inDegree[id] = node.InDegree
	}

	// 入次数0のノードをキューに追加
	queue := []string{}
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	// 安定したソート順序のためにソート
	sort.Strings(queue)

	result := []string{}
	for len(queue) > 0 {
		// キューの先頭を取り出す
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// 依存しているノードの入次数を減らす
		node := graph.Nodes[current]
		if node == nil {
			continue
		}

		newQueue := []string{}
		for _, depID := range node.Dependents {
			inDegree[depID]--
			if inDegree[depID] == 0 {
				newQueue = append(newQueue, depID)
			}
		}

		// 安定したソート順序のためにソート
		sort.Strings(newQueue)
		queue = append(queue, newQueue...)
	}

	// サイクル検出: 全ノードが結果に含まれていなければサイクルあり
	if len(result) != len(graph.Nodes) {
		return nil, fmt.Errorf("cycle detected in task dependencies")
	}

	return result, nil
}

// GetBlockedTasks は依存が満たされていないタスクのIDリストを返す
func (m *TaskGraphManager) GetBlockedTasks() ([]string, error) {
	graph, err := m.BuildGraph()
	if err != nil {
		return nil, err
	}

	blocked := []string{}
	for _, edge := range graph.Edges {
		if !edge.Satisfied {
			blocked = append(blocked, edge.To)
		}
	}

	// 重複を除去
	seen := make(map[string]bool)
	unique := []string{}
	for _, id := range blocked {
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}

	sort.Strings(unique)
	return unique, nil
}

// GetReadyTasks は実行可能なタスク（PENDING で全依存が満たされている）のIDリストを返す
func (m *TaskGraphManager) GetReadyTasks() ([]string, error) {
	graph, err := m.BuildGraph()
	if err != nil {
		return nil, err
	}

	blockedSet := make(map[string]bool)
	for _, edge := range graph.Edges {
		if !edge.Satisfied {
			blockedSet[edge.To] = true
		}
	}

	ready := []string{}
	for id, node := range graph.Nodes {
		if node.Task.Status == TaskStatusPending && !blockedSet[id] {
			ready = append(ready, id)
		}
	}

	sort.Strings(ready)
	return ready, nil
}

// DetectCycle はグラフにサイクルがあるかどうかを検出する
func (m *TaskGraphManager) DetectCycle() (bool, []string, error) {
	_, err := m.GetExecutionOrder()
	if err != nil {
		if err.Error() == "cycle detected in task dependencies" {
			// サイクルに関与しているノードを特定
			graph, buildErr := m.BuildGraph()
			if buildErr != nil {
				return true, nil, buildErr
			}

			// 入次数が残っているノードがサイクルに関与
			inDegree := make(map[string]int)
			for id, node := range graph.Nodes {
				inDegree[id] = node.InDegree
			}

			// Kahn's algorithm を再実行して残りを特定
			queue := []string{}
			for id, deg := range inDegree {
				if deg == 0 {
					queue = append(queue, id)
				}
			}

			for len(queue) > 0 {
				current := queue[0]
				queue = queue[1:]

				node := graph.Nodes[current]
				if node == nil {
					continue
				}

				for _, depID := range node.Dependents {
					inDegree[depID]--
					if inDegree[depID] == 0 {
						queue = append(queue, depID)
					}
				}
			}

			// 入次数が残っているノードがサイクルに関与
			cycleNodes := []string{}
			for id, deg := range inDegree {
				if deg > 0 {
					cycleNodes = append(cycleNodes, id)
				}
			}

			sort.Strings(cycleNodes)
			return true, cycleNodes, nil
		}
		return false, nil, err
	}

	return false, nil, nil
}

// GetTaskDependencyInfo は特定タスクの依存情報を返す
func (m *TaskGraphManager) GetTaskDependencyInfo(taskID string) (*TaskDependencyInfo, error) {
	graph, err := m.BuildGraph()
	if err != nil {
		return nil, err
	}

	node, exists := graph.Nodes[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	info := &TaskDependencyInfo{
		TaskID:              taskID,
		DependsOn:           node.Task.Dependencies,
		DependedBy:          node.Dependents,
		UnsatisfiedDeps:     []string{},
		AllDepsAreSatisfied: true,
	}

	// 満たされていない依存をチェック
	completedStatuses := map[TaskStatus]bool{
		TaskStatusSucceeded: true,
		TaskStatusCanceled:  true,
	}

	for _, depID := range node.Task.Dependencies {
		depNode, exists := graph.Nodes[depID]
		if !exists || !completedStatuses[depNode.Task.Status] {
			info.UnsatisfiedDeps = append(info.UnsatisfiedDeps, depID)
			info.AllDepsAreSatisfied = false
		}
	}

	return info, nil
}

// TaskDependencyInfo は特定タスクの依存関係情報を表す
type TaskDependencyInfo struct {
	TaskID              string   // タスクID
	DependsOn           []string // このタスクが依存しているタスクID
	DependedBy          []string // このタスクに依存しているタスクID
	UnsatisfiedDeps     []string // まだ完了していない依存タスクID
	AllDepsAreSatisfied bool     // 全ての依存が満たされているか
}
