package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yk-lab/toske/i18n"
)

var (
	// Version is the current version of toske
	// This can be overridden at build time using ldflags:
	// go build -ldflags "-X github.com/yk-lab/toske/cmd.Version=0.1.0"
	Version = "dev"

	// Commit is the git commit hash
	// This can be set at build time using ldflags:
	// go build -ldflags "-X github.com/yk-lab/toske/cmd.Commit=$(git rev-parse HEAD)"
	Commit = "unknown"

	// BuildDate is the build date
	// This can be set at build time using ldflags:
	// go build -ldflags "-X github.com/yk-lab/toske/cmd.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
	BuildDate = "unknown"
)

// ja: versionCmd は version コマンドを表します
// en: versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: i18n.T("version.short"),
	Long:  i18n.T("version.long"),
	Run: func(cmd *cobra.Command, args []string) {
		short, _ := cmd.Flags().GetBool("short")
		if short {
			fmt.Println(Version)
		} else {
			fmt.Printf("toske version %s\n", Version)
			fmt.Printf("  commit: %s\n", Commit)
			fmt.Printf("  built:  %s\n", BuildDate)
		}
	},
}

func init() {
	versionCmd.Flags().Bool("short", false, i18n.T("version.flag.short"))
	rootCmd.AddCommand(versionCmd)
}
