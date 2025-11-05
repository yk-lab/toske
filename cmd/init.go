package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long: `Initialize creates a new configuration file at the default location.
The default path is ~/.config/toske/config.yml

You can override the default path by setting the TOSKE_CONFIG environment variable.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runInit(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit() error {
	configPath := getConfigPath()

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		// File exists, ask for confirmation
		fmt.Printf("Configuration file already exists at: %s\n", configPath)
		fmt.Print("Do you want to overwrite it? [y/N]: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Initialization cancelled.")
			return nil
		}
	}

	// Create directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create config file with template
	content := getConfigTemplate()
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✓ Configuration file created successfully at: %s\n", configPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit the configuration file to add your projects")
	fmt.Printf("     toske edit\n")
	fmt.Println("  2. Validate your configuration")
	fmt.Printf("     toske validate\n")
	fmt.Println("  3. Backup your project files")
	fmt.Printf("     toske backup --project <project-name>\n")

	return nil
}

// getConfigPath returns the configuration file path
// Priority: TOSKE_CONFIG env var > default path
func getConfigPath() string {
	if configPath := os.Getenv("TOSKE_CONFIG"); configPath != "" {
		return configPath
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home dir cannot be determined
		return "./toske-config.yml"
	}

	return filepath.Join(homeDir, ".config", "toske", "config.yml")
}

// getConfigTemplate returns the default configuration template
func getConfigTemplate() string {
	return `version: 1.0.0
projects:
  - name: sample-project
    repo: git@github.com:user/sample-project.git
    branch: main
    backup_paths:
      - .env
      - db.sqlite3
      - config/
    backup_retention: 3

# Add more projects as needed:
#  - name: another-project
#    repo: https://github.com/user/another-project.git
#    branch: develop
#    backup_paths:
#      - .env.local
#      - data/
#    backup_retention: 5
`
}
