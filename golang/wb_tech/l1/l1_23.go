package l1

import (
	"fmt"
)

func Example_L1_23() {
	slice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	i := 2

	copy(slice[i:], slice[i+1:])

	slice = slice[:len(slice)-1]

	fmt.Println(slice)
}
