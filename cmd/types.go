package cmd

// ja: Config は設定ファイルの構造を表します
// en: Config represents the structure of the configuration file
type Config struct {
	Version  string    `mapstructure:"version" yaml:"version"`
	Projects []Project `mapstructure:"projects" yaml:"projects"`
}

// ja: Project はプロジェクト設定を表します
// en: Project represents a project configuration
type Project struct {
	Name            string   `mapstructure:"name" yaml:"name"`
	Repo            string   `mapstructure:"repo" yaml:"repo"`
	Branch          string   `mapstructure:"branch" yaml:"branch"`
	BackupPaths     []string `mapstructure:"backup_paths" yaml:"backup_paths,omitempty"`
	BackupRetention int      `mapstructure:"backup_retention" yaml:"backup_retention,omitempty"`
}
