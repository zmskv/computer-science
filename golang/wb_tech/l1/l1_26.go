package l1

import (
	"fmt"
	"strings"
)

func hasUniqueSymbols(s string) bool {
	s = strings.ToLower(s)
	seen := make(map[rune]bool)

	for _, r := range s {
		if seen[r] {
			return false
		}
		seen[r] = true
	}
	return true
}

func Example_L1_26() {
	fmt.Println(hasUniqueSymbols("abcd"))
	fmt.Println(hasUniqueSymbols("abCdefAaf"))
	fmt.Println(hasUniqueSymbols("aabcd"))
}
