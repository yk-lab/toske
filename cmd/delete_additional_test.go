package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// TestRunDelete_NoBackup tests deletion fails when no backup exists
func TestRunDelete_NoBackup(t *testing.T) {
	configData := `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    repository_path: /tmp/test-project
    backup_paths:
      - .env
`

	defer setupTestConfig(t, configData)()

	// Do NOT create backup - this is the test condition

	originalProjectName := deleteProjectName
	originalForce := deleteForce
	deleteProjectName = "test-project"
	deleteForce = true
	defer func() {
		deleteProjectName = originalProjectName
		deleteForce = originalForce
	}()

	err := runDelete()
	if err == nil {
		t.Error("Expected error for missing backup, but got nil")
	} else if !strings.Contains(err.Error(), "no backups") && !strings.Contains(err.Error(), "バックアップがありません") {
		t.Errorf("Expected 'no backups' error message, got: %v", err)
	}
}

// TestRunDelete_NoRepositoryPath tests deletion fails when repository_path is not set
func TestRunDelete_NoRepositoryPath(t *testing.T) {
	configData := `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    backup_paths:
      - .env
`

	defer setupTestConfig(t, configData)()
	defer setupTestBackup(t, "test-project")()

	originalProjectName := deleteProjectName
	originalForce := deleteForce
	deleteProjectName = "test-project"
	deleteForce = true
	defer func() {
		deleteProjectName = originalProjectName
		deleteForce = originalForce
	}()

	err := runDelete()
	if err == nil {
		t.Error("Expected error for missing repository_path, but got nil")
	} else if !strings.Contains(err.Error(), "repository_path") {
		t.Errorf("Expected 'repository_path' error message, got: %v", err)
	}
}

// TestRunDelete_RepositoryNotFound tests deletion succeeds even if repository directory doesn't exist
func TestRunDelete_RepositoryNotFound(t *testing.T) {
	configData := `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    repository_path: /tmp/nonexistent-repo-12345
    backup_paths:
      - .env
`

	defer setupTestConfig(t, configData)()
	defer setupTestBackup(t, "test-project")()

	originalProjectName := deleteProjectName
	originalForce := deleteForce
	deleteProjectName = "test-project"
	deleteForce = true
	defer func() {
		deleteProjectName = originalProjectName
		deleteForce = originalForce
	}()

	// Ensure the repository directory does not exist
	os.RemoveAll("/tmp/nonexistent-repo-12345")

	err := runDelete()
	if err != nil {
		t.Errorf("Expected deletion to succeed even with missing repository directory, got error: %v", err)
	}

	// Verify project was deleted from config
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
		t.Errorf("Expected project to be deleted from config, but found %d projects", len(config.Projects))
	}
}

// TestRunDelete_PhysicalDeletion tests that repository directory is physically deleted
func TestRunDelete_PhysicalDeletion(t *testing.T) {
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "test-repo")

	configData := `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    repository_path: ` + repoPath + `
    backup_paths:
      - .env
`

	defer setupTestConfig(t, configData)()
	defer setupTestBackup(t, "test-project")()
	defer setupRepositoryDir(t, repoPath)()

	// Verify repository exists before deletion
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		t.Fatal("Repository directory should exist before deletion")
	}

	originalProjectName := deleteProjectName
	originalForce := deleteForce
	deleteProjectName = "test-project"
	deleteForce = true
	defer func() {
		deleteProjectName = originalProjectName
		deleteForce = originalForce
	}()

	err := runDelete()
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify repository directory was deleted
	if _, err := os.Stat(repoPath); !os.IsNotExist(err) {
		t.Error("Repository directory should have been deleted")
	}

	// Verify project was deleted from config
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
		t.Errorf("Expected project to be deleted from config, but found %d projects", len(config.Projects))
	}
}
