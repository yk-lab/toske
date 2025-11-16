package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestRunPrune(t *testing.T) {
	tests := []struct {
		name              string
		projectName       string
		all               bool
		keep              int
		keepExplicit      bool
		configData        string
		setupBackups      func(t *testing.T, homeDir string)
		expectError       bool
		errorMessage      string
		validateResult    func(t *testing.T, homeDir string)
	}{
		{
			name:         "prune specific project with --keep",
			projectName:  "test-project",
			all:          false,
			keep:         2,
			keepExplicit: true,
			configData: `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    backup_paths:
      - .env
`,
			setupBackups: func(t *testing.T, homeDir string) {
				createTestBackups(t, homeDir, "test-project", 5)
			},
			expectError: false,
			validateResult: func(t *testing.T, homeDir string) {
				// Verify only 2 backups remain
				backupDir := filepath.Join(homeDir, ".config", "toske", "backups", "test-project")
				metadata := readBackupMetadata(t, backupDir)
				if len(metadata.Backups) != 2 {
					t.Errorf("Expected 2 backups after prune, got %d", len(metadata.Backups))
				}
			},
		},
		{
			name:         "prune specific project with backup_retention",
			projectName:  "test-project",
			all:          false,
			keep:         0,
			keepExplicit: false,
			configData: `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    backup_paths:
      - .env
    backup_retention: 3
`,
			setupBackups: func(t *testing.T, homeDir string) {
				createTestBackups(t, homeDir, "test-project", 6)
			},
			expectError: false,
			validateResult: func(t *testing.T, homeDir string) {
				backupDir := filepath.Join(homeDir, ".config", "toske", "backups", "test-project")
				metadata := readBackupMetadata(t, backupDir)
				if len(metadata.Backups) != 3 {
					t.Errorf("Expected 3 backups after prune, got %d", len(metadata.Backups))
				}
			},
		},
		{
			name:         "prune all projects",
			projectName:  "",
			all:          true,
			keep:         2,
			keepExplicit: true,
			configData: `version: 1.0.0
projects:
  - name: project-one
    repo: git@github.com:user/one.git
    branch: main
  - name: project-two
    repo: git@github.com:user/two.git
    branch: main
`,
			setupBackups: func(t *testing.T, homeDir string) {
				createTestBackups(t, homeDir, "project-one", 4)
				createTestBackups(t, homeDir, "project-two", 5)
			},
			expectError: false,
			validateResult: func(t *testing.T, homeDir string) {
				// Verify both projects have 2 backups
				backupDir1 := filepath.Join(homeDir, ".config", "toske", "backups", "project-one")
				metadata1 := readBackupMetadata(t, backupDir1)
				if len(metadata1.Backups) != 2 {
					t.Errorf("Expected 2 backups for project-one, got %d", len(metadata1.Backups))
				}

				backupDir2 := filepath.Join(homeDir, ".config", "toske", "backups", "project-two")
				metadata2 := readBackupMetadata(t, backupDir2)
				if len(metadata2.Backups) != 2 {
					t.Errorf("Expected 2 backups for project-two, got %d", len(metadata2.Backups))
				}
			},
		},
		{
			name:         "error when both --project and --all specified",
			projectName:  "test-project",
			all:          true,
			keep:         2,
			keepExplicit: true,
			configData: `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
`,
			setupBackups: func(t *testing.T, homeDir string) {},
			expectError:  true,
			errorMessage: "Cannot specify both",
		},
		{
			name:         "error when neither --project nor --all specified",
			projectName:  "",
			all:          false,
			keep:         2,
			keepExplicit: true,
			configData: `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
`,
			setupBackups: func(t *testing.T, homeDir string) {},
			expectError:  true,
			errorMessage: "Either --project or --all",
		},
		{
			name:         "error when project not found",
			projectName:  "nonexistent",
			all:          false,
			keep:         2,
			keepExplicit: true,
			configData: `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
`,
			setupBackups: func(t *testing.T, homeDir string) {},
			expectError:  true,
			errorMessage: "not found in configuration file",
		},
		{
			name:         "skip when no retention policy",
			projectName:  "test-project",
			all:          false,
			keep:         0,
			keepExplicit: false,
			configData: `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
`,
			setupBackups: func(t *testing.T, homeDir string) {
				createTestBackups(t, homeDir, "test-project", 3)
			},
			expectError: false,
			validateResult: func(t *testing.T, homeDir string) {
				// Verify backups are unchanged (not pruned)
				backupDir := filepath.Join(homeDir, ".config", "toske", "backups", "test-project")
				metadata := readBackupMetadata(t, backupDir)
				if len(metadata.Backups) != 3 {
					t.Errorf("Expected 3 backups (unchanged), got %d", len(metadata.Backups))
				}
			},
		},
		{
			name:         "--keep overrides backup_retention",
			projectName:  "test-project",
			all:          false,
			keep:         5,
			keepExplicit: true,
			configData: `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    backup_retention: 2
`,
			setupBackups: func(t *testing.T, homeDir string) {
				createTestBackups(t, homeDir, "test-project", 7)
			},
			expectError: false,
			validateResult: func(t *testing.T, homeDir string) {
				// Verify 5 backups remain (not 2)
				backupDir := filepath.Join(homeDir, ".config", "toske", "backups", "test-project")
				metadata := readBackupMetadata(t, backupDir)
				if len(metadata.Backups) != 5 {
					t.Errorf("Expected 5 backups (--keep should override), got %d", len(metadata.Backups))
				}
			},
		},
		{
			name:         "partial failure in --all mode",
			projectName:  "",
			all:          true,
			keep:         2,
			keepExplicit: true,
			configData: `version: 1.0.0
projects:
  - name: project-with-backups
    repo: git@github.com:user/one.git
    branch: main
  - name: project-no-backups
    repo: git@github.com:user/two.git
    branch: main
`,
			setupBackups: func(t *testing.T, homeDir string) {
				// Only create backups for one project
				createTestBackups(t, homeDir, "project-with-backups", 4)
			},
			expectError:  true,
			errorMessage: "Failed to prune",
		},
		{
			name:         "--all mode with some projects already within limit",
			projectName:  "",
			all:          true,
			keep:         5,
			keepExplicit: true,
			configData: `version: 1.0.0
projects:
  - name: project-needs-prune
    repo: git@github.com:user/one.git
    branch: main
  - name: project-within-limit
    repo: git@github.com:user/two.git
    branch: main
`,
			setupBackups: func(t *testing.T, homeDir string) {
				createTestBackups(t, homeDir, "project-needs-prune", 8)
				createTestBackups(t, homeDir, "project-within-limit", 3)
			},
			expectError: false,
			validateResult: func(t *testing.T, homeDir string) {
				// Verify project-needs-prune has 5 backups
				backupDir1 := filepath.Join(homeDir, ".config", "toske", "backups", "project-needs-prune")
				metadata1 := readBackupMetadata(t, backupDir1)
				if len(metadata1.Backups) != 5 {
					t.Errorf("Expected 5 backups for project-needs-prune, got %d", len(metadata1.Backups))
				}

				// Verify project-within-limit still has 3 backups (unchanged)
				backupDir2 := filepath.Join(homeDir, ".config", "toske", "backups", "project-within-limit")
				metadata2 := readBackupMetadata(t, backupDir2)
				if len(metadata2.Backups) != 3 {
					t.Errorf("Expected 3 backups for project-within-limit (unchanged), got %d", len(metadata2.Backups))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup config
			defer setupTestConfig(t, tt.configData)()

			// Setup test home directory
			tempHome := t.TempDir()
			originalHome := os.Getenv("HOME")
			os.Setenv("HOME", tempHome)
			defer os.Setenv("HOME", originalHome)

			// Setup backups if needed
			if tt.setupBackups != nil {
				tt.setupBackups(t, tempHome)
			}

			// Set flags
			originalProjectName := pruneProjectName
			originalAll := pruneAll
			originalKeep := pruneKeep
			pruneProjectName = tt.projectName
			pruneAll = tt.all
			pruneKeep = tt.keep
			defer func() {
				pruneProjectName = originalProjectName
				pruneAll = originalAll
				pruneKeep = originalKeep
			}()

			// Run prune
			err := runPrune(tt.keepExplicit)

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

				if tt.validateResult != nil {
					tt.validateResult(t, tempHome)
				}
			}
		})
	}
}

func TestPruneNoBackupDirectory(t *testing.T) {
	configData := `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    backup_retention: 3
`

	defer setupTestConfig(t, configData)()

	// Setup test home directory without creating backup directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Set flags
	originalProjectName := pruneProjectName
	originalKeep := pruneKeep
	pruneProjectName = "test-project"
	pruneKeep = 0
	defer func() {
		pruneProjectName = originalProjectName
		pruneKeep = originalKeep
	}()

	// Run prune - should fail because backup directory doesn't exist
	err := runPrune(false)
	if err == nil {
		t.Errorf("Expected error for non-existent backup directory but got nil")
	}
}

func TestPruneNoMetadata(t *testing.T) {
	configData := `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    backup_retention: 3
`

	defer setupTestConfig(t, configData)()

	// Setup test home directory with backup directory but no metadata
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Create backup directory without metadata file
	backupDir := filepath.Join(tempHome, ".config", "toske", "backups", "test-project")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}

	// Set flags
	originalProjectName := pruneProjectName
	originalKeep := pruneKeep
	pruneProjectName = "test-project"
	pruneKeep = 0
	defer func() {
		pruneProjectName = originalProjectName
		pruneKeep = originalKeep
	}()

	// Run prune - should fail because metadata doesn't exist
	err := runPrune(false)
	if err == nil {
		t.Errorf("Expected error for missing metadata but got nil")
	}
}

func TestPruneAlreadyWithinLimit(t *testing.T) {
	configData := `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
    backup_retention: 5
`

	defer setupTestConfig(t, configData)()

	// Setup test home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Create only 3 backups (less than retention limit of 5)
	createTestBackups(t, tempHome, "test-project", 3)

	// Set flags
	originalProjectName := pruneProjectName
	originalKeep := pruneKeep
	pruneProjectName = "test-project"
	pruneKeep = 0
	defer func() {
		pruneProjectName = originalProjectName
		pruneKeep = originalKeep
	}()

	// Run prune - should succeed with skip message
	err := runPrune(false)
	if err != nil {
		t.Errorf("Expected no error when backup count is already within limit, got: %v", err)
	}

	// Verify backups were not modified
	backupDir := filepath.Join(tempHome, ".config", "toske", "backups", "test-project")
	metadata := readBackupMetadata(t, backupDir)
	if len(metadata.Backups) != 3 {
		t.Errorf("Expected 3 backups to remain unchanged, got %d", len(metadata.Backups))
	}
}

func TestDetermineRetention(t *testing.T) {
	tests := []struct {
		name              string
		project           Project
		keepFlag          int
		keepExplicit      bool
		expectedRetention int
		expectedSkip      bool
	}{
		{
			name: "--keep flag takes priority",
			project: Project{
				Name:            "test",
				BackupRetention: 3,
			},
			keepFlag:          5,
			keepExplicit:      true,
			expectedRetention: 5,
			expectedSkip:      false,
		},
		{
			name: "use backup_retention when --keep not specified",
			project: Project{
				Name:            "test",
				BackupRetention: 4,
			},
			keepFlag:          0,
			keepExplicit:      false,
			expectedRetention: 4,
			expectedSkip:      false,
		},
		{
			name: "skip when neither is set",
			project: Project{
				Name:            "test",
				BackupRetention: 0,
			},
			keepFlag:          0,
			keepExplicit:      false,
			expectedRetention: 0,
			expectedSkip:      true,
		},
		{
			name: "--keep=1 is valid",
			project: Project{
				Name:            "test",
				BackupRetention: 0,
			},
			keepFlag:          1,
			keepExplicit:      true,
			expectedRetention: 1,
			expectedSkip:      false,
		},
		{
			name: "--keep=0 explicitly set",
			project: Project{
				Name:            "test",
				BackupRetention: 5,
			},
			keepFlag:          0,
			keepExplicit:      true,
			expectedRetention: 0,
			expectedSkip:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retention, skip, err := determineRetention(tt.project, tt.keepFlag, tt.keepExplicit)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if retention != tt.expectedRetention {
				t.Errorf("Expected retention %d, got %d", tt.expectedRetention, retention)
			}

			if skip != tt.expectedSkip {
				t.Errorf("Expected skip %v, got %v", tt.expectedSkip, skip)
			}
		})
	}
}

func TestPrunePreservesNewestBackups(t *testing.T) {
	configData := `version: 1.0.0
projects:
  - name: test-project
    repo: git@github.com:user/test.git
    branch: main
`

	defer setupTestConfig(t, configData)()

	// Setup test home directory
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Create 5 backups with specific timestamps
	backupDir := filepath.Join(tempHome, ".config", "toske", "backups", "test-project")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}

	baseTime := time.Now()
	metadata := BackupMetadata{
		Project: "test-project",
		Backups: []BackupRecord{
			{
				Filename:  "backup_newest.tar.gz",
				Timestamp: baseTime,
				Files:     []string{"file1"},
			},
			{
				Filename:  "backup_second.tar.gz",
				Timestamp: baseTime.Add(-1 * time.Hour),
				Files:     []string{"file1"},
			},
			{
				Filename:  "backup_third.tar.gz",
				Timestamp: baseTime.Add(-2 * time.Hour),
				Files:     []string{"file1"},
			},
			{
				Filename:  "backup_fourth.tar.gz",
				Timestamp: baseTime.Add(-3 * time.Hour),
				Files:     []string{"file1"},
			},
			{
				Filename:  "backup_oldest.tar.gz",
				Timestamp: baseTime.Add(-4 * time.Hour),
				Files:     []string{"file1"},
			},
		},
	}

	// Create actual backup files
	for _, backup := range metadata.Backups {
		backupPath := filepath.Join(backupDir, backup.Filename)
		if err := os.WriteFile(backupPath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create backup file: %v", err)
		}
	}

	// Write metadata
	data, err := yaml.Marshal(&metadata)
	if err != nil {
		t.Fatalf("Failed to marshal metadata: %v", err)
	}
	metadataPath := filepath.Join(backupDir, "backups.yaml")
	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		t.Fatalf("Failed to write metadata: %v", err)
	}

	// Set flags to keep only 2 backups
	originalProjectName := pruneProjectName
	originalKeep := pruneKeep
	pruneProjectName = "test-project"
	pruneKeep = 2
	defer func() {
		pruneProjectName = originalProjectName
		pruneKeep = originalKeep
	}()

	// Run prune
	err = runPrune(true) // keepExplicit = true because we're testing --keep=2
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify that only the 2 newest backups remain
	newMetadata := readBackupMetadata(t, backupDir)
	if len(newMetadata.Backups) != 2 {
		t.Fatalf("Expected 2 backups, got %d", len(newMetadata.Backups))
	}

	if newMetadata.Backups[0].Filename != "backup_newest.tar.gz" {
		t.Errorf("Expected newest backup to be preserved, got %s", newMetadata.Backups[0].Filename)
	}
	if newMetadata.Backups[1].Filename != "backup_second.tar.gz" {
		t.Errorf("Expected second newest backup to be preserved, got %s", newMetadata.Backups[1].Filename)
	}

	// Verify old backup files are deleted
	for _, filename := range []string{"backup_third.tar.gz", "backup_fourth.tar.gz", "backup_oldest.tar.gz"} {
		backupPath := filepath.Join(backupDir, filename)
		if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
			t.Errorf("Expected old backup %s to be deleted, but it still exists", filename)
		}
	}
}

// Helper function to create test backups
func createTestBackups(t *testing.T, homeDir, projectName string, count int) {
	backupDir := filepath.Join(homeDir, ".config", "toske", "backups", projectName)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}

	metadata := BackupMetadata{
		Project: projectName,
		Backups: []BackupRecord{},
	}

	baseTime := time.Now()
	for i := 0; i < count; i++ {
		filename := "backup_" + time.Now().Format("20060102_150405.000000") + ".tar.gz"
		timestamp := baseTime.Add(-time.Duration(i) * time.Hour)

		// Create backup file
		backupPath := filepath.Join(backupDir, filename)
		if err := os.WriteFile(backupPath, []byte("test backup"), 0644); err != nil {
			t.Fatalf("Failed to create backup file: %v", err)
		}

		// Add to metadata
		metadata.Backups = append(metadata.Backups, BackupRecord{
			Filename:  filename,
			Timestamp: timestamp,
			Files:     []string{"test.txt"},
		})

		// Sleep briefly to ensure unique filenames
		time.Sleep(1 * time.Millisecond)
	}

	// Write metadata file
	data, err := yaml.Marshal(&metadata)
	if err != nil {
		t.Fatalf("Failed to marshal metadata: %v", err)
	}

	metadataPath := filepath.Join(backupDir, "backups.yaml")
	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		t.Fatalf("Failed to write metadata file: %v", err)
	}
}

// Helper function to read backup metadata
func readBackupMetadata(t *testing.T, backupDir string) BackupMetadata {
	metadataPath := filepath.Join(backupDir, "backups.yaml")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatalf("Failed to read metadata: %v", err)
	}

	var metadata BackupMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}

	return metadata
}
