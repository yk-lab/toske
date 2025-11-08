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
		"help.unknownCommand":    "Unknown command '%s'",
		"help.displayError":      "Error displaying help: %v",

		// Init command
		"init.short": "Initialize configuration file",
		"init.long": `Initialize creates a new configuration file at the default location.
The default path is ~/.config/toske/config.yml

You can override the default path by setting the TOSKE_CONFIG environment variable.`,

		// Version command
		"version.short":      "Print the version number of toske",
		"version.long":       "Print the version number of toske along with build information",
		"version.flag.short": "Print only the version number",

		// Edit command
		"edit.short":         "Edit the configuration file",
		"edit.long":          "Open the configuration file in your default editor.\nThe editor is determined by the EDITOR environment variable, or falls back to vi/vim/nano.",
		"edit.noConfig":      "Configuration file does not exist: %s\nRun 'toske init' to create one.",
		"edit.noEditor":      "No suitable editor found. Please set the EDITOR environment variable.",
		"edit.editorError":   "Failed to open editor: %v",
		"edit.openingEditor": "Opening configuration file in %s...",
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

		// Validate command
		"validate.short":                    "Validate the configuration file",
		"validate.long":                     "Validate checks the configuration file for syntax errors and ensures all required fields are present.",
		"validate.noConfig":                 "Configuration file does not exist: %s\nRun 'toske init' to create one.",
		"validate.checking":                 "Checking configuration file: %s",
		"validate.readError":                "Failed to read configuration file: %v",
		"validate.parseError":               "Failed to parse configuration file: %v",
		"validate.success":                  "✓ Configuration file is valid!",
		"validate.projectCount":             "  Found %d project(s) configured",
		"validate.error.noVersion":          "Configuration error: 'version' field is required",
		"validate.error.noProjects":         "Configuration error: at least one project must be defined",
		"validate.error.projectNoName":      "Configuration error: project #%d is missing the 'name' field",
		"validate.error.duplicateName":      "Configuration error: duplicate project name '%s'",
		"validate.error.projectNoRepo":      "Configuration error: project '%s' is missing the 'repo' field",
		"validate.error.projectNoBranch":    "Configuration error: project '%s' is missing the 'branch' field",
		"validate.error.invalidRetention":   "Configuration error: project '%s' has invalid backup_retention value: %d (must be >= 0)",

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
		"help.unknownCommand":    "不明なコマンド '%s'",
		"help.displayError":      "ヘルプの表示エラー: %v",

		// Init command
		"init.short": "設定ファイルを初期化",
		"init.long": `デフォルトの場所に新しい設定ファイルを作成します。
デフォルトパス: ~/.config/toske/config.yml

TOSKE_CONFIG 環境変数を設定することで、デフォルトパスを上書きできます。`,

		// Version command
		"version.short":      "toske のバージョン番号を表示",
		"version.long":       "toske のバージョン番号とビルド情報を表示します",
		"version.flag.short": "バージョン番号のみを表示",

		// Edit command
		"edit.short":         "設定ファイルを編集",
		"edit.long":          "設定ファイルをデフォルトエディタで開きます。\nエディタは EDITOR 環境変数で決定されます。設定されていない場合は vi/vim/nano を使用します。",
		"edit.noConfig":      "設定ファイルが存在しません: %s\n'toske init' を実行して作成してください。",
		"edit.noEditor":      "適切なエディタが見つかりません。EDITOR 環境変数を設定してください。",
		"edit.editorError":   "エディタの起動に失敗しました: %v",
		"edit.openingEditor": "設定ファイルを %s で開いています...",
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

		// Validate command
		"validate.short":                    "設定ファイルを検証",
		"validate.long":                     "設定ファイルの構文エラーをチェックし、すべての必須フィールドが存在することを確認します。",
		"validate.noConfig":                 "設定ファイルが存在しません: %s\n'toske init' を実行して作成してください。",
		"validate.checking":                 "設定ファイルを確認しています: %s",
		"validate.readError":                "設定ファイルの読み込みに失敗しました: %v",
		"validate.parseError":               "設定ファイルのパースに失敗しました: %v",
		"validate.success":                  "✓ 設定ファイルは正常です！",
		"validate.projectCount":             "  %d 個のプロジェクトが設定されています",
		"validate.error.noVersion":          "設定エラー: 'version' フィールドは必須です",
		"validate.error.noProjects":         "設定エラー: 少なくとも1つのプロジェクトを定義する必要があります",
		"validate.error.projectNoName":      "設定エラー: プロジェクト #%d に 'name' フィールドがありません",
		"validate.error.duplicateName":      "設定エラー: プロジェクト名 '%s' が重複しています",
		"validate.error.projectNoRepo":      "設定エラー: プロジェクト '%s' に 'repo' フィールドがありません",
		"validate.error.projectNoBranch":    "設定エラー: プロジェクト '%s' に 'branch' フィールドがありません",
		"validate.error.invalidRetention":   "設定エラー: プロジェクト '%s' の backup_retention 値が無効です: %d (0以上である必要があります)",

		// Config
		"config.legacyWarning":       "⚠️  警告: レガシーの設定ファイル位置を使用しています。",
		"config.legacyWarningDetail": "   次のコマンドで ~/.config/toske/config.yml に移行してください: mv ~/.toske.yaml ~/.config/toske/config.yml",

		// Common
		"common.error": "エラー: %v",
	},
}
