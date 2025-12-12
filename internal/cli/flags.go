package cli

import (
	"flag"
	"io"
)

// Flags holds command-line arguments
type Flags struct {
	MetaModel string
}

// ParseFlags parses command-line arguments
func ParseFlags(args []string, output io.Writer) (*Flags, error) {
	fs := flag.NewFlagSet("agent-runner", flag.ContinueOnError)
	fs.SetOutput(output)

	var flags Flags
	fs.StringVar(&flags.MetaModel, "meta-model", "", "Meta agent LLM model ID")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	return &flags, nil
}

// ResolveMetaModel determines the final model ID based on priority:
// 1. CLI flag
// 2. Task YAML configuration
// 3. Default value (gpt-5.2 for Meta-agent)
func ResolveMetaModel(cliModel, yamlModel string) string {
	if cliModel != "" {
		return cliModel
	}
	if yamlModel != "" {
		return yamlModel
	}
	return "gpt-5.2"
}
