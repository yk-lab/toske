package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yk-lab/toske/i18n"
)

// ja: listCmd は list コマンドを表します
// en: listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: i18n.T("list.short"),
	Long:  i18n.T("list.long"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runList(); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("common.error")+"\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList() error {
	// ja: 設定ファイルパスを決定
	// en: Determine config file path
	configPath := cfgFile
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}

	// ja: 設定ファイルが存在するかチェック
	// en: Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf(i18n.T("list.noConfig"), configPath)
	}

	// ja: 設定ファイルを読み込む
	// en: Load configuration file
	v := viper.New()
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf(i18n.T("list.readError"), err)
	}

	// ja: 設定を構造体にアンマーシャル
	// en: Unmarshal config into struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return fmt.Errorf(i18n.T("list.parseError"), err)
	}

	// ja: プロジェクトが存在しない場合
	// en: If no projects exist
	if len(config.Projects) == 0 {
		fmt.Println(i18n.T("list.noProjects"))
		return nil
	}

	// ja: プロジェクト一覧を表示
	// en: Display project list
	fmt.Println(i18n.T("list.header"))
	fmt.Println()
	for _, project := range config.Projects {
		fmt.Printf("  • %s\n", project.Name)
		fmt.Printf("    %s: %s\n", i18n.T("list.repo"), project.Repo)
		fmt.Printf("    %s: %s\n", i18n.T("list.branch"), project.Branch)
		if len(project.BackupPaths) > 0 {
			fmt.Printf("    %s:\n", i18n.T("list.backupPaths"))
			for _, path := range project.BackupPaths {
				fmt.Printf("      - %s\n", path)
			}
		}
		if project.BackupRetention > 0 {
			fmt.Printf("    %s: %d\n", i18n.T("list.retention"), project.BackupRetention)
		}
		fmt.Println()
	}

	fmt.Printf(i18n.T("list.total")+"\n", len(config.Projects))

	return nil
}
