package main

func removeDuplicates(nums []int) int {
	ans := make([]int, 0)

	prev := nums[0]
	ans = append(ans, prev)

	for i := 1; i < len(nums); i++ {
		if nums[i] != prev {
			ans = append(ans, nums[i])
		}
		prev = nums[i]
	}

	for j := 0; j < len(ans); j++ {
		nums[j] = ans[j]
	}

	return len(ans)
}


// ассимптотическая сложность O(n + m), где n - длина nums, m - количество уникальных цифр в массиве nums
// сложность по памяти O(n), где n - количество уникальных цифр 
