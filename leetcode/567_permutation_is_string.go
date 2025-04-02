package leetcode

func checkInclusion(s1 string, s2 string) bool {
	if len(s1) > len(s2) {
		return false
	}

	s1_arr := make([]int, 26)
	s2_arr := make([]int, 26)

	for i := 0; i < len(s1); i++ {
		s1_arr[s1[i]-'a']++
		s2_arr[s2[i]-'a']++
	}

	for i := 0; i < len(s2)-len(s1); i++ {
		if match(s1_arr, s2_arr) {
			return true
		}
		s2_arr[s2[i+len(s1)]-'a']++
		s2_arr[s2[i]-'a']--
	}

	return match(s1_arr, s2_arr)

}

func match(s1, s2 []int) bool {
	for i := 0; i < len(s2); i++ {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}


// ассимптотическая сложность O(n) - где n длина второй строки
// сложность по памяти O(1)