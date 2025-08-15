package l1

import "fmt"

func Example_L1_9() {
	in := make(chan int)
	out := make(chan int)

	go func() {
		nums := []int{1, 2, 3, 4, 5}
		for _, n := range nums {
			in <- n
		}
		close(in)
	}()

	go func() {
		for x := range in {
			out <- x * 2
		}
		close(out)
	}()

	for result := range out {
		fmt.Println(result)
	}
}
