package leetcode

func maxPower(s string) int {
	mxLength := -1000
	prev := s[0]
	cnt := 1

	for i := 1; i < len(s); i++ {
		if prev != s[i] {
			mxLength = max(mxLength, cnt)
			cnt = 0
		}
		prev = s[i]
		cnt++
	}

	mxLength = max(mxLength, cnt)

	return mxLength

}

// ассимптотическая сложность O(n), где n - длина строки
// сложность по памяти O(1)
