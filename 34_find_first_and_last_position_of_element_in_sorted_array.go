package main

func searchRange(nums []int, target int) []int {
	if len(nums) == 0 {
		return []int{-1, -1}
	}
	first := foundFirst(nums, target)
	last := foundLast(nums, target)

	if first > last {
		return []int{-1, -1}
	}

	return []int{first, last}

}

func foundFirst(nums []int, target int) int {
	lt, rt := 0, len(nums)-1

	for lt <= rt {
		mid := lt + (rt-lt)/2

		if nums[mid] >= target {
			rt = mid - 1
		} else if nums[mid] < target {
			lt = mid + 1
		}
	}

	return lt
}

func foundLast(nums []int, target int) int {
	lt, rt := 0, len(nums)-1

	for lt <= rt {
		mid := lt + (rt-lt)/2

		if nums[mid] <= target {
			lt = mid + 1
		} else if nums[mid] > target {
			rt = mid - 1
		}
	}

	return rt
}

// ассимптотическая сложность O(log n), где n - длина nums
// сложность по памяти O(1)