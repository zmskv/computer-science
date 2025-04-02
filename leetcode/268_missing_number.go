package leetcode

func missingNumber(nums []int) int {
	res := len(nums)
	for i := 0; i < len(nums); i++ {
		res += i - nums[i]
	}
	return res
}

// ассимптотическа сложность O(n), где n - длина nums
// сложность по памяти O(1)


// есть проще решение через суммы:
// посчитать сумму промежутка [0, n] и вычесть сумму исходного массива из этого получим заветное число