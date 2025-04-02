package leetcode

func intersect(nums1 []int, nums2 []int) []int {
	m := make(map[int]int)

	for i := 0; i < len(nums1); i++ {
		if _, exist := m[nums1[i]]; exist {
			m[nums1[i]]++
		} else {
			m[nums1[i]] = 1
		}
	}
	ans := make([]int, 0)

	for j := 0; j < len(nums2); j++ {
		if m[nums2[j]] > 0 {
			ans = append(ans, nums2[j])
			m[nums2[j]]--
		}
	}

	return ans

}

// ассимптотическая сложность O(max(n, m)), где n и m длины массивов nums1 и nums2 соответственно
// сложность по памяти O(n), гдe n - количество повторяющихся чисел в обоих массивах
