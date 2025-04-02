package leetcode

func merge(nums1 []int, m int, nums2 []int, n int) {

	i := m - 1
	j := n - 1
	k := m + n - 1

	for i >= 0 {
		if (i >= 0) && nums1[i] > nums2[j] {
			nums1[k] = nums1[i]
			k--
			i--
		} else {
			nums1[k] = nums2[j]
			k--
			j--
		}
	}
}

// ассимптотическая сложность O(n), где n - длина массива nums1
// сложность по памяти O(1)