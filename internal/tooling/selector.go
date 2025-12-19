package tooling

import (
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/biwakonbu/agent-runner/pkg/config"
)

const (
	CategoryMeta      = "meta"
	CategoryTask      = "task"
	CategoryPlan      = "plan"
	CategoryExecution = "execution"
	CategoryWorker    = "worker"
)

// Selector picks tool candidates based on profiles, weights, and cooldowns.
type Selector struct {
	cfg       *config.ToolingConfig
	profile   *config.ToolProfile
	cooldowns map[string]map[string]time.Time
	lastIndex map[string]int
	rng       *rand.Rand
	mu        sync.Mutex
}

func NewSelector(cfg *config.ToolingConfig) *Selector {
	s := &Selector{
		cfg:       cfg,
		cooldowns: make(map[string]map[string]time.Time),
		lastIndex: make(map[string]int),
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	s.profile = s.resolveProfile()
	return s
}

func (s *Selector) resolveProfile() *config.ToolProfile {
	if s.cfg == nil {
		return nil
	}
	if s.cfg.ActiveProfile != "" {
		for i := range s.cfg.Profiles {
			if s.cfg.Profiles[i].ID == s.cfg.ActiveProfile {
				return &s.cfg.Profiles[i]
			}
		}
	}
	if len(s.cfg.Profiles) > 0 {
		return &s.cfg.Profiles[0]
	}
	return nil
}

func (s *Selector) ForceCandidate() (config.ToolCandidate, bool) {
	if s.cfg == nil || !s.cfg.Force.Enabled {
		return config.ToolCandidate{}, false
	}
	if s.cfg.Force.Tool == "" {
		return config.ToolCandidate{}, false
	}
	return config.ToolCandidate{
		Tool:  s.cfg.Force.Tool,
		Model: s.cfg.Force.Model,
	}, true
}

func (s *Selector) Category(category string) (*config.ToolCategoryConfig, bool) {
	if s.profile == nil {
		return nil, false
	}
	if s.profile.Categories == nil {
		return nil, false
	}
	if cfg, ok := s.profile.Categories[category]; ok {
		return &cfg, true
	}
	if category != CategoryMeta {
		if cfg, ok := s.profile.Categories[CategoryMeta]; ok {
			return &cfg, true
		}
	}
	return nil, false
}

// Select picks a candidate for the category (force overrides apply).
func (s *Selector) Select(category string) (config.ToolCandidate, bool) {
	if forced, ok := s.ForceCandidate(); ok {
		return forced, true
	}
	cfg, ok := s.Category(category)
	if !ok || len(cfg.Candidates) == 0 {
		return config.ToolCandidate{}, false
	}
	candidates := s.availableCandidates(category, cfg.Candidates)
	if len(candidates) == 0 {
		return config.ToolCandidate{}, false
	}
	switch strings.ToLower(cfg.Strategy) {
	case "round_robin":
		return s.pickRoundRobin(category, candidates), true
	default:
		return s.pickWeighted(candidates), true
	}
}

func (s *Selector) availableCandidates(category string, candidates []config.ToolCandidate) []config.ToolCandidate {
	now := time.Now()
	var filtered []config.ToolCandidate
	for _, c := range candidates {
		if !isCandidateAvailable(c) {
			continue
		}
		key := candidateKey(c)
		if until, ok := s.cooldowns[category][key]; ok && until.After(now) {
			continue
		}
		filtered = append(filtered, c)
	}
	return filtered
}

func (s *Selector) pickWeighted(candidates []config.ToolCandidate) config.ToolCandidate {
	total := 0
	for _, c := range candidates {
		w := c.Weight
		if w <= 0 {
			w = 1
		}
		total += w
	}
	if total <= 0 {
		return candidates[0]
	}
	r := s.rng.Intn(total)
	acc := 0
	for _, c := range candidates {
		w := c.Weight
		if w <= 0 {
			w = 1
		}
		acc += w
		if r < acc {
			return c
		}
	}
	return candidates[0]
}

func (s *Selector) pickRoundRobin(category string, candidates []config.ToolCandidate) config.ToolCandidate {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := s.lastIndex[category]
	if idx >= len(candidates) {
		idx = 0
	}
	selected := candidates[idx]
	s.lastIndex[category] = (idx + 1) % len(candidates)
	return selected
}

func (s *Selector) MarkRateLimited(category string, candidate config.ToolCandidate, cooldownSec int) {
	if cooldownSec <= 0 {
		cooldownSec = 120
	}
	if s.cooldowns[category] == nil {
		s.cooldowns[category] = make(map[string]time.Time)
	}
	s.cooldowns[category][candidateKey(candidate)] = time.Now().Add(time.Duration(cooldownSec) * time.Second)
}

func (s *Selector) ShouldFallbackOnRateLimit(category string) bool {
	cfg, ok := s.Category(category)
	if !ok {
		return false
	}
	if !cfg.FallbackOnRateLimit {
		return false
	}
	return true
}

func (s *Selector) CooldownSec(category string) int {
	cfg, ok := s.Category(category)
	if !ok {
		return 0
	}
	return cfg.CooldownSec
}

func candidateKey(c config.ToolCandidate) string {
	if c.Model == "" {
		return c.Tool
	}
	return c.Tool + ":" + c.Model
}

func isCandidateAvailable(c config.ToolCandidate) bool {
	tool := strings.ToLower(strings.TrimSpace(c.Tool))
	switch tool {
	case "openai-chat":
		return os.Getenv("OPENAI_API_KEY") != ""
	case "mock":
		return true
	default:
		path := strings.TrimSpace(c.CLIPath)
		if path == "" {
			path = defaultCLIPath(tool)
		}
		if path == "" {
			return false
		}
		_, err := exec.LookPath(path)
		return err == nil
	}
}

func defaultCLIPath(tool string) string {
	switch tool {
	case "codex-cli":
		return "codex"
	case "claude-code", "claude-code-cli":
		return "claude"
	case "gemini-cli":
		return "gemini"
	case "cursor-cli":
		return "cursor"
	default:
		return tool
	}
}
