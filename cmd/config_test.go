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

// TestEndsWithPath tests the endsWithPath helper function
func TestEndsWithPath(t *testing.T) {
	testCases := []struct {
		name     string
		fullPath string
		suffix   string
		expected bool
	}{
		{
			name:     "exact match",
			fullPath: filepath.Join("home", "user", ".config", "toske", "config.yml"),
			suffix:   filepath.Join("home", "user", ".config", "toske", "config.yml"),
			expected: true,
		},
		{
			name:     "valid suffix",
			fullPath: filepath.Join("home", "user", ".config", "toske", "config.yml"),
			suffix:   filepath.Join(".config", "toske", "config.yml"),
			expected: true,
		},
		{
			name:     "single component suffix",
			fullPath: filepath.Join("home", "user", "config.yml"),
			suffix:   "config.yml",
			expected: true,
		},
		{
			name:     "different paths with same basename - should not match",
			fullPath: filepath.Join("foo", "bar", "config.yml"),
			suffix:   filepath.Join("baz", "qux", "config.yml"),
			expected: false,
		},
		{
			name:     "suffix longer than full path",
			fullPath: filepath.Join("config", "file.yml"),
			suffix:   filepath.Join("very", "long", "config", "file.yml"),
			expected: false,
		},
		{
			name:     "partial component match - should not match",
			fullPath: filepath.Join("home", "user", ".config"),
			suffix:   filepath.Join("r", ".config"),
			expected: false,
		},
		{
			name:     "root path",
			fullPath: string(filepath.Separator) + filepath.Join("home", "user", "config.yml"),
			suffix:   filepath.Join("user", "config.yml"),
			expected: true,
		},
		{
			name:     "empty suffix",
			fullPath: filepath.Join("home", "user", "config.yml"),
			suffix:   "",
			expected: false,
		},
		{
			name:     "current directory",
			fullPath: ".",
			suffix:   ".",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := endsWithPath(tc.fullPath, tc.suffix)
			if result != tc.expected {
				t.Errorf("endsWithPath(%q, %q) = %v, expected %v",
					tc.fullPath, tc.suffix, result, tc.expected)
			}
		})
	}
}

// Helper function to check if a path ends with another path
// by comparing path components rather than raw strings
func endsWithPath(fullPath, suffix string) bool {
	// Clean both paths to normalize them
	fullPath = filepath.Clean(fullPath)
	suffix = filepath.Clean(suffix)

	// Split paths into components
	fullComponents := splitPath(fullPath)
	suffixComponents := splitPath(suffix)

	// If suffix has more components than fullPath, it can't match
	if len(suffixComponents) > len(fullComponents) {
		return false
	}

	// Compare the trailing N components of fullPath with all components of suffix
	startIndex := len(fullComponents) - len(suffixComponents)
	for i := 0; i < len(suffixComponents); i++ {
		if fullComponents[startIndex+i] != suffixComponents[i] {
			return false
		}
	}

	return true
}

// splitPath splits a filepath into its components, handling OS-specific separators
func splitPath(path string) []string {
	if path == "" || path == "." {
		return []string{"."}
	}

	// Use filepath.Split repeatedly to get all components
	var components []string
	for {
		dir, file := filepath.Split(path)
		if file != "" {
			components = append([]string{file}, components...)
		}

		if dir == "" || dir == string(filepath.Separator) || dir == "." {
			if dir == string(filepath.Separator) {
				components = append([]string{string(filepath.Separator)}, components...)
			}
			break
		}

		path = filepath.Clean(dir)
		if path == "." {
			break
		}
	}

	if len(components) == 0 {
		return []string{"."}
	}

	return components
}
