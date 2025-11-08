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

// init は validate サブコマンドを rootCmd に登録します。
func init() {
	rootCmd.AddCommand(validateCmd)
}

// runValidate は設定ファイルを読み込み、その構文と内容を検証します。
// 指定された cfgFile が空の場合はデフォルトの設定パスを使用し、ファイルの存在確認、読み取り、構造体へのアンマーシャル、及び validateConfig による検証を行います。
// 検証に成功すると成功メッセージとプロジェクト数を標準出力に表示し、失敗した場合は原因を示すエラーを返します。
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
		return fmt.Errorf(i18n.T("validate.noConfig"), configPath)
	}

	fmt.Printf(i18n.T("validate.checking")+"\n", configPath)

	// ja: 設定ファイルを読み込む
	// en: Load configuration file
	v := viper.New()
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf(i18n.T("validate.readError"), err)
	}

	// ja: 設定を構造体にアンマーシャル
	// en: Unmarshal config into struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return fmt.Errorf(i18n.T("validate.parseError"), err)
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
// validateConfig は Config の内容を検証し、不備があれば最初に検出した理由を表すエラーを返す。
// 検証内容は、バージョンが空でないこと、少なくとも1つのプロジェクトが存在すること、各プロジェクトについて名前の重複がないことおよびプロジェクト固有の検証（名前・リポジトリ・ブランチの必須チェック、バックアップ保持期間が負でないこと）を含む。
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
// validateProject は個別の Project を検証し、名前の有無・重複、リポジトリ、ブランチ、およびバックアップ保持期間が有効であることを確認します。
// 無効な項目が見つかった場合はローカライズされたエラーメッセージを含む error を返します。
func validateProject(project *Project, index int, projectNames map[string]bool) error {
	projectNum := index + 1

	// ja: プロジェクト名の検証
	// en: Validate project name
	if project.Name == "" {
		return fmt.Errorf(i18n.T("validate.error.projectNoName"), projectNum)
	}

	// ja: 重複した名前のチェック
	// en: Check for duplicate names
	if projectNames[project.Name] {
		return fmt.Errorf(i18n.T("validate.error.duplicateName"), project.Name)
	}

	// ja: リポジトリURLの検証
	// en: Validate repository URL
	if project.Repo == "" {
		return fmt.Errorf(i18n.T("validate.error.projectNoRepo"), project.Name)
	}

	// ja: ブランチ名の検証
	// en: Validate branch name
	if project.Branch == "" {
		return fmt.Errorf(i18n.T("validate.error.projectNoBranch"), project.Name)
	}

	// ja: backup_retention の検証（0以上である必要がある）
	// en: Validate backup_retention (must be >= 0)
	if project.BackupRetention < 0 {
		return fmt.Errorf(i18n.T("validate.error.invalidRetention"), project.Name, project.BackupRetention)
	}

	return nil
}