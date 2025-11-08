package cmd

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestRunBackup(t *testing.T) {
	tests := []struct {
		name         string
		projectName  string
		configData   string
		setupFiles   func(string) error
		expectError  bool
		errorMessage string
	}{
		{
			name:        "successful backup",
			projectName: "test-project",
			configData: `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    backup_paths:
      - .env
      - db.sqlite3
    backup_retention: 3
`,
			setupFiles: func(baseDir string) error {
				if err := os.WriteFile(filepath.Join(baseDir, ".env"), []byte("TEST=value"), 0644); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(baseDir, "db.sqlite3"), []byte("database"), 0644)
			},
			expectError: false,
		},
		{
			name:        "backup with directory",
			projectName: "dir-project",
			configData: `version: 1.0.0
projects:
  - name: dir-project
    repo: git@github.com:user/dir.git
    branch: main
    backup_paths:
      - config/
    backup_retention: 2
`,
			setupFiles: func(baseDir string) error {
				if err := os.MkdirAll(filepath.Join(baseDir, "config"), 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(baseDir, "config", "app.conf"), []byte("config"), 0644)
			},
			expectError: false,
		},
		{
			name:        "project not found",
			projectName: "nonexistent",
			configData: `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    backup_paths:
      - .env
`,
			setupFiles:   func(baseDir string) error { return nil },
			expectError:  true,
			errorMessage: "not found in configuration file",
		},
		{
			name:        "no backup paths",
			projectName: "no-paths-project",
			configData: `version: 1.0.0
projects:
  - name: no-paths-project
    repo: git@github.com:user/test.git
    branch: main
    backup_paths: []
`,
			setupFiles:   func(baseDir string) error { return nil },
			expectError:  true,
			errorMessage: "has no backup_paths configured",
		},
		{
			name:         "missing project flag",
			projectName:  "",
			configData:   `version: 1.0.0\nprojects: []`,
			setupFiles:   func(baseDir string) error { return nil },
			expectError:  true,
			errorMessage: "Project name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup temporary directories
			tempDir := t.TempDir()
			workDir := filepath.Join(tempDir, "work")
			if err := os.MkdirAll(workDir, 0755); err != nil {
				t.Fatalf("Failed to create work directory: %v", err)
			}

			// Setup config
			defer setupTestConfig(t, tt.configData)()

			// Setup test files
			originalWd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get working directory: %v", err)
			}
			defer os.Chdir(originalWd)

			if err := os.Chdir(workDir); err != nil {
				t.Fatalf("Failed to change directory: %v", err)
			}

			if err := tt.setupFiles(workDir); err != nil {
				t.Fatalf("Failed to setup test files: %v", err)
			}

			// Set project name
			originalProjectName := projectName
			projectName = tt.projectName
			defer func() { projectName = originalProjectName }()

			// Run backup
			err = runBackup()

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
			}
		})
	}
}

func TestBackupArchiveContents(t *testing.T) {
	// Setup temporary directories
	tempDir := t.TempDir()
	workDir := filepath.Join(tempDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}

	// Create test files
	testFiles := map[string]string{
		".env":       "TEST=value",
		"db.sqlite3": "database content",
	}

	for filename, content := range testFiles {
		if err := os.WriteFile(filepath.Join(workDir, filename), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Setup config
	configData := `version: 1.0.0
projects:
  - name: archive-test
    repo: git@github.com:user/test.git
    branch: main
    backup_paths:
      - .env
      - db.sqlite3
    backup_retention: 3
`
	defer setupTestConfig(t, configData)()

	// Change to work directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Set project name
	originalProjectName := projectName
	projectName = "archive-test"
	defer func() { projectName = originalProjectName }()

	// Run backup
	if err := runBackup(); err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Verify backup archive exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	backupDir := filepath.Join(homeDir, ".config", "toske", "backups", "archive-test")
	metadataPath := filepath.Join(backupDir, "backups.yaml")

	// Read metadata
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatalf("Failed to read metadata file: %v", err)
	}

	var metadata BackupMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		t.Fatalf("Failed to parse metadata: %v", err)
	}

	if len(metadata.Backups) == 0 {
		t.Fatal("No backups found in metadata")
	}

	// Check the latest backup
	latestBackup := metadata.Backups[0]
	archivePath := filepath.Join(backupDir, latestBackup.Filename)

	// Verify archive exists
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Fatalf("Backup archive does not exist: %s", archivePath)
	}

	// Verify archive contents
	archiveFile, err := os.Open(archivePath)
	if err != nil {
		t.Fatalf("Failed to open archive: %v", err)
	}
	defer archiveFile.Close()

	gzipReader, err := gzip.NewReader(archiveFile)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	foundFiles := make(map[string]bool)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to read tar header: %v", err)
		}

		foundFiles[header.Name] = true
	}

	// Verify all expected files are in the archive
	expectedFiles := []string{".env", "db.sqlite3"}
	for _, expected := range expectedFiles {
		if !foundFiles[expected] {
			t.Errorf("Expected file '%s' not found in archive", expected)
		}
	}
}

func TestBackupRetention(t *testing.T) {
	// Setup temporary directories
	tempDir := t.TempDir()
	workDir := filepath.Join(tempDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}

	// Create test file
	if err := os.WriteFile(filepath.Join(workDir, ".env"), []byte("TEST=value"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Setup config with retention=2
	configData := `version: 1.0.0
projects:
  - name: retention-test
    repo: git@github.com:user/test.git
    branch: main
    backup_paths:
      - .env
    backup_retention: 2
`
	defer setupTestConfig(t, configData)()

	// Change to work directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Set project name
	originalProjectName := projectName
	projectName = "retention-test"
	defer func() { projectName = originalProjectName }()

	// Create 3 backups (with sleep to ensure different timestamps)
	for i := 0; i < 3; i++ {
		if err := runBackup(); err != nil {
			t.Fatalf("Backup %d failed: %v", i+1, err)
		}
		// Sleep to ensure different timestamps (format is up to seconds)
		if i < 2 {
			time.Sleep(1 * time.Second)
		}
	}

	// Verify only 2 backups are kept
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	backupDir := filepath.Join(homeDir, ".config", "toske", "backups", "retention-test")
	metadataPath := filepath.Join(backupDir, "backups.yaml")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatalf("Failed to read metadata file: %v", err)
	}

	var metadata BackupMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		t.Fatalf("Failed to parse metadata: %v", err)
	}

	if len(metadata.Backups) != 2 {
		t.Errorf("Expected 2 backups to be kept, but found %d", len(metadata.Backups))
	}

	// Verify only the 2 latest backup files exist
	for _, backup := range metadata.Backups {
		archivePath := filepath.Join(backupDir, backup.Filename)
		if _, err := os.Stat(archivePath); os.IsNotExist(err) {
			t.Errorf("Expected backup file %s to exist, but it doesn't", backup.Filename)
		}
	}
}
