package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yk-lab/toske/i18n"
	"gopkg.in/yaml.v3"
)

var (
	removeProjectName string
	removeForce       bool
)

// ja: removeCmd は remove コマンドを表します
// en: removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: i18n.T("remove.short"),
	Long:  i18n.T("remove.long"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runRemove(); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("common.error")+"\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().StringVarP(&removeProjectName, "project", "p", "", i18n.T("remove.flag.project"))
	removeCmd.Flags().BoolVarP(&removeForce, "force", "f", false, i18n.T("remove.flag.force"))
	removeCmd.MarkFlagRequired("project")
}

func runRemove() error {
	// ja: プロジェクト名の前後の空白を削除
	// en: Trim leading and trailing whitespace from project name
	removeProjectName = strings.TrimSpace(removeProjectName)

	// ja: プロジェクト名が空でないかチェック (Cobra の MarkFlagRequired のバックアップ)
	// en: Check project name is not empty (backup for Cobra's MarkFlagRequired)
	if removeProjectName == "" {
		return fmt.Errorf("%s", i18n.T("remove.noProjectFlag"))
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
		return fmt.Errorf(i18n.T("remove.noConfig"), configPath)
	}

	// ja: 設定ファイルを読み込む
	// en: Load configuration file
	config, err := loadRemoveConfig(configPath)
	if err != nil {
		return err
	}

	// ja: 指定されたプロジェクトを検索
	// en: Find the specified project
	projectIndex := findRemoveProjectIndex(config.Projects, removeProjectName)
	if projectIndex == -1 {
		return fmt.Errorf(i18n.T("remove.projectNotFound"), removeProjectName)
	}

	// ja: 削除確認
	// en: Confirm removal
	confirmed, err := confirmRemoval(removeProjectName, removeForce)
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println(i18n.T("remove.cancelled"))
		return nil
	}

	// ja: プロジェクトを削除
	// en: Remove the project
	config.Projects = append(config.Projects[:projectIndex], config.Projects[projectIndex+1:]...)

	// ja: 設定ファイルを保存
	// en: Save configuration file
	if err := saveRemoveConfig(configPath, &config); err != nil {
		return err
	}

	fmt.Printf(i18n.T("remove.success")+"\n", removeProjectName)

	return nil
}

// ja: loadRemoveConfig は設定ファイルを読み込みます
// en: loadRemoveConfig loads the configuration file
func loadRemoveConfig(configPath string) (Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf(i18n.T("remove.readError"), err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return Config{}, fmt.Errorf(i18n.T("remove.parseError"), err)
	}

	return config, nil
}

// ja: findRemoveProjectIndex はプロジェクト名からインデックスを検索します
// en: findRemoveProjectIndex finds the index of a project by name
func findRemoveProjectIndex(projects []Project, name string) int {
	for i := range projects {
		if projects[i].Name == name {
			return i
		}
	}
	return -1
}

// ja: confirmRemoval は削除の確認を行います
// en: confirmRemoval confirms the removal
func confirmRemoval(projectName string, force bool) (bool, error) {
	// ja: --force フラグが指定されている場合は確認をスキップ
	// en: Skip confirmation if --force flag is specified
	if force {
		return true, nil
	}

	// ja: 削除確認プロンプトを表示
	// en: Show confirmation prompt
	fmt.Printf(i18n.T("remove.confirmPrompt"), projectName)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf(i18n.T("remove.readInputError"), err)
	}

	// ja: 入力をトリムして小文字に変換
	// en: Trim and convert input to lowercase
	response = strings.ToLower(strings.TrimSpace(response))

	// ja: yまたはyes以外の場合はキャンセル
	// en: Cancel if response is not y or yes
	return response == "y" || response == "yes", nil
}

// ja: saveRemoveConfig は設定ファイルを保存します
// en: saveRemoveConfig saves the configuration file
func saveRemoveConfig(configPath string, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf(i18n.T("remove.marshalError"), err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf(i18n.T("remove.writeError"), err)
	}

	return nil
}
