package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		want        *Config
		wantErr     bool
		errContains string
	}{
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: true,
		},
		{
			name: "schema only",
			args: []string{"schema.json"},
			want: &Config{
				SchemaFile: "schema.json",
				Port:       DefaultPort,
			},
			wantErr: false,
		},
		{
			name: "schema with seed file",
			args: []string{"schema.json", "with", "seed.json"},
			want: &Config{
				SchemaFile: "schema.json",
				SeedFile:   "seed.json",
				Port:       DefaultPort,
			},
			wantErr: false,
		},
		{
			name: "schema with custom port",
			args: []string{"schema.json", "on", "3000"},
			want: &Config{
				SchemaFile: "schema.json",
				Port:       3000,
			},
			wantErr: false,
		},
		{
			name: "schema with seed and custom port",
			args: []string{"schema.json", "with", "seed.json", "on", "3000"},
			want: &Config{
				SchemaFile: "schema.json",
				SeedFile:   "seed.json",
				Port:       3000,
			},
			wantErr: false,
		},
		{
			name: "custom port with seed",
			args: []string{"schema.json", "on", "8081", "with", "seed.json"},
			want: &Config{
				SchemaFile: "schema.json",
				SeedFile:   "seed.json",
				Port:       8081,
			},
			wantErr: false,
		},
		{
			name: "help flag",
			args: []string{"--help"},
			want: &Config{
				ShowHelp: true,
				Port:     DefaultPort,
			},
			wantErr: false,
		},
		{
			name: "help flag short",
			args: []string{"-h"},
			want: &Config{
				ShowHelp: true,
				Port:     DefaultPort,
			},
			wantErr: false,
		},
		{
			name: "version flag",
			args: []string{"--version"},
			want: &Config{
				ShowVersion: true,
				Port:        DefaultPort,
			},
			wantErr: false,
		},
		{
			name: "version flag short",
			args: []string{"-v"},
			want: &Config{
				ShowVersion: true,
				Port:        DefaultPort,
			},
			wantErr: false,
		},
		{
			name:        "with without seed file",
			args:        []string{"schema.json", "with"},
			wantErr:     true,
			errContains: "expected seed file after 'with'",
		},
		{
			name:        "on without port",
			args:        []string{"schema.json", "on"},
			wantErr:     true,
			errContains: "expected port number after 'on'",
		},
		{
			name:        "invalid port",
			args:        []string{"schema.json", "on", "abc"},
			wantErr:     true,
			errContains: "invalid port",
		},
		{
			name:        "port too low",
			args:        []string{"schema.json", "on", "0"},
			wantErr:     true,
			errContains: "must be between 1 and 65535",
		},
		{
			name:        "port too high",
			args:        []string{"schema.json", "on", "99999"},
			wantErr:     true,
			errContains: "must be between 1 and 65535",
		},
		{
			name:        "unexpected argument",
			args:        []string{"schema.json", "invalid"},
			wantErr:     true,
			errContains: "unexpected argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("Parse() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if !tt.wantErr {
				if got.SchemaFile != tt.want.SchemaFile {
					t.Errorf("Parse() SchemaFile = %v, want %v", got.SchemaFile, tt.want.SchemaFile)
				}
				if got.SeedFile != tt.want.SeedFile {
					t.Errorf("Parse() SeedFile = %v, want %v", got.SeedFile, tt.want.SeedFile)
				}
				if got.Port != tt.want.Port {
					t.Errorf("Parse() Port = %v, want %v", got.Port, tt.want.Port)
				}
				if got.ShowHelp != tt.want.ShowHelp {
					t.Errorf("Parse() ShowHelp = %v, want %v", got.ShowHelp, tt.want.ShowHelp)
				}
				if got.ShowVersion != tt.want.ShowVersion {
					t.Errorf("Parse() ShowVersion = %v, want %v", got.ShowVersion, tt.want.ShowVersion)
				}
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	// Create temporary test files
	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "schema.json")
	seedFile := filepath.Join(tmpDir, "seed.json")

	// Create test files
	if err := os.WriteFile(schemaFile, []byte("{}"), 0o644); err != nil {
		t.Fatalf("failed to create test schema file: %v", err)
	}
	if err := os.WriteFile(seedFile, []byte("{}"), 0o644); err != nil {
		t.Fatalf("failed to create test seed file: %v", err)
	}

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid schema only",
			config: &Config{
				SchemaFile: schemaFile,
				Port:       8080,
			},
			wantErr: false,
		},
		{
			name: "valid schema and seed",
			config: &Config{
				SchemaFile: schemaFile,
				SeedFile:   seedFile,
				Port:       8080,
			},
			wantErr: false,
		},
		{
			name: "schema not found",
			config: &Config{
				SchemaFile: filepath.Join(tmpDir, "nonexistent.json"),
				Port:       8080,
			},
			wantErr: true,
		},
		{
			name: "seed not found",
			config: &Config{
				SchemaFile: schemaFile,
				SeedFile:   filepath.Join(tmpDir, "nonexistent_seed.json"),
				Port:       8080,
			},
			wantErr: true,
		},
		{
			name: "help flag skips validation",
			config: &Config{
				ShowHelp:   true,
				SchemaFile: "nonexistent.json",
				Port:       8080,
			},
			wantErr: false,
		},
		{
			name: "version flag skips validation",
			config: &Config{
				ShowVersion: true,
				SchemaFile:  "nonexistent.json",
				Port:        8080,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigString(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   string
	}{
		{
			name: "schema only",
			config: &Config{
				SchemaFile: "schema.json",
				Port:       8080,
			},
			want: "Schema: schema.json, Port: 8080",
		},
		{
			name: "schema with seed",
			config: &Config{
				SchemaFile: "schema.json",
				SeedFile:   "seed.json",
				Port:       3000,
			},
			want: "Schema: schema.json, Seed: seed.json, Port: 3000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.String()
			if got != tt.want {
				t.Errorf("Config.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
