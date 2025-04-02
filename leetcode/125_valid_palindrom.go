package leetcode

import (
	"strings"
	"unicode"
)

func cleanString(input string) string {
	var result strings.Builder

	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		}
	}

	return result.String()
}

func isPalindrome(s string) bool {

	validString := cleanString(s)
	validString = strings.ToLower(validString)

	lt := 0
	rt := len(validString) - 1

	for lt <= rt {
		if validString[lt] != validString[rt] {
			return false
		}
		lt++
		rt--
	}

	return true
}

// ассимптотическая сложность O(n), где n - количество символов в строке
// сложность по памяти O(n), из-за создания новой очищенной строки
