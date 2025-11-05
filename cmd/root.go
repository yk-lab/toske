package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/yk-lab/toske/i18n"
	"github.com/yk-lab/toske/utils"
)

var cfgFile string

// ja: rootCmd は、サブコマンドが指定されなかった場合の基本コマンドを表します
// en: rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "toske",
	Short: i18n.T("root.short"),
	Long:  getHeroMessage(),
	// ja: アクションが関連付けられている場合は、以下の行のコメントを解除してください
	// en: Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

func getHeroMessage() string {
	txt, err := utils.AAFromText("hero.txt")
	if err != nil {
		log.Printf("Error reading hero message file: %v", err)
		return "TOSKE"
	}
	return txt
}

// ja: Execute は、すべての子コマンドを root コマンドに追加し、適切にフラグを設定します。
// これは main.main() によって呼び出されます。rootCmd に対して一度だけ実行する必要があります。
// en: Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// ja: ここでフラグと設定を定義します。
	// Cobra は永続フラグをサポートしており、ここで定義された場合、アプリケーション全体でグローバルになります。
	// en: Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/toske/config.yml)")

	// ja: Cobra はローカルフラグもサポートしており、これはこのアクションが直接呼び出された場合にのみ実行されます。
	// en: Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// ja: initConfig は、設定ファイルと設定されている場合は環境変数を読み込みます。
// en: initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// ja: フラグから設定ファイルを使用する
		// en: Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// ja: デフォルトの設定ファイルパスを使用する（TOSKE_CONFIG 環境変数やフォールバックを処理）
		// en: Use default config path (handles TOSKE_CONFIG env var and fallback)
		viper.SetConfigFile(getDefaultConfigPath())
	}

	// ja: 環境変数を自動的に読み込む
	// en: read in environment variables that match
	viper.AutomaticEnv()

	// ja: 設定ファイルが見つかった場合は読み込む
	// en: If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		configFile := viper.ConfigFileUsed()
		fmt.Fprintln(os.Stderr, "Using config file:", configFile)

		// ja: レガシーパスを使用している場合は移行を促す警告を表示
		// en: Show migration warning if using legacy path
		if isLegacyConfigPath(configFile) {
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, i18n.T("config.legacyWarning"))
			fmt.Fprintln(os.Stderr, i18n.T("config.legacyWarningDetail"))
			fmt.Fprintln(os.Stderr, "")
		}
	}
}
