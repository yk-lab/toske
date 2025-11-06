package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yk-lab/toske/i18n"
)

// ja: editCmd は edit コマンドを表します
// en: editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: i18n.T("edit.short"),
	Long:  i18n.T("edit.long"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runEdit(); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("common.error")+"\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func runEdit() error {
	// ja: 設定ファイルパスを決定
	// en: Determine config file path
	configPath := cfgFile
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}

	// ja: 設定ファイルが存在するかチェック
	// en: Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		msg := fmt.Sprintf(i18n.T("edit.noConfig"), configPath)
		return fmt.Errorf("%s", msg)
	}

	// ja: エディタを決定
	// en: Determine editor
	editor := getEditor()
	if editor == "" {
		return fmt.Errorf("%s", i18n.T("edit.noEditor"))
	}

	// ja: エディタで設定ファイルを開く
	// en: Open config file in editor
	fmt.Printf(i18n.T("edit.openingEditor")+"\n", editor)

	// ja: エディタコマンドをパースして引数を分離
	// en: Parse editor command to separate arguments
	editorParts := strings.Fields(editor)
	editorCmd := exec.Command(editorParts[0], append(editorParts[1:], configPath)...)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		msg := fmt.Sprintf(i18n.T("edit.editorError"), err)
		return fmt.Errorf("%s", msg)
	}

	return nil
}

// ja: getEditor はユーザーの環境に応じたエディタを返します
// en: getEditor returns the editor based on user's environment
func getEditor() string {
	// ja: EDITOR 環境変数をチェック
	// en: Check EDITOR environment variable
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	// ja: VISUAL 環境変数をチェック（代替手段）
	// en: Check VISUAL environment variable (alternative)
	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual
	}

	// ja: デフォルトエディタを順番に試す
	// en: Try default editors in order
	editors := []string{"vim", "vi", "nano"}
	for _, editor := range editors {
		if _, err := exec.LookPath(editor); err == nil {
			return editor
		}
	}

	return ""
}
