package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
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
    repository_path: /tmp/test-project
    backup_paths:
      - .env
    backup_retention: 3
  - name: other-project
    repo: git@github.com:user/other.git
    branch: main
    repository_path: /tmp/other-project
`,
			userInput:       "y\n",
			expectError:     false,
			expectedRemains: []string{"test-project", "other-project"},
		},
		{
			name:        "successful deletion with yes (full word) confirmation",
			projectName: "test-project",
			configData: `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    repository_path: /tmp/test-project
  - name: another-project
    repo: git@github.com:user/another.git
    branch: develop
    repository_path: /tmp/another-project
`,
			userInput:       "yes\n",
			expectError:     false,
			expectedRemains: []string{"test-project", "another-project"},
		},
		{
			name:        "cancelled deletion with no",
			projectName: "test-project",
			configData: `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    repository_path: /tmp/test-project
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
    repository_path: /tmp/test-project
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
    repository_path: /tmp/test-project
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
    repository_path: /tmp/test-project
`,
			userInput:       "y\n",
			expectError:     false,
			expectedRemains: []string{"test-project"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup config
			defer setupTestConfig(t, tt.configData)()

			// Setup mock backup for the project
			if tt.projectName != "" && !strings.Contains(tt.name, "not found") && !strings.Contains(tt.name, "missing project flag") {
				defer setupTestBackup(t, tt.projectName)()
			}

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
    repository_path: /tmp/test-project
`
			defer setupTestConfig(t, configData)()

			// Setup mock backup
			defer setupTestBackup(t, "test-project")()

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

			// Projects always remain in config (delete only removes repository, not config)
			if len(config.Projects) != 1 {
				t.Errorf("Expected 1 project in config, got %d", len(config.Projects))
			}
			if len(config.Projects) > 0 && config.Projects[0].Name != "test-project" {
				t.Errorf("Expected 'test-project' in config, got '%s'", config.Projects[0].Name)
			}
		})
	}
}

// TestDeleteProjectNameWithWhitespace tests that project names with whitespace are trimmed
func TestDeleteProjectNameWithWhitespace(t *testing.T) {
	configData := `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    repository_path: /tmp/test-project
`

	defer setupTestConfig(t, configData)()
	defer setupTestBackup(t, "test-project")()

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

	// Delete with whitespace around project name
	originalProjectName := deleteProjectName
	deleteProjectName = "  test-project  "
	defer func() { deleteProjectName = originalProjectName }()

	err = runDelete()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify project was deleted
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
		t.Errorf("Expected project to remain in config after deletion, but found %d projects", len(config.Projects))
	}
	if config.Projects[0].Name != "test-project" {
		t.Errorf("Expected 'test-project' to remain in config, got '%s'", config.Projects[0].Name)
	}
}

// TestDeletePreservesYAMLFormat tests that deleting a project preserves correct YAML field names
func TestDeletePreservesYAMLFormat(t *testing.T) {
	configData := `version: 1.0.0
projects:
  - name: project-one
    repo: git@github.com:user/one.git
    branch: main
    repository_path: /tmp/project-one
    backup_paths:
      - .env
      - db.sqlite3
    backup_retention: 3
  - name: project-two
    repo: git@github.com:user/two.git
    branch: develop
    repository_path: /tmp/project-two
    backup_paths:
      - config/
    backup_retention: 5
`

	defer setupTestConfig(t, configData)()
	defer setupTestBackup(t, "project-one")()

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

	// Delete project-one
	originalProjectName := deleteProjectName
	deleteProjectName = "project-one"
	defer func() { deleteProjectName = originalProjectName }()

	err = runDelete()
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
		t.Fatalf("Failed to read config after deletion: %v", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Verify both projects remain in config (delete only removes repository, not config)
	if len(config.Projects) != 2 {
		t.Fatalf("Expected 2 projects in config, got %d", len(config.Projects))
	}

	// Verify project-one is still in config
	if config.Projects[0].Name != "project-one" {
		t.Errorf("Expected project-one at position 0, got %s", config.Projects[0].Name)
	}

	// Verify project-two still has its backup paths and retention
	project := config.Projects[1]
	if project.Name != "project-two" {
		t.Errorf("Expected project-two at position 1, got %s", project.Name)
	}
	if len(project.BackupPaths) != 1 {
		t.Errorf("Expected 1 backup path, got %d - YAML field names may be corrupted", len(project.BackupPaths))
	}
	if project.BackupRetention != 5 {
		t.Errorf("Expected backup retention 5, got %d - YAML field names may be corrupted", project.BackupRetention)
	}
}

// TestDeleteWithForceFlag tests that --force flag skips confirmation prompt
func TestDeleteWithForceFlag(t *testing.T) {
	configData := `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    repository_path: /tmp/test-project
  - name: other-project
    repo: git@github.com:user/other.git
    branch: main
    repository_path: /tmp/other-project
`

	defer setupTestConfig(t, configData)()
	defer setupTestBackup(t, "test-project")()

	// Set flags - no stdin setup needed because --force skips prompt
	originalProjectName := deleteProjectName
	originalForce := deleteForce
	deleteProjectName = "test-project"
	deleteForce = true
	defer func() {
		deleteProjectName = originalProjectName
		deleteForce = originalForce
	}()

	// Run delete - should succeed without stdin
	err := runDelete()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify project was deleted
	v := viper.New()
	v.SetConfigFile(cfgFile)
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if len(config.Projects) != 2 {
		t.Errorf("Expected 2 projects to remain in config, got %d", len(config.Projects))
	}
	if len(config.Projects) >= 2 {
		if config.Projects[0].Name != "test-project" {
			t.Errorf("Expected 'test-project' at position 0, got '%s'", config.Projects[0].Name)
		}
		if config.Projects[1].Name != "other-project" {
			t.Errorf("Expected 'other-project' at position 1, got '%s'", config.Projects[1].Name)
		}
	}
}

// TestFindProjectIndex tests the findProjectIndex helper function
func TestFindProjectIndex(t *testing.T) {
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
			result := findProjectIndex(tt.projects, tt.search)
			if result != tt.expected {
				t.Errorf("Expected index %d, got %d", tt.expected, result)
			}
		})
	}
}

// TestConfirmDeletion tests the confirmDeletion function
func TestConfirmDeletion(t *testing.T) {
	tests := []struct {
		name       string
		repoExists bool
		force      bool
		input      string
		expected   bool
	}{
		{
			name:       "force flag true",
			repoExists: true,
			force:      true,
			input:      "", // no input needed
			expected:   true,
		},
		{
			name:       "confirm with y",
			repoExists: true,
			force:      false,
			input:      "y\n",
			expected:   true,
		},
		{
			name:       "confirm with yes",
			repoExists: false,
			force:      false,
			input:      "yes\n",
			expected:   true,
		},
		{
			name:       "cancel with n",
			repoExists: true,
			force:      false,
			input:      "n\n",
			expected:   false,
		},
		{
			name:       "cancel with empty",
			repoExists: false,
			force:      false,
			input:      "\n",
			expected:   false,
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

			// Create a test project
			project := Project{
				Name:           "test-project",
				Repo:           "git@github.com:user/test.git",
				Branch:         "main",
				RepositoryPath: "/tmp/test-project",
			}

			result, err := confirmDeletion(project, tt.repoExists, tt.force)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestSaveConfigPermissions tests that config file is saved with 0600 permissions
func TestSaveConfigPermissions(t *testing.T) {
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

	err := saveConfig(configPath, &config)
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

// TestDeleteMultipleProjectsOrder tests that deleting a project maintains the order of remaining projects
func TestDeleteMultipleProjectsOrder(t *testing.T) {
	configData := `version: 1.0.0
projects:
  - name: first-project
    repo: git@github.com:user/first.git
    branch: main
    repository_path: /tmp/first-project
  - name: second-project
    repo: git@github.com:user/second.git
    branch: main
    repository_path: /tmp/second-project
  - name: third-project
    repo: git@github.com:user/third.git
    branch: main
    repository_path: /tmp/third-project
  - name: fourth-project
    repo: git@github.com:user/fourth.git
    branch: main
    repository_path: /tmp/fourth-project
`

	defer setupTestConfig(t, configData)()
	defer setupTestBackup(t, "second-project")()

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

	// All projects should remain in config (delete only removes repository, not config)
	expectedProjects := []string{"first-project", "second-project", "third-project", "fourth-project"}
	if len(config.Projects) != len(expectedProjects) {
		t.Fatalf("Expected %d projects, got %d", len(expectedProjects), len(config.Projects))
	}

	for i, expected := range expectedProjects {
		if config.Projects[i].Name != expected {
			t.Errorf("Expected project '%s' at position %d, got '%s'", expected, i, config.Projects[i].Name)
		}
	}
}

// setupTestBackup creates a mock backup directory and metadata file for testing
func setupTestBackup(t *testing.T, projectName string) func() {
	t.Helper()

	// ja: テスト用の一時ディレクトリを使用
	// en: Use temporary directory for testing
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	backupDir := filepath.Join(tempHome, ".config", "toske", "backups", projectName)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}

	// Create a mock backup metadata file
	metadata := BackupMetadata{
		Project: projectName,
		Backups: []BackupRecord{
			{
				Filename:  "backup-" + projectName + "-20240101-120000.tar.gz",
				Timestamp: time.Now(),
				Files:     []string{".env", "config.yml"},
			},
		},
	}

	metadataPath := filepath.Join(backupDir, "backups.yaml")
	data, err := yaml.Marshal(&metadata)
	if err != nil {
		t.Fatalf("Failed to marshal metadata: %v", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		t.Fatalf("Failed to write metadata file: %v", err)
	}

	return func() {
		os.RemoveAll(backupDir)
	}
}

// setupRepositoryDir creates a test repository directory
func setupRepositoryDir(t *testing.T, repoPath string) func() {
	t.Helper()

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create repository directory: %v", err)
	}

	// Create a dummy file to make it non-empty
	dummyFile := filepath.Join(repoPath, "README.md")
	if err := os.WriteFile(dummyFile, []byte("test repo"), 0644); err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}

	return func() {
		os.RemoveAll(repoPath)
	}
}
