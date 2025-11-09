package cmd

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yk-lab/toske/i18n"
	"gopkg.in/yaml.v3"
)

var (
	restoreProjectName string
	backupIndex        int
	forceRestore       bool
)

// ja: restoreCmd は restore コマンドを表します
// en: restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: i18n.T("restore.short"),
	Long:  i18n.T("restore.long"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runRestore(); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("common.error")+"\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().StringVarP(&restoreProjectName, "project", "p", "", i18n.T("restore.flag.project"))
	restoreCmd.Flags().IntVarP(&backupIndex, "backup", "b", 1, i18n.T("restore.flag.backup"))
	restoreCmd.Flags().BoolVarP(&forceRestore, "force", "f", false, i18n.T("restore.flag.force"))
}

func runRestore() error {
	// ja: プロジェクト名が指定されているかチェック
	// en: Check if project name is specified
	if restoreProjectName == "" {
		return fmt.Errorf("%s", i18n.T("restore.noProjectFlag"))
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
		return fmt.Errorf(i18n.T("restore.noConfig"), configPath)
	}

	// ja: 設定ファイルを読み込む
	// en: Load configuration file
	v := viper.New()
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf(i18n.T("restore.readError"), err)
	}

	// ja: 設定を構造体にアンマーシャル
	// en: Unmarshal config into struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return fmt.Errorf(i18n.T("restore.parseError"), err)
	}

	// ja: 指定されたプロジェクトを検索
	// en: Find the specified project
	var project *Project
	for i := range config.Projects {
		if config.Projects[i].Name == restoreProjectName {
			project = &config.Projects[i]
			break
		}
	}

	if project == nil {
		return fmt.Errorf(i18n.T("restore.projectNotFound"), restoreProjectName)
	}

	// ja: バックアップディレクトリを取得
	// en: Get backup directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	backupDir := filepath.Join(homeDir, ".config", "toske", "backups", project.Name)

	// ja: バックアップディレクトリが存在するかチェック
	// en: Check if backup directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return fmt.Errorf(i18n.T("restore.noBackupDir"), project.Name)
	}

	// ja: メタデータファイルを読み込む
	// en: Load metadata file
	metadataPath := filepath.Join(backupDir, "backups.yaml")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return fmt.Errorf(i18n.T("restore.noMetadata"), project.Name)
	}

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf(i18n.T("restore.readMetadataError"), err)
	}

	var metadata BackupMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return fmt.Errorf(i18n.T("restore.parseMetadataError"), err)
	}

	// ja: バックアップが存在するかチェック
	// en: Check if backups exist
	if len(metadata.Backups) == 0 {
		return fmt.Errorf(i18n.T("restore.noBackups"), project.Name)
	}

	// ja: バックアップインデックスが有効かチェック
	// en: Check if backup index is valid
	if backupIndex < 1 || backupIndex > len(metadata.Backups) {
		return fmt.Errorf(i18n.T("restore.invalidBackupIndex"), backupIndex, len(metadata.Backups))
	}

	// ja: 復元するバックアップを選択（1-indexed）
	// en: Select backup to restore (1-indexed)
	selectedBackup := metadata.Backups[backupIndex-1]
	archivePath := filepath.Join(backupDir, selectedBackup.Filename)

	// ja: アーカイブファイルが存在するかチェック
	// en: Check if archive file exists
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		return fmt.Errorf(i18n.T("restore.backupNotFound"), selectedBackup.Filename)
	}

	// ja: 選択したバックアップ情報を表示
	// en: Display selected backup information
	fmt.Printf(i18n.T("restore.selectingBackup")+"\n", selectedBackup.Filename, selectedBackup.Timestamp.Format("2006-01-02 15:04:05"))

	// ja: 確認プロンプト（--force フラグが指定されていない場合）
	// en: Confirmation prompt (if --force flag is not specified)
	if !forceRestore {
		fmt.Println(i18n.T("restore.confirmOverwrite"))
		fmt.Print(i18n.T("restore.confirmPrompt"))

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf(i18n.T("restore.readInputError"), err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println(i18n.T("restore.cancelled"))
			return nil
		}
	}

	// ja: ファイルを復元
	// en: Restore files
	fmt.Println(i18n.T("restore.restoringFiles"))

	fileCount, err := extractBackupArchive(archivePath)
	if err != nil {
		return fmt.Errorf(i18n.T("restore.extractError"), err)
	}

	fmt.Println()
	fmt.Println(i18n.T("restore.success"))
	fmt.Printf(i18n.T("restore.restoredFiles")+"\n", fileCount)

	return nil
}

// ja: extractBackupArchive はバックアップアーカイブを展開します
// en: extractBackupArchive extracts a backup archive
func extractBackupArchive(archivePath string) (int, error) {
	// ja: アーカイブファイルを開く
	// en: Open archive file
	archiveFile, err := os.Open(archivePath)
	if err != nil {
		return 0, fmt.Errorf(i18n.T("restore.openArchiveError"), err)
	}
	defer archiveFile.Close()

	// ja: gzip リーダーを作成
	// en: Create gzip reader
	gzipReader, err := gzip.NewReader(archiveFile)
	if err != nil {
		return 0, err
	}
	defer gzipReader.Close()

	// ja: tar リーダーを作成
	// en: Create tar reader
	tarReader := tar.NewReader(gzipReader)

	// ja: カレントディレクトリを取得
	// en: Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return 0, err
	}

	fileCount := 0

	// ja: アーカイブ内の各ファイルを処理
	// en: Process each file in the archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}

		// ja: ディレクトリエントリはスキップ（ファイル作成時に自動的に作成される）
		// en: Skip directory entries (they will be created automatically when creating files)
		if header.Typeflag == tar.TypeDir {
			continue
		}

		// ja: ファイルパスを決定
		// en: Determine file path
		targetPath := filepath.Join(currentDir, header.Name)

		// ja: セキュリティチェック：パストラバーサル攻撃を防ぐ
		// en: Security check: prevent path traversal attacks
		if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(currentDir)) {
			continue
		}

		fmt.Printf(i18n.T("restore.extractingFile")+"\n", header.Name)

		// ja: ディレクトリを作成
		// en: Create directory
		targetDir := filepath.Dir(targetPath)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return 0, fmt.Errorf(i18n.T("restore.createDirError"), err)
		}

		// ja: ファイルを作成
		// en: Create file
		outFile, err := os.Create(targetPath)
		if err != nil {
			return 0, fmt.Errorf(i18n.T("restore.writeFileError"), err)
		}

		// ja: ファイル内容をコピー
		// en: Copy file contents
		if _, err := io.Copy(outFile, tarReader); err != nil {
			outFile.Close()
			return 0, err
		}
		outFile.Close()

		// ja: ファイルのパーミッションを設定
		// en: Set file permissions
		if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
			return 0, err
		}

		fileCount++
	}

	return fileCount, nil
}
