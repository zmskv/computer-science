package l1

import "fmt"

func binarySearch(arr []int, target int) int {
	left := 0
	right := len(arr) - 1

	for left <= right {
		mid := (left + right) / 2

		if arr[mid] == target {
			return mid
		}
		if arr[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return -1
}

func Example_L1_17() {
	arr := []int{1, 3, 5, 7, 9, 15}
	fmt.Println(binarySearch(arr, 7))
}
