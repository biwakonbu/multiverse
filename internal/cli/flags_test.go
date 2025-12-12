package cli

import (
	"bytes"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *Flags
		wantErr bool
	}{
		{
			name: "no flags",
			args: []string{},
			want: &Flags{MetaModel: ""},
		},
		{
			name: "meta-model flag",
			args: []string{"--meta-model=gpt-5.2-mini"},
			want: &Flags{MetaModel: "gpt-5.2-mini"},
		},
		{
			name:    "unknown flag",
			args:    []string{"--unknown"},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			got, err := ParseFlags(tt.args, &buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.MetaModel != tt.want.MetaModel {
					t.Errorf("ParseFlags() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestResolveMetaModel(t *testing.T) {
	tests := []struct {
		name      string
		cliModel  string
		yamlModel string
		want      string
	}{
		{
			name:      "CLI takes precedence",
			cliModel:  "cli-model",
			yamlModel: "yaml-model",
			want:      "cli-model",
		},
		{
			name:      "YAML used if CLI empty",
			cliModel:  "",
			yamlModel: "yaml-model",
			want:      "yaml-model",
		},
		{
			name:      "Default used if both empty",
			cliModel:  "",
			yamlModel: "",
			want:      "gpt-5.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveMetaModel(tt.cliModel, tt.yamlModel); got != tt.want {
				t.Errorf("ResolveMetaModel() = %v, want %v", got, tt.want)
			}
		})
	}
}
