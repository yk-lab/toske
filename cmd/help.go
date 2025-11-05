package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yk-lab/toske/i18n"
)

// ja: helpCmd は help コマンドを表します
// en: helpCmd represents the help command
var helpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: i18n.T("help.short"),
	Long:  i18n.T("help.long"),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// ja: 引数がない場合は、すべてのコマンドのヘルプを表示
			// en: Display help for all commands if no arguments
			showAllCommandsHelp(cmd)
		} else {
			// ja: 特定のコマンドのヘルプを表示
			// en: Display help for a specific command
			showCommandHelp(cmd, args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(helpCmd)
}

// ja: showAllCommandsHelp はすべてのコマンドのヘルプを表示します
// en: showAllCommandsHelp displays help for all commands
func showAllCommandsHelp(cmd *cobra.Command) {
	root := cmd.Root()

	// ja: 使用法を表示
	// en: Display usage
	fmt.Printf("%s\n", i18n.T("help.usage"))
	fmt.Printf("  %s\n\n", root.UseLine())

	// ja: 説明を表示
	// en: Display description
	if root.Long != "" {
		fmt.Printf("%s\n\n", root.Long)
	} else if root.Short != "" {
		fmt.Printf("%s\n\n", root.Short)
	}

	// ja: 利用可能なコマンドを表示
	// en: Display available commands
	fmt.Printf("%s\n", i18n.T("help.availableCommands"))

	commands := root.Commands()
	if len(commands) > 0 {
		maxLen := 0
		for _, c := range commands {
			if !c.IsAvailableCommand() || c.Hidden {
				continue
			}
			if len(c.Name()) > maxLen {
				maxLen = len(c.Name())
			}
		}

		for _, c := range commands {
			if !c.IsAvailableCommand() || c.Hidden {
				continue
			}
			fmt.Printf("  %-*s  %s\n", maxLen, c.Name(), c.Short)
		}
	}

	// ja: フラグを表示
	// en: Display flags
	if root.HasAvailableFlags() {
		fmt.Printf("\n%s\n", i18n.T("help.flags"))
		fmt.Printf("%s", root.Flags().FlagUsages())
	}

	// ja: 追加のヘルプ情報を表示
	// en: Display additional help information
	fmt.Printf("%s\n", i18n.T("help.additionalHelp"))
}

// ja: showCommandHelp は特定のコマンドのヘルプを表示します
// en: showCommandHelp displays help for a specific command
func showCommandHelp(cmd *cobra.Command, commandName string) {
	root := cmd.Root()

	// ja: 指定されたコマンドを検索
	// en: Find the specified command
	var targetCmd *cobra.Command
	for _, c := range root.Commands() {
		if c.Name() == commandName || contains(c.Aliases, commandName) {
			targetCmd = c
			break
		}
	}

	if targetCmd == nil {
		fmt.Printf("Unknown command '%s'\n", commandName)
		fmt.Printf("%s\n", i18n.T("help.additionalHelp"))
		return
	}

	// ja: コマンドのヘルプを表示
	// en: Display command help
	targetCmd.Help()
}

// ja: contains は文字列のスライスに指定された文字列が含まれているかチェックします
// en: contains checks if a string slice contains a specified string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, str) {
			return true
		}
	}
	return false
}
