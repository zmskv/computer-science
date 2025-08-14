package l1

import (
	"fmt"
	"time"
)

func Example_L1_5() {
	ch := make(chan any)

	go func() {
		counter := 1
		for {
			ch <- counter
			counter++
			time.Sleep(500 * time.Millisecond)
		}
	}()

	go func() {
		for val := range ch {
			fmt.Println(val)
		}
	}()

	timeLimit := 3 * time.Second
	<-time.After(timeLimit)
	close(ch)
}
