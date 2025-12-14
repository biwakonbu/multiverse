package meta

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// extractJSON extracts JSON content from LLM response, handling markdown code blocks
// and Codex CLI output which includes header information before the JSON
func extractJSON(response string) string {
	response = strings.TrimSpace(response)

	// Method 1: Try to extract from markdown code block (```json ... ```)
	reMarkdown := regexp.MustCompile("(?s)```json\\s*\\n(.+?)\\n```")
	matches := reMarkdown.FindStringSubmatch(response)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Method 2: Try generic code block extraction (``` ... ```)
	reGeneric := regexp.MustCompile("(?s)```\\s*\\n(.+?)\\n```")
	matches = reGeneric.FindStringSubmatch(response)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Method 3: Strip leading/trailing backticks if present
	if strings.HasPrefix(response, "```") && strings.HasSuffix(response, "```") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		return strings.TrimSpace(response)
	}

	// Method 4: Extract JSON object starting with "{" from Codex CLI output
	// Codex CLI includes header info (version, workdir, model, etc.) before the actual JSON
	// Look for the first "{" that starts a JSON object
	if idx := strings.Index(response, "{"); idx >= 0 {
		// Find the matching closing brace
		jsonStr := response[idx:]
		// Validate it's actually JSON by finding balanced braces
		braceCount := 0
		endIdx := -1
		for i, ch := range jsonStr {
			if ch == '{' {
				braceCount++
			} else if ch == '}' {
				braceCount--
				if braceCount == 0 {
					endIdx = i + 1
					break
				}
			}
		}
		if endIdx > 0 {
			return strings.TrimSpace(jsonStr[:endIdx])
		}
	}

	return response
}

// extractYAML extracts YAML content from LLM response, handling markdown code blocks
// and Codex CLI output which includes header information before the YAML
func extractYAML(response string) string {
	response = strings.TrimSpace(response)

	// Method 1: Try to extract from markdown code block (```yaml ... ```)
	reMarkdown := regexp.MustCompile("(?s)```ya?ml\\s*\\n(.+?)\\n```")
	matches := reMarkdown.FindStringSubmatch(response)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Method 2: Try generic code block extraction (``` ... ```)
	reGeneric := regexp.MustCompile("(?s)```\\s*\\n(.+?)\\n```")
	matches = reGeneric.FindStringSubmatch(response)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Method 3: Strip leading/trailing backticks if present (e.g. `yaml ... ` or ``` ... ``` without newlines)
	// This handles cases where LLM might output inline code or malformed blocks
	if strings.HasPrefix(response, "```") && strings.HasSuffix(response, "```") {
		response = strings.TrimPrefix(response, "```yaml")
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		return strings.TrimSpace(response)
	}

	// Method 4: Extract YAML starting with "type:" from Codex CLI output
	// Codex CLI includes header info (version, workdir, model, etc.) before the actual YAML
	// Look for "type: " at the beginning of a line and extract from there
	reTypeYAML := regexp.MustCompile(`(?m)^type:\s+\w+`)
	loc := reTypeYAML.FindStringIndex(response)
	if loc != nil {
		return strings.TrimSpace(response[loc[0]:])
	}

	return response
}

// jsonToYAML translates JSON string to YAML string
// This is used to maintain compatibility with existing YAML parsing logic
func jsonToYAML(jsonStr string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON for conversion: %w", err)
	}

	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML for conversion: %w", err)
	}

	return string(yamlBytes), nil
}

// buildDecomposeUserPrompt builds the user prompt for decompose request
func buildDecomposeUserPrompt(req *DecomposeRequest) string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "User Input:\n%s\n\n", req.UserInput)

	fmt.Fprintf(b, "Context:\n")
	fmt.Fprintf(b, "Workspace: %s\n", req.Context.WorkspacePath)

	if len(req.Context.ExistingTasks) > 0 {
		fmt.Fprintf(b, "Existing Tasks:\n")
		for _, t := range req.Context.ExistingTasks {
			fmt.Fprintf(b, "- %s: %s (%s)\n", t.ID, t.Title, t.Status)
		}
	}

	if len(req.Context.ConversationHistory) > 0 {
		fmt.Fprintf(b, "\nConversation History:\n")
		for _, msg := range req.Context.ConversationHistory {
			// Limit content length for history to avoid hitting token limits
			content := msg.Content
			if len(content) > 300 {
				content = content[:300] + "..."
			}
			fmt.Fprintf(b, "- [%s] %s\n", msg.Role, content)
		}
	}

	return b.String()
}

// statusPriority returns priority for deterministic sorting (lower = higher priority)
// PRD 13.3 #2: RUNNING > BLOCKED > PENDING/READY > others
func statusPriority(status string) int {
	switch status {
	case "RUNNING":
		return 0
	case "BLOCKED":
		return 1
	case "PENDING", "READY":
		return 2
	default:
		return 3 // SUCCEEDED, FAILED, COMPLETED, etc.
	}
}

// buildPlanPatchUserPrompt builds the user prompt for plan patch request
// QH-001: Includes full structured context (WBS node_index, conversation history, task details)
func buildPlanPatchUserPrompt(req *PlanPatchRequest) string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "User Input:\n%s\n\n", req.UserInput)

	fmt.Fprintf(b, "Context:\n")

	// Existing tasks with full facet information (max 200 tasks)
	// PRD 13.3 #2: 決定論的ソート (status優先 + ID昇順)
	if len(req.Context.ExistingTasks) > 0 {
		fmt.Fprintf(b, "Existing Tasks:\n")
		tasks := req.Context.ExistingTasks
		if len(tasks) > 200 {
			// Sort by status priority (RUNNING > BLOCKED > PENDING > SUCCEEDED) + ID
			sortedTasks := make([]ExistingTaskSummary, len(tasks))
			copy(sortedTasks, tasks)
			sort.SliceStable(sortedTasks, func(i, j int) bool {
				pi, pj := statusPriority(sortedTasks[i].Status), statusPriority(sortedTasks[j].Status)
				if pi != pj {
					return pi < pj
				}
				return sortedTasks[i].ID < sortedTasks[j].ID
			})
			tasks = sortedTasks[:200]
			fmt.Fprintf(b, "(showing first 200 of %d tasks, prioritized by status)\n", len(req.Context.ExistingTasks))
		}
		for _, t := range tasks {
			deps := "none"
			if len(t.Dependencies) > 0 {
				deps = strings.Join(t.Dependencies, ",")
			}
			parent := "root"
			if t.ParentID != nil && *t.ParentID != "" {
				parent = *t.ParentID
			}
			fmt.Fprintf(b, "- %s: %s (%s) [phase=%s, milestone=%s, level=%d, deps=%s, parent=%s]\n",
				t.ID, t.Title, t.Status, t.PhaseName, t.Milestone, t.WBSLevel, deps, parent)
		}
	}

	// WBS structure with node_index (PRD 12.1 / meta-protocol.md 10.2)
	// QH-001: Deterministic trimming to max 200 nodes via BFS from root
	if req.Context.ExistingWBS != nil {
		nodes := req.Context.ExistingWBS.NodeIndex
		const maxWBSNodes = 200
		if len(nodes) > maxWBSNodes {
			nodes = trimWBSNodesBFS(nodes, req.Context.ExistingWBS.RootNodeID, maxWBSNodes)
			fmt.Fprintf(b, "\nWBS Structure (Root: %s, showing %d of %d nodes):\n",
				req.Context.ExistingWBS.RootNodeID, len(nodes), len(req.Context.ExistingWBS.NodeIndex))
		} else {
			fmt.Fprintf(b, "\nWBS Structure (Root: %s):\n", req.Context.ExistingWBS.RootNodeID)
		}
		for _, n := range nodes {
			parent := "root"
			if n.ParentID != nil && *n.ParentID != "" {
				parent = *n.ParentID
			}
			children := "none"
			if len(n.Children) > 0 {
				children = strings.Join(n.Children, ",")
			}
			fmt.Fprintf(b, "  - %s: parent=%s, children=[%s]\n", n.NodeID, parent, children)
		}
	}

	// Conversation history (max 10 messages, each truncated to 300 chars)
	if len(req.Context.ConversationHistory) > 0 {
		fmt.Fprintf(b, "\nConversation History:\n")
		msgs := req.Context.ConversationHistory
		if len(msgs) > 10 {
			msgs = msgs[len(msgs)-10:]
		}
		for _, m := range msgs {
			content := m.Content
			if len(content) > 300 {
				content = content[:300] + "..."
			}
			fmt.Fprintf(b, "- [%s] %s\n", m.Role, content)
		}
	}

	return b.String()
}

// trimWBSNodesBFS trims WBS nodes to a maximum count using BFS from root.
// QH-001 (PRD 13.3 #1): Deterministic subset selection for large WBS.
// Returns nodes in BFS order, ensuring root and its descendants are prioritized.
func trimWBSNodesBFS(nodes []WBSNodeIndex, rootNodeID string, maxNodes int) []WBSNodeIndex {
	if len(nodes) <= maxNodes {
		return nodes
	}

	// Build lookup maps
	nodeByID := make(map[string]WBSNodeIndex, len(nodes))
	for _, n := range nodes {
		nodeByID[n.NodeID] = n
	}

	// BFS from root
	result := make([]WBSNodeIndex, 0, maxNodes)
	seen := make(map[string]struct{})
	queue := []string{rootNodeID}

	for len(queue) > 0 && len(result) < maxNodes {
		nodeID := queue[0]
		queue = queue[1:]

		if _, ok := seen[nodeID]; ok {
			continue
		}
		seen[nodeID] = struct{}{}

		node, exists := nodeByID[nodeID]
		if !exists {
			continue
		}
		result = append(result, node)

		// Add children to queue (maintains child order for determinism)
		for _, childID := range node.Children {
			if _, ok := seen[childID]; !ok {
				queue = append(queue, childID)
			}
		}
	}

	return result
}
