package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestRunRemove(t *testing.T) {
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
			name:        "successful removal with yes confirmation",
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
			name:        "successful removal with yes (full word) confirmation",
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
			name:        "cancelled removal with no",
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
			name:        "cancelled removal with enter (default no)",
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
			name:        "remove last remaining project",
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
			originalProjectName := removeProjectName
			removeProjectName = tt.projectName
			defer func() { removeProjectName = originalProjectName }()

			// Run remove
			err := runRemove()

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
					t.Fatalf("Failed to read config after removal: %v", err)
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

func TestRunRemoveNoConfig(t *testing.T) {
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
	originalProjectName := removeProjectName
	removeProjectName = "test-project"
	defer func() { removeProjectName = originalProjectName }()

	// Run remove - should fail because config doesn't exist
	err := runRemove()
	if err == nil {
		t.Errorf("Expected error for non-existent config file but got nil")
	}
}

func TestRunRemoveInvalidYAML(t *testing.T) {
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
	originalProjectName := removeProjectName
	removeProjectName = "test-project"
	defer func() { removeProjectName = originalProjectName }()

	// Run remove - should fail on parsing
	err := runRemove()
	if err == nil {
		t.Errorf("Expected error for invalid YAML but got nil")
	}
}

func TestRemoveConfirmationInput(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		shouldRemove bool
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
			originalProjectName := removeProjectName
			removeProjectName = "test-project"
			defer func() { removeProjectName = originalProjectName }()

			// Run remove
			err = runRemove()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify project was removed or not
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

			if tt.shouldRemove && projectExists {
				t.Errorf("Expected project to be removed, but it still exists")
			}
			if !tt.shouldRemove && !projectExists {
				t.Errorf("Expected project to remain, but it was removed")
			}
		})
	}
}

// TestRemoveProjectNameWithWhitespace tests that project names with whitespace are trimmed
func TestRemoveProjectNameWithWhitespace(t *testing.T) {
	configData := `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
`

	defer setupTestConfig(t, configData)()

	// Setup stdin to confirm removal
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

	// Remove with whitespace around project name
	originalProjectName := removeProjectName
	removeProjectName = "  test-project  "
	defer func() { removeProjectName = originalProjectName }()

	err = runRemove()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify project was removed
	v := viper.New()
	v.SetConfigFile(cfgFile)
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if len(config.Projects) != 0 {
		t.Errorf("Expected project to be removed despite whitespace in flag, but found %d projects", len(config.Projects))
	}
}

// TestRemovePreservesYAMLFormat tests that removing a project preserves correct YAML field names
func TestRemovePreservesYAMLFormat(t *testing.T) {
	configData := `version: 1.0.0
projects:
  - name: project-one
    repo: git@github.com:user/one.git
    branch: main
    backup_paths:
      - .env
      - db.sqlite3
    backup_retention: 3
  - name: project-two
    repo: git@github.com:user/two.git
    branch: develop
    backup_paths:
      - config/
    backup_retention: 5
`

	defer setupTestConfig(t, configData)()

	// Setup stdin to confirm removal
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

	// Remove project-one
	originalProjectName := removeProjectName
	removeProjectName = "project-one"
	defer func() { removeProjectName = originalProjectName }()

	err = runRemove()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Read the raw YAML file to verify field names
	rawYAML, err := os.ReadFile(cfgFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	yamlContent := string(rawYAML)

	// Verify that the YAML uses snake_case field names, not Go field names
	if !strings.Contains(yamlContent, "backup_paths:") {
		t.Error("Expected YAML to contain 'backup_paths:' but it doesn't")
	}
	if !strings.Contains(yamlContent, "backup_retention:") {
		t.Error("Expected YAML to contain 'backup_retention:' but it doesn't")
	}

	// Verify it doesn't use Go field names
	if strings.Contains(yamlContent, "backuppaths") {
		t.Error("YAML should not contain Go field name 'backuppaths'")
	}
	if strings.Contains(yamlContent, "backupretention") {
		t.Error("YAML should not contain Go field name 'backupretention'")
	}

	// Verify that we can still read the config with viper
	v := viper.New()
	v.SetConfigFile(cfgFile)
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config after removal: %v", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Verify project-two still has its backup paths and retention
	if len(config.Projects) != 1 {
		t.Fatalf("Expected 1 project, got %d", len(config.Projects))
	}

	project := config.Projects[0]
	if project.Name != "project-two" {
		t.Errorf("Expected project-two, got %s", project.Name)
	}
	if len(project.BackupPaths) != 1 {
		t.Errorf("Expected 1 backup path, got %d - YAML field names may be corrupted", len(project.BackupPaths))
	}
	if project.BackupRetention != 5 {
		t.Errorf("Expected backup retention 5, got %d - YAML field names may be corrupted", project.BackupRetention)
	}
}

// TestRemoveWithForceFlag tests that --force flag skips confirmation prompt
func TestRemoveWithForceFlag(t *testing.T) {
	configData := `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
  - name: other-project
    repo: git@github.com:user/other.git
    branch: main
`

	defer setupTestConfig(t, configData)()

	// Set flags - no stdin setup needed because --force skips prompt
	originalProjectName := removeProjectName
	originalForce := removeForce
	removeProjectName = "test-project"
	removeForce = true
	defer func() {
		removeProjectName = originalProjectName
		removeForce = originalForce
	}()

	// Run remove - should succeed without stdin
	err := runRemove()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify project was removed
	v := viper.New()
	v.SetConfigFile(cfgFile)
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if len(config.Projects) != 1 {
		t.Errorf("Expected 1 remaining project, got %d", len(config.Projects))
	}
	if len(config.Projects) > 0 && config.Projects[0].Name != "other-project" {
		t.Errorf("Expected 'other-project' to remain, got '%s'", config.Projects[0].Name)
	}
}

// TestFindRemoveProjectIndex tests the findRemoveProjectIndex helper function
func TestFindRemoveProjectIndex(t *testing.T) {
	projects := []Project{
		{Name: "first"},
		{Name: "second"},
		{Name: "third"},
	}

	tests := []struct {
		name     string
		projects []Project
		search   string
		expected int
	}{
		{
			name:     "find first project",
			projects: projects,
			search:   "first",
			expected: 0,
		},
		{
			name:     "find middle project",
			projects: projects,
			search:   "second",
			expected: 1,
		},
		{
			name:     "find last project",
			projects: projects,
			search:   "third",
			expected: 2,
		},
		{
			name:     "project not found",
			projects: projects,
			search:   "nonexistent",
			expected: -1,
		},
		{
			name:     "empty projects list",
			projects: []Project{},
			search:   "any",
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findRemoveProjectIndex(tt.projects, tt.search)
			if result != tt.expected {
				t.Errorf("Expected index %d, got %d", tt.expected, result)
			}
		})
	}
}

// TestConfirmRemoval tests the confirmRemoval function
func TestConfirmRemoval(t *testing.T) {
	tests := []struct {
		name     string
		force    bool
		input    string
		expected bool
	}{
		{
			name:     "force flag true",
			force:    true,
			input:    "", // no input needed
			expected: true,
		},
		{
			name:     "confirm with y",
			force:    false,
			input:    "y\n",
			expected: true,
		},
		{
			name:     "confirm with yes",
			force:    false,
			input:    "yes\n",
			expected: true,
		},
		{
			name:     "cancel with n",
			force:    false,
			input:    "n\n",
			expected: false,
		},
		{
			name:     "cancel with empty",
			force:    false,
			input:    "\n",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.force {
				// Setup stdin mock
				tmpfile, err := os.CreateTemp("", "stdin")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				defer os.Remove(tmpfile.Name())

				if _, err := tmpfile.WriteString(tt.input); err != nil {
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

			result, err := confirmRemoval("test-project", tt.force)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestSaveRemoveConfigPermissions tests that config file is saved with 0600 permissions
func TestSaveRemoveConfigPermissions(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yml")

	config := Config{
		Version: "1.0.0",
		Projects: []Project{
			{
				Name:   "test",
				Repo:   "git@github.com:user/test.git",
				Branch: "main",
			},
		},
	}

	err := saveRemoveConfig(configPath, &config)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	// Check file permissions
	mode := info.Mode().Perm()
	expected := os.FileMode(0600)
	if mode != expected {
		t.Errorf("Expected file permissions %o, got %o", expected, mode)
	}
}

// TestRemoveMultipleProjectsOrder tests that removing a project maintains the order of remaining projects
func TestRemoveMultipleProjectsOrder(t *testing.T) {
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

	// Setup stdin to confirm removal
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

	// Remove second-project
	originalProjectName := removeProjectName
	removeProjectName = "second-project"
	defer func() { removeProjectName = originalProjectName }()

	err = runRemove()
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
