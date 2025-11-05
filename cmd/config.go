package cmd

import (
	"os"
	"path/filepath"
)

// ja: getDefaultConfigPath はデフォルトの設定ファイルパスを返します
// ja: 優先順位: TOSKE_CONFIG 環境変数 > ~/.config/toske/config.yml > ./toske-config.yml (フォールバック)
// en: getDefaultConfigPath returns the default configuration file path
// en: Priority: TOSKE_CONFIG env var > ~/.config/toske/config.yml > ./toske-config.yml (fallback)
func getDefaultConfigPath() string {
	// ja: 最初に環境変数をチェック
	// en: Check environment variable first
	if configPath := os.Getenv("TOSKE_CONFIG"); configPath != "" {
		return configPath
	}

	// ja: ホームディレクトリ内の XDG 準拠パスを使用
	// en: Try to use XDG-compliant path in home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// ja: ホームディレクトリが取得できない場合はカレントディレクトリにフォールバック
		// en: Fallback to current directory if home dir cannot be determined
		return "./toske-config.yml"
	}

	return filepath.Join(homeDir, ".config", "toske", "config.yml")
}
