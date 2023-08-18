package solc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompilerConfigSanitizeArguments(t *testing.T) {
	config := &CompilerConfig{}

	tests := []struct {
		name    string
		args    []string
		want    []string
		wantErr string
	}{
		{
			name:    "Valid Arguments",
			args:    []string{"--optimize-runs", "-"},
			want:    []string{"--optimize-runs", "-"},
			wantErr: "",
		},
		{
			name:    "Invalid Argument",
			args:    []string{"--optimize", "--invalid"},
			want:    nil,
			wantErr: "invalid argument: --invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := config.SanitizeArguments(tt.args)
			assert.Equal(t, tt.want, got)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func TestCompilerConfigValidate(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		compilerVersion string
		wantErr         string
	}{
		{
			name:            "Valid Arguments",
			compilerVersion: "0.8.21",
			args:            []string{"--overwrite", "--combined-json", "--optimize", "200", "-"},
			wantErr:         "",
		},
		{
			name:            "Missing Required Argument",
			compilerVersion: "0.8.21",
			args:            []string{"--overwrite", "--combined-json"},
			wantErr:         "missing required argument: -",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &CompilerConfig{CompilerVersion: tt.compilerVersion, Arguments: tt.args}
			err := config.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func TestCompilerConfigVersion(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		version string
		wantErr string
	}{
		{
			name:    "Valid Compiler Version",
			args:    []string{"--overwrite", "--combined-json", "--optimize", "200", "-"},
			version: "0.8.21",
			wantErr: "",
		},
		{
			name:    "Invalid Compiler Version",
			args:    []string{"--overwrite", "--combined-json", "--optimize", "200", "-"},
			version: "0.00",
			wantErr: "invalid compiler version: 0.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &CompilerConfig{CompilerVersion: tt.version, Arguments: tt.args}
			err := config.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func TestConfigFunctions(t *testing.T) {
	config := &CompilerConfig{Arguments: []string{"--json"}}

	// Test SetArguments
	newArgs := []string{"-"}
	config.SetArguments(newArgs)
	assert.Equal(t, config.GetArguments(), newArgs)

	// Test AppendArguments
	appendArgs := []string{"--json"}
	config.AppendArguments(appendArgs...)
	assert.Equal(t, config.GetArguments(), []string{"-", "--json"})
}

func TestNewDefaultConfig(t *testing.T) {
	tests := []struct {
		name            string
		compilerVersion string
		wantErr         bool
	}{
		{
			name:            "default config",
			compilerVersion: "",
			wantErr:         true,
		},
		{
			name:            "default config",
			compilerVersion: "0.4.11",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewDefaultCompilerConfig(tt.compilerVersion)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				config.SetCompilerVersion("0.4.12")
				assert.Equal(t, config.GetCompilerVersion(), "0.4.12")
			}
		})
	}
}
