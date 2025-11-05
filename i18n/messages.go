package i18n

// messages contains all translated messages
// Structure: messages[language][key] = translated text
var messages = map[string]map[string]string{
	"en": {
		// Root command
		"root.short": "A brief description of your application",

		// Help command
		"help.short": "Display help information about commands",
		"help.long": `Display detailed help information about available commands and their usage.

Use 'toske help [command]' to get detailed information about a specific command.`,
		"help.availableCommands": "Available Commands:",
		"help.flags":             "Flags:",
		"help.usage":             "Usage:",
		"help.examples":          "Examples:",
		"help.additionalHelp":    "\nUse \"toske [command] --help\" for more information about a command.",

		// Init command
		"init.short": "Initialize configuration file",
		"init.long": `Initialize creates a new configuration file at the default location.
The default path is ~/.config/toske/config.yml

You can override the default path by setting the TOSKE_CONFIG environment variable.`,
		"init.fileExists":          "Configuration file already exists at: %s",
		"init.overwritePrompt":     "Do you want to overwrite it? [y/N]: ",
		"init.cancelled":           "Initialization cancelled.",
		"init.createDirError":      "failed to create config directory: %w",
		"init.writeFileError":      "failed to write config file: %w",
		"init.readInputError":      "failed to read input: %w",
		"init.success":             "✓ Configuration file created successfully at: %s",
		"init.nextSteps":           "\nNext steps:",
		"init.nextSteps.edit":      "  1. Edit the configuration file to add your projects",
		"init.nextSteps.editCmd":   "     toske edit",
		"init.nextSteps.validate":  "  2. Validate your configuration",
		"init.nextSteps.validateCmd": "     toske validate",
		"init.nextSteps.backup":    "  3. Backup your project files",
		"init.nextSteps.backupCmd": "     toske backup --project <project-name>",

		// Config
		"config.legacyWarning":       "⚠️  WARNING: You are using a legacy configuration file location.",
		"config.legacyWarningDetail": "   Please migrate to ~/.config/toske/config.yml by running: mv ~/.toske.yaml ~/.config/toske/config.yml",

		// Common
		"common.error": "Error: %v",
	},
	"ja": {
		// Root command
		"root.short": "アプリケーションの簡単な説明",

		// Help command
		"help.short": "コマンドのヘルプ情報を表示",
		"help.long": `利用可能なコマンドとその使用方法の詳細なヘルプ情報を表示します。

特定のコマンドの詳細情報を取得するには 'toske help [コマンド]' を使用してください。`,
		"help.availableCommands": "利用可能なコマンド:",
		"help.flags":             "フラグ:",
		"help.usage":             "使用法:",
		"help.examples":          "例:",
		"help.additionalHelp":    "\nコマンドの詳細情報は \"toske [コマンド] --help\" を使用してください。",

		// Init command
		"init.short": "設定ファイルを初期化",
		"init.long": `デフォルトの場所に新しい設定ファイルを作成します。
デフォルトパス: ~/.config/toske/config.yml

TOSKE_CONFIG 環境変数を設定することで、デフォルトパスを上書きできます。`,
		"init.fileExists":          "設定ファイルは既に存在します: %s",
		"init.overwritePrompt":     "上書きしますか？ [y/N]: ",
		"init.cancelled":           "初期化をキャンセルしました。",
		"init.createDirError":      "設定ディレクトリの作成に失敗しました: %w",
		"init.writeFileError":      "設定ファイルの書き込みに失敗しました: %w",
		"init.readInputError":      "入力の読み取りに失敗しました: %w",
		"init.success":             "✓ 設定ファイルを正常に作成しました: %s",
		"init.nextSteps":           "\n次のステップ:",
		"init.nextSteps.edit":      "  1. 設定ファイルを編集してプロジェクトを追加",
		"init.nextSteps.editCmd":   "     toske edit",
		"init.nextSteps.validate":  "  2. 設定ファイルを検証",
		"init.nextSteps.validateCmd": "     toske validate",
		"init.nextSteps.backup":    "  3. プロジェクトファイルをバックアップ",
		"init.nextSteps.backupCmd": "     toske backup --project <プロジェクト名>",

		// Config
		"config.legacyWarning":       "⚠️  警告: レガシーの設定ファイル位置を使用しています。",
		"config.legacyWarningDetail": "   次のコマンドで ~/.config/toske/config.yml に移行してください: mv ~/.toske.yaml ~/.config/toske/config.yml",

		// Common
		"common.error": "エラー: %v",
	},
}
