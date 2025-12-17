package agenttools

import (
	"context"
	"strings"
	"testing"
)

func TestRegistry_Implementations(t *testing.T) {
	implementedKinds := []string{"claude-code", "claude-code-cli", "cursor-cli"}

	for _, kind := range implementedKinds {
		t.Run(kind, func(t *testing.T) {
			// 1. Verify we can create the provider
			p, err := New(kind, ProviderConfig{Kind: kind})
			if err != nil {
				t.Fatalf("New(%q) failed: %v", kind, err)
			}

			if p.Kind() != kind {
				t.Errorf("Kind() = %q, want %q", p.Kind(), kind)
			}

			// 2. Verify Build returns a valid plan (not "not implemented" error)
			ctx := context.Background()
			req := Request{Prompt: "hello"}
			plan, err := p.Build(ctx, req)
			if err != nil {
				t.Fatalf("Build() failed: %v (expected success for implemented provider)", err)
			}

			// 3. Verify basic plan structure
			if plan.Command == "" {
				t.Error("Build() returned empty Command")
			}
		})
	}
}

func TestRegistry_Codex(t *testing.T) {
	// Verify Codex is also registered
	kind := "codex-cli"
	p, err := New(kind, ProviderConfig{Kind: kind})
	if err != nil {
		t.Fatalf("New(%q) failed: %v", kind, err)
	}
	if p.Kind() != kind {
		t.Errorf("Kind() = %q, want %q", p.Kind(), kind)
	}
}

func TestCodexProvider_Build_DefaultFlags(t *testing.T) {
	p := NewCodexProvider(ProviderConfig{Kind: "codex-cli"})

	ctx := context.Background()
	req := Request{Prompt: "test prompt"}
	plan, err := p.Build(ctx, req)
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// コマンドが codex であること
	if plan.Command != "codex" {
		t.Errorf("Command = %q, want %q", plan.Command, "codex")
	}

	// 必須フラグの検証
	args := strings.Join(plan.Args, " ")

	// exec サブコマンド
	if !strings.Contains(args, "exec") {
		t.Error("Args should contain 'exec' subcommand")
	}

	// サンドボックス無効化フラグ
	if !strings.Contains(args, "--dangerously-bypass-approvals-and-sandbox") {
		t.Error("Args should contain '--dangerously-bypass-approvals-and-sandbox'")
	}

	// 作業ディレクトリ（-C フラグ）
	if !strings.Contains(args, "-C /workspace/project") {
		t.Errorf("Args should contain '-C /workspace/project', got: %s", args)
	}

	// JSON 出力
	if !strings.Contains(args, "--json") {
		t.Error("Args should contain '--json'")
	}

	// デフォルトモデル
	if !strings.Contains(args, "-m gpt-5.1-codex") {
		t.Errorf("Args should contain '-m gpt-5.1-codex', got: %s", args)
	}

	// デフォルト思考の深さ
	if !strings.Contains(args, "-c reasoning_effort=medium") {
		t.Errorf("Args should contain '-c reasoning_effort=medium', got: %s", args)
	}

	// プロンプトが最後にあること
	if !strings.HasSuffix(args, "test prompt") {
		t.Errorf("Args should end with prompt, got: %s", args)
	}
}

func TestCodexProvider_Build_CustomModel(t *testing.T) {
	p := NewCodexProvider(ProviderConfig{Kind: "codex-cli"})

	ctx := context.Background()
	req := Request{
		Prompt: "test",
		Model:  "o3",
	}
	plan, err := p.Build(ctx, req)
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	args := strings.Join(plan.Args, " ")
	if !strings.Contains(args, "-m o3") {
		t.Errorf("Args should contain '-m o3', got: %s", args)
	}
}

func TestCodexProvider_Build_ModelAliasIsNormalized(t *testing.T) {
	p := NewCodexProvider(ProviderConfig{Kind: "codex-cli"})

	ctx := context.Background()
	req := Request{
		Prompt: "test",
		Model:  "5.1-codex-mini",
	}
	plan, err := p.Build(ctx, req)
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	args := strings.Join(plan.Args, " ")
	if !strings.Contains(args, "-m gpt-5.1-codex-mini") {
		t.Errorf("Args should contain '-m gpt-5.1-codex-mini', got: %s", args)
	}
}

func TestCodexProvider_Build_ReasoningEffort(t *testing.T) {
	p := NewCodexProvider(ProviderConfig{Kind: "codex-cli"})

	ctx := context.Background()
	req := Request{
		Prompt:          "test",
		ReasoningEffort: "high",
	}
	plan, err := p.Build(ctx, req)
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	args := strings.Join(plan.Args, " ")
	if !strings.Contains(args, "-c reasoning_effort=high") {
		t.Errorf("Args should contain '-c reasoning_effort=high', got: %s", args)
	}
}

func TestCodexProvider_Build_InvalidReasoningEffortFallback(t *testing.T) {
	p := NewCodexProvider(ProviderConfig{Kind: "codex-cli"})

	ctx := context.Background()

	t.Run("xhigh is clamped to high", func(t *testing.T) {
		req := Request{
			Prompt:          "test",
			ReasoningEffort: "xhigh",
		}
		plan, err := p.Build(ctx, req)
		if err != nil {
			t.Fatalf("Build() failed: %v", err)
		}
		args := strings.Join(plan.Args, " ")
		if !strings.Contains(args, "-c reasoning_effort=high") {
			t.Errorf("Args should contain '-c reasoning_effort=high', got: %s", args)
		}
	})

	t.Run("unknown value falls back to medium", func(t *testing.T) {
		req := Request{
			Prompt:          "test",
			ReasoningEffort: "weird",
		}
		plan, err := p.Build(ctx, req)
		if err != nil {
			t.Fatalf("Build() failed: %v", err)
		}
		args := strings.Join(plan.Args, " ")
		if !strings.Contains(args, "-c reasoning_effort=medium") {
			t.Errorf("Args should contain '-c reasoning_effort=medium', got: %s", args)
		}
	})
}

func TestCodexProvider_Build_Stdin(t *testing.T) {
	p := NewCodexProvider(ProviderConfig{Kind: "codex-cli"})

	ctx := context.Background()
	req := Request{
		Prompt:   "test prompt via stdin",
		UseStdin: true,
	}
	plan, err := p.Build(ctx, req)
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// 最後の引数が "-" であること
	lastArg := plan.Args[len(plan.Args)-1]
	if lastArg != "-" {
		t.Errorf("Last arg should be '-', got: %s", lastArg)
	}

	// Stdin にプロンプトが設定されていること
	if plan.Stdin != "test prompt via stdin" {
		t.Errorf("Stdin should be prompt, got: %s", plan.Stdin)
	}
}

func TestCodexProvider_Build_ChatModeRejected(t *testing.T) {
	p := NewCodexProvider(ProviderConfig{Kind: "codex-cli"})

	ctx := context.Background()
	req := Request{
		Prompt: "test",
		Mode:   "chat",
	}
	_, err := p.Build(ctx, req)
	if err == nil {
		t.Fatal("Build() should have failed for chat mode")
	}

	if !strings.Contains(err.Error(), "only 'exec' is supported") {
		t.Errorf("Error should mention exec only, got: %v", err)
	}
}

func TestCodexProvider_Build_NoOldFlags(t *testing.T) {
	p := NewCodexProvider(ProviderConfig{Kind: "codex-cli"})

	ctx := context.Background()
	temp := 0.5
	maxTokens := 1000
	req := Request{
		Prompt:      "test",
		Temperature: &temp,
		MaxTokens:   &maxTokens,
	}
	plan, err := p.Build(ctx, req)
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	args := strings.Join(plan.Args, " ")

	// 古いフラグが存在しないこと
	if strings.Contains(args, "--cwd") {
		t.Error("Args should NOT contain '--cwd' (use '-C' instead)")
	}
	if strings.Contains(args, "--temperature") {
		t.Error("Args should NOT contain '--temperature' (use '-c temperature=X' instead)")
	}
	if strings.Contains(args, "--max-tokens") {
		t.Error("Args should NOT contain '--max-tokens' (use '-c max_tokens=X' instead)")
	}
	if strings.Contains(args, "--stdin") {
		t.Error("Args should NOT contain '--stdin' (use '-' as PROMPT instead)")
	}
	if strings.Contains(args, "--sandbox") {
		t.Error("Args should NOT contain '--sandbox' (use '--dangerously-bypass-approvals-and-sandbox' instead)")
	}

	// 正しいフラグが使われていること
	if !strings.Contains(args, "-c temperature=") {
		t.Errorf("Args should contain '-c temperature=', got: %s", args)
	}
	if !strings.Contains(args, "-c max_tokens=") {
		t.Errorf("Args should contain '-c max_tokens=', got: %s", args)
	}
}

func TestRegistry_Gemini(t *testing.T) {
	kind := "gemini-cli"
	p, err := New(kind, ProviderConfig{Kind: kind})
	if err != nil {
		t.Fatalf("New(%q) failed: %v", kind, err)
	}
	if p.Kind() != kind {
		t.Errorf("Kind() = %q, want %q", p.Kind(), kind)
	}
}

func TestGeminiProvider_Build(t *testing.T) {
	p := NewGeminiProvider(ProviderConfig{Kind: "gemini-cli"})
	ctx := context.Background()
	req := Request{
		Prompt: "hello gemini",
	}
	plan, err := p.Build(ctx, req)
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	if plan.Command != "gemini" {
		t.Errorf("Command = %q, want %q", plan.Command, "gemini")
	}

	args := strings.Join(plan.Args, " ")
	if !strings.Contains(args, "--model gemini-2.5-pro") {
		t.Errorf("Default model missing, got: %s", args)
	}
	if !strings.HasSuffix(args, "hello gemini") {
		t.Errorf("Prompt should be last arg, got: %s", args)
	}
}

func TestGeminiProvider_Build_Options(t *testing.T) {
	p := NewGeminiProvider(ProviderConfig{Kind: "gemini-cli"})
	ctx := context.Background()
	temp := 0.7
	maxTokens := 2048
	req := Request{
		Prompt:      "test",
		Model:       "gemini-ultra",
		Temperature: &temp,
		MaxTokens:   &maxTokens,
		ToolSpecific: map[string]interface{}{
			"json_output": true,
		},
		UseStdin: true,
	}
	plan, err := p.Build(ctx, req)
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	args := strings.Join(plan.Args, " ")

	if !strings.Contains(args, "--model gemini-ultra") {
		t.Errorf("Model override failed, got: %s", args)
	}
	if !strings.Contains(args, "--temperature 0.70") {
		t.Errorf("Temperature failed, got: %s", args)
	}
	if !strings.Contains(args, "--max-output-tokens 2048") {
		t.Errorf("Max tokens failed, got: %s", args)
	}
	if !strings.Contains(args, "--json") {
		t.Errorf("JSON output failed, got: %s", args)
	}

	if plan.Stdin != "test" {
		t.Errorf("Stdin content mismatch, got: %q", plan.Stdin)
	}
	if plan.Args[len(plan.Args)-1] != "-" {
		t.Errorf("Last arg should be '-', got: %s", plan.Args[len(plan.Args)-1])
	}
}
