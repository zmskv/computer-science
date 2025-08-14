package l1

import (
	"fmt"
	"time"
)

func Worker(id int, ch <-chan any) {
	for data := range ch {
		fmt.Printf("Worker %d: %v\n", id, data)
	}
}

func Example_L1_3(n int) {
	ch := make(chan any)

	for i := 1; i <= n; i++ {
		go Worker(i, ch)
	}

	counter := 1
	for {
		ch <- counter
		counter++
		time.Sleep(500 * time.Millisecond)
	}
}
