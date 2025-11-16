package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yk-lab/toske/i18n"
	"gopkg.in/yaml.v3"
)

var (
	pruneProjectName string
	pruneAll         bool
	pruneKeep        int
)

// ja: errNoPruneNeeded は保持件数以内でprune不要な場合のセンチネルエラー
// en: errNoPruneNeeded is a sentinel error when pruning is not needed (within retention limit)
var errNoPruneNeeded = errors.New("prune: no prune needed")

// ja: pruneCmd は prune コマンドを表します
// en: pruneCmd represents the prune command
var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: i18n.T("prune.short"),
	Long:  i18n.T("prune.long"),
	Run: func(cmd *cobra.Command, args []string) {
		keepExplicit := cmd.Flags().Changed("keep")
		if err := runPrune(keepExplicit); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("common.error")+"\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(pruneCmd)
	pruneCmd.Flags().StringVarP(&pruneProjectName, "project", "p", "", i18n.T("prune.flag.project"))
	pruneCmd.Flags().BoolVar(&pruneAll, "all", false, i18n.T("prune.flag.all"))
	pruneCmd.Flags().IntVarP(&pruneKeep, "keep", "k", 0, i18n.T("prune.flag.keep"))
}

func runPrune(keepExplicit bool) error {
	// ja: --project と --all の両方が指定されている場合はエラー
	// en: Error if both --project and --all are specified
	if pruneProjectName != "" && pruneAll {
		return fmt.Errorf("%s", i18n.T("prune.bothFlags"))
	}

	// ja: --project も --all も指定されていない場合はエラー
	// en: Error if neither --project nor --all is specified
	if pruneProjectName == "" && !pruneAll {
		return fmt.Errorf("%s", i18n.T("prune.noFlags"))
	}

	// ja: --keep が負の値の場合はエラー
	// en: Error if --keep is negative
	if pruneKeep < 0 {
		return fmt.Errorf("%s", i18n.T("prune.invalidKeep"))
	}

	// ja: 設定ファイルパスを決定
	// en: Determine config file path
	configPath := cfgFile
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}

	// ja: 設定ファイルが存在するかチェック
	// en: Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf(i18n.T("prune.noConfig"), configPath)
	}

	// ja: 設定ファイルを読み込む
	// en: Load configuration file
	v := viper.New()
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf(i18n.T("prune.readError"), err)
	}

	// ja: 設定を構造体にアンマーシャル
	// en: Unmarshal config into struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return fmt.Errorf(i18n.T("prune.parseError"), err)
	}

	// ja: プロジェクトが存在しない場合
	// en: If no projects exist
	if len(config.Projects) == 0 {
		return fmt.Errorf("%s", i18n.T("prune.noProjects"))
	}

	// ja: ホームディレクトリを取得
	// en: Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if pruneAll {
		// ja: すべてのプロジェクトを対象
		// en: Target all projects
		fmt.Println(i18n.T("prune.pruningAll"))
		fmt.Println()

		successCount := 0
		skippedCount := 0
		errorCount := 0

		for _, project := range config.Projects {
			fmt.Printf(i18n.T("prune.processingProject")+"\n", project.Name)

			backupDir := filepath.Join(homeDir, ".config", "toske", "backups", project.Name)
			retention, skip, err := determineRetention(project, pruneKeep, keepExplicit)

			if err != nil {
				fmt.Fprintf(os.Stderr, "  "+i18n.T("prune.error")+"\n", err)
				errorCount++
				fmt.Println()
				continue
			}

			if skip {
				fmt.Println("  " + i18n.T("prune.noRetentionSkip"))
				skippedCount++
				fmt.Println()
				continue
			}

			if err := pruneProjectBackups(backupDir, retention); err != nil {
				if errors.Is(err, errNoPruneNeeded) {
					// ja: 保持件数以内の場合はスキップとして扱う
					// en: Treat as skipped when already within retention limit
					// ja: メタデータを読み込んで実際のバックアップ数を取得
					// en: Load metadata to get actual backup count
					metadataPath := filepath.Join(backupDir, "backups.yaml")
					data, readErr := os.ReadFile(metadataPath)
					backupCount := 0
					if readErr == nil {
						var metadata BackupMetadata
						if unmarshalErr := yaml.Unmarshal(data, &metadata); unmarshalErr == nil {
							backupCount = len(metadata.Backups)
						}
					}
					fmt.Printf("  "+i18n.T("prune.noPruneNeeded")+"\n", backupCount, retention)
					skippedCount++
				} else {
					fmt.Fprintf(os.Stderr, "  "+i18n.T("prune.error")+"\n", err)
					errorCount++
				}
			} else {
				fmt.Printf("  "+i18n.T("prune.pruned")+"\n", retention)
				successCount++
			}
			fmt.Println()
		}

		fmt.Println(i18n.T("prune.summary"))
		fmt.Printf(i18n.T("prune.summarySuccess")+"\n", successCount)
		fmt.Printf(i18n.T("prune.summarySkipped")+"\n", skippedCount)
		if errorCount > 0 {
			fmt.Printf(i18n.T("prune.summaryError")+"\n", errorCount)
			return fmt.Errorf(i18n.T("prune.partialFailure"), errorCount)
		}
	} else {
		// ja: 指定されたプロジェクトを検索
		// en: Find the specified project
		var project *Project
		for i := range config.Projects {
			if config.Projects[i].Name == pruneProjectName {
				project = &config.Projects[i]
				break
			}
		}

		if project == nil {
			return fmt.Errorf(i18n.T("prune.projectNotFound"), pruneProjectName)
		}

		backupDir := filepath.Join(homeDir, ".config", "toske", "backups", project.Name)
		retention, skip, err := determineRetention(*project, pruneKeep, keepExplicit)

		if err != nil {
			return err
		}

		if skip {
			fmt.Printf(i18n.T("prune.noRetentionWarning")+"\n", project.Name)
			return nil
		}

		fmt.Printf(i18n.T("prune.pruningProject")+"\n", project.Name, retention)

		if err := pruneProjectBackups(backupDir, retention); err != nil {
			if errors.Is(err, errNoPruneNeeded) {
				// ja: 保持件数以内の場合は正常終了
				// en: Return success when already within retention limit
				fmt.Println()
				// ja: メタデータを読み込んでバックアップ数を取得
				// en: Load metadata to get backup count
				metadataPath := filepath.Join(backupDir, "backups.yaml")
				data, readErr := os.ReadFile(metadataPath)
				if readErr == nil {
					var metadata BackupMetadata
					if unmarshalErr := yaml.Unmarshal(data, &metadata); unmarshalErr == nil {
						fmt.Printf(i18n.T("prune.noPruneNeeded")+"\n", len(metadata.Backups), retention)
					}
				}
				return nil
			}
			return err
		}

		fmt.Println()
		fmt.Println(i18n.T("prune.success"))
	}

	return nil
}

// ja: determineRetention は保持件数を決定します
// en: determineRetention determines the retention count
// Returns (retention count, skip flag, error)
func determineRetention(project Project, keepFlag int, keepExplicit bool) (int, bool, error) {
	// ja: --keep フラグが明示的に指定されている場合はそれを優先（値が0でも）
	// en: Prioritize --keep flag if explicitly specified (even if value is 0)
	if keepExplicit {
		return keepFlag, false, nil
	}

	// ja: backup_retention 設定を使用
	// en: Use backup_retention setting
	if project.BackupRetention > 0 {
		return project.BackupRetention, false, nil
	}

	// ja: どちらも設定されていない場合はスキップ
	// en: Skip if neither is set
	return 0, true, nil
}

// ja: pruneProjectBackups はプロジェクトのバックアップをpruneします
// en: pruneProjectBackups prunes backups for a project
func pruneProjectBackups(backupDir string, retention int) error {
	// ja: バックアップディレクトリが存在するかチェック
	// en: Check if backup directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return fmt.Errorf("%s", i18n.T("prune.noBackupDir"))
	}

	metadataPath := filepath.Join(backupDir, "backups.yaml")

	// ja: メタデータファイルが存在するかチェック
	// en: Check if metadata file exists
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return fmt.Errorf("%s", i18n.T("prune.noMetadata"))
	}

	// ja: メタデータを読み込む
	// en: Load metadata
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf(i18n.T("prune.readMetadataError"), err)
	}

	var metadata BackupMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return fmt.Errorf(i18n.T("prune.parseMetadataError"), err)
	}

	// ja: バックアップがない場合
	// en: If no backups exist
	if len(metadata.Backups) == 0 {
		return fmt.Errorf("%s", i18n.T("prune.noBackups"))
	}

	// ja: 保持件数以下の場合は削除不要
	// en: No need to delete if backup count is below retention
	if len(metadata.Backups) <= retention {
		return errNoPruneNeeded
	}

	// ja: 削除対象のバックアップ数を計算
	// en: Calculate number of backups to delete
	deleteCount := len(metadata.Backups) - retention

	// ja: 保持件数を超えるバックアップを削除
	// en: Delete backups exceeding retention count
	for _, backup := range metadata.Backups[retention:] {
		archivePath := filepath.Join(backupDir, backup.Filename)
		if err := os.Remove(archivePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf(i18n.T("prune.deleteError"), backup.Filename, err)
		}
	}

	// ja: メタデータを更新
	// en: Update metadata
	metadata.Backups = metadata.Backups[:retention]
	data, err = yaml.Marshal(&metadata)
	if err != nil {
		return fmt.Errorf(i18n.T("prune.marshalError"), err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf(i18n.T("prune.writeError"), err)
	}

	fmt.Printf("  "+i18n.T("prune.deleted")+"\n", deleteCount)

	return nil
}
