package l2

import (
	"errors"
	"strings"
	"unicode"
)

// Unpack string
// Example: a4bc2d5e -> aaaabccddddde
func Unpack(s string) (string, error) {
	var b strings.Builder

	var (
		prevRune rune
		hasPrev  bool
		escaped  bool
	)

	for _, r := range s {
		if escaped {
			if hasPrev {
				b.WriteRune(prevRune)
			}
			prevRune, hasPrev = r, true
			escaped = false
			continue
		}

		if r == '\\' {
			escaped = true
			continue
		}

		if unicode.IsDigit(r) {
			if !hasPrev {
				return "", errors.New("invalid string")
			}
			count := int(r - '0')
			for i := 0; i < count; i++ {
				b.WriteRune(prevRune)
			}
			hasPrev = false
			continue
		}

		if hasPrev {
			b.WriteRune(prevRune)
		}
		prevRune, hasPrev = r, true
	}

	if escaped {
		return "", errors.New("invalid string")
	}
	if hasPrev {
		b.WriteRune(prevRune)
	}

	return b.String(), nil
}
