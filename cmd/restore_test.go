package cmd

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestRunRestore(t *testing.T) {
	tests := []struct {
		name         string
		projectName  string
		configData   string
		setupBackup  func(string) error
		backupIndex  int
		forceRestore bool
		expectError  bool
		errorMessage string
	}{
		{
			name:        "successful restore",
			projectName: "test-project",
			configData: `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    backup_paths:
      - .env
    backup_retention: 3
`,
			setupBackup: func(tempDir string) error {
				return createTestBackup(tempDir, "test-project", []testFile{
					{name: ".env", content: "TEST=value"},
				})
			},
			backupIndex:  1,
			forceRestore: true,
			expectError:  false,
		},
		{
			name:        "restore with multiple files",
			projectName: "multi-file-project",
			configData: `version: 1.0.0
projects:
  - name: multi-file-project
    repo: git@github.com:user/multi.git
    branch: main
    backup_paths:
      - .env
      - db.sqlite3
    backup_retention: 3
`,
			setupBackup: func(tempDir string) error {
				return createTestBackup(tempDir, "multi-file-project", []testFile{
					{name: ".env", content: "TEST=value"},
					{name: "db.sqlite3", content: "database"},
				})
			},
			backupIndex:  1,
			forceRestore: true,
			expectError:  false,
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
			setupBackup:  func(tempDir string) error { return nil },
			backupIndex:  1,
			forceRestore: true,
			expectError:  true,
			errorMessage: "not found in configuration file",
		},
		{
			name:         "missing project flag",
			projectName:  "",
			configData:   `version: 1.0.0\nprojects: []`,
			setupBackup:  func(tempDir string) error { return nil },
			backupIndex:  1,
			forceRestore: true,
			expectError:  true,
			errorMessage: "Project name is required",
		},
		{
			name:        "no backup directory",
			projectName: "no-backup-project",
			configData: `version: 1.0.0
projects:
  - name: no-backup-project
    repo: git@github.com:user/test.git
    branch: main
    backup_paths:
      - .env
`,
			setupBackup:  func(tempDir string) error { return nil },
			backupIndex:  1,
			forceRestore: true,
			expectError:  true,
			errorMessage: "No backup directory found",
		},
		{
			name:        "invalid backup index",
			projectName: "test-project",
			configData: `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    backup_paths:
      - .env
`,
			setupBackup: func(tempDir string) error {
				return createTestBackup(tempDir, "test-project", []testFile{
					{name: ".env", content: "TEST=value"},
				})
			},
			backupIndex:  99,
			forceRestore: true,
			expectError:  true,
			errorMessage: "Invalid backup index",
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

			// ja: テスト環境を分離するため、HOMEを一時ディレクトリに設定
			// en: Isolate test environment by setting HOME to temp directory
			t.Setenv("HOME", tempDir)
			t.Setenv("USERPROFILE", tempDir) // Windows support

			// Setup config
			defer setupTestConfig(t, tt.configData)()

			// Setup test backup
			if err := tt.setupBackup(tempDir); err != nil {
				t.Fatalf("Failed to setup test backup: %v", err)
			}

			// Change to work directory
			originalWd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get working directory: %v", err)
			}
			defer os.Chdir(originalWd)

			if err := os.Chdir(workDir); err != nil {
				t.Fatalf("Failed to change directory: %v", err)
			}

			// Set restore parameters
			originalProjectName := restoreProjectName
			originalBackupIndex := backupIndex
			originalForceRestore := forceRestore
			restoreProjectName = tt.projectName
			backupIndex = tt.backupIndex
			forceRestore = tt.forceRestore
			defer func() {
				restoreProjectName = originalProjectName
				backupIndex = originalBackupIndex
				forceRestore = originalForceRestore
			}()

			// Run restore
			err = runRestore()

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

func TestRestoreFileContents(t *testing.T) {
	// Setup temporary directories
	tempDir := t.TempDir()
	workDir := filepath.Join(tempDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}

	// ja: テスト環境を分離するため、HOMEを一時ディレクトリに設定
	// en: Isolate test environment by setting HOME to temp directory
	t.Setenv("HOME", tempDir)
	t.Setenv("USERPROFILE", tempDir) // Windows support

	// Setup config
	configData := `version: 1.0.0
projects:
  - name: content-test
    repo: git@github.com:user/test.git
    branch: main
    backup_paths:
      - .env
      - config.json
    backup_retention: 3
`
	defer setupTestConfig(t, configData)()

	// Create test backup
	testFiles := []testFile{
		{name: ".env", content: "TEST_ENV=restored"},
		{name: "config.json", content: `{"key": "value"}`},
	}
	if err := createTestBackup(tempDir, "content-test", testFiles); err != nil {
		t.Fatalf("Failed to create test backup: %v", err)
	}

	// Change to work directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Set restore parameters
	originalProjectName := restoreProjectName
	originalBackupIndex := backupIndex
	originalForceRestore := forceRestore
	restoreProjectName = "content-test"
	backupIndex = 1
	forceRestore = true
	defer func() {
		restoreProjectName = originalProjectName
		backupIndex = originalBackupIndex
		forceRestore = originalForceRestore
	}()

	// Run restore
	if err := runRestore(); err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Verify restored file contents
	for _, tf := range testFiles {
		data, err := os.ReadFile(filepath.Join(workDir, tf.name))
		if err != nil {
			t.Errorf("Failed to read restored file %s: %v", tf.name, err)
			continue
		}

		if string(data) != tf.content {
			t.Errorf("File %s content mismatch. Expected: %s, Got: %s", tf.name, tf.content, string(data))
		}
	}
}

func TestRestoreWithMultipleBackups(t *testing.T) {
	// Setup temporary directories
	tempDir := t.TempDir()
	workDir := filepath.Join(tempDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}

	// ja: テスト環境を分離するため、HOMEを一時ディレクトリに設定
	// en: Isolate test environment by setting HOME to temp directory
	t.Setenv("HOME", tempDir)
	t.Setenv("USERPROFILE", tempDir) // Windows support

	// Setup config
	configData := `version: 1.0.0
projects:
  - name: multi-backup-test
    repo: git@github.com:user/test.git
    branch: main
    backup_paths:
      - .env
    backup_retention: 5
`
	defer setupTestConfig(t, configData)()

	// Create multiple backups with different content (oldest to newest)
	backups := []struct {
		content string
		delay   time.Duration
	}{
		{content: "VERSION=1.0", delay: 0},
		{content: "VERSION=2.0", delay: 10 * time.Millisecond},
		{content: "VERSION=3.0", delay: 20 * time.Millisecond},
	}

	for i, b := range backups {
		if i > 0 {
			time.Sleep(b.delay)
		}
		if err := createTestBackup(tempDir, "multi-backup-test", []testFile{
			{name: ".env", content: b.content},
		}); err != nil {
			t.Fatalf("Failed to create backup %d: %v", i+1, err)
		}
	}

	// Change to work directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Test restoring latest backup (index 1)
	t.Run("restore latest backup", func(t *testing.T) {
		originalProjectName := restoreProjectName
		originalBackupIndex := backupIndex
		originalForceRestore := forceRestore
		restoreProjectName = "multi-backup-test"
		backupIndex = 1
		forceRestore = true
		defer func() {
			restoreProjectName = originalProjectName
			backupIndex = originalBackupIndex
			forceRestore = originalForceRestore
		}()

		if err := runRestore(); err != nil {
			t.Fatalf("Restore failed: %v", err)
		}

		data, err := os.ReadFile(filepath.Join(workDir, ".env"))
		if err != nil {
			t.Fatalf("Failed to read restored file: %v", err)
		}

		// Latest backup should have VERSION=3.0
		if string(data) != "VERSION=3.0" {
			t.Errorf("Expected latest backup content 'VERSION=3.0', got: %s", string(data))
		}
	})

	// Test restoring second latest backup (index 2)
	t.Run("restore second backup", func(t *testing.T) {
		originalProjectName := restoreProjectName
		originalBackupIndex := backupIndex
		originalForceRestore := forceRestore
		restoreProjectName = "multi-backup-test"
		backupIndex = 2
		forceRestore = true
		defer func() {
			restoreProjectName = originalProjectName
			backupIndex = originalBackupIndex
			forceRestore = originalForceRestore
		}()

		if err := runRestore(); err != nil {
			t.Fatalf("Restore failed: %v", err)
		}

		data, err := os.ReadFile(filepath.Join(workDir, ".env"))
		if err != nil {
			t.Fatalf("Failed to read restored file: %v", err)
		}

		// Second backup should have VERSION=2.0
		if string(data) != "VERSION=2.0" {
			t.Errorf("Expected second backup content 'VERSION=2.0', got: %s", string(data))
		}
	})
}

func TestExtractBackupArchive(t *testing.T) {
	// Setup temporary directory
	tempDir := t.TempDir()
	workDir := filepath.Join(tempDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}

	// Create a test archive
	archivePath := filepath.Join(tempDir, "test.tar.gz")
	testFiles := []testFile{
		{name: ".env", content: "TEST=value"},
		{name: "config/app.conf", content: "config content"},
	}

	if err := createTestArchive(archivePath, testFiles); err != nil {
		t.Fatalf("Failed to create test archive: %v", err)
	}

	// Change to work directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Extract archive
	fileCount, err := extractBackupArchive(archivePath)
	if err != nil {
		t.Fatalf("Failed to extract archive: %v", err)
	}

	if fileCount != len(testFiles) {
		t.Errorf("Expected %d files to be extracted, got %d", len(testFiles), fileCount)
	}

	// Verify extracted files
	for _, tf := range testFiles {
		filePath := filepath.Join(workDir, tf.name)
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Failed to read extracted file %s: %v", tf.name, err)
			continue
		}

		if string(data) != tf.content {
			t.Errorf("File %s content mismatch. Expected: %s, Got: %s", tf.name, tf.content, string(data))
		}
	}
}

func TestExtractBackupArchivePathTraversal(t *testing.T) {
	// Setup temporary directory
	tempDir := t.TempDir()
	workDir := filepath.Join(tempDir, "work")
	siblingDir := filepath.Join(tempDir, "sibling")

	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}
	if err := os.MkdirAll(siblingDir, 0755); err != nil {
		t.Fatalf("Failed to create sibling directory: %v", err)
	}

	// Create an archive with path traversal attempts
	archivePath := filepath.Join(tempDir, "malicious.tar.gz")
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("Failed to create archive file: %v", err)
	}
	defer archiveFile.Close()

	gzipWriter := gzip.NewWriter(archiveFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Attempt various path traversal attacks
	maliciousFiles := []struct {
		name    string
		content string
	}{
		{name: "../escape.txt", content: "escaped"},
		{name: "../../escape2.txt", content: "escaped2"},
		{name: "../sibling/attack.txt", content: "attacked"},
		{name: filepath.Join(tempDir, "absolute_attack.txt"), content: "absolute path attack"},
	}

	for _, mf := range maliciousFiles {
		header := &tar.Header{
			Name:    mf.name,
			Size:    int64(len(mf.content)),
			Mode:    0644,
			ModTime: time.Now(),
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatalf("Failed to write tar header: %v", err)
		}
		if _, err := tarWriter.Write([]byte(mf.content)); err != nil {
			t.Fatalf("Failed to write tar content: %v", err)
		}
	}

	tarWriter.Close()
	gzipWriter.Close()
	archiveFile.Close()

	// Change to work directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Extract archive - should skip all malicious files
	fileCount, err := extractBackupArchive(archivePath)
	if err != nil {
		t.Fatalf("Failed to extract archive: %v", err)
	}

	// Should extract 0 files (all were malicious)
	if fileCount != 0 {
		t.Errorf("Expected 0 files to be extracted (all malicious), got %d", fileCount)
	}

	// Verify no files escaped the work directory
	escapedFiles := []string{
		filepath.Join(tempDir, "escape.txt"),
		filepath.Join(tempDir, "escape2.txt"),
		filepath.Join(siblingDir, "attack.txt"),
		filepath.Join(tempDir, "absolute_attack.txt"),
	}

	for _, escapedFile := range escapedFiles {
		data, err := os.ReadFile(escapedFile)
		if err == nil {
			t.Errorf("Security vulnerability: file escaped to %s with content: %s", escapedFile, string(data))
		}
	}

	// Verify work directory is still empty
	entries, err := os.ReadDir(workDir)
	if err != nil {
		t.Fatalf("Failed to read work directory: %v", err)
	}
	if len(entries) > 0 {
		t.Errorf("Expected work directory to be empty, but found %d entries", len(entries))
	}
}

// Helper types and functions

type testFile struct {
	name    string
	content string
}

// createTestBackup creates a test backup in the specified temp directory
func createTestBackup(tempDir, projectName string, files []testFile) error {
	backupDir := filepath.Join(tempDir, ".config", "toske", "backups", projectName)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}

	// Create archive
	timestamp := time.Now()
	archiveFilename := "backup_" + timestamp.Format("20060102_150405.000000") + ".tar.gz"
	archivePath := filepath.Join(backupDir, archiveFilename)

	if err := createTestArchive(archivePath, files); err != nil {
		return err
	}

	// Read existing metadata if it exists
	metadataPath := filepath.Join(backupDir, "backups.yaml")
	var metadata BackupMetadata
	if data, err := os.ReadFile(metadataPath); err == nil {
		if err := yaml.Unmarshal(data, &metadata); err != nil {
			return err
		}
	}

	// Set project name
	metadata.Project = projectName

	// Add new backup record
	var fileNames []string
	for _, f := range files {
		fileNames = append(fileNames, f.name)
	}

	newBackup := BackupRecord{
		Filename:  archiveFilename,
		Timestamp: timestamp,
		Files:     fileNames,
	}

	// Prepend new backup (newest first)
	metadata.Backups = append([]BackupRecord{newBackup}, metadata.Backups...)

	// Write updated metadata
	data, err := yaml.Marshal(&metadata)
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, data, 0644)
}

// createTestArchive creates a tar.gz archive with the specified files
func createTestArchive(archivePath string, files []testFile) error {
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer archiveFile.Close()

	gzipWriter := gzip.NewWriter(archiveFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	for _, f := range files {
		// Create directory for file if needed
		if dir := filepath.Dir(f.name); dir != "." {
			header := &tar.Header{
				Name:     dir + "/",
				Typeflag: tar.TypeDir,
				Mode:     0755,
				ModTime:  time.Now(),
			}
			if err := tarWriter.WriteHeader(header); err != nil {
				return err
			}
		}

		// Add file
		header := &tar.Header{
			Name:    filepath.ToSlash(f.name),
			Size:    int64(len(f.content)),
			Mode:    0644,
			ModTime: time.Now(),
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if _, err := tarWriter.Write([]byte(f.content)); err != nil {
			return err
		}
	}

	return nil
}
