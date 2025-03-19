# リポジトリアーカイブツール 初期バージョンのサブコマンド一覧と機能定義

## サブコマンド一覧

| サブコマンド | 説明                                 | オプション（例）                     | 実装優先度 |
|--------------|--------------------------------------|---------------------------------------|------------|
| `init`       | YAML設定ファイルを初期作成する       | なし                                  | 高    |
| `backup`     | 設定したファイルをバックアップする   | `-p, --project <project_name>`        | 高          |
| `delete`     | リポジトリを削除する（バックアップ済みが前提） | `-p, --project <project_name>`        | 高          |
| `restore`    | 再クローン＆バックアップファイル復元 | `-p, --project <project_name>`        | 高          |
| `list`       | 登録済みプロジェクト一覧を表示する   | なし                                  | 高          |
| `validate`   | YAML設定ファイルをJSON Schemaで検証する   | なし                                  | 高          |
| `remove`     | プロジェクトをバックアップ対象から削除する | `-p, --project <project_name>`        | 中          |
| `prune`      | 古いバックアップファイルを整理する   | `-p, --project <project_name>` \\ `--all` \\ `--keep <件数>` | 中 |
| `edit`       | 設定ファイルをデフォルトのエディタで開く | なし                                  | 高 |
| `help`       | ヘルプを表示する                     | なし                                  | 高          |
| `version`    | バージョン情報を表示する             | なし                                  | 高          |

## 各サブコマンドの詳細

### init

- YAML設定ファイルをデフォルトパスに作成する。
- デフォルトパス: `~/.config/archive-tool/config.yml`

例:

```bash
archive-tool init
```

### backup

- 設定に記載されたファイルをバックアップする。

```bash
archive-tool backup --project project-a
```

### delete

- バックアップが完了していることを確認した上でリポジトリを削除する。

```bash
archive-tool delete --project project-a
```

### restore

- リポジトリを再クローンし、最新バックアップを復元する。

```bash
archive-tool restore --project project-a
```

### prune

- 古いバックアップを削除して最新バックアップのみ保持。
- YAML設定の保持件数をデフォルト値とする。

```bash
archive-tool prune --project project-a --keep 3
archive-tool prune --all
```

#### 📌 prune --all と --keep の挙動定義

pruneコマンドにおいて--allオプションが指定された場合は、すべてのプロジェクトのバックアップを整理対象にします。このとき、--keepオプションとの兼ね合いは次のようにします。

基本的な挙動ルール

| 条件 | 挙動 |
|------|------|
| --all指定なし（プロジェクト指定） | YAML設定値または--keep指定値を使用 |
| --all指定あり、--keep指定なし | 各プロジェクトごとのYAML設定値を使用（未設定の場合はスキップ） |
| --all指定あり、--keep指定あり | --keep指定値をすべてのプロジェクトに適用 |

##### 🔖 具体例による挙動説明

例えば次のYAML設定の場合：

```yaml
projects:
  - name: project-a
    backup_retention: 3

  - name: project-b
    backup_retention: 2

  - name: project-c
    # backup_retention 未設定
```

ケース別の挙動例：

| コマンド実行例 | 結果 |
|----------------|------|
| `prune --all` | project-a: 3件、project-b: 2件、project-c: スキップ |
| `prune --all --keep 1` | project-a: 1件、project-b: 1件、project-c: 1件 |
| `prune --project project-a` | project-a: 3件 |
| `prune --project project-c` | エラー（--keepを指定する必要あり） |
| `prune --project project-c --keep 1` | project-c: 1件 |

##### 🚩 エラーケースの挙動について（安全策）

- YAMLにも--keepにも値が指定されていないプロジェクトは、安全を考慮してスキップし、最後に警告表示をします。

```plaintext
警告: project-c は保持件数未設定のためスキップされました。
```

このようにすることで、不用意なバックアップ削除を防げます。

##### 🎯 結論（推奨する挙動まとめ）

- prune --all使用時は、原則YAMLの値を使用する。

- --keep指定時は、すべてのプロジェクトに対し、指定の値を優先して適用する。
- YAMLとコマンドラインオプションがともに未指定の場合、対象プロジェクトをスキップし警告する（安全性重視）。

### remove

- YAML設定ファイルから特定プロジェクトを削除。

```bash
archive-tool remove --project project-a
```

### validate

- YAML設定ファイルをJSON Schemaで検証する。

```bash
archive-tool validate
```

### edit

- YAML設定ファイルをデフォルトのエディタ（環境変数`EDITOR`またはデフォルトの`vi`）で開く。

```bash
archive-tool edit
```

```bash
export EDITOR="code --wait"
archive-tool edit
```

### list

- YAML設定ファイルに登録された全プロジェクト名を一覧表示。

```bash
archive-tool list
```

### help

- ヘルプを表示。

```bash
archive-tool help
```

### version

- バージョン情報を表示。

```bash
archive-tool version
```

## 設定ファイルのパスについて

- デフォルトは `~/.config/archive-tool/config.yml`
- 環境変数 `ARCHIVE_TOOL_CONFIG` が設定されている場合はそちらを優先

例:

```bash
export ARCHIVE_TOOL_CONFIG="/path/to/config.yml"
```

## 前提条件

- Gitがインストール済み。
- YAML設定ファイルはJSON Schemaを用いてバリデーションを行う。
