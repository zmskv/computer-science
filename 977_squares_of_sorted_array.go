package main

func sortedSquares(nums []int) []int {
	n := len(nums)
	ans := make([]int, n)
	lt, rt := 0, n-1
	pos := n - 1

	for lt <= rt {
		if nums[lt]*nums[lt] > nums[rt]*nums[rt] {
			ans[pos] = nums[lt] * nums[lt]
			lt++
		} else {
			ans[pos] = nums[rt] * nums[rt]
			rt--
		}
		pos--
	}

	return ans
}

// ассимптотическая сложность O(n), где n - количество элементов в массиве
// O(n) по памяти, где n - длина массива ans
