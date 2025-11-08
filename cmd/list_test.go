package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunList(t *testing.T) {
	tests := []struct {
		name        string
		configData  string
		expectError bool
	}{
		{
			name: "valid config with single project",
			configData: `version: 1.0.0
projects:
  - name: sample-project
    repo: git@github.com:user/sample-project.git
    branch: main
    backup_paths:
      - .env
      - db.sqlite3
    backup_retention: 3
`,
			expectError: false,
		},
		{
			name: "valid config with multiple projects",
			configData: `version: 1.0.0
projects:
  - name: project-one
    repo: git@github.com:user/project-one.git
    branch: main
    backup_paths:
      - .env
    backup_retention: 3
  - name: project-two
    repo: https://github.com/user/project-two.git
    branch: develop
    backup_paths:
      - .env
      - db/
    backup_retention: 5
`,
			expectError: false,
		},
		{
			name: "empty projects list",
			configData: `version: 1.0.0
projects: []
`,
			expectError: false,
		},
		{
			name: "invalid yaml",
			configData: `version: 1.0.0
projects:
  - name: sample-project
    repo: git@github.com:user/sample-project.git
    branch: main
  invalid yaml here
`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "config.yml")

			if err := os.WriteFile(configPath, []byte(tt.configData), 0644); err != nil {
				t.Fatalf("Failed to create test config file: %v", err)
			}

			// Set cfgFile to use our temporary config
			originalCfgFile := cfgFile
			cfgFile = configPath
			defer func() {
				cfgFile = originalCfgFile
			}()

			// Run list
			err := runList()
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestRunListNoConfig(t *testing.T) {
	// Create temporary directory without config file
	tempDir := t.TempDir()
	nonExistentPath := filepath.Join(tempDir, "nonexistent.yml")

	// Set cfgFile to non-existent path
	originalCfgFile := cfgFile
	cfgFile = nonExistentPath
	defer func() {
		cfgFile = originalCfgFile
	}()

	// Run list - should fail because config doesn't exist
	err := runList()
	if err == nil {
		t.Errorf("Expected error for non-existent config file but got nil")
	}
}

func TestRunListWithMinimalProject(t *testing.T) {
	// Test with minimal project configuration (no optional fields)
	configData := `version: 1.0.0
projects:
  - name: minimal-project
    repo: git@github.com:user/minimal.git
    branch: main
`

	// Create temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yml")

	if err := os.WriteFile(configPath, []byte(configData), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Set cfgFile to use our temporary config
	originalCfgFile := cfgFile
	cfgFile = configPath
	defer func() {
		cfgFile = originalCfgFile
	}()

	// Run list - should succeed
	err := runList()
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}
}
