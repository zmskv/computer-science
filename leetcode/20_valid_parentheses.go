package leetcode

func isValid(s string) bool {
	m := map[rune]rune{
		')': '(',
		'}': '{',
		']': '[',
	}
	stack := make([]rune, 0)

	for _, ch := range s {
		if _, exist := m[ch]; !exist {
			stack = append(stack, ch)
		} else {
			if len(stack) == 0 {
				return false
			} else if stack[len(stack)-1] == m[ch] {
				stack = stack[:len(stack)-1]
			} else {
				return false
			}
		}
	}

	return len(stack) == 0
}

// ассимптотическая сложность O(n), где n - длина строки s
// сложность по памяти O(n), гдe n - длина стека
