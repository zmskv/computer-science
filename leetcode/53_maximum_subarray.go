package leetcode

import "math"

func maxSubArray(nums []int) int {
	sum := 0
	mxSum := math.MinInt

	for _, val := range nums {
		sum = max(sum+val, val)
		mxSum = max(mxSum, sum)
	}

	return mxSum
}

// сложность O(n), где n - длина nums
// сложность по памяти O(1)
