# 設定ファイルのスキーマサンプル

```yaml
projects:
  - name: project-a
    repo: git@github.com:user/project-a.git
    branch: main
    backup_paths:
      - .env
      - db.sqlite3
      - migrations/
      - config/
    backup_retention: 5

  - name: project-b
    repo: https://github.com/user/project-b.git
    branch: develop
    backup_paths:
      - .env.local
      - data/
```

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://example.com/archive-tool/config.schema.json",
  "type": "object",
  "required": ["projects"],
  "properties": {
    "projects": {
      "type": "array",
      "description": "バックアップ対象プロジェクト一覧",
      "items": {
        "type": "object",
        "required": ["name", "repo", "branch", "backup_paths"],
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
        "additionalItems": false
      }
    }
  },
  "additionalProperties": false
}
```
