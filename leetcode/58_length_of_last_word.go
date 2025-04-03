package leetcode

import (
	"strings"
)

func lengthOfLastWord(s string) int {
	s = strings.TrimSpace(s)
	res := strings.Split(s, " ")

	return len(res[len(res)-1])
}

// сложность O(n) из-за strings.Split, где n - длина строки s
// сложность по памяти O(1)
