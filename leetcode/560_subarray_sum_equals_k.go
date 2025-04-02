package main

func subarraySum(nums []int, k int) int {
	m := make(map[int]int, 0)
	m[0] = 1
	curSum := 0
	count := 0

	for i := 0; i < len(nums); i++ {
		curSum += nums[i]

		if val, exist := m[curSum-k]; exist {
			count += val
		}

		m[curSum]++
	}

	return count

}

// O(n), где n - длина входного массива
// O(n), где n - размер мапы
