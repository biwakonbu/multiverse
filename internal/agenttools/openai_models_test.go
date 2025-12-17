package agenttools

import "testing"

func TestKnownOpenAIModels_IncludesDefaultsAndMini(t *testing.T) {
	want := []string{
		DefaultMetaModel,
		DefaultCodexModel,
		"gpt-5.1-codex-mini",
	}

	set := make(map[string]struct{}, len(KnownOpenAIModels))
	for _, m := range KnownOpenAIModels {
		set[m.ID] = struct{}{}
	}

	for _, id := range want {
		if _, ok := set[id]; !ok {
			t.Fatalf("KnownOpenAIModels is missing %q", id)
		}
	}
}

func TestResolveOpenAIModelID_ResolvesAlias(t *testing.T) {
	if got := ResolveOpenAIModelID("5.1-codex-mini"); got != "gpt-5.1-codex-mini" {
		t.Fatalf("ResolveOpenAIModelID() = %q, want %q", got, "gpt-5.1-codex-mini")
	}
}
