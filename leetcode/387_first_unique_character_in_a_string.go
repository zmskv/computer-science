package leetcode

func firstUniqChar(s string) int {
	m := make(map[rune]int)

	for _, ch := range s {
		if _, exist := m[ch]; exist {
			m[ch]++
		} else {
			m[ch] = 1
		}
	}

	for ind, ch := range s {
		if m[ch] == 1 {
			return ind
		}
	}




	return -1
}

// ассимптотическая сложность O(n), где n - длина строки 
// сложность по памяти O(n), где n - количество элементов в мапе
