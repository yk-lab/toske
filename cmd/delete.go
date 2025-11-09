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

var deleteProjectName string

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
}

func runDelete() error {
	// ja: プロジェクト名が指定されているかチェック
	// en: Check if project name is specified
	if deleteProjectName == "" {
		return fmt.Errorf("%s", i18n.T("delete.noProjectFlag"))
	}

	// ja: プロジェクト名の前後の空白を削除
	// en: Trim leading and trailing whitespace from project name
	deleteProjectName = strings.TrimSpace(deleteProjectName)

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
	v := viper.New()
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf(i18n.T("delete.readError"), err)
	}

	// ja: 設定を構造体にアンマーシャル
	// en: Unmarshal config into struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return fmt.Errorf(i18n.T("delete.parseError"), err)
	}

	// ja: 指定されたプロジェクトを検索
	// en: Find the specified project
	projectIndex := -1
	for i := range config.Projects {
		if config.Projects[i].Name == deleteProjectName {
			projectIndex = i
			break
		}
	}

	if projectIndex == -1 {
		return fmt.Errorf(i18n.T("delete.projectNotFound"), deleteProjectName)
	}

	// ja: 削除確認プロンプトを表示
	// en: Show confirmation prompt
	fmt.Printf(i18n.T("delete.confirmPrompt"), deleteProjectName)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf(i18n.T("delete.readInputError"), err)
	}

	// ja: 入力をトリムして小文字に変換
	// en: Trim and convert input to lowercase
	response = strings.ToLower(strings.TrimSpace(response))

	// ja: yまたはyes以外の場合はキャンセル
	// en: Cancel if response is not y or yes
	if response != "y" && response != "yes" {
		fmt.Println(i18n.T("delete.cancelled"))
		return nil
	}

	// ja: プロジェクトを削除
	// en: Delete the project
	config.Projects = append(config.Projects[:projectIndex], config.Projects[projectIndex+1:]...)

	// ja: 設定ファイルを更新
	// en: Update configuration file
	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf(i18n.T("delete.writeError"), err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf(i18n.T("delete.writeError"), err)
	}

	fmt.Printf(i18n.T("delete.success")+"\n", deleteProjectName)

	return nil
}
