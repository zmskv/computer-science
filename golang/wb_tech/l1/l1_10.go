package l1

import (
	"fmt"
	"math"
)

func Example_L1_10() {
	temps := []float64{-25.4, -27.0, 13.0, 19.0, 15.5, 24.5, -21.0, 32.5}
	groups := make(map[int][]float64)

	for _, t := range temps {
		key := int(math.Floor(t/10) * 10)
		groups[key] = append(groups[key], t)
	}

	fmt.Println(groups)
}
