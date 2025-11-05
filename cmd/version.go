package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yk-lab/toske/i18n"
)

// Version is the current version of toske
const Version = "0.1.0"

// ja: versionCmd は version コマンドを表します
// en: versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: i18n.T("version.short"),
	Long:  i18n.T("version.long"),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("toske version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
