package l1

import (
	"fmt"
)

func setBit(num int64, pos int, value int) int64 {
	if value == 1 {
		num |= (1 << pos)
	} else {
		num &^= (1 << pos)
	}
	return num
}

func Example_L1_8() {
	var n int64 = 5
	fmt.Println(setBit(n, 1, 0))
}
