package agenttools

import "testing"

func TestKnownClaudeModels_IncludesClaude45Family(t *testing.T) {
	want := []string{
		"claude-haiku-4-5",
		"claude-sonnet-4-5",
		"claude-opus-4-5",
	}

	set := make(map[string]struct{}, len(KnownClaudeModels))
	for _, m := range KnownClaudeModels {
		set[m] = struct{}{}
	}

	for _, m := range want {
		if _, ok := set[m]; !ok {
			t.Fatalf("KnownClaudeModels is missing %q", m)
		}
	}
}

func TestDefaultClaudeModel_IsInKnownClaudeModels(t *testing.T) {
	set := make(map[string]struct{}, len(KnownClaudeModels))
	for _, m := range KnownClaudeModels {
		set[m] = struct{}{}
	}
	if _, ok := set[DefaultClaudeModel]; !ok {
		t.Fatalf("DefaultClaudeModel %q is not in KnownClaudeModels", DefaultClaudeModel)
	}
}
