package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yk-lab/toske/i18n"
)

// ja: Config は設定ファイルの構造を表します
// en: Config represents the structure of the configuration file
type Config struct {
	Version  string    `mapstructure:"version"`
	Projects []Project `mapstructure:"projects"`
}

// ja: Project はプロジェクト設定を表します
// en: Project represents a project configuration
type Project struct {
	Name            string   `mapstructure:"name"`
	Repo            string   `mapstructure:"repo"`
	Branch          string   `mapstructure:"branch"`
	BackupPaths     []string `mapstructure:"backup_paths"`
	BackupRetention int      `mapstructure:"backup_retention"`
}

// ja: validateCmd は validate コマンドを表します
// en: validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: i18n.T("validate.short"),
	Long:  i18n.T("validate.long"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runValidate(); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("common.error")+"\n", err)
			os.Exit(1)
		}
	},
}

// init は validateCmd を rootCmd に登録します。
func init() {
	rootCmd.AddCommand(validateCmd)
}

// runValidate は設定ファイルを決定して読み込み、構文および内容を検証します。
// 設定ファイルパスは明示的な cfgFile を優先し、未指定時はデフォルトパスを使用します。
// ファイルが存在しない、読み込みに失敗する、パースに失敗する、または設定検証で問題がある場合は error を返します。
// 成功時は検証成功メッセージとプロジェクト数を標準出力に表示します。
func runValidate() error {
	// ja: 設定ファイルパスを決定
	// en: Determine config file path
	configPath := cfgFile
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}

	// ja: 設定ファイルが存在するかチェック
	// en: Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		msg := fmt.Sprintf(i18n.T("validate.noConfig"), configPath)
		return fmt.Errorf("%s", msg)
	}

	fmt.Printf(i18n.T("validate.checking")+"\n", configPath)

	// ja: 設定ファイルを読み込む
	// en: Load configuration file
	v := viper.New()
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		msg := fmt.Sprintf(i18n.T("validate.readError"), err)
		return fmt.Errorf("%s", msg)
	}

	// ja: 設定を構造体にアンマーシャル
	// en: Unmarshal config into struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		msg := fmt.Sprintf(i18n.T("validate.parseError"), err)
		return fmt.Errorf("%s", msg)
	}

	// ja: 設定を検証
	// en: Validate configuration
	if err := validateConfig(&config); err != nil {
		return err
	}

	fmt.Println(i18n.T("validate.success"))
	fmt.Printf(i18n.T("validate.projectCount")+"\n", len(config.Projects))

	return nil
}

// ja: validateConfig は設定ファイルの内容を検証します
// validateConfig は設定全体の内容を検証します。
// バージョンが空でないこと、プロジェクトが少なくとも1件存在することを確認し、各プロジェクトについて名前の重複、必須フィールド（名前・リポジトリ・ブランチ）およびバックアップ保持期間が0以上であることを検証します。
// 検証に失敗した場合は該当するエラーを返します。
func validateConfig(config *Config) error {
	// ja: バージョンの検証
	// en: Validate version
	if config.Version == "" {
		return fmt.Errorf("%s", i18n.T("validate.error.noVersion"))
	}

	// ja: プロジェクトの検証
	// en: Validate projects
	if len(config.Projects) == 0 {
		return fmt.Errorf("%s", i18n.T("validate.error.noProjects"))
	}

	// ja: 各プロジェクトを検証
	// en: Validate each project
	projectNames := make(map[string]bool)
	for i, project := range config.Projects {
		if err := validateProject(&project, i, projectNames); err != nil {
			return err
		}
		projectNames[project.Name] = true
	}

	return nil
}

// ja: validateProject は個々のプロジェクト設定を検証します
// validateProject は単一の Project 設定を検証します。
// 空の名前、重複する名前、空のリポジトリまたはブランチ、または 0 未満の backup_retention がある場合に
// 説明付きの error を返します。正常なら nil を返します。
func validateProject(project *Project, index int, projectNames map[string]bool) error {
	projectNum := index + 1

	// ja: プロジェクト名の検証
	// en: Validate project name
	if project.Name == "" {
		msg := fmt.Sprintf(i18n.T("validate.error.projectNoName"), projectNum)
		return fmt.Errorf("%s", msg)
	}

	// ja: 重複した名前のチェック
	// en: Check for duplicate names
	if projectNames[project.Name] {
		msg := fmt.Sprintf(i18n.T("validate.error.duplicateName"), project.Name)
		return fmt.Errorf("%s", msg)
	}

	// ja: リポジトリURLの検証
	// en: Validate repository URL
	if project.Repo == "" {
		msg := fmt.Sprintf(i18n.T("validate.error.projectNoRepo"), project.Name)
		return fmt.Errorf("%s", msg)
	}

	// ja: ブランチ名の検証
	// en: Validate branch name
	if project.Branch == "" {
		msg := fmt.Sprintf(i18n.T("validate.error.projectNoBranch"), project.Name)
		return fmt.Errorf("%s", msg)
	}

	// ja: backup_retention の検証（0以上である必要がある）
	// en: Validate backup_retention (must be >= 0)
	if project.BackupRetention < 0 {
		msg := fmt.Sprintf(i18n.T("validate.error.invalidRetention"), project.Name, project.BackupRetention)
		return fmt.Errorf("%s", msg)
	}

	return nil
}