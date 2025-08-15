package l1

import "fmt"

func quickSort(arr []int, left, right int) {
	if left >= right {
		return
	}

	pivot := arr[(left+right)/2]
	i := left
	j := right

	for i <= j {
		for arr[i] < pivot {
			i++
		}
		for arr[j] > pivot {
			j--
		}
		if i <= j {
			arr[i], arr[j] = arr[j], arr[i]
			i++
			j--
		}
	}
	if left < j {
		quickSort(arr, left, j)
	}
	if i < right {
		quickSort(arr, i, right)
	}
}

func Example_L1_16() {
	arr := []int{5, 3, 8, 4, 2, 7, 1, 6}
	quickSort(arr, 0, len(arr)-1)
	fmt.Println(arr)
}
