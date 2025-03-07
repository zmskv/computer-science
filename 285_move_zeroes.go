package main

func moveZeroes(nums []int) []int {

	lastNonZeroFoundAt := 0

	for i := 0; i < len(nums); i++ {
		if nums[i] != 0 {
			nums[lastNonZeroFoundAt] = nums[i]
			lastNonZeroFoundAt++
		}
	}

	for i := lastNonZeroFoundAt; i < len(nums); i++ {
		nums[i] = 0
	}

	return nums
}

// ассимптотическая сложность О(n + m), где n - количество чисел, m - количество нулей
// сложность по памяти O(1)
