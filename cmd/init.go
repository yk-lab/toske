package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yk-lab/toske/i18n"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: i18n.T("init.short"),
	Long:  i18n.T("init.long"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runInit(); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("common.error")+"\n", err)
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
		fmt.Printf(i18n.T("init.fileExists")+"\n", configPath)
		fmt.Print(i18n.T("init.overwritePrompt"))

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf(i18n.T("init.readInputError"), err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println(i18n.T("init.cancelled"))
			return nil
		}
	}

	// Create directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf(i18n.T("init.createDirError"), err)
	}

	// Create config file with template
	content := getConfigTemplate()
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf(i18n.T("init.writeFileError"), err)
	}

	fmt.Printf(i18n.T("init.success")+"\n", configPath)
	fmt.Println(i18n.T("init.nextSteps"))
	fmt.Println(i18n.T("init.nextSteps.edit"))
	fmt.Println(i18n.T("init.nextSteps.editCmd"))
	fmt.Println(i18n.T("init.nextSteps.validate"))
	fmt.Println(i18n.T("init.nextSteps.validateCmd"))
	fmt.Println(i18n.T("init.nextSteps.backup"))
	fmt.Println(i18n.T("init.nextSteps.backupCmd"))

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
