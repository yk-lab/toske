package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yk-lab/toske/i18n"
	"gopkg.in/yaml.v3"
)

var projectName string

// ja: BackupMetadata はバックアップのメタデータを表します
// en: BackupMetadata represents backup metadata
type BackupMetadata struct {
	Project string         `yaml:"project"`
	Backups []BackupRecord `yaml:"backups"`
}

// ja: BackupRecord は個々のバックアップ記録を表します
// en: BackupRecord represents an individual backup record
type BackupRecord struct {
	Filename  string    `yaml:"filename"`
	Timestamp time.Time `yaml:"timestamp"`
	Files     []string  `yaml:"files"`
}

// ja: backupCmd は backup コマンドを表します
// en: backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: i18n.T("backup.short"),
	Long:  i18n.T("backup.long"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runBackup(); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("common.error")+"\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.Flags().StringVarP(&projectName, "project", "p", "", i18n.T("backup.flag.project"))
}

func runBackup() error {
	// ja: プロジェクト名が指定されているかチェック
	// en: Check if project name is specified
	if projectName == "" {
		return fmt.Errorf("%s", i18n.T("backup.noProjectFlag"))
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
		return fmt.Errorf(i18n.T("backup.noConfig"), configPath)
	}

	// ja: 設定ファイルを読み込む
	// en: Load configuration file
	v := viper.New()
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf(i18n.T("backup.readError"), err)
	}

	// ja: 設定を構造体にアンマーシャル
	// en: Unmarshal config into struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return fmt.Errorf(i18n.T("backup.parseError"), err)
	}

	// ja: 指定されたプロジェクトを検索
	// en: Find the specified project
	var project *Project
	for i := range config.Projects {
		if config.Projects[i].Name == projectName {
			project = &config.Projects[i]
			break
		}
	}

	if project == nil {
		return fmt.Errorf(i18n.T("backup.projectNotFound"), projectName)
	}

	// ja: バックアップ対象ファイルがあるかチェック
	// en: Check if there are files to backup
	if len(project.BackupPaths) == 0 {
		return fmt.Errorf(i18n.T("backup.noBackupPaths"), projectName)
	}

	fmt.Printf(i18n.T("backup.creatingBackup")+"\n", project.Name)

	// ja: バックアップディレクトリを作成
	// en: Create backup directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	backupDir := filepath.Join(homeDir, ".config", "toske", "backups", project.Name)
	fmt.Printf(i18n.T("backup.creatingDir")+"\n", backupDir)

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf(i18n.T("backup.createDirError"), err)
	}

	// ja: バックアップアーカイブを作成
	// en: Create backup archive
	timestamp := time.Now()
	archiveFilename := fmt.Sprintf("backup_%s.tar.gz", timestamp.Format("20060102_150405"))
	archivePath := filepath.Join(backupDir, archiveFilename)

	fmt.Printf(i18n.T("backup.creatingArchive")+"\n", archiveFilename)

	backedUpFiles, err := createBackupArchive(archivePath, project.BackupPaths)
	if err != nil {
		return fmt.Errorf(i18n.T("backup.archiveError"), err)
	}

	// ja: メタデータファイルを更新
	// en: Update metadata file
	fmt.Println(i18n.T("backup.updatingMetadata"))
	if err := updateMetadata(backupDir, project.Name, archiveFilename, timestamp, backedUpFiles); err != nil {
		return fmt.Errorf(i18n.T("backup.metadataError"), err)
	}

	// ja: backup_retention に基づいて古いバックアップを削除
	// en: Prune old backups based on backup_retention
	if project.BackupRetention > 0 {
		fmt.Printf(i18n.T("backup.pruningOldBackups")+"\n", project.BackupRetention)
		if err := pruneOldBackups(backupDir, project.BackupRetention); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("backup.pruneError")+"\n", err)
		}
	}

	fmt.Println()
	fmt.Println(i18n.T("backup.success"))
	fmt.Printf(i18n.T("backup.backupLocation")+"\n", archivePath)

	return nil
}

// ja: createBackupArchive はバックアップアーカイブを作成します
// en: createBackupArchive creates a backup archive
func createBackupArchive(archivePath string, backupPaths []string) ([]string, error) {
	// ja: アーカイブファイルを作成
	// en: Create archive file
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return nil, err
	}
	defer archiveFile.Close()

	// ja: gzip ライターを作成
	// en: Create gzip writer
	gzipWriter := gzip.NewWriter(archiveFile)
	defer gzipWriter.Close()

	// ja: tar ライターを作成
	// en: Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// ja: カレントディレクトリを取得
	// en: Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	var backedUpFiles []string

	// ja: 各バックアップ対象パスを処理
	// en: Process each backup path
	for _, backupPath := range backupPaths {
		fullPath := filepath.Join(currentDir, backupPath)

		// ja: ファイルまたはディレクトリが存在するかチェック
		// en: Check if file or directory exists
		info, err := os.Stat(fullPath)
		if os.IsNotExist(err) {
			fmt.Printf(i18n.T("backup.fileNotFound")+"\n", backupPath)
			continue
		}
		if err != nil {
			return nil, err
		}

		// ja: ファイルまたはディレクトリをアーカイブに追加
		// en: Add file or directory to archive
		if info.IsDir() {
			err = addDirToArchive(tarWriter, fullPath, backupPath)
		} else {
			err = addFileToArchive(tarWriter, fullPath, backupPath)
			fmt.Printf(i18n.T("backup.addingFile")+"\n", backupPath)
		}

		if err != nil {
			return nil, err
		}

		backedUpFiles = append(backedUpFiles, backupPath)
	}

	return backedUpFiles, nil
}

// ja: addFileToArchive はファイルをアーカイブに追加します
// en: addFileToArchive adds a file to the archive
func addFileToArchive(tarWriter *tar.Writer, fullPath, archivePath string) error {
	file, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name:    archivePath,
		Size:    info.Size(),
		Mode:    int64(info.Mode()),
		ModTime: info.ModTime(),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tarWriter, file)
	return err
}

// ja: addDirToArchive はディレクトリを再帰的にアーカイブに追加します
// en: addDirToArchive recursively adds a directory to the archive
func addDirToArchive(tarWriter *tar.Writer, fullPath, archivePath string) error {
	return filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// ja: ディレクトリ自体はスキップ
		// en: Skip directories themselves
		if info.IsDir() {
			return nil
		}

		// ja: アーカイブ内のパスを計算
		// en: Calculate path in archive
		relPath, err := filepath.Rel(fullPath, path)
		if err != nil {
			return err
		}
		archiveFilePath := filepath.Join(archivePath, relPath)

		fmt.Printf(i18n.T("backup.addingFile")+"\n", archiveFilePath)

		return addFileToArchive(tarWriter, path, archiveFilePath)
	})
}

// ja: updateMetadata はメタデータファイルを更新します
// en: updateMetadata updates the metadata file
func updateMetadata(backupDir, projectName, archiveFilename string, timestamp time.Time, files []string) error {
	metadataPath := filepath.Join(backupDir, "backups.yaml")

	// ja: 既存のメタデータを読み込む
	// en: Load existing metadata
	var metadata BackupMetadata
	if data, err := os.ReadFile(metadataPath); err == nil {
		if err := yaml.Unmarshal(data, &metadata); err != nil {
			return err
		}
	}

	// ja: プロジェクト名を設定
	// en: Set project name
	metadata.Project = projectName

	// ja: 新しいバックアップ記録を追加
	// en: Add new backup record
	metadata.Backups = append(metadata.Backups, BackupRecord{
		Filename:  archiveFilename,
		Timestamp: timestamp,
		Files:     files,
	})

	// ja: タイムスタンプでソート（新しい順）
	// en: Sort by timestamp (newest first)
	sort.Slice(metadata.Backups, func(i, j int) bool {
		return metadata.Backups[i].Timestamp.After(metadata.Backups[j].Timestamp)
	})

	// ja: メタデータファイルに書き込む
	// en: Write metadata file
	data, err := yaml.Marshal(&metadata)
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, data, 0644)
}

// ja: pruneOldBackups は古いバックアップを削除します
// en: pruneOldBackups removes old backups
func pruneOldBackups(backupDir string, retention int) error {
	metadataPath := filepath.Join(backupDir, "backups.yaml")

	// ja: メタデータを読み込む
	// en: Load metadata
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return err
	}

	var metadata BackupMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return err
	}

	// ja: 保持件数を超えるバックアップを削除
	// en: Delete backups exceeding retention count
	if len(metadata.Backups) > retention {
		for _, backup := range metadata.Backups[retention:] {
			archivePath := filepath.Join(backupDir, backup.Filename)
			if err := os.Remove(archivePath); err != nil && !os.IsNotExist(err) {
				return err
			}
		}

		// ja: メタデータを更新
		// en: Update metadata
		metadata.Backups = metadata.Backups[:retention]
		data, err := yaml.Marshal(&metadata)
		if err != nil {
			return err
		}

		if err := os.WriteFile(metadataPath, data, 0644); err != nil {
			return err
		}
	}

	return nil
}
