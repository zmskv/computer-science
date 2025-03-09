package main

func missingNumber(nums []int) int {
	res := len(nums)
	for i := 0; i < len(nums); i++ {
		res += i - nums[i]
	}
	return res
}

// ассимптотическа сложность O(n), где n - длина nums
// сложность по памяти O(1)
