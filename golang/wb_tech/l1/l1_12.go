package l1

import "fmt"

func Example_L1_12() {
	words := []string{"cat", "cat", "dog", "cat", "tree"}

	set := make(map[string]struct{})

	for _, word := range words {
		set[word] = struct{}{}
	}

	for word := range set {
		fmt.Println(word)
	}
}
