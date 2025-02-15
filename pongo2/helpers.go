package pongo2

import (
	"encoding/base32"
	"github.com/google/uuid"
	"strings"
	"unicode"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// convertToCamelCase converts a kebab-case string to camelCase.
func convertToCamelCase(input string) string {
	var result []rune
	capitalizeNext := false

	for _, char := range input {
		if char == '-' {
			capitalizeNext = true
		} else {
			if capitalizeNext {
				result = append(result, unicode.ToUpper(char))
				capitalizeNext = false
			} else {
				result = append(result, char)
			}
		}
	}

	return string(result)
}

func convertToKebabCase(input string) string {
	var result []rune

	for i, char := range input {
		if unicode.IsUpper(char) {
			if i > 0 {
				result = append(result, '-')
			}
			result = append(result, unicode.ToLower(char))
		} else {
			result = append(result, char)
		}
	}

	return string(result)
}

// genRandomId generates a random ID.
// It is a base32 encoded UUID (26 characters long)
func genRandomId() string {
	UUID := uuid.New()
	return strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(UUID[:]))
}
