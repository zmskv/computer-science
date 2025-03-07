package main

func twoSum2(numbers []int, target int) []int {
	lt, rt := 0, len(numbers)-1

	for lt < rt {
		sum := numbers[lt] + numbers[rt]

		if sum == target {
			break
		} else if sum > target {
			rt--
		} else if sum < target {
			lt++
		}
	}

	return []int{lt + 1, rt + 1}
}


// ассимптотическая сложность O(n), где n - длина numbers
// сложность по памяти O(1)