package l1

import (
	"fmt"
	"math/big"
)

func Example_L1l_22() {
	a := new(big.Int)
	b := new(big.Int)

	a.SetString("3000000", 10)
	b.SetString("2000000", 10)

	sum := new(big.Int).Add(a, b)
	diff := new(big.Int).Sub(a, b)
	prod := new(big.Int).Mul(a, b)
	quot := new(big.Int).Div(a, b)

	fmt.Println("a =", a)
	fmt.Println("b =", b)
	fmt.Println("a + b =", sum)
	fmt.Println("a - b =", diff)
	fmt.Println("a * b =", prod)
	fmt.Println("a / b =", quot)
}
