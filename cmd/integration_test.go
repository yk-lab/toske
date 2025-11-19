package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// TestDeleteAndRestoreIntegration tests the full workflow:
// 1. Create backup
// 2. Delete repository (config retained)
// 3. Restore from backup
func TestDeleteAndRestoreIntegration(t *testing.T) {
	// Setup: Create a temporary directory structure
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "test-repo")
	configPath := filepath.Join(tempDir, "config.yml")

	// Create repository directory with test files
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create repository directory: %v", err)
	}

	// Create test files in repository
	testFile1 := filepath.Join(repoPath, ".env")
	testFile2 := filepath.Join(repoPath, "config.yml")
	testContent1 := "DATABASE_URL=postgres://localhost/test"
	testContent2 := "app:\n  name: test-app\n  version: 1.0.0"

	if err := os.WriteFile(testFile1, []byte(testContent1), 0644); err != nil {
		t.Fatalf("Failed to create test file 1: %v", err)
	}
	if err := os.WriteFile(testFile2, []byte(testContent2), 0644); err != nil {
		t.Fatalf("Failed to create test file 2: %v", err)
	}

	// Create config file
	configContent := `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    repository_path: ` + repoPath + `
    backup_paths:
      - .env
      - config.yml
    backup_retention: 3
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Set global config file path
	originalCfgFile := cfgFile
	cfgFile = configPath
	defer func() { cfgFile = originalCfgFile }()

	// Change to repository directory for backup
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Fatalf("chdir back to %s: %v", originalWd, err)
		}
	}()

	if err := os.Chdir(repoPath); err != nil {
		t.Fatalf("Failed to change to repository directory: %v", err)
	}

	// Step 1: Create backup
	t.Log("Step 1: Creating backup...")
	originalProjectName := projectName
	projectName = "test-project"
	defer func() { projectName = originalProjectName }()

	err = runBackup()
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Verify backup was created
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}
	backupDir := filepath.Join(homeDir, ".config", "toske", "backups", "test-project")
	defer os.RemoveAll(backupDir) // Cleanup backups after test

	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		t.Fatal("Backup directory was not created")
	}

	// Step 2: Delete repository (force mode to skip confirmation)
	t.Log("Step 2: Deleting repository...")
	originalDeleteProjectName := deleteProjectName
	originalDeleteForce := deleteForce
	deleteProjectName = "test-project"
	deleteForce = true
	defer func() {
		deleteProjectName = originalDeleteProjectName
		deleteForce = originalDeleteForce
	}()

	err = runDelete()
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify repository directory was deleted
	if _, err := os.Stat(repoPath); !os.IsNotExist(err) {
		t.Error("Repository directory should have been deleted")
	}

	// Verify test files were deleted
	if _, err := os.Stat(testFile1); !os.IsNotExist(err) {
		t.Error("Test file 1 should have been deleted with repository")
	}
	if _, err := os.Stat(testFile2); !os.IsNotExist(err) {
		t.Error("Test file 2 should have been deleted with repository")
	}

	// Verify config still exists and contains the project
	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config after deletion: %v", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if len(config.Projects) != 1 {
		t.Fatalf("Expected 1 project in config after deletion, got %d", len(config.Projects))
	}
	if config.Projects[0].Name != "test-project" {
		t.Errorf("Expected 'test-project' in config, got '%s'", config.Projects[0].Name)
	}

	// Step 3: Restore from backup
	t.Log("Step 3: Restoring from backup...")

	// Restore needs to run from the repository directory (or we need to create it first)
	// Since the directory was deleted, we need to recreate it
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to recreate repository directory: %v", err)
	}

	// Change to repository directory
	if err := os.Chdir(repoPath); err != nil {
		t.Fatalf("Failed to change to repository directory: %v", err)
	}

	// Restore from backup (force mode to skip confirmation)
	originalRestoreProjectName := restoreProjectName
	originalForceRestore := forceRestore
	restoreProjectName = "test-project"
	forceRestore = true
	defer func() {
		restoreProjectName = originalRestoreProjectName
		forceRestore = originalForceRestore
	}()

	err = runRestore()
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Step 4: Verify files were restored
	t.Log("Step 4: Verifying restored files...")

	// Check if files exist
	if _, err := os.Stat(testFile1); os.IsNotExist(err) {
		t.Error("Test file 1 should have been restored")
	}
	if _, err := os.Stat(testFile2); os.IsNotExist(err) {
		t.Error("Test file 2 should have been restored")
	}

	// Verify file contents
	content1, err := os.ReadFile(testFile1)
	if err != nil {
		t.Fatalf("Failed to read restored file 1: %v", err)
	}
	if string(content1) != testContent1 {
		t.Errorf("File 1 content mismatch.\nExpected: %s\nGot: %s", testContent1, string(content1))
	}

	content2, err := os.ReadFile(testFile2)
	if err != nil {
		t.Fatalf("Failed to read restored file 2: %v", err)
	}
	if string(content2) != testContent2 {
		t.Errorf("File 2 content mismatch.\nExpected: %s\nGot: %s", testContent2, string(content2))
	}

	t.Log("✓ Integration test passed: delete → restore workflow works correctly")
}

// TestDeleteWithoutBackupFails tests that delete command fails when no backup exists
func TestDeleteWithoutBackupFails(t *testing.T) {
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "test-repo")
	configPath := filepath.Join(tempDir, "config.yml")

	// Create repository directory
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create repository directory: %v", err)
	}

	// Create config file
	configContent := `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    repository_path: ` + repoPath + `
    backup_paths:
      - .env
    backup_retention: 3
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Set global config file path
	originalCfgFile := cfgFile
	cfgFile = configPath
	defer func() { cfgFile = originalCfgFile }()

	// Try to delete without creating backup first
	originalDeleteProjectName := deleteProjectName
	originalDeleteForce := deleteForce
	deleteProjectName = "test-project"
	deleteForce = true
	defer func() {
		deleteProjectName = originalDeleteProjectName
		deleteForce = originalDeleteForce
	}()

	err := runDelete()
	if err == nil {
		t.Error("Expected delete to fail when no backup exists, but it succeeded")
	}

	// Verify error message mentions backup
	if !strings.Contains(err.Error(), "backup") && !strings.Contains(err.Error(), "バックアップ") {
		t.Errorf("Expected error message to mention backup, got: %v", err)
	}

	// Verify repository was NOT deleted
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		t.Error("Repository should NOT have been deleted when backup doesn't exist")
	}
}

// TestMultipleDeleteRestoreCycles tests multiple delete/restore cycles
func TestMultipleDeleteRestoreCycles(t *testing.T) {
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "test-repo")
	configPath := filepath.Join(tempDir, "config.yml")

	// Create repository directory
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create repository directory: %v", err)
	}

	// Create test file
	testFile := filepath.Join(repoPath, ".env")
	testContent := "CYCLE=1"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create config file
	configContent := `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    repository_path: ` + repoPath + `
    backup_paths:
      - .env
    backup_retention: 5
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Set global config
	originalCfgFile := cfgFile
	cfgFile = configPath
	defer func() { cfgFile = originalCfgFile }()

	// Setup backup directory cleanup
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}
	backupDir := filepath.Join(homeDir, ".config", "toske", "backups", "test-project")
	defer os.RemoveAll(backupDir)

	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Fatalf("chdir back to %s: %v", originalWd, err)
		}
	}()

	// Perform multiple cycles
	for cycle := 1; cycle <= 3; cycle++ {
		t.Logf("Cycle %d: backup → delete → restore", cycle)

		// Update file content for this cycle
		cycleContent := fmt.Sprintf("CYCLE=%d", cycle)
		if err := os.WriteFile(testFile, []byte(cycleContent), 0644); err != nil {
			t.Fatalf("Cycle %d: Failed to update test file: %v", cycle, err)
		}

		// Backup (from repository directory)
		if err := os.Chdir(repoPath); err != nil {
			t.Fatalf("Cycle %d: Failed to change to repo directory: %v", cycle, err)
		}

		originalProjectName := projectName
		projectName = "test-project"
		if err := runBackup(); err != nil {
			t.Fatalf("Cycle %d: Backup failed: %v", cycle, err)
		}
		projectName = originalProjectName

		if err := os.Chdir(originalWd); err != nil {
			t.Fatalf("Cycle %d: chdir back to %s: %v", cycle, originalWd, err)
		}

		// Delete
		originalDeleteProjectName := deleteProjectName
		originalDeleteForce := deleteForce
		deleteProjectName = "test-project"
		deleteForce = true
		if err := runDelete(); err != nil {
			t.Fatalf("Cycle %d: Delete failed: %v", cycle, err)
		}
		deleteProjectName = originalDeleteProjectName
		deleteForce = originalDeleteForce

		// Verify deletion
		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Errorf("Cycle %d: File should have been deleted", cycle)
		}

		// Recreate directory for restore
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			t.Fatalf("Cycle %d: Failed to recreate directory: %v", cycle, err)
		}

		// Restore (from repository directory)
		if err := os.Chdir(repoPath); err != nil {
			t.Fatalf("Cycle %d: Failed to change to repo directory: %v", cycle, err)
		}

		originalRestoreProjectName := restoreProjectName
		originalForceRestore := forceRestore
		restoreProjectName = "test-project"
		forceRestore = true
		if err := runRestore(); err != nil {
			t.Fatalf("Cycle %d: Restore failed: %v", cycle, err)
		}
		restoreProjectName = originalRestoreProjectName
		forceRestore = originalForceRestore

		if err := os.Chdir(originalWd); err != nil {
			t.Fatalf("Cycle %d: chdir back to %s: %v", cycle, originalWd, err)
		}

		// Verify restoration
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Cycle %d: Failed to read restored file: %v", cycle, err)
		}
		if string(content) != cycleContent {
			t.Errorf("Cycle %d: Content mismatch. Expected: %s, Got: %s", cycle, cycleContent, string(content))
		}
	}

	t.Log("✓ Multiple delete/restore cycles completed successfully")
}
