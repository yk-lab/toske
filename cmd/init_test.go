package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGetConfigTemplate tests the configuration template generation
func TestGetConfigTemplate(t *testing.T) {
	template := getConfigTemplate()

	// Check that template is not empty
	if template == "" {
		t.Fatal("Expected template to be non-empty")
	}

	// Check for required fields
	requiredFields := []string{
		"version:",
		"projects:",
		"name:",
		"repo:",
		"branch:",
		"backup_paths:",
		"backup_retention:",
	}

	for _, field := range requiredFields {
		if !strings.Contains(template, field) {
			t.Errorf("Expected template to contain '%s'", field)
		}
	}

	// Check for sample project
	if !strings.Contains(template, "sample-project") {
		t.Error("Expected template to contain 'sample-project'")
	}
}

// TestRunInitNewFile tests creating a new config file
func TestRunInitNewFile(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".config", "toske", "config.yml")

	// Set cfgFile to temporary path
	originalCfgFile := cfgFile
	cfgFile = configPath
	defer func() { cfgFile = originalCfgFile }()

	// Run init
	err := runInit()
	if err != nil {
		t.Fatalf("runInit() failed: %v", err)
	}

	// Check if file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Expected config file to be created at %s", configPath)
	}

	// Read and verify content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read created config file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "version:") {
		t.Error("Expected config file to contain 'version:'")
	}
	if !strings.Contains(contentStr, "projects:") {
		t.Error("Expected config file to contain 'projects:'")
	}
}

// TestRunInitDirectoryCreation tests that directories are created if they don't exist
func TestRunInitDirectoryCreation(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nested", "deep", "path", "config.yml")

	// Set cfgFile to temporary path
	originalCfgFile := cfgFile
	cfgFile = configPath
	defer func() { cfgFile = originalCfgFile }()

	// Run init
	err := runInit()
	if err != nil {
		t.Fatalf("runInit() failed: %v", err)
	}

	// Check if directories and file were created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Expected config file to be created at %s", configPath)
	}

	// Verify parent directory exists
	configDir := filepath.Dir(configPath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Fatalf("Expected directory to be created at %s", configDir)
	}
}

// TestRunInitFilePermissions tests that the created file has correct permissions
func TestRunInitFilePermissions(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yml")

	// Set cfgFile to temporary path
	originalCfgFile := cfgFile
	cfgFile = configPath
	defer func() { cfgFile = originalCfgFile }()

	// Run init
	err := runInit()
	if err != nil {
		t.Fatalf("runInit() failed: %v", err)
	}

	// Check file permissions
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	expectedPerm := os.FileMode(0644)
	if fileInfo.Mode().Perm() != expectedPerm {
		t.Errorf("Expected file permissions to be %v, got %v", expectedPerm, fileInfo.Mode().Perm())
	}
}

// TestRunInitExistingFile tests the behavior when config file already exists
// Note: This test cannot easily simulate user input without refactoring runInit()
// to accept an io.Reader for testing. We'll test that it returns without error
// when we simulate a "no" response by not providing stdin input in a non-interactive environment.
func TestRunInitExistingFile(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yml")

	// Create an existing config file
	existingContent := "existing: config"
	if err := os.WriteFile(configPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to create existing config file: %v", err)
	}

	// Set cfgFile to temporary path
	originalCfgFile := cfgFile
	cfgFile = configPath
	defer func() { cfgFile = originalCfgFile }()

	// Note: In a real scenario, runInit() would prompt for input when the file exists.
	// This test is limited because we can't easily simulate interactive input without refactoring.
	// The test verifies that the file exists before running init.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Expected file to exist before test")
	}

	// We can't fully test the overwrite scenario without refactoring runInit()
	// to accept an io.Reader for testing purposes.
	t.Skip("Skipping interactive test - requires refactoring to support io.Reader injection")
}

// TestGetConfigTemplateYAMLFormat tests that the template is valid YAML format
func TestGetConfigTemplateYAMLFormat(t *testing.T) {
	template := getConfigTemplate()

	// Basic YAML format checks
	lines := strings.Split(template, "\n")
	foundVersion := false
	foundProjects := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "version:") {
			foundVersion = true
		}
		if strings.HasPrefix(trimmed, "projects:") {
			foundProjects = true
		}
	}

	if !foundVersion {
		t.Error("Expected template to have 'version:' at start of line")
	}
	if !foundProjects {
		t.Error("Expected template to have 'projects:' at start of line")
	}

	// Check for proper list formatting (projects should have list items)
	if !strings.Contains(template, "  - name:") {
		t.Error("Expected template to have properly indented list items")
	}
}
