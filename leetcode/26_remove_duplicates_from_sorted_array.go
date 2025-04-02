package main

func removeDuplicates(nums []int) int {
	ans := make([]int, 0)

	ans = append(ans, nums[0])

	for i := 1; i < len(nums); i++ {
		if nums[i] != nums[i-1] {
			ans = append(ans, nums[i])
		}
		nums[i-1] = nums[i]
	}

	for j := 0; j < len(ans); j++ {
		nums[j] = ans[j]
	}

	return len(ans)
}

// ассимптотическая сложность O(n + m), где n - длина nums, m - количество уникальных цифр в массиве nums
// сложность по памяти O(n), где n - количество уникальных цифр
