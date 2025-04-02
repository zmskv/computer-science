package leetcode

func twoSum(nums []int, target int) []int {
	m := make(map[int]int)

	for i := 0; i < len(nums); i++ {
		first := nums[i]
		first_ind := i
		second := target - first

		if second_ind, ok := m[second]; ok {
			return []int{second_ind, first_ind}
		}
		m[nums[i]] = i
	}
	return nil

}


// ассимптотическая сложность O(n), где n - длина nums
// сложность по памяти O(n)