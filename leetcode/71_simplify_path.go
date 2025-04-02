package leetcode

import (
	"strings"
)

func simplifyPath(path string) string {
	components := strings.Split(path, "/")
	stack := make([]string, 0)

	for _, c := range components {
		if c == "." || c == "" {
			continue
		} else if c == ".." {
			if len(stack) != 0 {
				stack = stack[:len(stack)-1]
			}
		} else {
			stack = append(stack, c)
		}
	}

	return "/" + strings.Join(stack, "/")
}

// ассимптотическая сложность O(n), где n - длина строки
// сложность по памяти O(n), где n - длина стека
