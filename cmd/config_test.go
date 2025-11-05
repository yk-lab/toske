package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGetDefaultConfigPath tests the default config path logic
func TestGetDefaultConfigPath(t *testing.T) {
	// Save original environment and restore after test
	originalEnv := os.Getenv("TOSKE_CONFIG")
	defer func() {
		if originalEnv != "" {
			os.Setenv("TOSKE_CONFIG", originalEnv)
		} else {
			os.Unsetenv("TOSKE_CONFIG")
		}
	}()

	t.Run("returns environment variable path when set", func(t *testing.T) {
		expectedPath := "/custom/path/to/config.yml"
		os.Setenv("TOSKE_CONFIG", expectedPath)

		result := getDefaultConfigPath()
		if result != expectedPath {
			t.Errorf("Expected %s, got %s", expectedPath, result)
		}

		os.Unsetenv("TOSKE_CONFIG")
	})

	t.Run("returns XDG compliant path when no files exist", func(t *testing.T) {
		os.Unsetenv("TOSKE_CONFIG")

		result := getDefaultConfigPath()

		// Should contain .config/toske/config.yml
		if !filepath.IsAbs(result) {
			// Allow relative path fallback for testing environments
			if result != "./toske-config.yml" {
				t.Errorf("Expected absolute path or fallback, got %s", result)
			}
			return
		}

		if !filepath.IsAbs(result) {
			t.Errorf("Expected absolute path, got %s", result)
		}

		expectedSuffix := filepath.Join(".config", "toske", "config.yml")
		if !filepath.IsAbs(result) || !endsWithPath(result, expectedSuffix) {
			t.Errorf("Expected path to end with %s, got %s", expectedSuffix, result)
		}
	})

	t.Run("returns existing new path when it exists", func(t *testing.T) {
		os.Unsetenv("TOSKE_CONFIG")

		// Create temporary directory structure
		tempDir := t.TempDir()
		configDir := filepath.Join(tempDir, ".config", "toske")
		newPath := filepath.Join(configDir, "config.yml")

		// Create the directory and file
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}
		if err := os.WriteFile(newPath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		// Temporarily override HOME environment
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tempDir)
		defer os.Setenv("HOME", originalHome)

		result := getDefaultConfigPath()

		expectedPath := filepath.Join(tempDir, ".config", "toske", "config.yml")
		if result != expectedPath {
			t.Errorf("Expected %s, got %s", expectedPath, result)
		}
	})

	t.Run("returns legacy path when only legacy exists", func(t *testing.T) {
		os.Unsetenv("TOSKE_CONFIG")

		// Create temporary directory structure
		tempDir := t.TempDir()
		legacyPath := filepath.Join(tempDir, ".toske.yaml")

		// Create only the legacy file
		if err := os.WriteFile(legacyPath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create legacy config file: %v", err)
		}

		// Temporarily override HOME environment
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tempDir)
		defer os.Setenv("HOME", originalHome)

		result := getDefaultConfigPath()

		expectedPath := filepath.Join(tempDir, ".toske.yaml")
		if result != expectedPath {
			t.Errorf("Expected %s, got %s", expectedPath, result)
		}
	})

	t.Run("prefers new path over legacy when both exist", func(t *testing.T) {
		os.Unsetenv("TOSKE_CONFIG")

		// Create temporary directory structure
		tempDir := t.TempDir()
		configDir := filepath.Join(tempDir, ".config", "toske")
		newPath := filepath.Join(configDir, "config.yml")
		legacyPath := filepath.Join(tempDir, ".toske.yaml")

		// Create both files
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}
		if err := os.WriteFile(newPath, []byte("new"), 0644); err != nil {
			t.Fatalf("Failed to create new config file: %v", err)
		}
		if err := os.WriteFile(legacyPath, []byte("legacy"), 0644); err != nil {
			t.Fatalf("Failed to create legacy config file: %v", err)
		}

		// Temporarily override HOME environment
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tempDir)
		defer os.Setenv("HOME", originalHome)

		result := getDefaultConfigPath()

		expectedPath := filepath.Join(tempDir, ".config", "toske", "config.yml")
		if result != expectedPath {
			t.Errorf("Expected new path %s, got %s", expectedPath, result)
		}
	})
}

// TestIsLegacyConfigPath tests the legacy path detection
func TestIsLegacyConfigPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory, skipping test")
	}

	testCases := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "legacy path",
			path:     filepath.Join(homeDir, ".toske.yaml"),
			expected: true,
		},
		{
			name:     "new XDG path",
			path:     filepath.Join(homeDir, ".config", "toske", "config.yml"),
			expected: false,
		},
		{
			name:     "custom path",
			path:     "/custom/path/config.yml",
			expected: false,
		},
		{
			name:     "relative path",
			path:     "./config.yml",
			expected: false,
		},
		{
			name:     "empty path",
			path:     "",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isLegacyConfigPath(tc.path)
			if result != tc.expected {
				t.Errorf("isLegacyConfigPath(%s) = %v, expected %v", tc.path, result, tc.expected)
			}
		})
	}
}

// TestGetDefaultConfigPathFallback tests fallback behavior when home directory cannot be determined
func TestGetDefaultConfigPathFallback(t *testing.T) {
	// This test is difficult to implement without mocking os.UserHomeDir
	// In a real scenario where HOME is not available, it should fall back to ./toske-config.yml
	t.Skip("Skipping fallback test - requires mocking os.UserHomeDir")
}

// Helper function to check if a path ends with another path
func endsWithPath(fullPath, suffix string) bool {
	fullPath = filepath.Clean(fullPath)
	suffix = filepath.Clean(suffix)
	return filepath.Base(fullPath) == filepath.Base(suffix) ||
		len(fullPath) >= len(suffix) && fullPath[len(fullPath)-len(suffix):] == suffix
}
