package l1

import "fmt"

func Example_L1_13_XOR() {
	a := 5
	b := 7

	a = a ^ b
	b = a ^ b
	a = a ^ b

	fmt.Println(a, b)
}
