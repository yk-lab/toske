package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &Config{
				Version: "1.0.0",
				Projects: []Project{
					{
						Name:            "test-project",
						Repo:            "git@github.com:user/repo.git",
						Branch:          "main",
						BackupPaths:     []string{".env"},
						BackupRetention: 3,
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing version",
			config: &Config{
				Projects: []Project{
					{
						Name:   "test-project",
						Repo:   "git@github.com:user/repo.git",
						Branch: "main",
					},
				},
			},
			expectError: true,
		},
		{
			name: "no projects",
			config: &Config{
				Version:  "1.0.0",
				Projects: []Project{},
			},
			expectError: true,
		},
		{
			name: "project without name",
			config: &Config{
				Version: "1.0.0",
				Projects: []Project{
					{
						Repo:   "git@github.com:user/repo.git",
						Branch: "main",
					},
				},
			},
			expectError: true,
		},
		{
			name: "project without repo",
			config: &Config{
				Version: "1.0.0",
				Projects: []Project{
					{
						Name:   "test-project",
						Branch: "main",
					},
				},
			},
			expectError: true,
		},
		{
			name: "project without branch",
			config: &Config{
				Version: "1.0.0",
				Projects: []Project{
					{
						Name: "test-project",
						Repo: "git@github.com:user/repo.git",
					},
				},
			},
			expectError: true,
		},
		{
			name: "duplicate project names",
			config: &Config{
				Version: "1.0.0",
				Projects: []Project{
					{
						Name:   "test-project",
						Repo:   "git@github.com:user/repo1.git",
						Branch: "main",
					},
					{
						Name:   "test-project",
						Repo:   "git@github.com:user/repo2.git",
						Branch: "main",
					},
				},
			},
			expectError: true,
		},
		{
			name: "negative backup retention",
			config: &Config{
				Version: "1.0.0",
				Projects: []Project{
					{
						Name:            "test-project",
						Repo:            "git@github.com:user/repo.git",
						Branch:          "main",
						BackupRetention: -1,
					},
				},
			},
			expectError: true,
		},
		{
			name: "multiple valid projects",
			config: &Config{
				Version: "1.0.0",
				Projects: []Project{
					{
						Name:            "project1",
						Repo:            "git@github.com:user/repo1.git",
						Branch:          "main",
						BackupPaths:     []string{".env"},
						BackupRetention: 3,
					},
					{
						Name:            "project2",
						Repo:            "https://github.com/user/repo2.git",
						Branch:          "develop",
						BackupPaths:     []string{".env", "db/"},
						BackupRetention: 5,
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateProject(t *testing.T) {
	tests := []struct {
		name         string
		project      *Project
		index        int
		projectNames map[string]bool
		expectError  bool
	}{
		{
			name: "valid project",
			project: &Project{
				Name:            "test-project",
				Repo:            "git@github.com:user/repo.git",
				Branch:          "main",
				BackupPaths:     []string{".env"},
				BackupRetention: 3,
			},
			index:        0,
			projectNames: make(map[string]bool),
			expectError:  false,
		},
		{
			name: "project without name",
			project: &Project{
				Repo:   "git@github.com:user/repo.git",
				Branch: "main",
			},
			index:        0,
			projectNames: make(map[string]bool),
			expectError:  true,
		},
		{
			name: "duplicate project name",
			project: &Project{
				Name:   "test-project",
				Repo:   "git@github.com:user/repo.git",
				Branch: "main",
			},
			index: 0,
			projectNames: map[string]bool{
				"test-project": true,
			},
			expectError: true,
		},
		{
			name: "project without repo",
			project: &Project{
				Name:   "test-project",
				Branch: "main",
			},
			index:        0,
			projectNames: make(map[string]bool),
			expectError:  true,
		},
		{
			name: "project without branch",
			project: &Project{
				Name: "test-project",
				Repo: "git@github.com:user/repo.git",
			},
			index:        0,
			projectNames: make(map[string]bool),
			expectError:  true,
		},
		{
			name: "negative backup retention",
			project: &Project{
				Name:            "test-project",
				Repo:            "git@github.com:user/repo.git",
				Branch:          "main",
				BackupRetention: -1,
			},
			index:        0,
			projectNames: make(map[string]bool),
			expectError:  true,
		},
		{
			name: "zero backup retention (valid)",
			project: &Project{
				Name:            "test-project",
				Repo:            "git@github.com:user/repo.git",
				Branch:          "main",
				BackupRetention: 0,
			},
			index:        0,
			projectNames: make(map[string]bool),
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProject(tt.project, tt.index, tt.projectNames)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestRunValidate(t *testing.T) {
	tests := []struct {
		name        string
		configData  string
		expectError bool
	}{
		{
			name: "valid config",
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
			name: "invalid yaml",
			configData: `version: 1.0.0
projects:
  - name: sample-project
    repo: git@github.com:user/sample-project.git
    branch: main
    backup_paths:
      - .env
      - db.sqlite3
    backup_retention: 3
  invalid yaml here
`,
			expectError: true,
		},
		{
			name: "missing required fields",
			configData: `version: 1.0.0
projects:
  - name: sample-project
    backup_paths:
      - .env
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

			// Run validate
			err := runValidate()
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestRunValidateNoConfig(t *testing.T) {
	// Create temporary directory without config file
	tempDir := t.TempDir()
	nonExistentPath := filepath.Join(tempDir, "nonexistent.yml")

	// Set cfgFile to non-existent path
	originalCfgFile := cfgFile
	cfgFile = nonExistentPath
	defer func() {
		cfgFile = originalCfgFile
	}()

	// Run validate - should fail because config doesn't exist
	err := runValidate()
	if err == nil {
		t.Errorf("Expected error for non-existent config file but got nil")
	}
}
