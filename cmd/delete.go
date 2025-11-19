package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yk-lab/toske/i18n"
	"gopkg.in/yaml.v3"
)

var (
	deleteProjectName string
	deleteForce       bool
)

// ja: deleteCmd は delete コマンドを表します
// en: deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: i18n.T("delete.short"),
	Long:  i18n.T("delete.long"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDelete(); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("common.error")+"\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVarP(&deleteProjectName, "project", "p", "", i18n.T("delete.flag.project"))
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, i18n.T("delete.flag.force"))
	deleteCmd.MarkFlagRequired("project")
}

func runDelete() error {
	// ja: プロジェクト名の前後の空白を削除
	// en: Trim leading and trailing whitespace from project name
	deleteProjectName = strings.TrimSpace(deleteProjectName)

	// ja: プロジェクト名が空でないかチェック (Cobra の MarkFlagRequired のバックアップ)
	// en: Check project name is not empty (backup for Cobra's MarkFlagRequired)
	if deleteProjectName == "" {
		return fmt.Errorf("%s", i18n.T("delete.noProjectFlag"))
	}

	// ja: 設定ファイルパスを決定
	// en: Determine config file path
	configPath := cfgFile
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}

	// ja: 設定ファイルが存在するかチェック
	// en: Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf(i18n.T("delete.noConfig"), configPath)
	}

	// ja: 設定ファイルを読み込む
	// en: Load configuration file
	config, err := loadDeleteConfig(configPath)
	if err != nil {
		return err
	}

	// ja: 指定されたプロジェクトを検索
	// en: Find the specified project
	projectIndex := findProjectIndex(config.Projects, deleteProjectName)
	if projectIndex == -1 {
		return fmt.Errorf(i18n.T("delete.projectNotFound"), deleteProjectName)
	}

	project := config.Projects[projectIndex]

	// ja: バックアップ存在確認
	// en: Check if backups exist
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	backupDir := filepath.Join(homeDir, ".config", "toske", "backups", project.Name)
	if !hasBackups(backupDir) {
		return fmt.Errorf(i18n.T("delete.noBackup"), project.Name, project.Name)
	}

	// ja: repository_path 確認
	// en: Check repository_path
	if project.RepositoryPath == "" {
		return fmt.Errorf(i18n.T("delete.noRepositoryPath"), project.Name)
	}

	// ja: リポジトリディレクトリ存在確認
	// en: Check if repository directory exists
	repoExists := false
	if stat, err := os.Stat(project.RepositoryPath); err == nil && stat.IsDir() {
		repoExists = true
	}

	// ja: 削除確認
	// en: Confirm deletion
	confirmed, err := confirmDeletion(project, repoExists, deleteForce)
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println(i18n.T("delete.cancelled"))
		return nil
	}

	// ja: リポジトリディレクトリ物理削除
	// en: Physically delete repository directory
	if repoExists {
		fmt.Printf(i18n.T("delete.removingRepository")+"\n", project.RepositoryPath)
		if err := os.RemoveAll(project.RepositoryPath); err != nil {
			return fmt.Errorf(i18n.T("delete.removeRepoError"), err)
		}
		fmt.Println(i18n.T("delete.repositoryRemoved"))
	}

	// ja: 設定は保持する（restore コマンドで復元できるようにするため）
	// en: Keep configuration (to allow restore command to work)
	fmt.Printf(i18n.T("delete.success")+"\n", deleteProjectName)

	return nil
}

// ja: loadDeleteConfig は設定ファイルを読み込みます
// en: loadDeleteConfig loads the configuration file
func loadDeleteConfig(configPath string) (Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf(i18n.T("delete.readError"), err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return Config{}, fmt.Errorf(i18n.T("delete.parseError"), err)
	}

	return config, nil
}

// ja: findProjectIndex はプロジェクト名からインデックスを検索します
// en: findProjectIndex finds the index of a project by name
func findProjectIndex(projects []Project, name string) int {
	for i := range projects {
		if projects[i].Name == name {
			return i
		}
	}
	return -1
}

// ja: confirmDeletion は削除の確認を行います
// en: confirmDeletion confirms the deletion
func confirmDeletion(project Project, repoExists bool, force bool) (bool, error) {
	// ja: --force フラグが指定されている場合は確認をスキップ
	// en: Skip confirmation if --force flag is specified
	if force {
		return true, nil
	}

	// ja: 削除確認プロンプトを表示
	// en: Show confirmation prompt
	fmt.Println(i18n.T("delete.confirmHeader"))
	fmt.Printf("  "+i18n.T("delete.confirmProject")+"\n", project.Name)

	if repoExists {
		fmt.Printf("  "+i18n.T("delete.confirmRepoPath")+"\n", project.RepositoryPath)
	} else {
		fmt.Printf("  "+i18n.T("delete.confirmNoRepo")+"\n", project.RepositoryPath)
	}

	fmt.Println()
	fmt.Print(i18n.T("delete.confirmPrompt"))

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf(i18n.T("delete.readInputError"), err)
	}

	// ja: 入力をトリムして小文字に変換
	// en: Trim and convert input to lowercase
	response = strings.ToLower(strings.TrimSpace(response))

	// ja: yまたはyes以外の場合はキャンセル
	// en: Cancel if response is not y or yes
	return response == "y" || response == "yes", nil
}

// ja: saveConfig は設定ファイルを保存します
// en: saveConfig saves the configuration file
func saveConfig(configPath string, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf(i18n.T("delete.marshalError"), err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf(i18n.T("delete.writeError"), err)
	}

	return nil
}

// ja: hasBackups はプロジェクトにバックアップが存在するかチェックします
// en: hasBackups checks if backups exist for the project
func hasBackups(backupDir string) bool {
	metadataPath := filepath.Join(backupDir, "backups.yaml")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return false
	}

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return false
	}

	var metadata BackupMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return false
	}

	return len(metadata.Backups) > 0
}
