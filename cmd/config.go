package cmd

import (
	"os"
	"path/filepath"
)

// ja: getDefaultConfigPath はデフォルトの設定ファイルパスを返します
// ja: 優先順位: TOSKE_CONFIG 環境変数 > 新パス（存在時） > レガシーパス（存在時） > 新パス（デフォルト）
// en: getDefaultConfigPath returns the default configuration file path
// en: Priority: TOSKE_CONFIG env var > new path (if exists) > legacy path (if exists) > new path (default)
func getDefaultConfigPath() string {
	// ja: 最初に環境変数をチェック
	// en: Check environment variable first
	if configPath := os.Getenv("TOSKE_CONFIG"); configPath != "" {
		return configPath
	}

	// ja: ホームディレクトリを取得
	// en: Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// ja: ホームディレクトリが取得できない場合はカレントディレクトリにフォールバック
		// en: Fallback to current directory if home dir cannot be determined
		return "./toske-config.yml"
	}

	// ja: 新しい XDG 準拠パス
	// en: New XDG-compliant path
	newPath := filepath.Join(homeDir, ".config", "toske", "config.yml")

	// ja: レガシーパス（後方互換性のため）
	// en: Legacy path (for backward compatibility)
	legacyPath := filepath.Join(homeDir, ".toske.yaml")

	// ja: 新しいパスが存在する場合はそれを使用
	// en: Use new path if it exists
	if _, err := os.Stat(newPath); err == nil {
		return newPath
	}

	// ja: レガシーパスが存在する場合はそれを使用（既存ユーザーのため）
	// en: Use legacy path if it exists (for existing users)
	if _, err := os.Stat(legacyPath); err == nil {
		return legacyPath
	}

	// ja: どちらも存在しない場合は新しいパスをデフォルトとして返す
	// en: Return new path as default if neither exists
	return newPath
}

// ja: isLegacyConfigPath は指定されたパスがレガシーパスかどうかを判定します
// en: isLegacyConfigPath checks if the given path is a legacy config path
func isLegacyConfigPath(configPath string) bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	legacyPath := filepath.Join(homeDir, ".toske.yaml")
	return configPath == legacyPath
}
