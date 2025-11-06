package cmd

import (
	"os"
	"os/exec"
	"testing"
)

// ja: hasDefaultEditor はデフォルトエディタが利用可能かチェックします
// en: hasDefaultEditor checks if any default editor is available
func hasDefaultEditor() bool {
	editors := []string{"vim", "vi", "nano"}
	for _, editor := range editors {
		if _, err := exec.LookPath(editor); err == nil {
			return true
		}
	}
	return false
}

func TestGetEditor(t *testing.T) {
	// ja: 元の環境変数を保存
	// en: Save original environment variables
	originalEditor, editorWasSet := os.LookupEnv("EDITOR")
	originalVisual, visualWasSet := os.LookupEnv("VISUAL")
	defer func() {
		// ja: テスト後に環境変数を復元
		// en: Restore environment variables after test
		if editorWasSet {
			os.Setenv("EDITOR", originalEditor)
		} else {
			os.Unsetenv("EDITOR")
		}
		if visualWasSet {
			os.Setenv("VISUAL", originalVisual)
		} else {
			os.Unsetenv("VISUAL")
		}
	}()

	tests := []struct {
		name          string
		editorEnv     string
		visualEnv     string
		expectNonEmpty bool
	}{
		{
			name:          "EDITOR environment variable is set",
			editorEnv:     "emacs",
			visualEnv:     "",
			expectNonEmpty: true,
		},
		{
			name:          "VISUAL environment variable is set",
			editorEnv:     "",
			visualEnv:     "nano",
			expectNonEmpty: true,
		},
		{
			name:          "Both environment variables are empty",
			editorEnv:     "",
			visualEnv:     "",
			expectNonEmpty: true, // Should find default editor (vim/vi/nano)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ja: 両方の環境変数が空の場合、デフォルトエディタの存在をチェック
			// en: If both environment variables are empty, check for default editor
			if tt.editorEnv == "" && tt.visualEnv == "" && !hasDefaultEditor() {
				t.Skip("Skipping test: no default editor (vim/vi/nano) found on system")
			}

			// ja: 環境変数を設定
			// en: Set environment variables
			os.Setenv("EDITOR", tt.editorEnv)
			os.Setenv("VISUAL", tt.visualEnv)

			// ja: エディタを取得
			// en: Get editor
			editor := getEditor()

			if tt.expectNonEmpty && editor == "" {
				t.Errorf("Expected non-empty editor, got empty string")
			}

			// ja: EDITOR が設定されている場合は、それが返されることを確認
			// en: If EDITOR is set, verify it's returned
			if tt.editorEnv != "" && editor != tt.editorEnv {
				t.Errorf("Expected editor %s, got %s", tt.editorEnv, editor)
			}

			// ja: EDITOR が空で VISUAL が設定されている場合は、VISUAL が返されることを確認
			// en: If EDITOR is empty and VISUAL is set, verify VISUAL is returned
			if tt.editorEnv == "" && tt.visualEnv != "" && editor != tt.visualEnv {
				t.Errorf("Expected editor %s, got %s", tt.visualEnv, editor)
			}
		})
	}
}

func TestGetEditorPriority(t *testing.T) {
	// ja: 元の環境変数を保存
	// en: Save original environment variables
	originalEditor, editorWasSet := os.LookupEnv("EDITOR")
	originalVisual, visualWasSet := os.LookupEnv("VISUAL")
	defer func() {
		// ja: テスト後に環境変数を復元
		// en: Restore environment variables after test
		if editorWasSet {
			os.Setenv("EDITOR", originalEditor)
		} else {
			os.Unsetenv("EDITOR")
		}
		if visualWasSet {
			os.Setenv("VISUAL", originalVisual)
		} else {
			os.Unsetenv("VISUAL")
		}
	}()

	// ja: EDITOR と VISUAL の両方が設定されている場合、EDITOR が優先される
	// en: When both EDITOR and VISUAL are set, EDITOR takes priority
	os.Setenv("EDITOR", "emacs")
	os.Setenv("VISUAL", "nano")

	editor := getEditor()
	if editor != "emacs" {
		t.Errorf("Expected EDITOR to take priority, got %s instead of emacs", editor)
	}
}

func TestGetEditorWithFlags(t *testing.T) {
	// ja: 元の環境変数を保存
	// en: Save original environment variables
	originalEditor, editorWasSet := os.LookupEnv("EDITOR")
	originalVisual, visualWasSet := os.LookupEnv("VISUAL")
	defer func() {
		// ja: テスト後に環境変数を復元
		// en: Restore environment variables after test
		if editorWasSet {
			os.Setenv("EDITOR", originalEditor)
		} else {
			os.Unsetenv("EDITOR")
		}
		if visualWasSet {
			os.Setenv("VISUAL", originalVisual)
		} else {
			os.Unsetenv("VISUAL")
		}
	}()

	tests := []struct {
		name         string
		editorEnv    string
		visualEnv    string
		expectedFull string
	}{
		{
			name:         "EDITOR with flags",
			editorEnv:    "code --wait",
			visualEnv:    "",
			expectedFull: "code --wait",
		},
		{
			name:         "VISUAL with flags",
			editorEnv:    "",
			visualEnv:    "subl -w",
			expectedFull: "subl -w",
		},
		{
			name:         "EDITOR with multiple flags",
			editorEnv:    "emacs -nw --no-splash",
			visualEnv:    "",
			expectedFull: "emacs -nw --no-splash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ja: 環境変数を設定
			// en: Set environment variables
			os.Setenv("EDITOR", tt.editorEnv)
			os.Setenv("VISUAL", tt.visualEnv)

			// ja: エディタを取得
			// en: Get editor
			editor := getEditor()

			// ja: フラグ付きのエディタコマンドが正しく返されることを確認
			// en: Verify that editor command with flags is returned correctly
			if editor != tt.expectedFull {
				t.Errorf("Expected editor %q, got %q", tt.expectedFull, editor)
			}
		})
	}
}
