package cmd

// ja: Config は設定ファイルの構造を表します
// en: Config represents the structure of the configuration file
type Config struct {
	Version  string    `mapstructure:"version"`
	Projects []Project `mapstructure:"projects"`
}

// ja: Project はプロジェクト設定を表します
// en: Project represents a project configuration
type Project struct {
	Name            string   `mapstructure:"name"`
	Repo            string   `mapstructure:"repo"`
	Branch          string   `mapstructure:"branch"`
	BackupPaths     []string `mapstructure:"backup_paths"`
	BackupRetention int      `mapstructure:"backup_retention"`
}
