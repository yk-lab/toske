# 設定ファイルのスキーマサンプル

```yaml
version: 1.0.0
projects:
  - name: project-a
    repo: git@github.com:user/project-a.git
    branch: main
    repository_path: /Users/username/projects/project-a
    backup_paths:
      - .env
      - db.sqlite3
      - migrations/
      - config/
    backup_retention: 5

  - name: project-b
    repo: https://github.com/user/project-b.git
    branch: develop
    repository_path: /Users/username/projects/project-b
    backup_paths:
      - .env.local
      - data/
```

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://example.com/archive-tool/config.schema.json",
  "type": "object",
  "additionalProperties": false,
  "required": ["version", "projects"],
  "properties": {
    "version": {
      "type": "string",
      "description": "設定ファイルのバージョン",
      "default": "1.0.0"
    },
    "projects": {
      "type": "array",
      "description": "バックアップ対象プロジェクト一覧",
      "items": {
        "type": "object",
        "required": ["name", "repo", "branch", "repository_path", "backup_paths"],
        "properties": {
          "name": {
            "type": "string",
            "description": "プロジェクトの識別名（任意の名称）"
          },
          "repo": {
            "type": "string",
            "format": "uri",
            "description": "GitリポジトリのURL（SSHまたはHTTPS）"
          },
          "branch": {
            "type": "string",
            "description": "使用するGitのブランチ名"
          },
          "repository_path": {
            "type": "string",
            "description": "ローカルリポジトリの絶対パス"
          },
          "backup_paths": {
            "type": "array",
            "description": "バックアップするファイルまたはディレクトリパスのリスト。ディレクトリの場合は再帰的にバックアップされる。",
            "items": {
              "type": "string"
            },
            "minItems": 1
          },
          "backup_retention": {
            "type": "integer",
            "minimum": 1,
            "default": 3,
            "description": "バックアップを保持する件数（デフォルトは3）"
          }
        },
        "additionalProperties": false
      }
    }
  },
}
```
