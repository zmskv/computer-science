package main

func longestSubarray(nums []int) int {
	zeroCnt := 0
	mxLenght := 0
	start := 0

	for i := 0; i < len(nums); i++ {
		if nums[i] == 0 {
			zeroCnt++
		}

		for zeroCnt > 1 {
			if nums[start] == 0 {
				zeroCnt--
			}
			start++
		}

		mxLenght = max(mxLenght, i-start)
	}

	return mxLenght
}

// ассимптотическая сложность О(n), где n - количество чисел
// сложность по памяти O(n), где n - количество чисел
