package persistence

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// --- Interfaces ---

type DesignRepository interface {
	LoadWBS() (*WBS, error)
	SaveWBS(wbs *WBS) error
	GetNode(nodeID string) (*NodeDesign, error)
	SaveNode(node *NodeDesign) error
}

type StateRepository interface {
	LoadNodesRuntime() (*NodesRuntime, error)
	SaveNodesRuntime(state *NodesRuntime) error
	LoadTasks() (*TasksState, error)
	SaveTasks(state *TasksState) error
	LoadAgents() (*AgentsState, error)
	SaveAgents(state *AgentsState) error
}

type HistoryRepository interface {
	AppendAction(action *Action) error
	ListActions(from, to time.Time) ([]Action, error)
}

type WorkspaceRepository interface {
	Init() error
	Design() DesignRepository
	State() StateRepository
	History() HistoryRepository
	Snapshot() SnapshotRepository
	BaseDir() string
}

// --- Implementations ---

type workspaceRepoImpl struct {
	baseDir  string
	design   *designRepoImpl
	state    *stateRepoImpl
	history  *historyRepoImpl
	snapshot *snapshotRepoImpl
}

func NewWorkspaceRepository(baseDir string) WorkspaceRepository {
	return &workspaceRepoImpl{
		baseDir: baseDir,
		design:  &designRepoImpl{baseDir: filepath.Join(baseDir, "design")},
		state:   &stateRepoImpl{baseDir: filepath.Join(baseDir, "state")},
		history: &historyRepoImpl{baseDir: filepath.Join(baseDir, "history")},
		snapshot: &snapshotRepoImpl{
			baseDir:  filepath.Join(baseDir, "snapshots"),
			stateDir: filepath.Join(baseDir, "state"),
		},
	}
}

func (r *workspaceRepoImpl) Init() error {
	dirs := []string{
		r.design.baseDir,
		filepath.Join(r.design.baseDir, "nodes"),
		r.state.baseDir,
		r.history.baseDir,
		filepath.Join(r.baseDir, "snapshots"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	return nil
}

func (r *workspaceRepoImpl) Design() DesignRepository     { return r.design }
func (r *workspaceRepoImpl) State() StateRepository       { return r.state }
func (r *workspaceRepoImpl) History() HistoryRepository   { return r.history }
func (r *workspaceRepoImpl) Snapshot() SnapshotRepository { return r.snapshot }
func (r *workspaceRepoImpl) BaseDir() string              { return r.baseDir }

// --- Design Repo ---

type designRepoImpl struct {
	baseDir string
}

func (r *designRepoImpl) LoadWBS() (*WBS, error) {
	path := filepath.Join(r.baseDir, "wbs.json")
	var wbs WBS
	if err := readJSON(path, &wbs); err != nil {
		return nil, err
	}
	return &wbs, nil
}

func (r *designRepoImpl) SaveWBS(wbs *WBS) error {
	path := filepath.Join(r.baseDir, "wbs.json")
	return writeJSON(path, wbs)
}

func (r *designRepoImpl) GetNode(nodeID string) (*NodeDesign, error) {
	path := filepath.Join(r.baseDir, "nodes", nodeID+".json")
	var node NodeDesign
	if err := readJSON(path, &node); err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *designRepoImpl) SaveNode(node *NodeDesign) error {
	path := filepath.Join(r.baseDir, "nodes", node.NodeID+".json")
	return writeJSON(path, node)
}

// --- State Repo ---

type stateRepoImpl struct {
	baseDir string
}

func (r *stateRepoImpl) LoadNodesRuntime() (*NodesRuntime, error) {
	path := filepath.Join(r.baseDir, "nodes-runtime.json")
	var s NodesRuntime
	if err := readJSON(path, &s); err != nil {
		if os.IsNotExist(err) {
			return &NodesRuntime{Nodes: []NodeRuntime{}}, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *stateRepoImpl) SaveNodesRuntime(state *NodesRuntime) error {
	path := filepath.Join(r.baseDir, "nodes-runtime.json")
	return writeJSON(path, state)
}

func (r *stateRepoImpl) LoadTasks() (*TasksState, error) {
	path := filepath.Join(r.baseDir, "tasks.json")
	var s TasksState
	if err := readJSON(path, &s); err != nil {
		if os.IsNotExist(err) {
			return &TasksState{Tasks: []TaskState{}}, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *stateRepoImpl) SaveTasks(state *TasksState) error {
	path := filepath.Join(r.baseDir, "tasks.json")
	return writeJSON(path, state)
}

func (r *stateRepoImpl) LoadAgents() (*AgentsState, error) {
	path := filepath.Join(r.baseDir, "agents.json")
	var s AgentsState
	if err := readJSON(path, &s); err != nil {
		if os.IsNotExist(err) {
			return &AgentsState{Agents: []AgentState{}}, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *stateRepoImpl) SaveAgents(state *AgentsState) error {
	path := filepath.Join(r.baseDir, "agents.json")
	return writeJSON(path, state)
}

// --- History Repo ---

type historyRepoImpl struct {
	baseDir string
}

func (r *historyRepoImpl) AppendAction(action *Action) error {
	filename := fmt.Sprintf("actions-%s.jsonl", action.At.Format("20060102"))
	path := filepath.Join(r.baseDir, filename)

	bytes, err := json.Marshal(action)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(bytes); err != nil {
		return err
	}
	if _, err := f.WriteString("\n"); err != nil {
		return err
	}
	return nil
}

func (r *historyRepoImpl) ListActions(from, to time.Time) ([]Action, error) {
	var actions []Action

	// Find all action files
	pattern := filepath.Join(r.baseDir, "actions-*.jsonl")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	for _, path := range files {
		// Optimization: Parse date from filename and check range before opening?
		// Filename format: actions-20060102.jsonl
		// ... Skip for now, simpler to just read valid JSONL files.

		f, err := os.Open(path)
		if err != nil {
			continue // Skip unreadable
		}

		dec := json.NewDecoder(f)
		for {
			var a Action
			if err := dec.Decode(&a); err != nil {
				if err == io.EOF {
					break
				}
				// Skip malformed lines? Or fail?
				// Continuing is safer for history reading.
				continue
			}
			if (a.At.Equal(from) || a.At.After(from)) && (a.At.Equal(to) || a.At.Before(to)) {
				actions = append(actions, a)
			}
		}
		_ = f.Close()
	}

	// Sort by time
	sort.Slice(actions, func(i, j int) bool {
		return actions[i].At.Before(actions[j].At)
	})

	return actions, nil
}

// --- Helper Functions ---

func readJSON(path string, v interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return json.NewDecoder(f).Decode(v)
}

func writeJSON(path string, v interface{}) error {
	tmpPath := path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return err
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	_ = f.Close() // Close before rename (windows compat/correctness)

	return os.Rename(tmpPath, path)
}
