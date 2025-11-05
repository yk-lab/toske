package i18n

import (
	"os"
	"strings"
	"sync"
)

// SupportedLanguages is the list of supported language codes
var SupportedLanguages = []string{"en", "ja"}

var (
	// currentLanguage holds the detected language
	currentLanguage string
	// languageMu protects currentLanguage from concurrent access
	languageMu sync.RWMutex
)

func init() {
	currentLanguage = detectLanguage()
}

// detectLanguage detects the user's preferred language from environment variables
// Priority: TOSKE_LANG > LANG > LC_ALL > default (en)
func detectLanguage() string {
	// Check TOSKE_LANG first (tool-specific override)
	if lang := os.Getenv("TOSKE_LANG"); lang != "" {
		return normalizeLanguage(lang)
	}

	// Check LANG environment variable
	if lang := os.Getenv("LANG"); lang != "" {
		return normalizeLanguage(lang)
	}

	// Check LC_ALL as fallback
	if lang := os.Getenv("LC_ALL"); lang != "" {
		return normalizeLanguage(lang)
	}

	// Default to English
	return "en"
}

// normalizeLanguage extracts the language code from locale string
// e.g., "ja_JP.UTF-8" -> "ja", "en_US" -> "en"
func normalizeLanguage(locale string) string {
	// Split by underscore or dot
	parts := strings.FieldsFunc(locale, func(r rune) bool {
		return r == '_' || r == '.' || r == '-'
	})

	if len(parts) == 0 {
		return "en"
	}

	lang := strings.ToLower(parts[0])

	// Check if the language is supported
	for _, supported := range SupportedLanguages {
		if lang == supported {
			return lang
		}
	}

	// Default to English if not supported
	return "en"
}

// T returns the translated message for the given key
// If the key is not found in the current language, it falls back to English
func T(key string) string {
	languageMu.RLock()
	lang := currentLanguage
	languageMu.RUnlock()

	// Try to get message in current language
	if msg, ok := messages[lang][key]; ok {
		return msg
	}

	// Fallback to English
	if msg, ok := messages["en"][key]; ok {
		return msg
	}

	// If not found at all, return the key itself as a fallback
	return key
}

// GetLanguage returns the current language code
func GetLanguage() string {
	languageMu.RLock()
	defer languageMu.RUnlock()
	return currentLanguage
}

// SetLanguage sets the current language (useful for testing)
func SetLanguage(lang string) {
	languageMu.Lock()
	defer languageMu.Unlock()
	currentLanguage = normalizeLanguage(lang)
}
