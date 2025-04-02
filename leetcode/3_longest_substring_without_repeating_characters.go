package main

func lengthOfLongestSubstring(s string) int {
	table := make(map[rune]int)
	lt, rt := 0, 0
	res := 0

	if len(s) == 1 {
		return 1
	}

	for rt < len(s) {
		if _, ok := table[rune(s[rt])]; !ok {
			table[rune(s[rt])]++
			rt++
		} else {
			delete(table, rune(s[lt]))
			lt++
		}
		res = max(len(table), res)
	}

	return res
}


// ассимптотическая сложность O(n), где n - длина строки
// сложность по памяти O(n), где n - количество символов в строке