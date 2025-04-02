package leetcode

func groupAnagrams(strs []string) [][]string {
	m := make(map[[26]rune][]string)

	for _, word := range strs {
		count := [26]rune{}
		for _, ch := range word {
			count[ch-'a']++
		}
		m[count] = append(m[count], word)
	}

	ans := make([][]string, 0)

	for _, val := range m {
		ans = append(ans, val)
	}

	return ans

}

// ассимптотическая сложность O(n * k), где n - количество строк, k - количество букв в строках
// сложность по памяти O(n * k), где n - количество списков, k - количество слов
