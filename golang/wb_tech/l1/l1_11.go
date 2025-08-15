package l1

import "fmt"

func Example_L1_11() {
	A := []int{1, 2, 3}
	B := []int{2, 3, 4}

	set := make(map[int]struct{})
	var intersection []int

	for _, v := range A {
		set[v] = struct{}{}
	}

	for _, v := range B {
		if _, ok := set[v]; ok {
			intersection = append(intersection, v)
		}
	}

	fmt.Println(intersection) 
}
