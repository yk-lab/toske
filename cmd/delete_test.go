package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestRunDelete(t *testing.T) {
	tests := []struct {
		name            string
		projectName     string
		configData      string
		userInput       string
		expectError     bool
		errorMessage    string
		expectedRemains []string
	}{
		{
			name:        "successful deletion with yes confirmation",
			projectName: "test-project",
			configData: `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    backup_paths:
      - .env
    backup_retention: 3
  - name: other-project
    repo: git@github.com:user/other.git
    branch: main
`,
			userInput:       "y\n",
			expectError:     false,
			expectedRemains: []string{"other-project"},
		},
		{
			name:        "successful deletion with yes (full word) confirmation",
			projectName: "test-project",
			configData: `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
  - name: another-project
    repo: git@github.com:user/another.git
    branch: develop
`,
			userInput:       "yes\n",
			expectError:     false,
			expectedRemains: []string{"another-project"},
		},
		{
			name:        "cancelled deletion with no",
			projectName: "test-project",
			configData: `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
`,
			userInput:       "n\n",
			expectError:     false,
			expectedRemains: []string{"test-project"},
		},
		{
			name:        "cancelled deletion with enter (default no)",
			projectName: "test-project",
			configData: `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
`,
			userInput:       "\n",
			expectError:     false,
			expectedRemains: []string{"test-project"},
		},
		{
			name:        "project not found",
			projectName: "nonexistent",
			configData: `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
`,
			userInput:    "",
			expectError:  true,
			errorMessage: "not found in configuration file",
		},
		{
			name:         "missing project flag",
			projectName:  "",
			configData:   `version: 1.0.0\nprojects: []`,
			userInput:    "",
			expectError:  true,
			errorMessage: "Project name is required",
		},
		{
			name:        "delete last remaining project",
			projectName: "test-project",
			configData: `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
`,
			userInput:       "y\n",
			expectError:     false,
			expectedRemains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup config
			defer setupTestConfig(t, tt.configData)()

			// Setup stdin mock
			if tt.userInput != "" {
				// Create a temporary file for stdin
				tmpfile, err := os.CreateTemp("", "stdin")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				defer os.Remove(tmpfile.Name())

				if _, err := tmpfile.WriteString(tt.userInput); err != nil {
					t.Fatalf("Failed to write to temp file: %v", err)
				}

				if _, err := tmpfile.Seek(0, 0); err != nil {
					t.Fatalf("Failed to seek temp file: %v", err)
				}

				oldStdin := os.Stdin
				os.Stdin = tmpfile
				defer func() {
					os.Stdin = oldStdin
					tmpfile.Close()
				}()
			}

			// Set project name
			originalProjectName := deleteProjectName
			deleteProjectName = tt.projectName
			defer func() { deleteProjectName = originalProjectName }()

			// Run delete
			err := runDelete()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.errorMessage != "" && !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("Expected error message to contain '%s', got: %v", tt.errorMessage, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				// Verify remaining projects
				v := viper.New()
				v.SetConfigFile(cfgFile)
				if err := v.ReadInConfig(); err != nil {
					t.Fatalf("Failed to read config after deletion: %v", err)
				}

				var config Config
				if err := v.Unmarshal(&config); err != nil {
					t.Fatalf("Failed to unmarshal config: %v", err)
				}

				if len(config.Projects) != len(tt.expectedRemains) {
					t.Errorf("Expected %d remaining projects, got %d", len(tt.expectedRemains), len(config.Projects))
				}

				for i, expectedName := range tt.expectedRemains {
					if i >= len(config.Projects) {
						t.Errorf("Expected project '%s' to remain, but it doesn't exist", expectedName)
						continue
					}
					if config.Projects[i].Name != expectedName {
						t.Errorf("Expected project '%s' at position %d, got '%s'", expectedName, i, config.Projects[i].Name)
					}
				}
			}
		})
	}
}

func TestRunDeleteNoConfig(t *testing.T) {
	// Create temporary directory without config file
	tempDir := t.TempDir()
	nonExistentPath := filepath.Join(tempDir, "nonexistent.yml")

	// Set cfgFile to non-existent path
	originalCfgFile := cfgFile
	cfgFile = nonExistentPath
	defer func() {
		cfgFile = originalCfgFile
	}()

	// Set project name
	originalProjectName := deleteProjectName
	deleteProjectName = "test-project"
	defer func() { deleteProjectName = originalProjectName }()

	// Run delete - should fail because config doesn't exist
	err := runDelete()
	if err == nil {
		t.Errorf("Expected error for non-existent config file but got nil")
	}
}

func TestRunDeleteInvalidYAML(t *testing.T) {
	// Create temporary config with invalid YAML
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yml")

	invalidYAML := `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
  invalid yaml here
`

	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	originalCfgFile := cfgFile
	cfgFile = configPath
	defer func() { cfgFile = originalCfgFile }()

	// Set project name
	originalProjectName := deleteProjectName
	deleteProjectName = "test-project"
	defer func() { deleteProjectName = originalProjectName }()

	// Run delete - should fail on parsing
	err := runDelete()
	if err == nil {
		t.Errorf("Expected error for invalid YAML but got nil")
	}
}

func TestDeleteConfirmationInput(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		shouldDelete  bool
	}{
		{"lowercase y", "y", true},
		{"uppercase Y", "Y", true},
		{"lowercase yes", "yes", true},
		{"uppercase YES", "YES", true},
		{"mixed case Yes", "Yes", true},
		{"lowercase n", "n", false},
		{"uppercase N", "N", false},
		{"lowercase no", "no", false},
		{"uppercase NO", "NO", false},
		{"empty input", "", false},
		{"whitespace", "  ", false},
		{"random text", "maybe", false},
		{"y with whitespace", " y ", true},
		{"yes with whitespace", " yes ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test config
			configData := `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
`
			defer setupTestConfig(t, configData)()

			// Setup stdin mock
			tmpfile, err := os.CreateTemp("", "stdin")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.WriteString(tt.input + "\n"); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}

			if _, err := tmpfile.Seek(0, 0); err != nil {
				t.Fatalf("Failed to seek temp file: %v", err)
			}

			oldStdin := os.Stdin
			os.Stdin = tmpfile
			defer func() {
				os.Stdin = oldStdin
				tmpfile.Close()
			}()

			// Set project name
			originalProjectName := deleteProjectName
			deleteProjectName = "test-project"
			defer func() { deleteProjectName = originalProjectName }()

			// Run delete
			err = runDelete()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify project was deleted or not
			v := viper.New()
			v.SetConfigFile(cfgFile)
			if err := v.ReadInConfig(); err != nil {
				t.Fatalf("Failed to read config: %v", err)
			}

			var config Config
			if err := v.Unmarshal(&config); err != nil {
				t.Fatalf("Failed to unmarshal config: %v", err)
			}

			projectExists := false
			for _, project := range config.Projects {
				if project.Name == "test-project" {
					projectExists = true
					break
				}
			}

			if tt.shouldDelete && projectExists {
				t.Errorf("Expected project to be deleted, but it still exists")
			}
			if !tt.shouldDelete && !projectExists {
				t.Errorf("Expected project to remain, but it was deleted")
			}
		})
	}
}

// TestDeleteMultipleProjectsOrder tests that deleting a project maintains the order of remaining projects
func TestDeleteMultipleProjectsOrder(t *testing.T) {
	configData := `version: 1.0.0
projects:
  - name: first-project
    repo: git@github.com:user/first.git
    branch: main
  - name: second-project
    repo: git@github.com:user/second.git
    branch: main
  - name: third-project
    repo: git@github.com:user/third.git
    branch: main
  - name: fourth-project
    repo: git@github.com:user/fourth.git
    branch: main
`

	defer setupTestConfig(t, configData)()

	// Setup stdin to confirm deletion
	tmpfile, err := os.CreateTemp("", "stdin")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString("y\n"); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	if _, err := tmpfile.Seek(0, 0); err != nil {
		t.Fatalf("Failed to seek temp file: %v", err)
	}

	oldStdin := os.Stdin
	os.Stdin = tmpfile
	defer func() {
		os.Stdin = oldStdin
		tmpfile.Close()
	}()

	// Delete second-project
	originalProjectName := deleteProjectName
	deleteProjectName = "second-project"
	defer func() { deleteProjectName = originalProjectName }()

	err = runDelete()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify remaining projects are in correct order
	v := viper.New()
	v.SetConfigFile(cfgFile)
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	expectedProjects := []string{"first-project", "third-project", "fourth-project"}
	if len(config.Projects) != len(expectedProjects) {
		t.Fatalf("Expected %d projects, got %d", len(expectedProjects), len(config.Projects))
	}

	for i, expected := range expectedProjects {
		if config.Projects[i].Name != expected {
			t.Errorf("Expected project '%s' at position %d, got '%s'", expected, i, config.Projects[i].Name)
		}
	}
}
