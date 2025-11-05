# ビルドと動作確認の手順 / Build and Verification Guide

このドキュメントでは、Toskeのビルド方法と動作確認の手順を説明します。

This document describes how to build and verify the Toske application.

---

## 📋 目次 / Table of Contents

- [必要な環境 / Requirements](#必要な環境--requirements)
- [ビルド手順 / Build Instructions](#ビルド手順--build-instructions)
- [動作確認手順 / Verification Steps](#動作確認手順--verification-steps)
- [トラブルシューティング / Troubleshooting](#トラブルシューティング--troubleshooting)

---

## 必要な環境 / Requirements

### 日本語

- **Go**: バージョン 1.24.1 以上
- **Git**: リポジトリのクローン用
- **OS**: Linux, macOS, Windows (Go がサポートする任意のプラットフォーム)

### English

- **Go**: Version 1.24.1 or higher
- **Git**: For cloning the repository
- **OS**: Linux, macOS, Windows (any platform supported by Go)

### 環境確認 / Environment Check

```bash
# Goのバージョン確認 / Check Go version
go version

# 正しいバージョンがインストールされていることを確認してください
# Ensure the correct version is installed
# 期待される出力 / Expected output: go version go1.24.x ...
```

---

## ビルド手順 / Build Instructions

### 日本語

#### 1. リポジトリのクローン

```bash
git clone https://github.com/yk-lab/toske.git
cd toske
```

#### 2. 依存関係のダウンロード

```bash
go mod download
```

#### 3. ビルド

```bash
# 基本的なビルド（カレントディレクトリに実行ファイルを作成）
go build -o toske

# または、Go の標準的なインストール方法を使用
go install
```

#### 4. ビルド成果物の確認

```bash
# ビルドが成功したか確認
ls -lh toske

# 実行ファイルの情報を表示
file toske
```

### English

#### 1. Clone the Repository

```bash
git clone https://github.com/yk-lab/toske.git
cd toske
```

#### 2. Download Dependencies

```bash
go mod download
```

#### 3. Build

```bash
# Basic build (creates executable in current directory)
go build -o toske

# Or use Go's standard installation method
go install
```

#### 4. Verify Build Output

```bash
# Check if the build was successful
ls -lh toske

# Display file information
file toske
```

---

## 動作確認手順 / Verification Steps

### 日本語

#### 1. 基本動作確認

```bash
# ヘルプメッセージを表示
./toske --help
```

**期待される出力:**
- Toskeのロゴ（アスキーアート）
- 利用可能なコマンド一覧
- グローバルフラグの説明

#### 2. サブコマンドの確認

```bash
# initコマンドのヘルプを表示
./toske init --help

# completionコマンドのヘルプを表示
./toske completion --help
```

#### 3. 設定ファイルの初期化テスト

```bash
# テスト用の一時ディレクトリを作成
mkdir -p /tmp/toske-test
cd /tmp/toske-test

# 設定ファイルを初期化
/path/to/toske init

# 設定ファイルが作成されたか確認
ls -la ~/.config/toske/config.yml
```

#### 4. 設定ファイルのカスタムパステスト

```bash
# カスタムパスで設定ファイルを指定
/path/to/toske --config /tmp/toske-test/custom-config.yml init

# カスタムパスに設定ファイルが作成されたか確認
ls -la /tmp/toske-test/custom-config.yml
```

#### 5. 環境変数のテスト

```bash
# TOSKE_CONFIG環境変数を使用
export TOSKE_CONFIG=/tmp/toske-test/env-config.yml
/path/to/toske init

# 環境変数で指定したパスに設定ファイルが作成されたか確認
ls -la /tmp/toske-test/env-config.yml
```

### English

#### 1. Basic Operation Check

```bash
# Display help message
./toske --help
```

**Expected output:**
- Toske logo (ASCII art)
- List of available commands
- Description of global flags

#### 2. Verify Subcommands

```bash
# Display help for init command
./toske init --help

# Display help for completion command
./toske completion --help
```

#### 3. Configuration File Initialization Test

```bash
# Create a temporary directory for testing
mkdir -p /tmp/toske-test
cd /tmp/toske-test

# Initialize configuration file
/path/to/toske init

# Verify the configuration file was created
ls -la ~/.config/toske/config.yml
```

#### 4. Custom Path Configuration Test

```bash
# Specify a custom path for the configuration file
/path/to/toske --config /tmp/toske-test/custom-config.yml init

# Verify the configuration file was created at the custom path
ls -la /tmp/toske-test/custom-config.yml
```

#### 5. Environment Variable Test

```bash
# Use the TOSKE_CONFIG environment variable
export TOSKE_CONFIG=/tmp/toske-test/env-config.yml
/path/to/toske init

# Verify the configuration file was created at the path specified by the environment variable
ls -la /tmp/toske-test/env-config.yml
```

---

## クロスプラットフォームビルド / Cross-Platform Build

### 日本語

異なるOS/アーキテクチャ向けにビルドする場合:

```bash
# Linux (64-bit)
GOOS=linux GOARCH=amd64 go build -o toske-linux-amd64

# macOS (64-bit Intel)
GOOS=darwin GOARCH=amd64 go build -o toske-darwin-amd64

# macOS (ARM64 - Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o toske-darwin-arm64

# Windows (64-bit)
GOOS=windows GOARCH=amd64 go build -o toske-windows-amd64.exe
```

### English

To build for different OS/architectures:

```bash
# Linux (64-bit)
GOOS=linux GOARCH=amd64 go build -o toske-linux-amd64

# macOS (64-bit Intel)
GOOS=darwin GOARCH=amd64 go build -o toske-darwin-amd64

# macOS (ARM64 - Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o toske-darwin-arm64

# Windows (64-bit)
GOOS=windows GOARCH=amd64 go build -o toske-windows-amd64.exe
```

---

## テスト実行 / Running Tests

### 日本語

```bash
# 全てのテストを実行
go test ./...

# カバレッジ付きでテストを実行
go test -cover ./...

# 詳細な出力でテストを実行
go test -v ./...
```

**注意:** 現在、テストファイルが存在しない場合は、`[no test files]` というメッセージが表示されます。

### English

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

**Note:** If no test files exist, you will see a `[no test files]` message.

---

## トラブルシューティング / Troubleshooting

### 日本語

#### ビルドエラー

**問題:** `go: module ... not found`

**解決方法:**
```bash
# 依存関係を再取得
go mod tidy
go mod download
```

**問題:** Goのバージョンが古い

**解決方法:**
```bash
# Goを最新バージョンにアップデート
# 公式サイトからダウンロード: https://golang.org/dl/
```

#### 実行時エラー

**問題:** `permission denied`

**解決方法:**
```bash
# 実行権限を付与
chmod +x toske
```

**問題:** 設定ファイルが見つからない

**解決方法:**
```bash
# 設定ファイルを初期化
./toske init

# または環境変数を設定
export TOSKE_CONFIG=/path/to/your/config.yml
```

### English

#### Build Errors

**Issue:** `go: module ... not found`

**Solution:**
```bash
# Re-fetch dependencies
go mod tidy
go mod download
```

**Issue:** Go version is too old

**Solution:**
```bash
# Update Go to the latest version
# Download from official site: https://golang.org/dl/
```

#### Runtime Errors

**Issue:** `permission denied`

**Solution:**
```bash
# Grant execute permission
chmod +x toske
```

**Issue:** Configuration file not found

**Solution:**
```bash
# Initialize configuration file
./toske init

# Or set environment variable
export TOSKE_CONFIG=/path/to/your/config.yml
```

---

## 開発者向け情報 / Developer Information

### 日本語

#### デバッグビルド

```bash
# デバッグ情報を含めてビルド
go build -gcflags="all=-N -l" -o toske-debug

# デバッガで実行（例: dlv）
dlv exec ./toske-debug
```

#### ベンダリング

```bash
# vendor ディレクトリに依存関係をコピー
go mod vendor

# vendor を使用してビルド
go build -mod=vendor -o toske
```

#### コードの静的解析

```bash
# go vetで問題をチェック
go vet ./...

# golangci-lintで詳細な解析（インストールが必要）
golangci-lint run
```

### English

#### Debug Build

```bash
# Build with debug information
go build -gcflags="all=-N -l" -o toske-debug

# Run with debugger (e.g., dlv)
dlv exec ./toske-debug
```

#### Vendoring

```bash
# Copy dependencies to vendor directory
go mod vendor

# Build using vendor
go build -mod=vendor -o toske
```

#### Static Code Analysis

```bash
# Check for issues with go vet
go vet ./...

# Detailed analysis with golangci-lint (requires installation)
golangci-lint run
```

---

## 参考リンク / References

- [Go Documentation](https://golang.org/doc/)
- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Viper Configuration](https://github.com/spf13/viper)
- [Project Requirements](./requirements.md)
- [Commands Design v1](./commands_v1.md)
- [Commands Design v2](./commands_v2.md)

---

## ライセンス / License

このプロジェクトのライセンスについては、[LICENSE](../LICENSE) ファイルを参照してください。

For information about the project's license, please refer to the [LICENSE](../LICENSE) file.
